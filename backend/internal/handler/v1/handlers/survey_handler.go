package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/medical-app/backend/internal/entity"
	"github.com/medical-app/backend/internal/handler/middleware"
	"github.com/medical-app/backend/internal/service"
	"github.com/medical-app/backend/pkg/response"
)

type SurveyHandler struct {
	svc   *service.SurveyService
	auth  *middleware.AuthMiddleware
	audit *middleware.AuditMiddleware
}

func NewSurveyHandler(svc *service.SurveyService, auth *middleware.AuthMiddleware, audit *middleware.AuditMiddleware) *SurveyHandler {
	return &SurveyHandler{svc: svc, auth: auth, audit: audit}
}

func (h *SurveyHandler) ListTemplates(c *fiber.Ctx) error {
	items, err := h.svc.ListTemplates(c.Context())
	if err != nil {
		return err
	}
	return response.Success(c, items)
}

func (h *SurveyHandler) GetTemplateByCode(c *fiber.Ctx) error {
	t, err := h.svc.GetTemplateByCode(c.Context(), c.Params("code"))
	if err != nil {
		return err
	}
	if t == nil {
		return response.NotFound(c, "Template not found")
	}
	return response.Success(c, t)
}

func (h *SurveyHandler) SubmitResponse(c *fiber.Ctx) error {
	var req entity.SurveyResponseCreate
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	created, err := h.svc.SubmitResponse(c.Context(), req)
	if err != nil {
		return err
	}

	h.audit.Log(c, entity.AuditActionCreate, entity.ResourceSurvey, &created.ID, nil, map[string]any{"template_id": created.TemplateID, "patient_id": created.PatientID})
	return response.Created(c, created)
}
