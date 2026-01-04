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

type refreshTokenRepository struct {
	db *pgxpool.Pool
	sb squirrel.StatementBuilderType
}

func NewRefreshTokenRepository(db *pgxpool.Pool) *refreshTokenRepository {
	return &refreshTokenRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *entity.RefreshToken) error {
	q := r.sb.Insert("refresh_tokens").
		Columns("id", "user_id", "token_hash", "expires_at", "created_at").
		Values(token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.CreatedAt)

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("build sql: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("insert refresh_token: %w", err)
	}
	return nil
}

func (r *refreshTokenRepository) GetActiveByHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error) {
	q := r.sb.Select(
		"id", "user_id", "token_hash", "expires_at", "revoked_at", "created_at",
	).From("refresh_tokens").
		Where(squirrel.Eq{"token_hash": tokenHash}).
		Where(squirrel.Expr("revoked_at IS NULL")).
		Where(squirrel.Expr("expires_at > NOW()"))

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	row := r.db.QueryRow(ctx, sql, args...)
	var t entity.RefreshToken
	if err := row.Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.RevokedAt, &t.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("select refresh_token: %w", err)
	}
	return &t, nil
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID, revokedAt time.Time) error {
	q := r.sb.Update("refresh_tokens").
		Set("revoked_at", revokedAt).
		Where(squirrel.Eq{"id": id})

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("build sql: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("revoke refresh_token: %w", err)
	}
	return nil
}

func (r *refreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID, revokedAt time.Time) error {
	q := r.sb.Update("refresh_tokens").
		Set("revoked_at", revokedAt).
		Where(squirrel.Eq{"user_id": userID}).
		Where(squirrel.Expr("revoked_at IS NULL"))

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("build sql: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("revoke all refresh_tokens: %w", err)
	}
	return nil
}
