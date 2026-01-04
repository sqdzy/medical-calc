package postgres

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/medical-app/backend/internal/entity"
)

type permissionRepository struct {
	db *pgxpool.Pool
	sb squirrel.StatementBuilderType
}

func NewPermissionRepository(db *pgxpool.Pool) *permissionRepository {
	return &permissionRepository{db: db, sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}
}

func (r *permissionRepository) ListByRoleID(ctx context.Context, roleID uuid.UUID) ([]*entity.Permission, error) {
	q := r.sb.Select("p.id", "p.name", "p.description", "p.created_at").
		From("permissions p").
		Join("role_permissions rp ON rp.permission_id = p.id").
		Where(squirrel.Eq{"rp.role_id": roleID})

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query permissions: %w", err)
	}
	defer rows.Close()

	var out []*entity.Permission
	for rows.Next() {
		var p entity.Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan permission: %w", err)
		}
		out = append(out, &p)
	}
	return out, nil
}
