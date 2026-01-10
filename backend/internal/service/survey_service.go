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
// Supports BVAS_V3 (boolean sum), DAS28_CRP (formula), BASDAI (formula),
// ASA (direct value), RCRI (boolean sum), GOLDMAN (boolean sum with weights), CAPRINI (boolean sum with weights).
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
	case "ASA":
		return calculateASA(template, responses)
	case "RCRI":
		return calculateRCRI(template, sections, responses)
	case "GOLDMAN":
		return calculateGoldman(template, sections, responses)
	case "CAPRINI":
		return calculateCaprini(template, sections, responses)
	default:
		// Try generic sum for unknown templates
		return calculateGenericSum(template, sections, responses)
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

// calculateASA handles ASA Physical Status Classification (direct value)
func calculateASA(template *entity.SurveyTemplate, responses map[string]interface{}) (float64, string, map[string]any, error) {
	breakdown := map[string]any{}

	// Get ASA class (1-6)
	asaClass := toFloat(responses["asa_class"])
	isEmergency := false
	if v, ok := responses["is_emergency"]; ok {
		isEmergency = toBool(v)
	}

	breakdown["asa_class"] = int(asaClass)
	breakdown["is_emergency"] = isEmergency

	// Determine category and risk description
	var category, description string
	var mortalityRate string

	switch int(asaClass) {
	case 1:
		category = "asa_1"
		description = "Здоровый пациент без системных заболеваний"
		mortalityRate = "0.1%"
	case 2:
		category = "asa_2"
		description = "Легкое системное заболевание без функциональных ограничений"
		mortalityRate = "0.2%"
	case 3:
		category = "asa_3"
		description = "Тяжелое системное заболевание с функциональными ограничениями"
		mortalityRate = "1.8%"
	case 4:
		category = "asa_4"
		description = "Тяжелое заболевание, постоянно угрожающее жизни"
		mortalityRate = "7.8%"
	case 5:
		category = "asa_5"
		description = "Умирающий пациент, не ожидающий выживания без операции"
		mortalityRate = "9.4%"
	case 6:
		category = "asa_6"
		description = "Донор органов с подтвержденной смертью мозга"
		mortalityRate = "N/A"
	default:
		category = "unknown"
		description = "Неопределенный класс ASA"
		mortalityRate = "N/A"
	}

	if isEmergency && asaClass < 6 {
		category += "_e"
		description += " (ЭКСТРЕННАЯ операция)"
		breakdown["emergency_modifier"] = "E"
	}

	breakdown["description"] = description
	breakdown["mortality_rate"] = mortalityRate

	return asaClass, category, breakdown, nil
}

// calculateRCRI handles Revised Cardiac Risk Index (Lee Index)
func calculateRCRI(template *entity.SurveyTemplate, sections []entity.SurveySection, responses map[string]interface{}) (float64, string, map[string]any, error) {
	breakdown := map[string]any{}
	riskFactors := []string{}

	// Count risk factors (6 boolean fields)
	fields := map[string]string{
		"high_risk_surgery": "Операция высокого риска",
		"ihd":               "Ишемическая болезнь сердца",
		"chf":               "Хроническая сердечная недостаточность",
		"cvd":               "Цереброваскулярное заболевание",
		"insulin_dm":        "Сахарный диабет на инсулинотерапии",
		"ckd":               "Хроническая болезнь почек (креатинин > 177 мкмоль/л)",
	}

	score := 0.0
	for fieldKey, fieldName := range fields {
		if toBool(responses[fieldKey]) {
			score++
			riskFactors = append(riskFactors, fieldName)
		}
	}

	breakdown["risk_factors"] = riskFactors
	breakdown["risk_factor_count"] = int(score)

	// Determine MACE risk category
	var category, maceRisk, description string
	switch int(score) {
	case 0:
		category = "class_i"
		maceRisk = "3.9%"
		description = "Класс I - минимальный риск MACE"
	case 1:
		category = "class_ii"
		maceRisk = "6.0%"
		description = "Класс II - низкий риск MACE"
	case 2:
		category = "class_iii"
		maceRisk = "10.1%"
		description = "Класс III - умеренный риск MACE"
	default:
		category = "class_iv"
		maceRisk = "15%+"
		description = "Класс IV - высокий риск MACE"
	}

	breakdown["mace_risk"] = maceRisk
	breakdown["description"] = description

	return score, category, breakdown, nil
}

