package user

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

type NextCodeFunc func(ctx context.Context, entity string) (string, error)

type Handler struct {
	repo     *Repository
	nextCode NextCodeFunc
	validate *validator.Validate
}

func NewHandler(repo *Repository, nextCode NextCodeFunc) *Handler {
	return &Handler{
		repo:     repo,
		nextCode: nextCode,
		validate: validator.New(),
	}
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "validation failed"})
	}

	code, err := h.nextCode(c.Context(), "user")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate user code"})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to hash password"})
	}

	user := &User{
		Code:         code,
		Email:        req.Email,
		PasswordHash: string(hash),
		FullName:     req.FullName,
		Role:         req.Role,
		IsActive:     true,
	}

	if err := h.repo.Create(c.Context(), user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "failed to create user",
			"details": err.Error(),
		})
	}

	user.PasswordHash = ""

	return c.Status(fiber.StatusCreated).JSON(user)
}

func (h *Handler) List(c *fiber.Ctx) error {
	users, err := h.repo.ListAll(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list users"})
	}
	for i := range users {
		users[i].PasswordHash = ""
	}
	return c.JSON(users)
}

func (h *Handler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	user, err := h.repo.FindByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}
	user.PasswordHash = ""
	return c.JSON(user)
}

func (h *Handler) UpdateRole(c *fiber.Ctx) error {
	id := c.Params("id")
	var req UpdateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "validation failed"})
	}
	if err := h.repo.UpdateRole(c.Context(), id, req.Role); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update role"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) Deactivate(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.repo.SetActive(c.Context(), id, false); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to deactivate user"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) Activate(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.repo.SetActive(c.Context(), id, true); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to activate user"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}