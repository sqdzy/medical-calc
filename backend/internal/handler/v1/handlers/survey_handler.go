package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/medical-app/backend/internal/entity"
	"github.com/medical-app/backend/internal/handler/middleware"
	"github.com/medical-app/backend/internal/service"
	"github.com/medical-app/backend/pkg/response"
)

type SurveyHandler struct {
	svc   *service.SurveyService
	ai    *service.AIAdviceService
	auth  *middleware.AuthMiddleware
	audit *middleware.AuditMiddleware
}

func NewSurveyHandler(svc *service.SurveyService, ai *service.AIAdviceService, auth *middleware.AuthMiddleware, audit *middleware.AuditMiddleware) *SurveyHandler {
	return &SurveyHandler{svc: svc, ai: ai, auth: auth, audit: audit}
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

type surveyCalculateRequest struct {
	Answers []struct {
		QuestionID string      `json:"question_id"`
		Value      interface{} `json:"value"`
	} `json:"answers"`
}

func (h *SurveyHandler) Calculate(c *fiber.Ctx) error {
	code := c.Params("code")
	if strings.TrimSpace(code) == "" {
		return response.BadRequest(c, "Invalid code")
	}

	var req surveyCalculateRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	t, err := h.svc.GetTemplateByCode(c.Context(), code)
	if err != nil {
		return err
	}
	if t == nil {
		return response.NotFound(c, "Template not found")
	}

	respMap := make(map[string]interface{}, len(req.Answers))
	for _, a := range req.Answers {
		qid := strings.TrimSpace(a.QuestionID)
		if qid == "" {
			continue
		}
		respMap[qid] = a.Value
	}

	score, category, breakdown, err := service.CalculateScore(t, respMap)
	if err != nil {
		return err
	}

	interpretation := category
	if desc, ok := breakdown["category_description"].(string); ok && strings.TrimSpace(desc) != "" {
		interpretation = desc
	}

	return response.Success(c, map[string]any{
		"score":          score,
		"interpretation": interpretation,
		"category":       category,
		"breakdown":      breakdown,
	})
}

type surveyAdviceRequest struct {
	Answers []struct {
		QuestionID string      `json:"question_id"`
		Value      interface{} `json:"value"`
	} `json:"answers"`
	Text string `json:"text"`
}

func (h *SurveyHandler) CreateAdvice(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return response.Unauthorized(c, "Unauthorized")
	}
	code := strings.TrimSpace(c.Params("code"))
	if code == "" {
		return response.BadRequest(c, "Invalid code")
	}
	if h.ai == nil {
		return response.BadRequest(c, "AI service not configured")
	}

	var req surveyAdviceRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	respMap := make(map[string]any, len(req.Answers))
	for _, a := range req.Answers {
		qid := strings.TrimSpace(a.QuestionID)
		if qid == "" {
			continue
		}
		respMap[qid] = a.Value
	}

	created, err := h.ai.CreateForUser(c.Context(), userID, code, respMap, req.Text)
	if err != nil {
		return err
	}
	return response.Created(c, created)
}

func (h *SurveyHandler) ListAdvice(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return response.Unauthorized(c, "Unauthorized")
	}
	if h.ai == nil {
		return response.Success(c, []*service.AIAdviceResult{})
	}
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)

	items, err := h.ai.ListForUser(c.Context(), userID, limit, offset)
	if err != nil {
		return err
	}
	return response.Success(c, items)
}
