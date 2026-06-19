package vendor

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/uswuth/vytora-backend/internal/entity/audit_trail"
	"github.com/uswuth/vytora-backend/internal/middleware"
	"github.com/uswuth/vytora-backend/internal/services"
)

type WorkflowHandler struct {
	vendorRepo *Repository
	auditRepo  *audit_trail.Repository
	nextCode   NextCodeFunc
}

func NewWorkflowHandler(vendorRepo *Repository, auditRepo *audit_trail.Repository, nextCode NextCodeFunc) *WorkflowHandler {
	return &WorkflowHandler{
		vendorRepo: vendorRepo,
		auditRepo:  auditRepo,
		nextCode:   nextCode,
	}
}

func (h *WorkflowHandler) logTransition(c *fiber.Ctx, vendorID uuid.UUID, oldStatus, newStatus string) error {
	claims := c.Locals(middleware.UserContextKey).(*services.Claims)
	changedBy, _ := uuid.Parse(claims.UserID)

	code, err := h.nextCode(c.Context(), "audit_trail")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate audit code"})
	}

	audit := &audit_trail.AuditTrail{
		Code:      code,
		TableName: "vendors",
		RecordID:  vendorID,
		Action:    "UPDATE",
		FieldName: "status",
		OldValue:  oldStatus,
		NewValue:  newStatus,
		ChangedBy: changedBy,
	}
	return h.auditRepo.Create(c.Context(), audit)
}

func (h *WorkflowHandler) Submit(c *fiber.Ctx) error {
	code := c.Params("code")
	v, err := h.vendorRepo.FindByCode(c.Context(), code)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "vendor not found"})
	}
	if v.Status != "Draft" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "vendor must be in Draft status"})
	}
	if err := h.vendorRepo.UpdateStatus(c.Context(), code, "Submitted"); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update status"})
	}
	if err := h.logTransition(c, v.ID, "Draft", "Submitted"); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *WorkflowHandler) ReviewRisk(c *fiber.Ctx) error {
	code := c.Params("code")
	v, err := h.vendorRepo.FindByCode(c.Context(), code)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "vendor not found"})
	}
	if v.Status != "Submitted" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "vendor must be in Submitted status"})
	}
	if err := h.vendorRepo.UpdateStatus(c.Context(), code, "RiskReview"); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update status"})
	}
	if err := h.logTransition(c, v.ID, "Submitted", "RiskReview"); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *WorkflowHandler) ReviewCompliance(c *fiber.Ctx) error {
	code := c.Params("code")
	v, err := h.vendorRepo.FindByCode(c.Context(), code)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "vendor not found"})
	}
	if v.Status != "RiskReview" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "vendor must be in RiskReview status"})
	}
	if err := h.vendorRepo.UpdateStatus(c.Context(), code, "ComplianceReview"); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update status"})
	}
	if err := h.logTransition(c, v.ID, "RiskReview", "ComplianceReview"); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *WorkflowHandler) Approve(c *fiber.Ctx) error {
	code := c.Params("code")
	v, err := h.vendorRepo.FindByCode(c.Context(), code)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "vendor not found"})
	}
	if v.Status != "ComplianceReview" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "vendor must be in ComplianceReview status"})
	}
	if err := h.vendorRepo.UpdateStatus(c.Context(), code, "Approved"); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update status"})
	}
	if err := h.logTransition(c, v.ID, "ComplianceReview", "Approved"); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *WorkflowHandler) Reject(c *fiber.Ctx) error {
	code := c.Params("code")
	v, err := h.vendorRepo.FindByCode(c.Context(), code)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "vendor not found"})
	}
	if v.Status == "Rejected" || v.Status == "Approved" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "vendor cannot be rejected from current status"})
	}
	oldStatus := v.Status
	if err := h.vendorRepo.UpdateStatus(c.Context(), code, "Rejected"); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update status"})
	}
	if err := h.logTransition(c, v.ID, oldStatus, "Rejected"); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}