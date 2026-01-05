package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/medical-app/backend/internal/entity"
	"github.com/medical-app/backend/internal/handler/middleware"
	"github.com/medical-app/backend/internal/service"
	"github.com/medical-app/backend/pkg/response"
)

type TherapyHandler struct {
	svc *service.TherapyService
}

func NewTherapyHandler(svc *service.TherapyService, _ any) *TherapyHandler {
	return &TherapyHandler{svc: svc}
}

func (h *TherapyHandler) CreateLog(c *fiber.Ctx) error {
	var req entity.TherapyLogCreate
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	// Resolve patient from authenticated user
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return response.Unauthorized(c, "authentication required")
	}

	created, err := h.svc.CreateLogForUser(c.Context(), userID, req)
	if err != nil {
		return err
	}
	return response.Created(c, created)
}

func (h *TherapyHandler) ListByPatient(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("patientId"))
	if err != nil {
		return response.BadRequest(c, "Invalid patientId")
	}
	patientID, ok, err := h.svc.ResolvePatientID(c.Context(), id)
	if err != nil {
		return err
	}
	if !ok {
		return response.NotFound(c, "Patient not found")
	}
	items, err := h.svc.ListByPatient(c.Context(), patientID, 100)
	if err != nil {
		return err
	}
	return response.Success(c, items)
}

func (h *TherapyHandler) DeleteLog(c *fiber.Ctx) error {
	logID, err := uuid.Parse(c.Params("logId"))
	if err != nil {
		return response.BadRequest(c, "Invalid logId")
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		return response.Unauthorized(c, "authentication required")
	}

	deleted, err := h.svc.DeleteLogForUser(c.Context(), userID, logID)
	if err != nil {
		return err
	}
	if !deleted {
		return response.NotFound(c, "Therapy log not found")
	}
	return response.NoContent(c)
}
