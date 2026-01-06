package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AIAdvice struct {
	ID         uuid.UUID       `json:"id"`
	PatientID  uuid.UUID       `json:"patient_id"`
	SurveyCode string          `json:"survey_code"`
	UserText   string          `json:"user_text,omitempty"`
	Score      *float64        `json:"score,omitempty"`
	Category   string          `json:"category,omitempty"`
	Details    json.RawMessage `json:"details,omitempty"`
	AdviceText string          `json:"advice_text"`
	CreatedAt  time.Time       `json:"created_at"`
}

type AIAdviceCreate struct {
	SurveyCode string         `json:"survey_code"`
	UserText   string         `json:"user_text"`
	Score      *float64       `json:"score,omitempty"`
	Category   string         `json:"category,omitempty"`
	Details    map[string]any `json:"details,omitempty"`
	AdviceText string         `json:"advice_text"`
	PatientID  uuid.UUID      `json:"-"`
}
