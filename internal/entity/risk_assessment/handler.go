package risk_assessment

import (
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
	repo         *Repository
	vendorRepo   *vendor.Repository
	nextCode     vendor.NextCodeFunc
	validate     *validator.Validate
}

func NewHandler(repo *Repository, vendorRepo *vendor.Repository, nextCode vendor.NextCodeFunc) *Handler {
	return &Handler{
		repo:       repo,
		vendorRepo: vendorRepo,
		nextCode:   nextCode,
		validate:   validator.New(),
	}
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateRiskAssessmentRequest
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

	assessmentDate, err := time.Parse("2006-01-02", req.AssessmentDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid assessment_date format"})
	}

	code, err := h.nextCode(c.Context(), "risk_assessment")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate code"})
	}

	claims := c.Locals(middleware.UserContextKey).(*services.Claims)
	reviewedBy, _ := uuid.Parse(claims.UserID)
	now := time.Now()

	ra := &RiskAssessment{
		Code:            code,
		VendorID:        v.ID,
		VendorCode:      v.Code,
		AssessmentDate:  assessmentDate,
		RiskLevel:       req.RiskLevel,
		Findings:        req.Findings,
		Recommendations: req.Recommendations,
		Status:          "Pending",
		ReviewedBy:      &reviewedBy,
		ReviewedAt:      &now,
	}

	if err := h.repo.Create(c.Context(), ra); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create risk assessment"})
	}

	return c.Status(fiber.StatusCreated).JSON(ra)
}

func (h *Handler) Get(c *fiber.Ctx) error {
	code := c.Params("code")
	ra, err := h.repo.FindByCode(c.Context(), code)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "risk assessment not found"})
	}
	return c.JSON(ra)
}

func (h *Handler) List(c *fiber.Ctx) error {
	vendorCode := c.Query("vendor_code")
	params := ListParams{
		VendorCode: vendorCode,
		RiskLevel:  c.Query("risk_level"),
		Status:     c.Query("status"),
	}
	if v := c.Query("limit"); v != "" {
		params.Limit, _ = strconv.Atoi(v)
	}
	if v := c.Query("offset"); v != "" {
		params.Offset, _ = strconv.Atoi(v)
	}

	assessments, total, err := h.repo.ListAll(c.Context(), params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list risk assessments"})
	}

	return c.JSON(fiber.Map{"data": assessments, "total": total})
}

func (h *Handler) Approve(c *fiber.Ctx) error {
	code := c.Params("code")
	ra, err := h.repo.FindByCode(c.Context(), code)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "risk assessment not found"})
	}

	ra.Status = "Approved"

	if err := h.repo.Update(c.Context(), ra); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to approve risk assessment"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}