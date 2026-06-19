package audit_trail

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	auditRepo *Repository
}

func NewHandler(auditRepo *Repository) *Handler {
	return &Handler{
		auditRepo: auditRepo,
	}
}

func (h *Handler) List(c *fiber.Ctx) error {
	params := ListParams{
		TableName:  c.Query("table"),
		RecordCode: c.Query("record_code"),
		Action:     c.Query("action"),
		ChangedBy:  c.Query("changed_by"),
		DateFrom:   c.Query("date_from"),
		DateTo:     c.Query("date_to"),
	}
	if v := c.Query("limit"); v != "" {
		params.Limit, _ = strconv.Atoi(v)
	}
	if v := c.Query("offset"); v != "" {
		params.Offset, _ = strconv.Atoi(v)
	}

	audits, total, err := h.auditRepo.List(c.Context(), params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list audit trail"})
	}

	return c.JSON(ListResponse{
		Data:  audits,
		Total: total,
	})
}