package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/medical-app/backend/internal/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID, at time.Time) error
	ListPermissions(ctx context.Context, userID uuid.UUID) ([]string, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *entity.RefreshToken) error
	GetActiveByHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error)
	Revoke(ctx context.Context, id uuid.UUID, revokedAt time.Time) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID, revokedAt time.Time) error
}

type AuditLogRepository interface {
	Create(ctx context.Context, entry *entity.AuditLog) error
}

type RoleRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Role, error)
	GetByName(ctx context.Context, name string) (*entity.Role, error)
}

type PermissionRepository interface {
	ListByRoleID(ctx context.Context, roleID uuid.UUID) ([]*entity.Permission, error)
}

type SurveyTemplateRepository interface {
	ListActive(ctx context.Context) ([]*entity.SurveyTemplate, error)
	GetByCode(ctx context.Context, code string) (*entity.SurveyTemplate, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.SurveyTemplate, error)
}

type SurveyResponseRepository interface {
	Create(ctx context.Context, resp *entity.SurveyResponse) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.SurveyResponse, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit int) ([]*entity.SurveyResponse, error)
	UpdateCalculated(ctx context.Context, id uuid.UUID, score float64, category string, interpretation string, aiSummary string, breakdown any) error
}

type DrugRepository interface {
	List(ctx context.Context, search string, limit int) ([]*entity.Drug, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Drug, error)
	Create(ctx context.Context, drug *entity.Drug) error
	Update(ctx context.Context, drug *entity.Drug) error
}

type TherapyLogRepository interface {
	Create(ctx context.Context, log *entity.TherapyLog) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit int) ([]*entity.TherapyLog, error)
}
