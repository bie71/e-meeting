package handlers

import (
	"e_metting/internal/models"
	"e_metting/internal/services"
	"log"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
	var raw models.SnackFilterRaw
	var pagination models.PaginationQuery
	if err := c.QueryParser(&raw); err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "invalid query " + err.Error(),
		})
	}

	if err := c.QueryParser(&pagination); err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "invalid query " + err.Error(),
		})
	}

	filter := models.SnackFilter{}

	if raw.Search != "" {
		filter.Search = &raw.Search
	}
	if raw.Category != "" {
		filter.Category = &raw.Category
	}

	if raw.MinPrice != "" {
		v, err := strconv.ParseFloat(raw.MinPrice, 64)
		if err == nil {
			filter.MinPrice = &v
		}
	}
	if raw.MaxPrice != "" {
		v, err := strconv.ParseFloat(raw.MaxPrice, 64)
		if err == nil {
			filter.MaxPrice = &v
		}
	}

	log.Printf("Parsed filter: %+v", filter)
	// Get snacks from service
	response, err := h.service.GetSnacks(&filter, &pagination)
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

func (h *SnackHandler) UpdateSnack(c *fiber.Ctx) error {
	var req models.Snack
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

	// Update snack
	err := h.service.UpdateSnack(req.ID, &req)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.JSON(models.SuccessResponse{
		Message: "Snack updated successfully",
	})
}

func (h *SnackHandler) DeleteSnack(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid snack ID " + err.Error(),
		})
	}

	err = h.service.DeleteSnack(id)
	if err != nil {
		switch err.Error() {
		case "snack not found":
			return c.Status(http.StatusNotFound).JSON(models.ErrorResponse{
				Error: "snack not found " + err.Error(),
			})
		default:
			return c.Status(http.StatusInternalServerError).JSON(models.ErrorResponse{
				Error: "Failed to delete snack " + err.Error(),
			})
		}
	}

	return c.Status(http.StatusOK).JSON(models.SuccessResponse{
		Message: "Snack deleted successfully",
	})
}

func (h *SnackHandler) GetSnackByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid snack ID " + err.Error(),
		})
	}

	snack, err := h.service.GetSnackByID(id)
	if err != nil {
		switch err.Error() {
		case "snack not found":
			return c.Status(http.StatusNotFound).JSON(models.ErrorResponse{
				Error: "snack not found " + err.Error(),
			})
		default:
			return c.Status(http.StatusInternalServerError).JSON(models.ErrorResponse{
				Error: "Failed to get snack " + err.Error(),
			})
		}
	}

	return c.JSON(snack)
}
