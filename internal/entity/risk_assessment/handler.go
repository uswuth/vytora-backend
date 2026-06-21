package risk_assessment

import (
	"strconv"
	"time"

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
	assessorID, _ := uuid.Parse(claims.UserID)

	ra := &RiskAssessment{
		Code:                 code,
		VendorID:             v.ID,
		VendorCode:           v.Code,
		AssessmentDate:       assessmentDate,
		AssessorID:           &assessorID,
		OverallRiskScore:     req.OverallRiskScore,
		RiskLevel:            req.RiskLevel,
		SecurityRiskScore:    req.SecurityRiskScore,
		FinancialRiskScore:   req.FinancialRiskScore,
		OperationalRiskScore: req.OperationalRiskScore,
		LegalRiskScore:       req.LegalRiskScore,
		Status:               "Draft",
		Notes:                req.Notes,
	}

	if err := h.repo.Create(c.Context(), ra); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create risk assessment"})
	}

	if err := h.auditLogger.LogCreateSimple(c.Context(), "risk_assessments", ra.ID, ra.Code, assessorID); err != nil {
		// Log but don't fail the request
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

func (h *Handler) Update(c *fiber.Ctx) error {
	code := c.Params("code")
	ra, err := h.repo.FindByCode(c.Context(), code)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "risk assessment not found"})
	}

	var req UpdateRiskAssessmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "validation failed"})
	}

	assessmentDate, err := time.Parse("2006-01-02", req.AssessmentDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid assessment_date format"})
	}

	claims := c.Locals(middleware.UserContextKey).(*services.Claims)
	assessorID, _ := uuid.Parse(claims.UserID)

	ra.AssessmentDate = assessmentDate
	ra.AssessorID = &assessorID
	ra.OverallRiskScore = req.OverallRiskScore
	ra.RiskLevel = req.RiskLevel
	ra.SecurityRiskScore = req.SecurityRiskScore
	ra.FinancialRiskScore = req.FinancialRiskScore
	ra.OperationalRiskScore = req.OperationalRiskScore
	ra.LegalRiskScore = req.LegalRiskScore
	ra.Status = req.Status
	ra.Notes = req.Notes

	if err := h.repo.Update(c.Context(), ra); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update risk assessment"})
	}

	if err := h.auditLogger.LogCreateSimple(c.Context(), "risk_assessments", ra.ID, ra.Code, assessorID); err != nil {
		// Log but don't fail the request
	}

	return c.JSON(ra)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	code := c.Params("code")

	ra, err := h.repo.FindByCode(c.Context(), code)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "risk assessment not found"})
	}

	if err := h.repo.Delete(c.Context(), code); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "risk assessment not found"})
	}

	claims := c.Locals(middleware.UserContextKey).(*services.Claims)
	changedBy, _ := uuid.Parse(claims.UserID)
	if err := h.auditLogger.LogDeleteSimple(c.Context(), "risk_assessments", ra.ID, ra.Code, changedBy); err != nil {
		// Log but don't fail the request
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) Approve(c *fiber.Ctx) error {
	code := c.Params("code")
	ra, err := h.repo.FindByCode(c.Context(), code)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "risk assessment not found"})
	}

	oldStatus := ra.Status
	ra.Status = "Approved"

	if err := h.repo.Update(c.Context(), ra); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to approve risk assessment"})
	}

	claims := c.Locals(middleware.UserContextKey).(*services.Claims)
	changedBy, _ := uuid.Parse(claims.UserID)
	if err := h.auditLogger.LogUpdateField(c.Context(), "risk_assessments", ra.ID, "status", oldStatus, "Approved", changedBy); err != nil {
		// Log but don't fail the request
	}

	return c.SendStatus(fiber.StatusNoContent)
}