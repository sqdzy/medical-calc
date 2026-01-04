package entity

import (
	"time"

	"github.com/google/uuid"
)

// TherapyLog represents a drug administration record
type TherapyLog struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	PatientID        uuid.UUID  `json:"patient_id" db:"patient_id"`
	DrugID           uuid.UUID  `json:"drug_id" db:"drug_id"`
	Dosage           string     `json:"dosage" db:"dosage"`
	DosageUnit       string     `json:"dosage_unit,omitempty" db:"dosage_unit"`
	Route            string     `json:"route,omitempty" db:"route"`
	AdministeredAt   *time.Time `json:"administered_at,omitempty" db:"administered_at"`
	NextScheduled    *time.Time `json:"next_scheduled,omitempty" db:"next_scheduled"`
	CycleNumber      int        `json:"cycle_number,omitempty" db:"cycle_number"`
	BatchNumber      string     `json:"batch_number,omitempty" db:"batch_number"`
	Site             string     `json:"site,omitempty" db:"site"`
	AdministeredByID *uuid.UUID `json:"administered_by,omitempty" db:"administered_by"`
	Status           string     `json:"status" db:"status"`
	AdverseReactions string     `json:"adverse_reactions,omitempty" db:"adverse_reactions"`
	Notes            string     `json:"notes,omitempty" db:"notes"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`

	// Joined fields
	Patient        *Patient `json:"patient,omitempty"`
	Drug           *Drug    `json:"drug,omitempty"`
	AdministeredBy *User    `json:"administered_by_user,omitempty"`
}

// TherapyLog status constants
const (
	TherapyStatusScheduled = "scheduled"
	TherapyStatusCompleted = "completed"
	TherapyStatusMissed    = "missed"
	TherapyStatusCancelled = "cancelled"
)

// Route constants
const (
	RouteSubcutaneous  = "subcutaneous"
	RouteIntravenous   = "intravenous"
	RouteIntramuscular = "intramuscular"
	RouteOral          = "oral"
)

// TherapyLogCreate represents data for creating a therapy log
type TherapyLogCreate struct {
	PatientID        uuid.UUID  `json:"patient_id" validate:"required"`
	DrugID           uuid.UUID  `json:"drug_id" validate:"required"`
	Dosage           string     `json:"dosage" validate:"required"`
	DosageUnit       string     `json:"dosage_unit,omitempty"`
	Route            string     `json:"route,omitempty"`
	AdministeredAt   *time.Time `json:"administered_at,omitempty"`
	NextScheduled    *time.Time `json:"next_scheduled,omitempty"`
	CycleNumber      int        `json:"cycle_number,omitempty"`
	BatchNumber      string     `json:"batch_number,omitempty"`
	Site             string     `json:"site,omitempty"`
	AdministeredByID *uuid.UUID `json:"administered_by,omitempty"`
	Status           string     `json:"status,omitempty"`
	Notes            string     `json:"notes,omitempty"`
}

// TherapyLogUpdate represents data for updating a therapy log
type TherapyLogUpdate struct {
	Dosage           *string    `json:"dosage,omitempty"`
	DosageUnit       *string    `json:"dosage_unit,omitempty"`
	Route            *string    `json:"route,omitempty"`
	AdministeredAt   *time.Time `json:"administered_at,omitempty"`
	NextScheduled    *time.Time `json:"next_scheduled,omitempty"`
	CycleNumber      *int       `json:"cycle_number,omitempty"`
	BatchNumber      *string    `json:"batch_number,omitempty"`
	Site             *string    `json:"site,omitempty"`
	AdministeredByID *uuid.UUID `json:"administered_by,omitempty"`
	Status           *string    `json:"status,omitempty"`
	AdverseReactions *string    `json:"adverse_reactions,omitempty"`
	Notes            *string    `json:"notes,omitempty"`
}

// TherapyLogFilter represents filter options for therapy logs
type TherapyLogFilter struct {
	PatientID     *uuid.UUID `query:"patient_id"`
	DrugID        *uuid.UUID `query:"drug_id"`
	Status        string     `query:"status"`
	DateFrom      *time.Time `query:"date_from"`
	DateTo        *time.Time `query:"date_to"`
	ScheduledFrom *time.Time `query:"scheduled_from"`
	ScheduledTo   *time.Time `query:"scheduled_to"`
	Page          int        `query:"page"`
	PerPage       int        `query:"per_page"`
}

// MedicalIndex represents a calculated medical index for a patient
type MedicalIndex struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	PatientID        uuid.UUID  `json:"patient_id" db:"patient_id"`
	IndexType        string     `json:"index_type" db:"index_type"`
	Value            float64    `json:"value" db:"value"`
	Category         string     `json:"category,omitempty" db:"category"`
	SurveyResponseID *uuid.UUID `json:"survey_response_id,omitempty" db:"survey_response_id"`
	Notes            string     `json:"notes,omitempty" db:"notes"`
	RecordedByID     *uuid.UUID `json:"recorded_by,omitempty" db:"recorded_by"`
	RecordedAt       time.Time  `json:"recorded_at" db:"recorded_at"`
}

// MedicalIndexCreate represents data for creating a medical index
type MedicalIndexCreate struct {
	PatientID        uuid.UUID  `json:"patient_id" validate:"required"`
	IndexType        string     `json:"index_type" validate:"required"`
	Value            float64    `json:"value" validate:"required"`
	Category         string     `json:"category,omitempty"`
	SurveyResponseID *uuid.UUID `json:"survey_response_id,omitempty"`
	Notes            string     `json:"notes,omitempty"`
	RecordedByID     *uuid.UUID `json:"recorded_by,omitempty"`
}

// CalendarEvent represents a therapy event for calendar view
type CalendarEvent struct {
	ID          uuid.UUID `json:"id"`
	PatientID   uuid.UUID `json:"patient_id"`
	PatientName string    `json:"patient_name"`
	DrugName    string    `json:"drug_name"`
	Dosage      string    `json:"dosage"`
	ScheduledAt time.Time `json:"scheduled_at"`
	Status      string    `json:"status"`
}
