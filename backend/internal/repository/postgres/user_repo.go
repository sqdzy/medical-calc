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

type userRepository struct {
	db *pgxpool.Pool
	sb squirrel.StatementBuilderType
}

func NewUserRepository(db *pgxpool.Pool) *userRepository {
	return &userRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	q := r.sb.Insert("users").
		Columns("id", "email", "password_hash", "first_name", "last_name", "phone", "role_id", "is_active", "email_verified").
		Values(user.ID, strings.ToLower(user.Email), user.PasswordHash, user.FirstName, user.LastName, user.Phone, user.RoleID, user.IsActive, user.EmailVerified)

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("build sql: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	q := r.sb.Select(
		"u.id", "u.email", "u.password_hash", "u.first_name", "u.last_name", "u.phone",
		"u.role_id", "u.is_active", "u.email_verified", "u.last_login_at", "u.created_at", "u.updated_at",
		"r.id", "r.name", "r.description", "r.created_at", "r.updated_at",
	).From("users u").
		LeftJoin("roles r ON r.id = u.role_id").
		Where(squirrel.Eq{"u.id": id})

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)

	var u entity.User
	var role entity.Role
	var roleID *uuid.UUID
	if err := row.Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Phone,
		&u.RoleID, &u.IsActive, &u.EmailVerified, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
		&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("select user: %w", err)
	}

	roleID = &role.ID
	if roleID != nil {
		u.Role = &role
	}

	return &u, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	q := r.sb.Select(
		"u.id", "u.email", "u.password_hash", "u.first_name", "u.last_name", "u.phone",
		"u.role_id", "u.is_active", "u.email_verified", "u.last_login_at", "u.created_at", "u.updated_at",
		"r.id", "r.name", "r.description", "r.created_at", "r.updated_at",
	).From("users u").
		LeftJoin("roles r ON r.id = u.role_id").
		Where(squirrel.Eq{"u.email": strings.ToLower(email)})

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)

	var u entity.User
	var role entity.Role
	if err := row.Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Phone,
		&u.RoleID, &u.IsActive, &u.EmailVerified, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
		&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("select user: %w", err)
	}

	u.Role = &role
	return &u, nil
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID, at time.Time) error {
	q := r.sb.Update("users").
		Set("last_login_at", at).
		Set("updated_at", time.Now().UTC()).
		Where(squirrel.Eq{"id": id})

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("build sql: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("update last_login_at: %w", err)
	}
	return nil
}

func (r *userRepository) ListPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	q := r.sb.Select("p.name").
		From("users u").
		Join("roles r ON r.id = u.role_id").
		Join("role_permissions rp ON rp.role_id = r.id").
		Join("permissions p ON p.id = rp.permission_id").
		Where(squirrel.Eq{"u.id": userID})

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query permissions: %w", err)
	}
	defer rows.Close()

	var perms []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan permission: %w", err)
		}
		perms = append(perms, name)
	}
	return perms, nil
}
