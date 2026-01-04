package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/medical-app/backend/internal/entity"
)

type surveyResponseRepository struct {
	db *pgxpool.Pool
	sb squirrel.StatementBuilderType
}

func NewSurveyResponseRepository(db *pgxpool.Pool) *surveyResponseRepository {
	return &surveyResponseRepository{db: db, sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}
}

func (r *surveyResponseRepository) Create(ctx context.Context, resp *entity.SurveyResponse) error {
	q := r.sb.Insert("survey_responses").
		Columns(
			"id", "template_id", "patient_id", "responses", "calculated_score",
			"score_breakdown", "interpretation", "ai_summary", "status",
			"submitted_at", "reviewed_by", "reviewed_at", "notes", "created_at",
		).
		Values(
			resp.ID, resp.TemplateID, resp.PatientID, resp.Responses, resp.CalculatedScore,
			resp.ScoreBreakdown, resp.Interpretation, resp.AISummary, resp.Status,
			resp.SubmittedAt, resp.ReviewedBy, resp.ReviewedAt, resp.Notes, resp.CreatedAt,
		)

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("build sql: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("insert survey response: %w", err)
	}
	return nil
}

func (r *surveyResponseRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.SurveyResponse, error) {
	q := r.sb.Select(
		"id", "template_id", "patient_id", "responses", "calculated_score",
		"score_breakdown", "interpretation", "ai_summary", "status",
		"submitted_at", "reviewed_by", "reviewed_at", "notes", "created_at",
	).From("survey_responses").Where(squirrel.Eq{"id": id})

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	var sr entity.SurveyResponse
	if err := row.Scan(
		&sr.ID, &sr.TemplateID, &sr.PatientID, &sr.Responses, &sr.CalculatedScore,
		&sr.ScoreBreakdown, &sr.Interpretation, &sr.AISummary, &sr.Status,
		&sr.SubmittedAt, &sr.ReviewedBy, &sr.ReviewedAt, &sr.Notes, &sr.CreatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("select response: %w", err)
	}
	return &sr, nil
}

func (r *surveyResponseRepository) ListByPatient(ctx context.Context, patientID uuid.UUID, limit int) ([]*entity.SurveyResponse, error) {
	if limit <= 0 {
		limit = 50
	}

	q := r.sb.Select(
		"id", "template_id", "patient_id", "responses", "calculated_score",
		"score_breakdown", "interpretation", "ai_summary", "status",
		"submitted_at", "reviewed_by", "reviewed_at", "notes", "created_at",
	).From("survey_responses").
		Where(squirrel.Eq{"patient_id": patientID}).
		OrderBy("submitted_at DESC").
		Limit(uint64(limit))

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query responses: %w", err)
	}
	defer rows.Close()

	var out []*entity.SurveyResponse
	for rows.Next() {
		var sr entity.SurveyResponse
		if err := rows.Scan(
			&sr.ID, &sr.TemplateID, &sr.PatientID, &sr.Responses, &sr.CalculatedScore,
			&sr.ScoreBreakdown, &sr.Interpretation, &sr.AISummary, &sr.Status,
			&sr.SubmittedAt, &sr.ReviewedBy, &sr.ReviewedAt, &sr.Notes, &sr.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan response: %w", err)
		}
		out = append(out, &sr)
	}
	return out, nil
}

func (r *surveyResponseRepository) UpdateCalculated(ctx context.Context, id uuid.UUID, score float64, category string, interpretation string, aiSummary string, breakdown any) error {
	breakdownJSON, _ := json.Marshal(breakdown)

	q := r.sb.Update("survey_responses").
		Set("calculated_score", score).
		Set("score_breakdown", breakdownJSON).
		Set("interpretation", interpretation).
		Set("ai_summary", aiSummary).
		Set("status", entity.SurveyStatusSubmitted).
		Where(squirrel.Eq{"id": id})

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("build sql: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("update calculated: %w", err)
	}
	_ = category
	_ = time.Now()
	return nil
}
