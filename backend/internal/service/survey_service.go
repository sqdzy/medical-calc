package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/medical-app/backend/internal/entity"
	"github.com/medical-app/backend/internal/external"
	"github.com/medical-app/backend/internal/repository"
)

type SurveyService struct {
	templateRepo repository.SurveyTemplateRepository
	responseRepo repository.SurveyResponseRepository
	gptClient    *external.YandexGPTClient
}

type SurveyDeps struct {
	TemplateRepo repository.SurveyTemplateRepository
	ResponseRepo repository.SurveyResponseRepository
	GPTClient    *external.YandexGPTClient
}

func NewSurveyService(d SurveyDeps) *SurveyService {
	return &SurveyService{
		templateRepo: d.TemplateRepo,
		responseRepo: d.ResponseRepo,
		gptClient:    d.GPTClient,
	}
}

func (s *SurveyService) ListTemplates(ctx context.Context) ([]*entity.SurveyTemplate, error) {
	return s.templateRepo.ListActive(ctx)
}

func (s *SurveyService) GetTemplateByCode(ctx context.Context, code string) (*entity.SurveyTemplate, error) {
	return s.templateRepo.GetByCode(ctx, code)
}

func (s *SurveyService) SubmitResponse(ctx context.Context, req entity.SurveyResponseCreate) (*entity.SurveyResponse, error) {
	template, err := s.templateRepo.GetByID(ctx, req.TemplateID)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, errors.New("survey template not found")
	}

	now := time.Now().UTC()
	respMap := req.Responses
	responsesJSON, _ := json.Marshal(respMap)

	sr := &entity.SurveyResponse{
		ID:          uuid.New(),
		TemplateID:  req.TemplateID,
		PatientID:   req.PatientID,
		Responses:   responsesJSON,
		Status:      entity.SurveyStatusSubmitted,
		SubmittedAt: now,
		CreatedAt:   now,
	}

	score, category, breakdown, err := CalculateScore(template, respMap)
	if err != nil {
		return nil, err
	}
	sr.CalculatedScore = &score
	breakdownJSON, _ := json.Marshal(breakdown)
	sr.ScoreBreakdown = breakdownJSON
	sr.Interpretation = fmt.Sprintf("%s (%s)", template.Code, category)

	// Enrich interpretation with GPT if available
	if s.gptClient != nil {
		// Add category to breakdown for richer context
		breakdown["category"] = category
		gptInterpretation, err := s.gptClient.InterpretSurvey(ctx, template.Name, score, breakdown)
		if err == nil && gptInterpretation != "" {
			sr.Interpretation = gptInterpretation
		}
	}

	if err := s.responseRepo.Create(ctx, sr); err != nil {
		return nil, err
	}

	sr.Template = template
	return sr, nil
}

// CalculateScore is the MVP scoring engine.
// Supports BVAS_V3 (boolean sum), DAS28_CRP (formula), BASDAI (formula).
func CalculateScore(template *entity.SurveyTemplate, responses map[string]interface{}) (score float64, category string, breakdown map[string]any, err error) {
	sections, err := template.GetSections()
	if err != nil {
		return 0, "", nil, err
	}

	breakdown = map[string]any{}

	switch template.Code {
	case "BVAS_V3":
		return calculateBVAS(template, sections, responses)
	case "DAS28_CRP":
		return calculateDAS28CRP(template, responses)
	case "BASDAI":
		return calculateBASDAI(template, responses)
	default:
		return 0, "not_implemented", breakdown, errors.New("scoring for this template is not implemented yet")
	}
}

