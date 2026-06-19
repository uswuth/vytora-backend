package contract

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/go-playground/validator/v10"
	"github.com/uswuth/vytora-backend/internal/entity/vendor"
)

type Handler struct {
	contractRepo *Repository
	vendorRepo   *vendor.Repository
	nextCode     vendor.NextCodeFunc
	validate     *validator.Validate
}

func NewHandler(contractRepo *Repository, vendorRepo *vendor.Repository, nextCode vendor.NextCodeFunc) *Handler {
	return &Handler{
		contractRepo: contractRepo,
		vendorRepo:   vendorRepo,
		nextCode:     nextCode,
		validate:     validator.New(),
	}
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateContractRequest
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

	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid start_date format"})
	}
	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid end_date format"})
	}

	code, err := h.nextCode(c.Context(), "contract")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate code"})
	}

	contract := &Contract{
		Code:           code,
		VendorID:       v.ID,
		VendorCode:     v.Code,
		ContractNumber: req.ContractNumber,
		StartDate:      start,
		EndDate:        end,
		ContractValue:  req.ContractValue,
		RenewalStatus:  req.RenewalStatus,
	}

	if err := h.contractRepo.Create(c.Context(), contract); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create contract"})
	}

	return c.Status(fiber.StatusCreated).JSON(contract)
}

func (h *Handler) Get(c *fiber.Ctx) error {
	code := c.Params("code")
	contract, err := h.contractRepo.FindByCode(c.Context(), code)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "contract not found"})
	}
	return c.JSON(contract)
}

func (h *Handler) List(c *fiber.Ctx) error {
	vendorCode := c.Query("vendor_code")
	v, err := h.vendorRepo.FindByCode(c.Context(), vendorCode)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "vendor not found"})
	}
	contracts, err := h.contractRepo.ListByVendor(c.Context(), v.ID.String())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list contracts"})
	}
	return c.JSON(contracts)
}

func (h *Handler) Expiring(c *fiber.Ctx) error {
	daysStr := c.Query("days")
	if daysStr == "" {
		daysStr = "30"
	}
	days, _ := strconv.Atoi(daysStr)

	contracts, err := h.contractRepo.Expiring(c.Context(), days)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get expiring contracts"})
	}
	return c.JSON(contracts)
}