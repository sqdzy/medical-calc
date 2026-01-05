package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/medical-app/backend/internal/entity"
)

type patientRepository struct {
	db *pgxpool.Pool
	sb squirrel.StatementBuilderType
}

func NewPatientRepository(db *pgxpool.Pool) *patientRepository {
	return &patientRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *patientRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Patient, error) {
	q := r.sb.Select(
		"id", "user_id", "full_name_encrypted", "birth_date_encrypted", "snils_encrypted",
		"gender", "diagnosis", "diagnosis_date", "attending_doctor_id", "notes",
		"created_at", "updated_at",
	).From("patients").Where(squirrel.Eq{"id": id})

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	var p entity.Patient
	if err := row.Scan(
		&p.ID, &p.UserID, &p.FullNameEncrypted, &p.BirthDateEncrypted, &p.SnilsEncrypted,
		&p.Gender, &p.Diagnosis, &p.DiagnosisDate, &p.AttendingDoctorID, &p.Notes,
		&p.CreatedAt, &p.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("select patient: %w", err)
	}
	return &p, nil
}

func (r *patientRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Patient, error) {
	q := r.sb.Select(
		"id", "user_id", "full_name_encrypted", "birth_date_encrypted", "snils_encrypted",
		"gender", "diagnosis", "diagnosis_date", "attending_doctor_id", "notes",
		"created_at", "updated_at",
	).From("patients").Where(squirrel.Eq{"user_id": userID})

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	var p entity.Patient
	if err := row.Scan(
		&p.ID, &p.UserID, &p.FullNameEncrypted, &p.BirthDateEncrypted, &p.SnilsEncrypted,
		&p.Gender, &p.Diagnosis, &p.DiagnosisDate, &p.AttendingDoctorID, &p.Notes,
		&p.CreatedAt, &p.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("select patient by user_id: %w", err)
	}
	return &p, nil
}

func (r *patientRepository) Create(ctx context.Context, patient *entity.Patient) error {
	now := time.Now().UTC()
	q := r.sb.Insert("patients").
		Columns(
			"id", "user_id", "full_name_encrypted", "birth_date_encrypted", "snils_encrypted",
			"gender", "diagnosis", "diagnosis_date", "attending_doctor_id", "notes",
			"created_at", "updated_at",
		).
		Values(
			patient.ID, patient.UserID, patient.FullNameEncrypted, patient.BirthDateEncrypted, patient.SnilsEncrypted,
			patient.Gender, patient.Diagnosis, patient.DiagnosisDate, patient.AttendingDoctorID, patient.Notes,
			now, now,
		)

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("build sql: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("insert patient: %w", err)
	}
	return nil
}
