package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/medical-app/backend/internal/entity"
)

type auditLogRepository struct {
	db *pgxpool.Pool
	sb squirrel.StatementBuilderType
}

func NewAuditLogRepository(db *pgxpool.Pool) *auditLogRepository {
	return &auditLogRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *auditLogRepository) Create(ctx context.Context, entry *entity.AuditLog) error {
	q := r.sb.Insert("audit_logs").
		Columns(
			"id", "user_id", "action", "resource_type", "resource_id",
			"old_value", "new_value", "ip_address", "user_agent", "request_id",
			"created_at",
		)

	var oldValue any
	var newValue any
	if len(entry.OldValue) > 0 {
		var tmp any
		if err := json.Unmarshal(entry.OldValue, &tmp); err == nil {
			oldValue = tmp
		}
	}
	if len(entry.NewValue) > 0 {
		var tmp any
		if err := json.Unmarshal(entry.NewValue, &tmp); err == nil {
			newValue = tmp
		}
	}

	q = q.Values(
		entry.ID,
		entry.UserID,
		entry.Action,
		entry.ResourceType,
		entry.ResourceID,
		oldValue,
		newValue,
		entry.IPAddress,
		entry.UserAgent,
		entry.RequestID,
		entry.CreatedAt,
	)

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("build sql: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}
