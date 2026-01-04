package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/medical-app/backend/internal/entity"
)

type drugRepository struct {
	db *pgxpool.Pool
	sb squirrel.StatementBuilderType
}

func NewDrugRepository(db *pgxpool.Pool) *drugRepository {
	return &drugRepository{db: db, sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}
}

func (r *drugRepository) List(ctx context.Context, search string, limit int) ([]*entity.Drug, error) {
	if limit <= 0 {
		limit = 50
	}

	q := r.sb.Select(
		"id", "name", "international_name", "trade_name", "ncbi_pubchem_id", "atc_code",
		"dosage_form", "manufacturer", "description", "contraindications", "is_active", "created_at", "updated_at",
	).From("drugs")

	if strings.TrimSpace(search) != "" {
		like := "%" + strings.ToLower(search) + "%"
		q = q.Where(squirrel.Or{
			squirrel.Like{"LOWER(name)": like},
			squirrel.Like{"LOWER(international_name)": like},
			squirrel.Like{"LOWER(trade_name)": like},
		})
	}

	q = q.OrderBy("name ASC").Limit(uint64(limit))

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query drugs: %w", err)
	}
	defer rows.Close()

	var out []*entity.Drug
	for rows.Next() {
		var d entity.Drug
		if err := rows.Scan(
			&d.ID, &d.Name, &d.InternationalName, &d.TradeName, &d.NCBIPubchemID, &d.ATCCode,
			&d.DosageForm, &d.Manufacturer, &d.Description, &d.Contraindications,
			&d.IsActive, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan drug: %w", err)
		}
		out = append(out, &d)
	}
	return out, nil
}

func (r *drugRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Drug, error) {
	q := r.sb.Select(
		"id", "name", "international_name", "trade_name", "ncbi_pubchem_id", "atc_code",
		"dosage_form", "manufacturer", "description", "contraindications", "is_active", "created_at", "updated_at",
	).From("drugs").Where(squirrel.Eq{"id": id})

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	var d entity.Drug
	if err := row.Scan(
		&d.ID, &d.Name, &d.InternationalName, &d.TradeName, &d.NCBIPubchemID, &d.ATCCode,
		&d.DosageForm, &d.Manufacturer, &d.Description, &d.Contraindications,
		&d.IsActive, &d.CreatedAt, &d.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("select drug: %w", err)
	}
	return &d, nil
}

func (r *drugRepository) Create(ctx context.Context, drug *entity.Drug) error {
	q := r.sb.Insert("drugs").
		Columns(
			"id", "name", "international_name", "trade_name", "ncbi_pubchem_id", "atc_code",
			"dosage_form", "manufacturer", "description", "contraindications", "is_active",
			"created_at", "updated_at",
		).
		Values(
			drug.ID, drug.Name, drug.InternationalName, drug.TradeName, drug.NCBIPubchemID, drug.ATCCode,
			drug.DosageForm, drug.Manufacturer, drug.Description, drug.Contraindications, drug.IsActive,
			time.Now().UTC(), time.Now().UTC(),
		)

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("build sql: %w", err)
	}
	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("insert drug: %w", err)
	}
	return nil
}

func (r *drugRepository) Update(ctx context.Context, drug *entity.Drug) error {
	q := r.sb.Update("drugs").
		Set("name", drug.Name).
		Set("international_name", drug.InternationalName).
		Set("trade_name", drug.TradeName).
		Set("ncbi_pubchem_id", drug.NCBIPubchemID).
		Set("atc_code", drug.ATCCode).
		Set("dosage_form", drug.DosageForm).
		Set("manufacturer", drug.Manufacturer).
		Set("description", drug.Description).
		Set("contraindications", drug.Contraindications).
		Set("is_active", drug.IsActive).
		Set("updated_at", time.Now().UTC()).
		Where(squirrel.Eq{"id": drug.ID})

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("build sql: %w", err)
	}
	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("update drug: %w", err)
	}
	return nil
}
