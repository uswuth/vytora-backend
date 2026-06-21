package report

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/uswuth/vytora-backend/internal/middleware"
	"github.com/uswuth/vytora-backend/internal/services"
)

type Handler struct {
	reportRepo *Repository
}

func NewHandler(reportRepo *Repository) *Handler {
	return &Handler{
		reportRepo: reportRepo,
	}
}

func getUserDeptID(c *fiber.Ctx) *uuid.UUID {
	claims, ok := c.Locals(middleware.UserContextKey).(*services.Claims)
	if !ok {
		return nil
	}
	if claims.Role == "department_manager" {
		uid, err := uuid.Parse(claims.UserID)
		if err == nil {
			return &uid
		}
	}
	return nil
}

func (h *Handler) Summary(c *fiber.Ctx) error {
	deptID := getUserDeptID(c)
	summary, err := h.reportRepo.GetSummary(c.Context(), deptID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate summary"})
	}
	return c.JSON(summary)
}

func (h *Handler) MonthlyOnboarding(c *fiber.Ctx) error {
	deptID := getUserDeptID(c)
	data, err := h.reportRepo.GetMonthlyOnboarding(c.Context(), deptID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get monthly onboarding"})
	}
	return c.JSON(data)
}

func (h *Handler) HighRiskVendors(c *fiber.Ctx) error {
	deptID := getUserDeptID(c)
	items, err := h.reportRepo.GetHighRiskVendors(c.Context(), deptID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get high-risk vendors"})
	}
	return c.JSON(items)
}

func (h *Handler) ExpiringContractsReport(c *fiber.Ctx) error {
	deptID := getUserDeptID(c)
	items, err := h.reportRepo.GetExpiringContractsReport(c.Context(), deptID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get expiring contracts report"})
	}
	return c.JSON(items)
}

func (h *Handler) ComplianceSummaryReport(c *fiber.Ctx) error {
	deptID := getUserDeptID(c)
	items, err := h.reportRepo.GetComplianceSummary(c.Context(), deptID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get compliance summary"})
	}
	return c.JSON(items)
}

func (h *Handler) TimeSeriesReport(c *fiber.Ctx) error {
	deptID := getUserDeptID(c)
	data, err := h.reportRepo.GetTimeSeries(c.Context(), deptID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get time series"})
	}
	return c.JSON(data)
}