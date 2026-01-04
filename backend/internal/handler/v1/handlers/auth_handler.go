package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/medical-app/backend/internal/entity"
	"github.com/medical-app/backend/internal/handler/middleware"
	"github.com/medical-app/backend/internal/service"
	"github.com/medical-app/backend/pkg/response"
)

type AuthHandler struct {
	auth  *service.AuthService
	audit *middleware.AuditMiddleware
}

func NewAuthHandler(auth *service.AuthService, audit *middleware.AuditMiddleware) *AuthHandler {
	return &AuthHandler{auth: auth, audit: audit}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req service.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	user, tokens, err := h.auth.Register(c.Context(), req)
	if err != nil {
		return err
	}

	h.audit.Log(c, entity.AuditActionCreate, entity.ResourceUser, &user.ID, nil, map[string]any{"email": user.Email, "role_id": user.RoleID})

	return response.Created(c, fiber.Map{
		"user":   user,
		"tokens": tokens,
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req service.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	user, tokens, err := h.auth.Login(c.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return response.Unauthorized(c, "Invalid credentials")
		}
		if errors.Is(err, service.ErrUserDisabled) {
			return response.Forbidden(c, "User disabled")
		}
		return err
	}

	h.audit.Log(c, entity.AuditActionLogin, entity.ResourceUser, &user.ID, nil, map[string]any{"email": user.Email})

	return response.Success(c, fiber.Map{
		"user":   user,
		"tokens": tokens,
	})
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	uid, ok := middleware.GetUserID(c)
	if !ok {
		return response.Unauthorized(c, "Unauthorized")
	}

	// For MVP, return only JWT subject.
	// Full profile will be fetched via UserService later.
	return response.Success(c, fiber.Map{
		"user_id": uid,
	})
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var req service.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	user, tokens, err := h.auth.Refresh(c.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidRefreshToken) {
			return response.Unauthorized(c, "Invalid or expired refresh token")
		}
		if errors.Is(err, service.ErrUserDisabled) {
			return response.Forbidden(c, "User disabled")
		}
		return err
	}

	return response.Success(c, fiber.Map{
		"user":   user,
		"tokens": tokens,
	})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	uid, ok := middleware.GetUserID(c)
	if !ok {
		return response.Unauthorized(c, "Unauthorized")
	}

	if err := h.auth.Logout(c.Context(), uid); err != nil {
		return err
	}

	return response.Success(c, fiber.Map{"message": "logged out"})
}

func parseUUIDParam(c *fiber.Ctx, key string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Params(key))
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}