// calculateGoldman handles Goldman Cardiac Risk Index (weighted boolean sum)
func calculateGoldman(template *entity.SurveyTemplate, sections []entity.SurveySection, responses map[string]interface{}) (float64, string, map[string]any, error) {
	return calculateWeightedBooleanSum(template, sections, responses, func(score float64) (string, string) {
		switch {
		case score <= 5:
			return "class_i", "Класс I - минимальный риск (0-5 баллов)"
		case score <= 12:
			return "class_ii", "Класс II - низкий риск (6-12 баллов)"
		case score <= 25:
			return "class_iii", "Класс III - умеренный риск (13-25 баллов)"
		default:
			return "class_iv", "Класс IV - высокий риск (>25 баллов)"
		}
	})
}

// calculateCaprini handles Caprini VTE Risk Score (weighted boolean sum)
func calculateCaprini(template *entity.SurveyTemplate, sections []entity.SurveySection, responses map[string]interface{}) (float64, string, map[string]any, error) {
	return calculateWeightedBooleanSum(template, sections, responses, func(score float64) (string, string) {
		switch {
		case score == 0:
			return "very_low", "Очень низкий риск ВТЭ (0 баллов)"
		case score <= 2:
			return "low", "Низкий риск ВТЭ (1-2 балла)"
		case score <= 4:
			return "moderate", "Умеренный риск ВТЭ (3-4 балла)"
		default:
			return "high", "Высокий риск ВТЭ (≥5 баллов)"
		}
	})
}

// calculateWeightedBooleanSum is a generic function for weighted boolean scoring
func calculateWeightedBooleanSum(template *entity.SurveyTemplate, sections []entity.SurveySection, responses map[string]interface{}, interpret func(float64) (string, string)) (float64, string, map[string]any, error) {
	breakdown := map[string]any{}
	positiveFactors := []string{}
	sectionScores := map[string]int{}

	var totalScore float64
	for _, sec := range sections {
		secScore := 0
		for _, q := range sec.Questions {
			if toBool(responses[q.ID]) {
				points := 1
				if q.Score > 0 {
					points = int(math.Round(q.Score))
				}
				secScore += points
				positiveFactors = append(positiveFactors, q.Text)
			}
		}
		sectionScores[sec.Title] = secScore
		totalScore += float64(secScore)
	}

	breakdown["section_scores"] = sectionScores
	breakdown["positive_factors"] = positiveFactors

	category, description := interpret(totalScore)
	breakdown["description"] = description

	return totalScore, category, breakdown, nil
}

// calculateGenericSum is a fallback for unknown templates using simple boolean sum
func calculateGenericSum(template *entity.SurveyTemplate, sections []entity.SurveySection, responses map[string]interface{}) (float64, string, map[string]any, error) {
	breakdown := map[string]any{}
	var totalScore float64

	for _, sec := range sections {
		for _, q := range sec.Questions {
			if toBool(responses[q.ID]) {
				points := 1
				if q.Score > 0 {
					points = int(math.Round(q.Score))
				}
				totalScore += float64(points)
			}
		}
	}

	return totalScore, "calculated", breakdown, nil
}

// toBool safely converts interface{} to bool
func toBool(v interface{}) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	case int:
		return val != 0
	case float64:
		return val != 0
	case string:
		return val == "true" || val == "1" || val == "yes"
	default:
		return false
	}
}
