package entity

import (
	"time"

	"github.com/google/uuid"
)

// Drug represents a medication in the system
type Drug struct {
	ID                uuid.UUID `json:"id" db:"id"`
	Name              string    `json:"name" db:"name"`
	InternationalName string    `json:"international_name,omitempty" db:"international_name"`
	TradeName         string    `json:"trade_name,omitempty" db:"trade_name"`
	NCBIPubchemID     string    `json:"ncbi_pubchem_id,omitempty" db:"ncbi_pubchem_id"`
	ATCCode           string    `json:"atc_code,omitempty" db:"atc_code"`
	DosageForm        string    `json:"dosage_form,omitempty" db:"dosage_form"`
	Manufacturer      string    `json:"manufacturer,omitempty" db:"manufacturer"`
	Description       string    `json:"description,omitempty" db:"description"`
	Contraindications string    `json:"contraindications,omitempty" db:"contraindications"`
	IsActive          bool      `json:"is_active" db:"is_active"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// DrugCreate represents data for creating a new drug
type DrugCreate struct {
	Name              string `json:"name" validate:"required"`
	InternationalName string `json:"international_name,omitempty"`
	TradeName         string `json:"trade_name,omitempty"`
	NCBIPubchemID     string `json:"ncbi_pubchem_id,omitempty"`
	ATCCode           string `json:"atc_code,omitempty"`
	DosageForm        string `json:"dosage_form,omitempty"`
	Manufacturer      string `json:"manufacturer,omitempty"`
	Description       string `json:"description,omitempty"`
	Contraindications string `json:"contraindications,omitempty"`
}

// DrugUpdate represents data for updating a drug
type DrugUpdate struct {
	Name              *string `json:"name,omitempty"`
	InternationalName *string `json:"international_name,omitempty"`
	TradeName         *string `json:"trade_name,omitempty"`
	NCBIPubchemID     *string `json:"ncbi_pubchem_id,omitempty"`
	ATCCode           *string `json:"atc_code,omitempty"`
	DosageForm        *string `json:"dosage_form,omitempty"`
	Manufacturer      *string `json:"manufacturer,omitempty"`
	Description       *string `json:"description,omitempty"`
	Contraindications *string `json:"contraindications,omitempty"`
	IsActive          *bool   `json:"is_active,omitempty"`
}

// DrugFilter represents filter options for drug list
type DrugFilter struct {
	Search   string `query:"search"`
	ATCCode  string `query:"atc_code"`
	IsActive *bool  `query:"is_active"`
	Page     int    `query:"page"`
	PerPage  int    `query:"per_page"`
}
