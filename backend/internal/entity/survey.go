package entity

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SurveyOption represents a selectable option for a question.
// New templates use objects with {value,label}. Older templates may still store options as strings.
type SurveyOption struct {
	Value float64 `json:"value"`
	Label string  `json:"label"`
}

// SurveyOptions supports unmarshalling both legacy string arrays and modern option objects.
type SurveyOptions []SurveyOption

func (o *SurveyOptions) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*o = nil
		return nil
	}

	// Preferred format: [{"value": 1, "label": "..."}, ...]
	var asObjects []SurveyOption
	if err := json.Unmarshal(data, &asObjects); err == nil {
		*o = asObjects
		return nil
	}

	// Legacy format: ["Option A", "Option B", ...]
	var asStrings []string
	if err := json.Unmarshal(data, &asStrings); err == nil {
		converted := make([]SurveyOption, 0, len(asStrings))
		for i, s := range asStrings {
			converted = append(converted, SurveyOption{Value: float64(i), Label: s})
		}
		*o = converted
		return nil
	}

	return fmt.Errorf("invalid survey options format: %s", string(data))
}

// SurveyTemplate represents a survey/questionnaire template
type SurveyTemplate struct {
	ID                  uuid.UUID       `json:"id" db:"id"`
	Code                string          `json:"code" db:"code"`
	Name                string          `json:"name" db:"name"`
	Description         string          `json:"description,omitempty" db:"description"`
	Category            string          `json:"category,omitempty" db:"category"`
	Questions           json.RawMessage `json:"questions" db:"questions"`
	ScoringLogic        json.RawMessage `json:"scoring_logic,omitempty" db:"scoring_logic"`
	InterpretationRules json.RawMessage `json:"interpretation_rules,omitempty" db:"interpretation_rules"`
	Version             int             `json:"version" db:"version"`
	IsActive            bool            `json:"is_active" db:"is_active"`
	CreatedBy           *uuid.UUID      `json:"created_by,omitempty" db:"created_by"`
	CreatedAt           time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at" db:"updated_at"`
}

// SurveyQuestion represents a question in a survey
type SurveyQuestion struct {
	ID       string                 `json:"id"`
	Text     string                 `json:"text"`
	Type     string                 `json:"type"` // boolean, number, scale, select, text
	Score    float64                `json:"score,omitempty"`
	Min      float64                `json:"min,omitempty"`
	Max      float64                `json:"max,omitempty"`
	Options  SurveyOptions          `json:"options,omitempty"`
	Labels   map[string]string      `json:"labels,omitempty"`
	Required bool                   `json:"required,omitempty"`
	Extra    map[string]interface{} `json:"extra,omitempty"`
}

// SurveySection represents a section of questions
type SurveySection struct {
	Section   string           `json:"section"`
	Title     string           `json:"title"`
	Questions []SurveyQuestion `json:"questions"`
}

// ScoringLogic represents how to calculate the survey score
type ScoringLogic struct {
	Type     string   `json:"type"` // "sum", "formula", "custom"
	Formula  string   `json:"formula,omitempty"`
	Sections []string `json:"sections,omitempty"`
}

// InterpretationRule represents score interpretation thresholds
type InterpretationRule struct {
	Min         float64 `json:"min"`
	Max         float64 `json:"max"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
}

// InterpretationRules wraps the array of rules
type InterpretationRules struct {
	Ranges []InterpretationRule `json:"ranges"`
}

// SurveyResponse represents a patient's survey submission
type SurveyResponse struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	TemplateID      uuid.UUID       `json:"template_id" db:"template_id"`
	PatientID       uuid.UUID       `json:"patient_id" db:"patient_id"`
	Responses       json.RawMessage `json:"responses" db:"responses"`
	CalculatedScore *float64        `json:"calculated_score,omitempty" db:"calculated_score"`
	ScoreBreakdown  json.RawMessage `json:"score_breakdown,omitempty" db:"score_breakdown"`
	Interpretation  string          `json:"interpretation,omitempty" db:"interpretation"`
	AISummary       string          `json:"ai_summary,omitempty" db:"ai_summary"`
	Status          string          `json:"status" db:"status"`
	SubmittedAt     time.Time       `json:"submitted_at" db:"submitted_at"`
	ReviewedBy      *uuid.UUID      `json:"reviewed_by,omitempty" db:"reviewed_by"`
	ReviewedAt      *time.Time      `json:"reviewed_at,omitempty" db:"reviewed_at"`
	Notes           string          `json:"notes,omitempty" db:"notes"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`

	// Joined fields
	Template *SurveyTemplate `json:"template,omitempty"`
	Patient  *Patient        `json:"patient,omitempty"`
	Reviewer *User           `json:"reviewer,omitempty"`
}

// Response status constants
const (
	SurveyStatusDraft     = "draft"
	SurveyStatusSubmitted = "submitted"
	SurveyStatusReviewed  = "reviewed"
)

// SurveyResponseCreate represents data for submitting a survey
type SurveyResponseCreate struct {
	TemplateID uuid.UUID              `json:"template_id" validate:"required"`
	PatientID  uuid.UUID              `json:"patient_id" validate:"required"`
	Responses  map[string]interface{} `json:"responses" validate:"required"`
	Status     string                 `json:"status,omitempty"`
}

// SurveyResponseFilter represents filter options for survey responses
type SurveyResponseFilter struct {
	PatientID  *uuid.UUID `query:"patient_id"`
	TemplateID *uuid.UUID `query:"template_id"`
	Status     string     `query:"status"`
	DateFrom   *time.Time `query:"date_from"`
	DateTo     *time.Time `query:"date_to"`
	Page       int        `query:"page"`
	PerPage    int        `query:"per_page"`
}

// GetSections parses the questions JSON into sections
func (t *SurveyTemplate) GetSections() ([]SurveySection, error) {
	var sections []SurveySection
	if err := json.Unmarshal(t.Questions, &sections); err != nil {
		return nil, err
	}
	return sections, nil
}

// GetScoringLogic parses the scoring logic JSON
func (t *SurveyTemplate) GetScoringLogic() (*ScoringLogic, error) {
	if t.ScoringLogic == nil {
		return nil, nil
	}
	var logic ScoringLogic
	if err := json.Unmarshal(t.ScoringLogic, &logic); err != nil {
		return nil, err
	}
	return &logic, nil
}

// GetInterpretationRules parses the interpretation rules JSON
func (t *SurveyTemplate) GetInterpretationRules() (*InterpretationRules, error) {
	if t.InterpretationRules == nil {
		return nil, nil
	}
	var rules InterpretationRules
	if err := json.Unmarshal(t.InterpretationRules, &rules); err != nil {
		return nil, err
	}
	return &rules, nil
}
