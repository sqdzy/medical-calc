package entity

import (
	"time"

	"github.com/google/uuid"
)

// Patient represents a patient in the system
type Patient struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	UserID             uuid.UUID  `json:"user_id" db:"user_id"`
	FullNameEncrypted  string     `json:"-" db:"full_name_encrypted"`
	BirthDateEncrypted string     `json:"-" db:"birth_date_encrypted"`
	SnilsEncrypted     string     `json:"-" db:"snils_encrypted"`
	Gender             string     `json:"gender,omitempty" db:"gender"`
	Diagnosis          string     `json:"diagnosis,omitempty" db:"diagnosis"`
	DiagnosisDate      *time.Time `json:"diagnosis_date,omitempty" db:"diagnosis_date"`
	AttendingDoctorID  *uuid.UUID `json:"attending_doctor_id,omitempty" db:"attending_doctor_id"`
	Notes              string     `json:"notes,omitempty" db:"notes"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`

	// Decrypted fields (populated by service layer)
	FullName  string `json:"full_name,omitempty"`
	BirthDate string `json:"birth_date,omitempty"`
	SNILS     string `json:"snils,omitempty"`

	// Joined fields
	User            *User `json:"user,omitempty"`
	AttendingDoctor *User `json:"attending_doctor,omitempty"`
}

// PatientCreate represents data for creating a new patient
type PatientCreate struct {
	UserID            uuid.UUID  `json:"user_id"`
	FullName          string     `json:"full_name" validate:"required"`
	BirthDate         string     `json:"birth_date" validate:"required"`
	SNILS             string     `json:"snils,omitempty"`
	Gender            string     `json:"gender,omitempty"`
	Diagnosis         string     `json:"diagnosis,omitempty"`
	DiagnosisDate     *time.Time `json:"diagnosis_date,omitempty"`
	AttendingDoctorID *uuid.UUID `json:"attending_doctor_id,omitempty"`
	Notes             string     `json:"notes,omitempty"`
}

// PatientUpdate represents data for updating a patient
type PatientUpdate struct {
	FullName          *string    `json:"full_name,omitempty"`
	BirthDate         *string    `json:"birth_date,omitempty"`
	SNILS             *string    `json:"snils,omitempty"`
	Gender            *string    `json:"gender,omitempty"`
	Diagnosis         *string    `json:"diagnosis,omitempty"`
	DiagnosisDate     *time.Time `json:"diagnosis_date,omitempty"`
	AttendingDoctorID *uuid.UUID `json:"attending_doctor_id,omitempty"`
	Notes             *string    `json:"notes,omitempty"`
}

// PatientListItem represents a patient in list view (minimal data)
type PatientListItem struct {
	ID                uuid.UUID  `json:"id"`
	FullName          string     `json:"full_name"`
	Gender            string     `json:"gender,omitempty"`
	Diagnosis         string     `json:"diagnosis,omitempty"`
	AttendingDoctorID *uuid.UUID `json:"attending_doctor_id,omitempty"`
	DoctorName        string     `json:"doctor_name,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

// PatientFilter represents filter options for patient list
type PatientFilter struct {
	DoctorID  *uuid.UUID `query:"doctor_id"`
	Diagnosis string     `query:"diagnosis"`
	Search    string     `query:"search"`
	Page      int        `query:"page"`
	PerPage   int        `query:"per_page"`
}
