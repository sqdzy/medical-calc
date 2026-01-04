package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/medical-app/backend/internal/service"
	"github.com/medical-app/backend/pkg/response"
)

type DrugHandler struct {
	svc *service.DrugService
}

func NewDrugHandler(svc *service.DrugService, _ any) *DrugHandler {
	return &DrugHandler{svc: svc}
}

func (h *DrugHandler) List(c *fiber.Ctx) error {
	search := c.Query("search")
	items, err := h.svc.List(c.Context(), search, 50)
	if err != nil {
		return err
	}
	return response.Success(c, items)
}

func (h *DrugHandler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid id")
	}
	item, err := h.svc.Get(c.Context(), id)
	if err != nil {
		return err
	}
	if item == nil {
		return response.NotFound(c, "Not found")
	}
	return response.Success(c, item)
}

// SearchPubChem searches for drugs in NCBI PubChem.
func (h *DrugHandler) SearchPubChem(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return response.BadRequest(c, "Query parameter 'q' is required")
	}
	cid, err := h.svc.SearchPubChem(c.Context(), query)
	if err != nil {
		return err
	}
	if cid == "" {
		return response.NotFound(c, "Drug not found in PubChem")
	}
	return response.Success(c, map[string]string{"cid": cid})
}

// VerifyPubChem verifies a drug exists in PubChem and returns compound info.
func (h *DrugHandler) VerifyPubChem(c *fiber.Ctx) error {
	name := c.Query("name")
	if name == "" {
		return response.BadRequest(c, "Query parameter 'name' is required")
	}
	info, err := h.svc.VerifyDrug(c.Context(), name)
	if err != nil {
		return err
	}
	if info == nil {
		return response.NotFound(c, "Drug not found in PubChem")
	}
	return response.Success(c, info)
}

// SearchPubMed searches for articles related to a drug in PubMed.
func (h *DrugHandler) SearchPubMed(c *fiber.Ctx) error {
	drugName := c.Query("drug")
	if drugName == "" {
		return response.BadRequest(c, "Query parameter 'drug' is required")
	}
	articles, err := h.svc.SearchPubMed(c.Context(), drugName)
	if err != nil {
		return err
	}
	return response.Success(c, articles)
}
