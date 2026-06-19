package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

// StructuredLogger returns a middleware that logs every request in JSON format.
func StructuredLogger(logger zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		logger.Info().
			Str("method", c.Method()).
			Str("url", c.OriginalURL()).
			Int("status", c.Response().StatusCode()).
			Dur("duration", time.Since(start)).
			Str("remote_addr", c.IP()).
			Msg("request completed")

		return err
	}
}