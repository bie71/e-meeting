package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func LoggerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Start timer
		start := time.Now()

		// Read request body
		var requestBody []byte
		if c.Request().Body() != nil {
			requestBody = c.Body()
			// You can't reassign the body in Fiber, so just log it
		}

		// Process request
		err := c.Next()

		// Stop timer
		duration := time.Since(start)

		// Log request details
		log.Info().
			Str("method", c.Method()).
			Str("path", c.OriginalURL()).
			Str("ip", c.IP()).
			Int("status", c.Response().StatusCode()).
			Dur("duration", duration).
			Str("user_agent", string(c.Request().Header.UserAgent())).
			Bytes("request_body", requestBody).
			Msg("HTTP request")

		return err
	}
}
