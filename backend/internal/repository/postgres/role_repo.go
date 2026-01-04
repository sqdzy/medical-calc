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

type roleRepository struct {
	db *pgxpool.Pool
	sb squirrel.StatementBuilderType
}

func NewRoleRepository(db *pgxpool.Pool) *roleRepository {
	return &roleRepository{db: db, sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}
}

func (r *roleRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Role, error) {
	q := r.sb.Select("id", "name", "description", "created_at", "updated_at").
		From("roles").
		Where(squirrel.Eq{"id": id})

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	var role entity.Role
	if err := row.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("select role: %w", err)
	}
	return &role, nil
}

func (r *roleRepository) GetByName(ctx context.Context, name string) (*entity.Role, error) {
	q := r.sb.Select("id", "name", "description", "created_at", "updated_at").
		From("roles").
		Where(squirrel.Eq{"name": name})

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	var role entity.Role
	if err := row.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("select role: %w", err)
	}
	return &role, nil
}
