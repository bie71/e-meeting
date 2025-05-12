package handlers

import (
	"os"
	"path/filepath"

	"e_metting/internal/models"

	"github.com/gofiber/fiber/v2"
)

func SwaggerUI(c *fiber.Ctx) error {
	swaggerPath := filepath.Join("docs", "swagger.json")
	swaggerFile, err := os.ReadFile(swaggerPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to load Swagger documentation",
		})
	}

	c.Set("Content-Type", "application/json")
	return c.Send(swaggerFile)
}

func RecoverPassword(c *fiber.Ctx) error {
	// Implement the logic to handle password recovery
	token := c.Query("token")

	return c.SendFile("./public/recover-password.html?token="+token, true) // Serve the password recovery page
}

func Login(c *fiber.Ctx) error {
	// Implement the logic to handle login
	return c.SendFile("./public/login.html", true) // Serve the login page
}
