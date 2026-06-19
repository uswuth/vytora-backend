package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/uswuth/vytora-backend/internal/services"
)

// RequirePermission checks if the authenticated user has the specified permission.
func RequirePermission(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, ok := c.Locals(UserContextKey).(*services.Claims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}

		allowedPermissions, exists := RolePermissions[claims.Role]
		if !exists {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "forbidden",
				"role": claims.Role,
				"permission_required": permission,
				"reason": "role not found in permissions map",
			})
		}

		for _, p := range allowedPermissions {
			if p == permission {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "forbidden",
			"role": claims.Role,
			"your_permissions": allowedPermissions,
			"permission_required": permission,
			"reason": "permission not granted to this role",
		})
	}
}