// calculateBVAS sums boolean scores per section.
func calculateBVAS(template *entity.SurveyTemplate, sections []entity.SurveySection, responses map[string]interface{}) (float64, string, map[string]any, error) {
	breakdown := map[string]any{}
	var score float64

	sectionScores := map[string]float64{}
	for _, sec := range sections {
		secScore := 0.0
		for _, q := range sec.Questions {
			val, ok := responses[q.ID]
			if !ok {
				continue
			}
			b, ok := val.(bool)
			if ok && b {
				secScore += q.Score
			}
		}
		sectionScores[sec.Section] = secScore
		score += secScore
	}
	breakdown["sections"] = sectionScores

	rules, _ := template.GetInterpretationRules()
	category := "unknown"
	if rules != nil {
		for _, r := range rules.Ranges {
			if score >= r.Min && score <= r.Max {
				category = r.Category
				breakdown["category_description"] = r.Description
				break
			}
		}
	}

	return round2(score), category, breakdown, nil
}

// calculateDAS28CRP implements DAS28-CRP formula:
// DAS28-CRP = 0.56*sqrt(TJC28) + 0.28*sqrt(SJC28) + 0.36*ln(CRP+1) + 0.014*GH + 0.96
// Required responses: tjc28, sjc28, crp, gh (0-100 VAS)
func calculateDAS28CRP(template *entity.SurveyTemplate, responses map[string]interface{}) (float64, string, map[string]any, error) {
	breakdown := map[string]any{}

	tjc28 := toFloat(responses["tjc28"])
	sjc28 := toFloat(responses["sjc28"])
	crp := toFloat(responses["crp"])
	gh := toFloat(responses["gh"])

	breakdown["tjc28"] = tjc28
	breakdown["sjc28"] = sjc28
	breakdown["crp"] = crp
	breakdown["gh"] = gh

	// DAS28-CRP formula
	score := 0.56*math.Sqrt(tjc28) + 0.28*math.Sqrt(sjc28) + 0.36*math.Log(crp+1) + 0.014*gh + 0.96
	score = round2(score)
	breakdown["formula"] = "0.56*sqrt(TJC28) + 0.28*sqrt(SJC28) + 0.36*ln(CRP+1) + 0.014*GH + 0.96"

	// Interpretation
	category := "unknown"
	var desc string
	switch {
	case score < 2.6:
		category = "remission"
		desc = "Ремиссия"
	case score < 3.2:
		category = "low_activity"
		desc = "Низкая активность"
	case score <= 5.1:
		category = "moderate_activity"
		desc = "Умеренная активность"
	default:
		category = "high_activity"
		desc = "Высокая активность"
	}
	breakdown["category_description"] = desc

	return score, category, breakdown, nil
}

// calculateBASDAI implements BASDAI formula:
// BASDAI = (Q1 + Q2 + Q3 + Q4 + (Q5+Q6)/2) / 5
// Each question is 0-10 VAS scale.
// Required responses: q1..q6
func calculateBASDAI(template *entity.SurveyTemplate, responses map[string]interface{}) (float64, string, map[string]any, error) {
	breakdown := map[string]any{}

	q1 := toFloat(responses["q1"])
	q2 := toFloat(responses["q2"])
	q3 := toFloat(responses["q3"])
	q4 := toFloat(responses["q4"])
	q5 := toFloat(responses["q5"])
	q6 := toFloat(responses["q6"])

	breakdown["q1"] = q1
	breakdown["q2"] = q2
	breakdown["q3"] = q3
	breakdown["q4"] = q4
	breakdown["q5"] = q5
	breakdown["q6"] = q6

	score := (q1 + q2 + q3 + q4 + (q5+q6)/2) / 5
	score = round2(score)
	breakdown["formula"] = "(Q1 + Q2 + Q3 + Q4 + (Q5+Q6)/2) / 5"

	// Interpretation
	category := "unknown"
	var desc string
	switch {
	case score < 4:
		category = "low_activity"
		desc = "Низкая активность заболевания"
	default:
		category = "high_activity"
		desc = "Высокая активность заболевания"
	}
	breakdown["category_description"] = desc

	return score, category, breakdown, nil
}

// toFloat safely converts interface{} to float64
func toFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case json.Number:
		f, _ := val.Float64()
		return f
	default:
		return 0
	}
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
