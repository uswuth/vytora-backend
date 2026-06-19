package middleware

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/uswuth/vytora-backend/internal/database"
)

type HealthChecker struct {
	logger    zerolog.Logger
	startTime time.Time
}

func NewHealthChecker(logger zerolog.Logger) *HealthChecker {
	return &HealthChecker{logger: logger, startTime: time.Now()}
}

func (h *HealthChecker) RegisterRoutes(app *fiber.App) {
	app.Get("/healthz", h.Healthz)
	app.Get("/readyz", h.Readyz)
}

func (h *HealthChecker) Healthz(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
		"uptime": time.Since(h.startTime).Seconds(),
	})
}

func (h *HealthChecker) Readyz(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()

	checks := fiber.Map{
		"database": h.checkDatabase(ctx),
		"jwt":      h.checkJWT(),
	}

	allHealthy := true
	for _, status := range checks {
		if s, ok := status.(string); ok && s != "ok" {
			allHealthy = false
			break
		}
	}

	status := "ready"
	code := fiber.StatusOK
	if !allHealthy {
		status = "not ready"
		code = fiber.StatusServiceUnavailable
	}

	return c.Status(code).JSON(fiber.Map{
		"status": status,
		"checks": checks,
	})
}

func (h *HealthChecker) checkDatabase(ctx context.Context) string {
	if database.Pool == nil {
		return "down"
	}
	if err := database.Pool.Ping(ctx); err != nil {
		h.logger.Error().Err(err).Msg("DB health check failed")
		return "down"
	}
	return "ok"
}

func (h *HealthChecker) checkJWT() string {
	return "ok"
}