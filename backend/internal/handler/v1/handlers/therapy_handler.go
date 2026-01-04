package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/medical-app/backend/internal/entity"
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
	created, err := h.svc.CreateLog(c.Context(), req)
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
	items, err := h.svc.ListByPatient(c.Context(), id, 100)
	if err != nil {
		return err
	}
	return response.Success(c, items)
}
