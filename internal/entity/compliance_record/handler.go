package compliance_record

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/uswuth/vytora-backend/internal/entity/vendor"
	"github.com/uswuth/vytora-backend/internal/middleware"
	"github.com/uswuth/vytora-backend/internal/services"
)

type Handler struct {
	compRepo   *Repository
	vendorRepo *vendor.Repository
	nextCode   func(ctx context.Context, entity string) (string, error)
	validate   *validator.Validate
}

func NewHandler(compRepo *Repository, vendorRepo *vendor.Repository, nextCode func(ctx context.Context, entity string) (string, error)) *Handler {
	return &Handler{
		compRepo:   compRepo,
		vendorRepo: vendorRepo,
		nextCode:   nextCode,
		validate:   validator.New(),
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

	from, err := time.Parse("2006-01-02", req.ValidFrom)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid valid_from format"})
	}
	until, err := time.Parse("2006-01-02", req.ValidUntil)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid valid_until format"})
	}

	status := "Pending"
	if time.Now().After(until) {
		status = "Expired"
	} else if time.Now().After(from) || time.Now().Equal(from) {
		status = "Approved"
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
		CertificationType: req.CertificationType,
		Status:            status,
		ValidFrom:         &from,
		ValidUntil:        &until,
		IssuedBy:          req.IssuedBy,
		EvidenceURL:       req.EvidenceURL,
		ReviewedBy:        &reviewedBy,
	}

	if err := h.compRepo.Create(c.Context(), cr); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create compliance record"})
	}

	cr.VendorCode = v.Code

	return c.Status(fiber.StatusCreated).JSON(cr)
}

func (h *Handler) Get(c *fiber.Ctx) error {
	code := c.Params("code")
	cr, err := h.compRepo.FindByCode(c.Context(), code)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "compliance record not found"})
	}
	return c.JSON(cr)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	code := c.Params("code")
	existing, err := h.compRepo.FindByCode(c.Context(), code)
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

	if req.ValidFrom != "" {
		t, _ := time.Parse("2006-01-02", req.ValidFrom)
		existing.ValidFrom = &t
	}
	if req.ValidUntil != "" {
		t, _ := time.Parse("2006-01-02", req.ValidUntil)
		existing.ValidUntil = &t
	}
	existing.CertificationType = req.CertificationType
	existing.Status = req.Status
	existing.IssuedBy = req.IssuedBy
	existing.EvidenceURL = req.EvidenceURL

	if err := h.compRepo.Update(c.Context(), existing); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update compliance record"})
	}

	return c.JSON(existing)
}

func (h *Handler) List(c *fiber.Ctx) error {
	vendorCode := c.Query("vendor_code")
	v, err := h.vendorRepo.FindByCode(c.Context(), vendorCode)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "vendor not found"})
	}
	records, err := h.compRepo.ListByVendor(c.Context(), v.ID.String())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list compliance records"})
	}
	return c.JSON(records)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	code := c.Params("code")
	if err := h.compRepo.Delete(c.Context(), code); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "compliance record not found"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) Expiring(c *fiber.Ctx) error {
	daysStr := c.Query("days")
	if daysStr == "" {
		daysStr = "30"
	}
	days, _ := strconv.Atoi(daysStr)

	records, err := h.compRepo.Expiring(c.Context(), days)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get expiring certifications"})
	}
	return c.JSON(records)
}