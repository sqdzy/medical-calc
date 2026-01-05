package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/medical-app/backend/internal/entity"
	"github.com/medical-app/backend/internal/repository"
	"github.com/medical-app/backend/pkg/validator"
)

type TherapyService struct {
	repo        repository.TherapyLogRepository
	patientRepo repository.PatientRepository
}

type TherapyDeps struct {
	Repo        repository.TherapyLogRepository
	PatientRepo repository.PatientRepository
}

func NewTherapyService(d TherapyDeps) *TherapyService {
	return &TherapyService{repo: d.Repo, patientRepo: d.PatientRepo}
}

// CreateLogForUser resolves/creates a patient record for the given userID, then creates the therapy log.
func (s *TherapyService) CreateLogForUser(ctx context.Context, userID uuid.UUID, req entity.TherapyLogCreate) (*entity.TherapyLog, error) {
	patient, err := s.patientRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if patient == nil {
		// Auto-create minimal patient record for this user
		patient = &entity.Patient{
			ID:                 uuid.New(),
			UserID:             userID,
			FullNameEncrypted:  "", // placeholder
			BirthDateEncrypted: "", // placeholder
		}
		if err := s.patientRepo.Create(ctx, patient); err != nil {
			return nil, err
		}
	}
	req.PatientID = patient.ID
	return s.CreateLog(ctx, req)
}

// CreateLog inserts a therapy log; caller must ensure req.PatientID is valid.
func (s *TherapyService) CreateLog(ctx context.Context, req entity.TherapyLogCreate) (*entity.TherapyLog, error) {
	v := validator.New()
	if req.PatientID == uuid.Nil {
		v.AddError("patient_id", "patient_id is required")
	}
	if req.DrugID == uuid.Nil {
		v.AddError("drug_id", "drug_id is required")
	}

	// Dosage must be numeric (clients may submit it as string or number; entity.Dosage already normalizes to string)
	rawDosage := strings.TrimSpace(string(req.Dosage))
	if rawDosage == "" {
		v.AddError("dosage", "dosage is required")
	} else {
		normalized := strings.ReplaceAll(rawDosage, ",", ".")
		val, err := strconv.ParseFloat(normalized, 64)
		if err != nil {
			v.AddError("dosage", "dosage must be a number")
		} else if val <= 0 {
			v.AddError("dosage", "dosage must be greater than 0")
		} else {
			req.Dosage = entity.Dosage(normalized)
		}
	}

	if strings.TrimSpace(req.DosageUnit) == "" {
		v.AddError("dosage_unit", "dosage_unit is required")
	}

	if req.Status != "" {
		v.OneOf("status", req.Status, []string{
			entity.TherapyStatusScheduled,
			entity.TherapyStatusCompleted,
			entity.TherapyStatusMissed,
			entity.TherapyStatusCancelled,
		}, "invalid status")
	}

	if v.HasErrors() {
		return nil, v.Errors()
	}

	now := time.Now().UTC()
	adminAt := req.AdministeredAt
	if adminAt == nil {
		if req.NextScheduled != nil {
			adminAt = req.NextScheduled
		} else {
			t := now
			adminAt = &t
		}
	}
	status := req.Status
	if status == "" {
		status = entity.TherapyStatusScheduled
	}
	logEntry := &entity.TherapyLog{
		ID:               uuid.New(),
		PatientID:        req.PatientID,
		DrugID:           req.DrugID,
		Dosage:           string(req.Dosage),
		DosageUnit:       req.DosageUnit,
		Route:            req.Route,
		AdministeredAt:   adminAt,
		NextScheduled:    req.NextScheduled,
		CycleNumber:      req.CycleNumber,
		BatchNumber:      req.BatchNumber,
		Site:             req.Site,
		AdministeredByID: req.AdministeredByID,
		Status:           status,
		Notes:            req.Notes,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.repo.Create(ctx, logEntry); err != nil {
		return nil, err
	}
	return logEntry, nil
}

func (s *TherapyService) ListByPatient(ctx context.Context, patientID uuid.UUID, limit int) ([]*entity.TherapyLog, error) {
	return s.repo.ListByPatient(ctx, patientID, limit)
}

// ResolvePatientID accepts either a patients.id or users.id and returns the corresponding patients.id.
func (s *TherapyService) ResolvePatientID(ctx context.Context, id uuid.UUID) (uuid.UUID, bool, error) {
	if s.patientRepo == nil {
		return id, true, nil
	}

	// First treat it as patients.id
	p, err := s.patientRepo.GetByID(ctx, id)
	if err != nil {
		return uuid.Nil, false, err
	}
	if p != nil {
		return p.ID, true, nil
	}

	// Fallback: treat it as users.id
	p, err = s.patientRepo.GetByUserID(ctx, id)
	if err != nil {
		return uuid.Nil, false, err
	}
	if p == nil {
		return uuid.Nil, false, nil
	}
	return p.ID, true, nil
}

// DeleteLogForUser deletes a therapy log if it belongs to the authenticated user.
func (s *TherapyService) DeleteLogForUser(ctx context.Context, userID uuid.UUID, logID uuid.UUID) (bool, error) {
	if userID == uuid.Nil {
		return false, errors.New("user_id is required")
	}
	if logID == uuid.Nil {
		return false, errors.New("log_id is required")
	}

	patient, err := s.patientRepo.GetByUserID(ctx, userID)
	if err != nil {
		return false, err
	}
	if patient == nil {
		return false, nil
	}
	return s.repo.DeleteByID(ctx, patient.ID, logID)
}
