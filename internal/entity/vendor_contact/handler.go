package vendor_contact

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/uswuth/vytora-backend/internal/entity/vendor"
	"github.com/uswuth/vytora-backend/internal/handlers"
)

type NextCodeFunc func(ctx context.Context, entity string) (string, error)

type Handler struct {
	contactRepo *Repository
	vendorRepo  *vendor.Repository
	nextCode    NextCodeFunc
	validate    *validator.Validate
}

func NewHandler(contactRepo *Repository, vendorRepo *vendor.Repository, nextCode NextCodeFunc) *Handler {
	return &Handler{
		contactRepo: contactRepo,
		vendorRepo:  vendorRepo,
		nextCode:    nextCode,
		validate:    validator.New(),
	}
}

// List returns all contacts for a vendor.
func (h *Handler) List(c *fiber.Ctx) error {
	vendorCode := c.Params("code")
	v, err := h.vendorRepo.FindByCode(c.Context(), vendorCode)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "vendor not found"})
	}

	contacts, err := h.contactRepo.ListByVendor(c.Context(), v.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list contacts"})
	}

	return c.JSON(contacts)
}

// Create adds a new contact to a vendor.
func (h *Handler) Create(c *fiber.Ctx) error {
	vendorCode := c.Params("code")
	v, err := h.vendorRepo.FindByCode(c.Context(), vendorCode)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "vendor not found"})
	}

	var req CreateContactRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "validation failed"})
	}

	code, err := h.nextCode(c.Context(), "vendor_contact")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate code"})
	}

	contact := &VendorContact{
		Code:     code,
		VendorID: v.ID,
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
	}

	if err := h.contactRepo.Create(c.Context(), contact); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create contact"})
	}

	return c.Status(fiber.StatusCreated).JSON(contact)
}

// Update modifies an existing contact.
func (h *Handler) Update(c *fiber.Ctx) error {
	contactID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid contact id"})
	}

	contact, err := h.contactRepo.FindByID(c.Context(), contactID)
	if err != nil {
		return handlers.HandleError(c, err)
	}
	if contact == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "contact not found"})
	}

	var req UpdateContactRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "validation failed"})
	}

	contact.Name = req.Name
	contact.Email = req.Email
	contact.Phone = req.Phone

	if err := h.contactRepo.Update(c.Context(), contact); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update contact"})
	}

	return c.JSON(contact)
}

// Delete removes a contact.
func (h *Handler) Delete(c *fiber.Ctx) error {
	contactID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid contact id"})
	}

	if err := h.contactRepo.Delete(c.Context(), contactID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "contact not found"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}