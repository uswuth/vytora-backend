package vendor

import (
	"context"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/uswuth/vytora-backend/internal/entity/audit_trail"
	"github.com/uswuth/vytora-backend/internal/entity/category"
	"github.com/uswuth/vytora-backend/internal/handlers"
	"github.com/uswuth/vytora-backend/internal/middleware"
	"github.com/uswuth/vytora-backend/internal/services"
)

type NextCodeFunc func(ctx context.Context, entity string) (string, error)

type Handler struct {
	vendorRepo   *Repository
	categoryRepo *category.Repository
	auditLogger  *audit_trail.Logger
	nextCode     NextCodeFunc
	validate     *validator.Validate
}

func NewHandler(vendorRepo *Repository, categoryRepo *category.Repository, auditLogger *audit_trail.Logger, nextCode NextCodeFunc) *Handler {
	return &Handler{
		vendorRepo:   vendorRepo,
		categoryRepo: categoryRepo,
		auditLogger:  auditLogger,
		nextCode:     nextCode,
		validate:     validator.New(),
	}
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateVendorRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "validation failed"})
	}

	if req.Category != "" {
		active, err := h.categoryRepo.IsActive(c.Context(), req.Category)
		if err != nil || !active {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "category not found or inactive"})
		}
	}

	code, err := h.nextCode(c.Context(), "vendor")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate vendor code"})
	}

	claims := c.Locals(middleware.UserContextKey).(*services.Claims)
	createdBy, _ := uuid.Parse(claims.UserID)

	vendor := &Vendor{
		Code:          code,
		Name:          req.Name,
		Category:      req.Category,
		ContactPerson: req.ContactPerson,
		ContactEmail:  req.ContactEmail,
		Country:       req.Country,
		RiskLevel:     req.RiskLevel,
		Status:        "Draft",
		CreatedBy:     createdBy,
	}

	if err := h.vendorRepo.Create(c.Context(), vendor); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create vendor"})
	}

	if err := h.auditLogger.LogCreate(c.Context(), "vendors", vendor.ID, createdBy, vendor); err != nil {
		// Log but don't fail the request
	}

	return c.Status(fiber.StatusCreated).JSON(vendor)
}

func (h *Handler) Get(c *fiber.Ctx) error {
	code := c.Params("code")
	vendor, err := h.vendorRepo.FindByCode(c.Context(), code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "vendor not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return c.JSON(vendor)
}

func (h *Handler) List(c *fiber.Ctx) error {
	claims := c.Locals(middleware.UserContextKey).(*services.Claims)

	params := ListParams{
		Search:    c.Query("search"),
		Category:  c.Query("category"),
		RiskLevel: c.Query("risk_level"),
		Status:    c.Query("status"),
		Country:   c.Query("country"),
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
	}

	// Department managers only see vendors they created
	if claims.Role == "department_manager" {
		uid, _ := uuid.Parse(claims.UserID)
		params.CreatedBy = &uid
	}
	if v := c.Query("limit"); v != "" {
		params.Limit, _ = strconv.Atoi(v)
	}
	if v := c.Query("offset"); v != "" {
		params.Offset, _ = strconv.Atoi(v)
	}

	vendors, total, err := h.vendorRepo.List(c.Context(), params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list vendors"})
	}

	return c.JSON(fiber.Map{"data": vendors, "total": total})
}

func (h *Handler) Update(c *fiber.Ctx) error {
	code := c.Params("code")

	var req UpdateVendorRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "validation failed"})
	}

	if req.Category != "" {
		active, err := h.categoryRepo.IsActive(c.Context(), req.Category)
		if err != nil || !active {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "category not found or inactive"})
		}
	}

	existing, err := h.vendorRepo.FindByCode(c.Context(), code)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "vendor not found"})
	}

	existing.Name = req.Name
	existing.Category = req.Category
	existing.ContactPerson = req.ContactPerson
	existing.ContactEmail = req.ContactEmail
	existing.Country = req.Country
	existing.RiskLevel = req.RiskLevel
	existing.Status = req.Status

	if err := h.vendorRepo.Update(c.Context(), existing); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update vendor"})
	}

	claims := c.Locals(middleware.UserContextKey).(*services.Claims)
	changedBy, _ := uuid.Parse(claims.UserID)
	if err := h.auditLogger.LogCreate(c.Context(), "vendors", existing.ID, changedBy, existing); err != nil {
		// Log but don't fail the request
	}

	return c.JSON(existing)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	code := c.Params("code")

	vendor, findErr := h.vendorRepo.FindByCode(c.Context(), code)
	if findErr != nil {
		if errors.Is(findErr, pgx.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "vendor not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to find vendor"})
	}

	if err := h.vendorRepo.Delete(c.Context(), code); err != nil {
		return handlers.HandleError(c, err)
	}

	claims := c.Locals(middleware.UserContextKey).(*services.Claims)
	changedBy, _ := uuid.Parse(claims.UserID)
	if err := h.auditLogger.LogDelete(c.Context(), "vendors", vendor.ID, vendor, changedBy); err != nil {
		// Log but don't fail the request
	}

	return c.SendStatus(fiber.StatusNoContent)
}
