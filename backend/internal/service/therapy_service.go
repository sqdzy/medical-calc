package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/medical-app/backend/internal/entity"
	"github.com/medical-app/backend/internal/repository"
)

type TherapyService struct {
	repo repository.TherapyLogRepository
}

type TherapyDeps struct {
	Repo repository.TherapyLogRepository
}

func NewTherapyService(d TherapyDeps) *TherapyService {
	return &TherapyService{repo: d.Repo}
}

func (s *TherapyService) CreateLog(ctx context.Context, req entity.TherapyLogCreate) (*entity.TherapyLog, error) {
	now := time.Now().UTC()
	status := req.Status
	if status == "" {
		status = entity.TherapyStatusScheduled
	}
	logEntry := &entity.TherapyLog{
		ID:               uuid.New(),
		PatientID:        req.PatientID,
		DrugID:           req.DrugID,
		Dosage:           req.Dosage,
		DosageUnit:       req.DosageUnit,
		Route:            req.Route,
		AdministeredAt:   req.AdministeredAt,
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
