package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthMiddleware struct {
	jwtSecret string
}

func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{jwtSecret: jwtSecret}
}

// RequireAuth enforces JWT auth.
func (m *AuthMiddleware) RequireAuth() fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey:  []byte(m.jwtSecret),
		TokenLookup: "header:Authorization",
		AuthScheme:  "Bearer",
		ContextKey:  "jwt",
	})
}

// OptionalAuth parses token if present but does not fail without it.
func (m *AuthMiddleware) OptionalAuth() fiber.Handler {
	mw := m.RequireAuth()
	return func(c *fiber.Ctx) error {
		h := c.Get("Authorization")
		if strings.TrimSpace(h) == "" {
			return c.Next()
		}
		if err := mw(c); err != nil {
			return c.Next()
		}
		return c.Next()
	}
}

// GetUserID extracts user ID from JWT claims.
func GetUserID(c *fiber.Ctx) (uuid.UUID, bool) {
	val := c.Locals("jwt")
	if val == nil {
		return uuid.Nil, false
	}
	t, ok := val.(*jwt.Token)
	if !ok {
		return uuid.Nil, false
	}
	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, false
	}
	sub, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(sub)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}
