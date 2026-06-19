package handlers

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/uswuth/vytora-backend/internal/entity/user"
	"github.com/uswuth/vytora-backend/internal/services"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userRepo    *user.Repository
	jwtService  *services.JWTService
	validate    *validator.Validate
}

func NewAuthHandler(userRepo *user.Repository, jwtService *services.JWTService) *AuthHandler {
	return &AuthHandler{
		userRepo:    userRepo,
		jwtService:  jwtService,
		validate:   validator.New(),
	}
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req user.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "validation failed"})
	}
	u, err := h.userRepo.FindByEmail(c.Context(), req.Email)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid email or password"})
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid email or password"})
	}
	if !u.IsActive {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "account is deactivated"})
	}
	token, err := h.jwtService.GenerateToken(u)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate token"})
	}
	resp := user.LoginResponse{
		Token: token,
		User:  *u,
	}
	return c.Status(fiber.StatusOK).JSON(resp)
}

func (h *AuthHandler) ExtendSession(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing authorization header"})
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid authorization format"})
	}

	newToken, ttl, err := h.jwtService.ExtendToken(parts[1])
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or expired token"})
	}

	return c.JSON(fiber.Map{
		"token":      newToken,
		"expires_in": ttl,
	})
}