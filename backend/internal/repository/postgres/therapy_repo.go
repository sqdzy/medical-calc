package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/medical-app/backend/internal/entity"
)

type therapyLogRepository struct {
	db *pgxpool.Pool
	sb squirrel.StatementBuilderType
}

func NewTherapyLogRepository(db *pgxpool.Pool) *therapyLogRepository {
	return &therapyLogRepository{db: db, sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}
}

func (r *therapyLogRepository) Create(ctx context.Context, logEntry *entity.TherapyLog) error {
	q := r.sb.Insert("therapy_logs").
		Columns(
			"id", "patient_id", "drug_id", "dosage", "dosage_unit", "route",
			"administered_at", "next_scheduled", "cycle_number", "batch_number", "site",
			"administered_by", "status", "adverse_reactions", "notes", "created_at", "updated_at",
		).
		Values(
			logEntry.ID, logEntry.PatientID, logEntry.DrugID, logEntry.Dosage, logEntry.DosageUnit, logEntry.Route,
			logEntry.AdministeredAt, logEntry.NextScheduled, logEntry.CycleNumber, logEntry.BatchNumber, logEntry.Site,
			logEntry.AdministeredByID, logEntry.Status, logEntry.AdverseReactions, logEntry.Notes,
			time.Now().UTC(), time.Now().UTC(),
		)

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("build sql: %w", err)
	}
	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("insert therapy log: %w", err)
	}
	return nil
}

func (r *therapyLogRepository) ListByPatient(ctx context.Context, patientID uuid.UUID, limit int) ([]*entity.TherapyLog, error) {
	if limit <= 0 {
		limit = 50
	}

	q := r.sb.Select(
		"tl.id", "tl.patient_id", "tl.drug_id", "d.name AS drug_name",
		"tl.dosage", "tl.dosage_unit", "tl.route",
		"tl.administered_at", "tl.next_scheduled", "tl.cycle_number", "tl.batch_number", "tl.site",
		"tl.administered_by", "tl.status", "tl.adverse_reactions", "tl.notes", "tl.created_at", "tl.updated_at",
	).From("therapy_logs tl").
		LeftJoin("drugs d ON d.id = tl.drug_id").
		Where(squirrel.Eq{"tl.patient_id": patientID}).
		OrderBy("ABS(EXTRACT(EPOCH FROM (COALESCE(tl.next_scheduled, tl.administered_at, tl.created_at) - NOW()))) ASC").
		OrderBy("COALESCE(tl.next_scheduled, tl.administered_at, tl.created_at) DESC").
		Limit(uint64(limit))

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query therapy logs: %w", err)
	}
	defer rows.Close()

	var out []*entity.TherapyLog
	for rows.Next() {
		var t entity.TherapyLog
		if err := rows.Scan(
			&t.ID, &t.PatientID, &t.DrugID, &t.DrugName, &t.Dosage, &t.DosageUnit, &t.Route,
			&t.AdministeredAt, &t.NextScheduled, &t.CycleNumber, &t.BatchNumber, &t.Site,
			&t.AdministeredByID, &t.Status, &t.AdverseReactions, &t.Notes, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan therapy log: %w", err)
		}
		out = append(out, &t)
	}
	return out, nil
}

func (r *therapyLogRepository) DeleteByID(ctx context.Context, patientID uuid.UUID, logID uuid.UUID) (bool, error) {
	q := r.sb.Delete("therapy_logs").
		Where(squirrel.Eq{"id": logID, "patient_id": patientID})

	sql, args, err := q.ToSql()
	if err != nil {
		return false, fmt.Errorf("build sql: %w", err)
	}

	ct, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return false, fmt.Errorf("delete therapy log: %w", err)
	}

	return ct.RowsAffected() > 0, nil
}
