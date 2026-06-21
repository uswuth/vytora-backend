package export

import (
	"encoding/csv"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/uswuth/vytora-backend/internal/entity/compliance_record"
	"github.com/uswuth/vytora-backend/internal/entity/contract"
	"github.com/uswuth/vytora-backend/internal/entity/risk_assessment"
	"github.com/uswuth/vytora-backend/internal/entity/vendor"
	"github.com/uswuth/vytora-backend/internal/middleware"
	"github.com/uswuth/vytora-backend/internal/services"
)

type Handler struct {
	vendorRepo   *vendor.Repository
	riskRepo     *risk_assessment.Repository
	compRepo     *compliance_record.Repository
	contractRepo *contract.Repository
}

func NewHandler(
	vendorRepo *vendor.Repository,
	riskRepo *risk_assessment.Repository,
	compRepo *compliance_record.Repository,
	contractRepo *contract.Repository,
) *Handler {
	return &Handler{
		vendorRepo:   vendorRepo,
		riskRepo:     riskRepo,
		compRepo:     compRepo,
		contractRepo: contractRepo,
	}
}

func writeCSV(c *fiber.Ctx, filename string, headers []string, rows [][]string) error {
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	writer := csv.NewWriter(c)
	defer writer.Flush()

	if err := writer.Write(headers); err != nil {
		return err
	}
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) getCreatedByParam(c *fiber.Ctx) *uuid.UUID {
	claims, ok := c.Locals(middleware.UserContextKey).(*services.Claims)
	if !ok || claims.Role != "department_manager" {
		return nil
	}
	uid, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil
	}
	return &uid
}

func (h *Handler) VendorsCSV(c *fiber.Ctx) error {
	params := vendor.ListParams{
		Limit:  10000,
		Offset: 0,
	}

	createdBy := h.getCreatedByParam(c)
	if createdBy != nil {
		params.CreatedBy = createdBy
	}

	vendors, _, err := h.vendorRepo.List(c.Context(), params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to export vendors"})
	}

	headers := []string{"Code", "Name", "Category", "Risk Level", "Status", "Country", "Created At", "Updated At"}
	var rows [][]string
	for _, v := range vendors {
		rows = append(rows, []string{
			v.Code, v.Name, v.Category, v.RiskLevel, v.Status, v.Country,
			v.CreatedAt.Format("2006-01-02 15:04:05"),
			v.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return writeCSV(c, "vendors.csv", headers, rows)
}

func (h *Handler) RisksCSV(c *fiber.Ctx) error {
	params := risk_assessment.ListParams{Limit: 10000}
	assessments, _, err := h.riskRepo.ListAll(c.Context(), params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to export risk assessments"})
	}

	headers := []string{"Code", "Vendor Code", "Assessment Date", "Overall Score", "Risk Level", "Security Score", "Financial Score", "Operational Score", "Legal Score", "Status"}
	var rows [][]string
	for _, ra := range assessments {
		rows = append(rows, []string{
			ra.Code, ra.VendorCode, ra.AssessmentDate.Format("2006-01-02"),
			strconv.FormatFloat(ra.OverallRiskScore, 'f', 2, 64),
			ra.RiskLevel,
			strconv.FormatFloat(ra.SecurityRiskScore, 'f', 2, 64),
			strconv.FormatFloat(ra.FinancialRiskScore, 'f', 2, 64),
			strconv.FormatFloat(ra.OperationalRiskScore, 'f', 2, 64),
			strconv.FormatFloat(ra.LegalRiskScore, 'f', 2, 64),
			ra.Status,
		})
	}

	return writeCSV(c, "risk_assessments.csv", headers, rows)
}

func (h *Handler) ComplianceCSV(c *fiber.Ctx) error {
	headers := []string{"Code", "Vendor Code", "Certification Type", "Status", "Valid From", "Valid Until", "Issued By"}
	var rows [][]string

	vendorParams := vendor.ListParams{Limit: 10000}
	vendors, _, err := h.vendorRepo.List(c.Context(), vendorParams)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to export compliance records"})
	}

	for _, v := range vendors {
		records, err := h.compRepo.ListByVendor(c.Context(), v.ID.String())
		if err != nil {
			continue
		}
		for _, cr := range records {
			validFrom := ""
			validUntil := ""
			if cr.ValidFrom != nil {
				validFrom = cr.ValidFrom.Format("2006-01-02")
			}
			if cr.ValidUntil != nil {
				validUntil = cr.ValidUntil.Format("2006-01-02")
			}
			rows = append(rows, []string{
				cr.Code, cr.VendorCode, cr.CertificationType, cr.Status,
				validFrom, validUntil, cr.IssuedBy,
			})
		}
	}

	return writeCSV(c, "compliance_records.csv", headers, rows)
}

func (h *Handler) ContractsCSV(c *fiber.Ctx) error {
	headers := []string{"Code", "Vendor Code", "Contract Number", "Start Date", "End Date", "Contract Value", "Renewal Status"}
	var rows [][]string

	vendorParams := vendor.ListParams{Limit: 10000}
	vendors, _, err := h.vendorRepo.List(c.Context(), vendorParams)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to export contracts"})
	}

	for _, v := range vendors {
		contracts, err := h.contractRepo.ListByVendor(c.Context(), v.ID.String())
		if err != nil {
			continue
		}
		for _, ctr := range contracts {
			valueStr := ""
			if ctr.ContractValue != nil {
				valueStr = strconv.FormatFloat(*ctr.ContractValue, 'f', 2, 64)
			}
			rows = append(rows, []string{
				ctr.Code, ctr.VendorCode, ctr.ContractNumber,
				ctr.StartDate.Format("2006-01-02"),
				ctr.EndDate.Format("2006-01-02"),
				valueStr,
				ctr.RenewalStatus,
			})
		}
	}

	return writeCSV(c, "contracts.csv", headers, rows)
}