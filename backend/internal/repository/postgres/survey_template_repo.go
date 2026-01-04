package postgres

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/medical-app/backend/internal/entity"
)

type surveyTemplateRepository struct {
	db *pgxpool.Pool
	sb squirrel.StatementBuilderType
}

func NewSurveyTemplateRepository(db *pgxpool.Pool) *surveyTemplateRepository {
	return &surveyTemplateRepository{db: db, sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}
}

func (r *surveyTemplateRepository) ListActive(ctx context.Context) ([]*entity.SurveyTemplate, error) {
	q := r.sb.Select(
		"id", "code", "name", "description", "category", "questions", "scoring_logic", "interpretation_rules",
		"version", "is_active", "created_by", "created_at", "updated_at",
	).From("survey_templates").Where(squirrel.Eq{"is_active": true}).OrderBy("name ASC")

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query templates: %w", err)
	}
	defer rows.Close()

	var out []*entity.SurveyTemplate
	for rows.Next() {
		var t entity.SurveyTemplate
		if err := rows.Scan(
			&t.ID, &t.Code, &t.Name, &t.Description, &t.Category,
			&t.Questions, &t.ScoringLogic, &t.InterpretationRules,
			&t.Version, &t.IsActive, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan template: %w", err)
		}
		out = append(out, &t)
	}

	return out, nil
}

func (r *surveyTemplateRepository) GetByCode(ctx context.Context, code string) (*entity.SurveyTemplate, error) {
	q := r.sb.Select(
		"id", "code", "name", "description", "category", "questions", "scoring_logic", "interpretation_rules",
		"version", "is_active", "created_by", "created_at", "updated_at",
	).From("survey_templates").Where(squirrel.Eq{"code": code})

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	var t entity.SurveyTemplate
	if err := row.Scan(
		&t.ID, &t.Code, &t.Name, &t.Description, &t.Category,
		&t.Questions, &t.ScoringLogic, &t.InterpretationRules,
		&t.Version, &t.IsActive, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("select template: %w", err)
	}
	return &t, nil
}

func (r *surveyTemplateRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.SurveyTemplate, error) {
	q := r.sb.Select(
		"id", "code", "name", "description", "category", "questions", "scoring_logic", "interpretation_rules",
		"version", "is_active", "created_by", "created_at", "updated_at",
	).From("survey_templates").Where(squirrel.Eq{"id": id})

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	var t entity.SurveyTemplate
	if err := row.Scan(
		&t.ID, &t.Code, &t.Name, &t.Description, &t.Category,
		&t.Questions, &t.ScoringLogic, &t.InterpretationRules,
		&t.Version, &t.IsActive, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("select template: %w", err)
	}
	return &t, nil
}
