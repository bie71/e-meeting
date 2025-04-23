package middleware

import (
	"e_metting/internal/models"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

func ValidateRequest[T any]() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var request T
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
				Error: "Invalid request body",
			})
		}

		if err := validate.Struct(request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
				Error: err.Error(),
			})
		}

		c.Locals("request", request)
		return c.Next()
	}
}
