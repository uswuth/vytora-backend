package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

var LoggerKey = "logger"
var RequestIDKey = "request_id"

func StructuredLogger(logger zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("X-Request-ID", requestID)
		c.Locals(RequestIDKey, requestID)
		start := time.Now()
		err := c.Next()
		ctxLogger := logger.With().
			Str("request_id", requestID).
			Str("service", "vrmp-api").
			Logger()
		ctxLogger.Info().
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", c.Response().StatusCode()).
			Dur("latency", time.Since(start)).
			Msg("request completed")
		c.Locals(LoggerKey, ctxLogger)
		return err
	}
}

func RequestIDMiddleware(c *fiber.Ctx) error {
	requestID := c.Get("X-Request-ID")
	if requestID == "" {
		requestID = uuid.New().String()
	}
	c.Set("X-Request-ID", requestID)
	c.Locals(RequestIDKey, requestID)
	return c.Next()
}

func GetRequestLogger(c *fiber.Ctx) zerolog.Logger {
	if l, ok := c.Locals(LoggerKey).(zerolog.Logger); ok {
		return l
	}
	return zerolog.Nop()
}