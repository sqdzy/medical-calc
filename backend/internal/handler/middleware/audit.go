package middleware

import (
	"encoding/json"
	"net"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/medical-app/backend/internal/entity"
	"github.com/medical-app/backend/internal/repository"
)

type AuditMiddleware struct {
	repo repository.AuditLogRepository
}

func NewAuditMiddleware(repo repository.AuditLogRepository) *AuditMiddleware {
	return &AuditMiddleware{repo: repo}
}

func (m *AuditMiddleware) Log(c *fiber.Ctx, action, resourceType string, resourceID *uuid.UUID, oldValue any, newValue any) {
	if m == nil || m.repo == nil {
		return
	}

	var oldJSON []byte
	var newJSON []byte
	if oldValue != nil {
		oldJSON, _ = json.Marshal(oldValue)
	}
	if newValue != nil {
		newJSON, _ = json.Marshal(newValue)
	}

	var userID *uuid.UUID
	if uid, ok := GetUserID(c); ok {
		userID = &uid
	}

	entry := &entity.AuditLog{
		ID:           uuid.New(),
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		OldValue:     oldJSON,
		NewValue:     newJSON,
		IPAddress:    net.ParseIP(c.IP()),
		UserAgent:    c.Get("User-Agent"),
		RequestID:    c.Get("X-Request-Id"),
		CreatedAt:    time.Now().UTC(),
	}

	_ = m.repo.Create(c.Context(), entry)
}
