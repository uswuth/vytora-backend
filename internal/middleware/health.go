package middleware

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/uswuth/vytora-backend/internal/database"
)

type HealthChecker struct {
	logger     zerolog.Logger
	startTime  time.Time
	allowedIPs map[string]bool
}

func NewHealthChecker(logger zerolog.Logger, allowedIPs []string) *HealthChecker {
	ipMap := make(map[string]bool)
	for _, ip := range allowedIPs {
		ipMap[ip] = true
		ipMap["127.0.0.1"] = true
		ipMap["::1"] = true
	}
	if len(ipMap) == 0 {
		ipMap["127.0.0.1"] = true
		ipMap["::1"] = true
	}
	return &HealthChecker{logger: logger, startTime: time.Now(), allowedIPs: ipMap}
}

func (h *HealthChecker) RegisterRoutes(app *fiber.App) {
	app.Get("/healthz", h.Healthz)
	app.Get("/readyz", h.Readyz)
}

func (h *HealthChecker) Healthz(c *fiber.Ctx) error {
	if !h.isAllowed(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
	}
	return c.JSON(fiber.Map{
		"status": "ok",
		"uptime": time.Since(h.startTime).Seconds(),
	})
}

func (h *HealthChecker) Readyz(c *fiber.Ctx) error {
	if !h.isAllowed(c) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
	}
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

func (h *HealthChecker) isAllowed(c *fiber.Ctx) bool {
	ip := c.IP()
	forwarded := c.Get("X-Forwarded-For")
	if forwarded != "" {
		ip = forwarded
	}
	return h.allowedIPs[ip]
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