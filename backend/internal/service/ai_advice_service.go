package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/medical-app/backend/internal/entity"
	"github.com/medical-app/backend/internal/external"
	"github.com/medical-app/backend/internal/repository"
)

const PatientAdviceDisclaimer = "Важно: это информационная справка, а не клиническая рекомендация и не заменяет консультацию врача. При ухудшении самочувствия, сильной боли, высокой температуре, одышке или других тревожных симптомах обратитесь за медицинской помощью."

func normalizeAdviceText(text string) string {
	out := strings.TrimSpace(text)
	if out == "" {
		return out
	}

	// If model included our exact disclaimer, remove it (we show it separately).
	out = strings.ReplaceAll(out, PatientAdviceDisclaimer, "")
	out = strings.TrimSpace(out)

	// Heuristic: model sometimes appends its own "Важно: ..." disclaimer.
	// If the last "Важно:" block looks like a disclaimer, strip it.
	lower := strings.ToLower(out)
	lastImportant := strings.LastIndex(lower, "\nважно:")
	if lastImportant == -1 {
		lastImportant = strings.LastIndex(lower, "важно:")
	}
	if lastImportant != -1 {
		// Only strip if it's near the end and looks like a safety disclaimer.
		if len(out)-lastImportant <= 700 {
			tail := lower[lastImportant:]
			if strings.Contains(tail, "не является") || strings.Contains(tail, "не клиничес") || strings.Contains(tail, "не замен") || strings.Contains(tail, "консульта") || strings.Contains(tail, "врач") {
				out = strings.TrimSpace(out[:lastImportant])
			}
		}
	}

	return out
}

type AIAdviceService struct {
	templateRepo repository.SurveyTemplateRepository
	patientRepo  repository.PatientRepository
	adviceRepo   repository.AIAdviceRepository
	gptClient    *external.YandexGPTClient
}

type AIAdviceDeps struct {
	TemplateRepo repository.SurveyTemplateRepository
	PatientRepo  repository.PatientRepository
	AdviceRepo   repository.AIAdviceRepository
	GPTClient    *external.YandexGPTClient
}

func NewAIAdviceService(d AIAdviceDeps) *AIAdviceService {
	return &AIAdviceService{
		templateRepo: d.TemplateRepo,
		patientRepo:  d.PatientRepo,
		adviceRepo:   d.AdviceRepo,
		gptClient:    d.GPTClient,
	}
}

