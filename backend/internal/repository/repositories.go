package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/medical-app/backend/internal/repository/postgres"
)

// Repositories groups all repositories.
type Repositories struct {
	User         UserRepository
	RefreshToken RefreshTokenRepository
	AuditLog     AuditLogRepository

	SurveyTemplate SurveyTemplateRepository
	SurveyResponse SurveyResponseRepository

	Drug       DrugRepository
	TherapyLog TherapyLogRepository

	Role       RoleRepository
	Permission PermissionRepository
}

// NewRepositories initializes all repository implementations.
func NewRepositories(db *pgxpool.Pool) *Repositories {
	return &Repositories{
		User:           postgres.NewUserRepository(db),
		RefreshToken:   postgres.NewRefreshTokenRepository(db),
		AuditLog:       postgres.NewAuditLogRepository(db),
		SurveyTemplate: postgres.NewSurveyTemplateRepository(db),
		SurveyResponse: postgres.NewSurveyResponseRepository(db),
		Drug:           postgres.NewDrugRepository(db),
		TherapyLog:     postgres.NewTherapyLogRepository(db),
		Role:           postgres.NewRoleRepository(db),
		Permission:     postgres.NewPermissionRepository(db),
	}
}
