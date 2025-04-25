package handlers

import (
	"e_metting/internal/models"
	"e_metting/internal/services"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type SnackHandler struct {
	service *services.SnackService
}

func NewSnackHandler(service *services.SnackService) *SnackHandler {
	return &SnackHandler{
		service: service,
	}
}

func (h *SnackHandler) GetSnacks(c *fiber.Ctx) error {
	// Get snacks from service
	response, err := h.service.GetSnacks()
	if err != nil {
		if err.Error() == "snacks not found" {
			return c.Status(http.StatusNotFound).JSON(models.ErrorResponse{
				Error: err.Error(),
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to fetch snacks",
		})
	}

	return c.JSON(response)
}
