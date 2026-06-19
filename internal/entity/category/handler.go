package category

import (
	"context"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/uswuth/vytora-backend/internal/middleware"
	"github.com/uswuth/vytora-backend/internal/services"
)

type NextCodeFunc func(ctx context.Context, entity string) (string, error)

type Handler struct {
	repo       *Repository
	nextCode   NextCodeFunc
	validate   *validator.Validate
}

func NewHandler(repo *Repository, nextCode NextCodeFunc) *Handler {
	return &Handler{
		repo:       repo,
		nextCode:   nextCode,
		validate:   validator.New(),
	}
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "validation failed"})
	}

	existing, _ := h.repo.FindByName(c.Context(), req.Name)
	if existing != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "category with this name already exists"})
	}

	code, err := h.nextCode(c.Context(), "category")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate code"})
	}

	claims := c.Locals(middleware.UserContextKey).(*services.Claims)
	createdBy, _ := uuid.Parse(claims.UserID)

	cat := &Category{
		Code:        code,
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Status:      req.Status,
		CreatedBy:   &createdBy,
	}

	if err := h.repo.Create(c.Context(), cat); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create category"})
	}

	return c.Status(fiber.StatusCreated).JSON(cat)
}

func (h *Handler) List(c *fiber.Ctx) error {
	params := ListParams{
		Search: c.Query("search"),
		Status: c.Query("status"),
	}
	if v := c.Query("limit"); v != "" {
		params.Limit, _ = strconv.Atoi(v)
	}
	if v := c.Query("offset"); v != "" {
		params.Offset, _ = strconv.Atoi(v)
	}

	cats, total, err := h.repo.List(c.Context(), params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list categories"})
	}

	return c.JSON(fiber.Map{"data": cats, "total": total})
}

func (h *Handler) Get(c *fiber.Ctx) error {
	code := c.Params("code")
	cat, err := h.repo.FindByCode(c.Context(), code)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "category not found"})
	}
	return c.JSON(cat)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	code := c.Params("code")
	cat, err := h.repo.FindByCode(c.Context(), code)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "category not found"})
	}

	var req UpdateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "validation failed"})
	}

	claims := c.Locals(middleware.UserContextKey).(*services.Claims)
	updatedBy, _ := uuid.Parse(claims.UserID)

	cat.DisplayName = req.DisplayName
	cat.Description = req.Description
	cat.Status = req.Status
	cat.UpdatedBy = &updatedBy

	if err := h.repo.Update(c.Context(), cat); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update category"})
	}

	return c.JSON(cat)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	code := c.Params("code")
	if err := h.repo.Delete(c.Context(), code); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete category"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}