type AIAdviceResult struct {
	ID         uuid.UUID `json:"id"`
	SurveyCode string    `json:"survey_code"`
	UserText   string    `json:"user_text,omitempty"`
	AdviceText string    `json:"advice_text"`
	Disclaimer string    `json:"disclaimer"`
	Score      *float64  `json:"score,omitempty"`
	Category   string    `json:"category,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

func (s *AIAdviceService) CreateForUser(ctx context.Context, userID uuid.UUID, surveyCode string, answers map[string]any, userText string) (*AIAdviceResult, error) {
	if userID == uuid.Nil {
		return nil, errors.New("user_id is required")
	}
	surveyCode = strings.TrimSpace(surveyCode)
	if surveyCode == "" {
		return nil, errors.New("survey_code is required")
	}

	t, err := s.templateRepo.GetByCode(ctx, surveyCode)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, errors.New("survey template not found")
	}

	score, category, breakdown, err := CalculateScore(t, answers)
	if err != nil {
		return nil, err
	}

	patient, err := s.patientRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if patient == nil {
		patient = &entity.Patient{ID: uuid.New(), UserID: userID, FullNameEncrypted: "", BirthDateEncrypted: ""}
		if err := s.patientRepo.Create(ctx, patient); err != nil {
			return nil, err
		}
	}

	adviceText := s.fallbackAdviceText(t, score, category, breakdown)
	if s.gptClient != nil {
		gptText, err := s.generatePatientAdvice(ctx, t, score, category, breakdown, userText)
		if err != nil {
			log.Printf("[AIAdvice] YandexGPT error (using fallback): %v", err)
		} else {
			trimmed := normalizeAdviceText(gptText)
			if trimmed != "" {
				adviceText = trimmed
			} else {
				log.Printf("[AIAdvice] YandexGPT returned empty response, using fallback")
			}
		}
	} else {
		log.Printf("[AIAdvice] gptClient is nil, using fallback")
	}

	detailsJSON, _ := json.Marshal(map[string]any{
		"score":     score,
		"category":  category,
		"breakdown": breakdown,
	})

	now := time.Now().UTC()
	item := &entity.AIAdvice{
		ID:         uuid.New(),
		PatientID:  patient.ID,
		SurveyCode: surveyCode,
		UserText:   strings.TrimSpace(userText),
		Score:      &score,
		Category:   category,
		Details:    detailsJSON,
		AdviceText: normalizeAdviceText(adviceText),
		CreatedAt:  now,
	}

	if err := s.adviceRepo.Create(ctx, item); err != nil {
		return nil, err
	}

	return &AIAdviceResult{
		ID:         item.ID,
		SurveyCode: item.SurveyCode,
		UserText:   item.UserText,
		AdviceText: item.AdviceText,
		Disclaimer: PatientAdviceDisclaimer,
		Score:      item.Score,
		Category:   item.Category,
		CreatedAt:  item.CreatedAt,
	}, nil
}

func (s *AIAdviceService) ListForUser(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*AIAdviceResult, error) {
	if userID == uuid.Nil {
		return nil, errors.New("user_id is required")
	}
	patient, err := s.patientRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if patient == nil {
		return []*AIAdviceResult{}, nil
	}

	items, err := s.adviceRepo.ListByPatient(ctx, patient.ID, limit, offset)
	if err != nil {
		return nil, err
	}

	out := make([]*AIAdviceResult, 0, len(items))
	for _, it := range items {
		out = append(out, &AIAdviceResult{
			ID:         it.ID,
			SurveyCode: it.SurveyCode,
			UserText:   it.UserText,
			AdviceText: normalizeAdviceText(it.AdviceText),
			Disclaimer: PatientAdviceDisclaimer,
			Score:      it.Score,
			Category:   it.Category,
			CreatedAt:  it.CreatedAt,
		})
	}
	return out, nil
}

func (s *AIAdviceService) fallbackAdviceText(t *entity.SurveyTemplate, score float64, category string, breakdown map[string]any) string {
	interpretation := category
	if desc, ok := breakdown["category_description"].(string); ok && strings.TrimSpace(desc) != "" {
		interpretation = desc
	}
	return fmt.Sprintf(
		"Результат опросника: %s (%.2f)\n\n%s",
		t.Name,
		score,
		interpretation,
	)
}

func (s *AIAdviceService) generatePatientAdvice(ctx context.Context, t *entity.SurveyTemplate, score float64, category string, breakdown map[string]any, userText string) (string, error) {
	systemPrompt := `Ты — медицинский информационный помощник для пациента.
Твои ответы должны быть безопасными и не содержать постановки диагноза, назначения рецептурных препаратов или дозировок.
Пиши простым русским языком.
Не добавляй дисклеймеры/предупреждения в стиле "Важно:" — приложение покажет стандартное предупреждение отдельно.
Структура ответа:
1) Короткое резюме результата (1-2 предложения)
2) Что это может означать (общими словами)
3) Что можно сделать сейчас (общие меры, без лечения/доз)
4) Когда обратиться к врачу срочно
5) Вопросы для обсуждения с врачом`

	userPrompt := fmt.Sprintf(
		"Опросник: %s (%s)\nИтоговый балл: %.2f\nКатегория: %s\nДетали: %v\n\nКомментарий пациента (если есть - учесть в ответе): %s\n",
		t.Name,
		t.Code,
		score,
		category,
		breakdown,
		strings.TrimSpace(userText),
	)

	text, err := s.gptClient.GenerateText(ctx, systemPrompt, userPrompt)
	if err != nil {
		return "", err
	}
	return text, nil
}
