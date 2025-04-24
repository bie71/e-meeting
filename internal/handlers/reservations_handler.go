package handlers

import (
	"e_metting/internal/models"
	"e_metting/internal/services"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type ReservationHandler struct {
	service *services.ReservationService
}

func NewReservationHandler(service *services.ReservationService) *ReservationHandler {
	return &ReservationHandler{
		service: service,
	}
}

func (h *ReservationHandler) GetReservationHistory(c *fiber.Ctx) error {
	var query models.ReservationHistoryQuery
	if err := c.QueryParser(&query); err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid query parameters",
		})
	}

	response, err := h.service.GetReservationHistory(&query)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to fetch reservation history",
		})
	}

	return c.JSON(response)
}

func (h *ReservationHandler) UpdateReservationStatus(c *fiber.Ctx) error {
	var req models.UpdateReservationStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid query parameters",
		})
	}
	updatedReservation, err := h.service.UpdateReservationStatus(&req)
	if err != nil {
		if err.Error() == "reservation not found" {
			return c.Status(http.StatusNotFound).JSON(models.ErrorResponse{
				Error: "reservation not found " + err.Error(),
			})
		}
		if strings.Contains(err.Error(), "invalid status") {
			return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
				Error: "invalid status " + err.Error(),
			})
		}

		return c.Status(http.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to update reservation status",
		})
	}

	return c.JSON(updatedReservation)
}
