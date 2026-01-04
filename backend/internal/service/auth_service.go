package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/medical-app/backend/internal/entity"
	"github.com/medical-app/backend/internal/repository"
	"github.com/medical-app/backend/pkg/crypto"
	"github.com/medical-app/backend/pkg/validator"
)

type AuthService struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	roleRepo         repository.RoleRepository

	jwtSecret  string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

type AuthDeps struct {
	UserRepo         repository.UserRepository
	RefreshTokenRepo repository.RefreshTokenRepository
	RoleRepo         repository.RoleRepository
	JWTSecret        string
	AccessTTL        time.Duration
	RefreshTTL       time.Duration
}

func NewAuthService(d AuthDeps) *AuthService {
	return &AuthService{
		userRepo:         d.UserRepo,
		refreshTokenRepo: d.RefreshTokenRepo,
		roleRepo:         d.RoleRepo,
		jwtSecret:        d.JWTSecret,
		accessTTL:        d.AccessTTL,
		refreshTTL:       d.RefreshTTL,
	}
}

type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Role      string `json:"role,omitempty"` // optional; if empty -> patient
}

type AuthTokens struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserDisabled        = errors.New("user disabled")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
)

func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*entity.User, *AuthTokens, error) {
	v := validator.New()
	v.Required("email", req.Email, "email is required")
	v.Email("email", req.Email, "invalid email")
	v.Required("password", req.Password, "password is required")
	v.Password("password", req.Password, "password must be at least 8 chars and include upper/lower/digit")
	if v.HasErrors() {
		return nil, nil, v.Errors()
	}

	existing, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, nil, err
	}
	if existing != nil {
		return nil, nil, errors.New("email already in use")
	}

	roleName := strings.TrimSpace(req.Role)
	if roleName == "" {
		roleName = entity.RolePatient
	}
	role, err := s.roleRepo.GetByName(ctx, roleName)
	if err != nil {
		return nil, nil, err
	}
	if role == nil {
		return nil, nil, errors.New("invalid role")
	}

	hash, err := crypto.HashPassword(req.Password)
	if err != nil {
		return nil, nil, err
	}

	now := time.Now().UTC()
	user := &entity.User{
		ID:           uuid.New(),
		Email:        strings.ToLower(req.Email),
		PasswordHash: hash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
		RoleID:       role.ID,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
		Role:         role,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, nil, err
	}

	tokens, err := s.issueTokens(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	user.Sanitize()
	return user, tokens, nil
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*entity.User, *AuthTokens, error) {
	v := validator.New()
	v.Required("email", req.Email, "email is required")
	v.Email("email", req.Email, "invalid email")
	v.Required("password", req.Password, "password is required")
	if v.HasErrors() {
		return nil, nil, v.Errors()
	}

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, nil, ErrUserDisabled
	}
	if !crypto.CheckPassword(req.Password, user.PasswordHash) {
		return nil, nil, ErrInvalidCredentials
	}

	_ = s.userRepo.UpdateLastLogin(ctx, user.ID, time.Now().UTC())

	tokens, err := s.issueTokens(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	perms, _ := s.userRepo.ListPermissions(ctx, user.ID)
	user.Permissions = perms
	user.Sanitize()

	return user, tokens, nil
}

func (s *AuthService) issueTokens(ctx context.Context, user *entity.User) (*AuthTokens, error) {
	now := time.Now().UTC()
	exp := now.Add(s.accessTTL)

	claims := jwt.MapClaims{
		"sub":  user.ID.String(),
		"role": user.RoleID.String(),
		"exp":  exp.Unix(),
		"iat":  now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	access, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}

	refresh, err := crypto.GenerateSecureToken(32)
	if err != nil {
		return nil, err
	}

	// Hash refresh token before storing
	refreshHash := crypto.SHA256Hash(refresh)

	refreshEntity := &entity.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: refreshHash,
		ExpiresAt: now.Add(s.refreshTTL),
		CreatedAt: now,
	}

	if err := s.refreshTokenRepo.Create(ctx, refreshEntity); err != nil {
		return nil, err
	}

	return &AuthTokens{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    exp,
	}, nil
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Refresh validates the refresh token, revokes it, and issues new tokens.
func (s *AuthService) Refresh(ctx context.Context, req RefreshRequest) (*entity.User, *AuthTokens, error) {
	if req.RefreshToken == "" {
		return nil, nil, ErrInvalidRefreshToken
	}

	tokenHash := crypto.SHA256Hash(req.RefreshToken)

	stored, err := s.refreshTokenRepo.GetActiveByHash(ctx, tokenHash)
	if err != nil {
		return nil, nil, err
	}
	if stored == nil {
		return nil, nil, ErrInvalidRefreshToken
	}

	// Revoke the used token (one-time use)
	_ = s.refreshTokenRepo.Revoke(ctx, stored.ID, time.Now().UTC())

	user, err := s.userRepo.GetByID(ctx, stored.UserID)
	if err != nil {
		return nil, nil, err
	}
	if user == nil || !user.IsActive {
		return nil, nil, ErrUserDisabled
	}

	tokens, err := s.issueTokens(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	perms, _ := s.userRepo.ListPermissions(ctx, user.ID)
	user.Permissions = perms
	user.Sanitize()

	return user, tokens, nil
}

// Logout revokes all refresh tokens for the user.
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.refreshTokenRepo.RevokeAllForUser(ctx, userID, time.Now().UTC())
}
