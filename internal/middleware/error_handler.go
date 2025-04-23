package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func ErrorHandlerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Defer untuk menangkap panic
		defer func() {
			if r := recover(); r != nil {
				log.Error().Interface("panic", r).Msg("Unhandled panic")
				c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
					Error: "Internal Server Error",
				})
			}
		}()

		err := c.Next()
		if err != nil {
			// Log error
			log.Error().Err(err).Msg("Request error")

			// Kirim response error
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			return c.Status(code).JSON(ErrorResponse{
				Error: err.Error(),
			})
		}

		return nil
	}
}
