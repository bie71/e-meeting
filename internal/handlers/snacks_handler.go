package handlers

import (
	"e_metting/internal/models"
	"e_metting/internal/services"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type SnackHandler struct {
	service   *services.SnackService
	validator *validator.Validate
}

func NewSnackHandler(service *services.SnackService, validator *validator.Validate) *SnackHandler {
	return &SnackHandler{
		service:   service,
		validator: validator,
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

func (h *SnackHandler) CreateSnack(c *fiber.Ctx) error {
	var req models.CreateSnackRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "invalid request body " + err.Error(),
		})
	}

	// Validate price is positive
	if req.Price <= 0 {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "price must be positive",
		})
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
			Error: err.Error(),
		})
	}

	// Create snack
	response, err := h.service.CreateSnack(&req)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.Status(http.StatusCreated).JSON(response)
}
