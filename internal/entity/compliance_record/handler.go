package compliance_record

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/uswuth/vytora-backend/internal/entity/audit_trail"
	"github.com/uswuth/vytora-backend/internal/entity/vendor"
	"github.com/uswuth/vytora-backend/internal/middleware"
	"github.com/uswuth/vytora-backend/internal/services"
)

type Handler struct {
	repo        *Repository
	vendorRepo  *vendor.Repository
	auditLogger *audit_trail.Logger
	nextCode    vendor.NextCodeFunc
	validate    *validator.Validate
}

func NewHandler(repo *Repository, vendorRepo *vendor.Repository, auditLogger *audit_trail.Logger, nextCode vendor.NextCodeFunc) *Handler {
	return &Handler{
		repo:        repo,
		vendorRepo:  vendorRepo,
		auditLogger: auditLogger,
		nextCode:    nextCode,
		validate:    validator.New(),
	}
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateComplianceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "validation failed"})
	}

	v, err := h.vendorRepo.FindByCode(c.Context(), req.VendorCode)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "vendor not found"})
	}

	code, err := h.nextCode(c.Context(), "compliance_record")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate code"})
	}

	claims := c.Locals(middleware.UserContextKey).(*services.Claims)
	reviewedBy, _ := uuid.Parse(claims.UserID)

	cr := &ComplianceRecord{
		Code:              code,
		VendorID:          v.ID,
		VendorCode:        v.Code,
		CertificationType: req.CertificationType,
		Status:            "Pending",
		IssuedBy:          req.IssuedBy,
		EvidenceURL:       req.EvidenceURL,
		ReviewedBy:        &reviewedBy,
	}

	if err := h.repo.Create(c.Context(), cr); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create compliance record"})
	}

	if err := h.auditLogger.LogCreateSimple(c.Context(), "compliance_records", cr.ID, cr.Code, reviewedBy); err != nil {
		// Log but don't fail the request
	}

	return c.Status(fiber.StatusCreated).JSON(cr)
}

func (h *Handler) Get(c *fiber.Ctx) error {
	code := c.Params("code")
	cr, err := h.repo.FindByCode(c.Context(), code)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "compliance record not found"})
	}
	return c.JSON(cr)
}

func (h *Handler) List(c *fiber.Ctx) error {
	vendorCode := c.Query("vendor_code")
	v, err := h.vendorRepo.FindByCode(c.Context(), vendorCode)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "vendor not found"})
	}
	records, err := h.repo.ListByVendor(c.Context(), v.ID.String())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list compliance records"})
	}
	return c.JSON(records)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	code := c.Params("code")
	cr, err := h.repo.FindByCode(c.Context(), code)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "compliance record not found"})
	}

	var req UpdateComplianceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "validation failed"})
	}

	claims := c.Locals(middleware.UserContextKey).(*services.Claims)
	reviewedBy, _ := uuid.Parse(claims.UserID)

	cr.CertificationType = req.CertificationType
	cr.Status = req.Status
	cr.IssuedBy = req.IssuedBy
	cr.EvidenceURL = req.EvidenceURL
	cr.ReviewedBy = &reviewedBy

	if err := h.repo.Update(c.Context(), cr); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update compliance record"})
	}

	if err := h.auditLogger.LogCreateSimple(c.Context(), "compliance_records", cr.ID, cr.Code, reviewedBy); err != nil {
		// Log but don't fail the request
	}

	return c.JSON(cr)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	code := c.Params("code")

	cr, err := h.repo.FindByCode(c.Context(), code)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "compliance record not found"})
	}

	if err := h.repo.Delete(c.Context(), code); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "compliance record not found"})
	}

	claims := c.Locals(middleware.UserContextKey).(*services.Claims)
	changedBy, _ := uuid.Parse(claims.UserID)
	if err := h.auditLogger.LogDeleteSimple(c.Context(), "compliance_records", cr.ID, cr.Code, changedBy); err != nil {
		// Log but don't fail the request
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) Expiring(c *fiber.Ctx) error {
	daysStr := c.Query("days")
	if daysStr == "" {
		daysStr = "30"
	}
	days, _ := strconv.Atoi(daysStr)

	records, err := h.repo.Expiring(c.Context(), days)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get expiring compliance records"})
	}
	return c.JSON(records)
}