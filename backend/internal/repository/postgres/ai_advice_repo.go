package postgres

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/medical-app/backend/internal/entity"
)

type aiAdviceRepository struct {
	db *pgxpool.Pool
	sb squirrel.StatementBuilderType
}

func NewAIAdviceRepository(db *pgxpool.Pool) *aiAdviceRepository {
	return &aiAdviceRepository{db: db, sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}
}

func (r *aiAdviceRepository) Create(ctx context.Context, item *entity.AIAdvice) error {
	q := r.sb.Insert("ai_advice").
		Columns("id", "patient_id", "survey_code", "user_text", "score", "category", "details", "advice_text", "created_at").
		Values(item.ID, item.PatientID, item.SurveyCode, item.UserText, item.Score, item.Category, item.Details, item.AdviceText, item.CreatedAt)

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("build sql: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("insert ai_advice: %w", err)
	}
	return nil
}

func (r *aiAdviceRepository) ListByPatient(ctx context.Context, patientID uuid.UUID, limit int, offset int) ([]*entity.AIAdvice, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	q := r.sb.Select(
		"id", "patient_id", "survey_code", "COALESCE(user_text, '')", "score", "COALESCE(category, '')", "details", "advice_text", "created_at",
	).
		From("ai_advice").
		Where(squirrel.Eq{"patient_id": patientID}).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query ai_advice: %w", err)
	}
	defer rows.Close()

	out := make([]*entity.AIAdvice, 0)
	for rows.Next() {
		var item entity.AIAdvice
		if err := rows.Scan(
			&item.ID,
			&item.PatientID,
			&item.SurveyCode,
			&item.UserText,
			&item.Score,
			&item.Category,
			&item.Details,
			&item.AdviceText,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan ai_advice: %w", err)
		}
		out = append(out, &item)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("rows ai_advice: %w", rows.Err())
	}
	return out, nil
}
