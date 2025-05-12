package handlers

import (
	"e_metting/internal/models"
	"e_metting/internal/services"
	"log"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
	log.Printf("Querying history from %s to %s", query.StartDatetime, query.EndDatetime)

	isAdmin, _ := c.Locals("isAdmin").(bool)
	query.IsAdmin = isAdmin
	if !isAdmin {
		authUserID, _ := c.Locals("userID").(string)
		query.UserID = uuid.MustParse(authUserID)

		log.Println("User ID from context:", query.UserID)
	}

	response, err := h.service.GetReservationHistory(&query)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to fetch reservation history " + err.Error(),
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

func (h *ReservationHandler) CalculateReservationCost(c *fiber.Ctx) error {
	var req models.ReservationCalculationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Validate time range
	if req.EndTime.Before(req.StartTime) {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "end_time cannot be before start_time",
		})
	}

	// Calculate costs
	response, err := h.service.CalculateReservationCost(&req)
	if err != nil {
		if err.Error() == "room not found or inactive" {
			return c.Status(http.StatusNotFound).JSON(models.ErrorResponse{
				Error: err.Error(),
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to calculate reservation cost",
		})
	}

	return c.JSON(response)
}

func (h *ReservationHandler) GetReservationByID(c *fiber.Ctx) error {
	// Parse reservation ID from URL
	reservationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid reservation ID",
		})

	}

	// Get reservation details from service
	reservation, err := h.service.GetReservationByID(reservationID)
	if err != nil {
		if err.Error() == "reservation not found" {
			return c.Status(http.StatusNotFound).JSON(models.ErrorResponse{
				Error: err.Error(),
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to fetch reservation details" + err.Error(),
		})
	}

	return c.JSON(reservation)
}

func (h *ReservationHandler) CreateReservation(c *fiber.Ctx) error {
	var req models.CreateReservationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Validate time range
	if req.EndTime.Before(req.StartTime) {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "end_time cannot be before start_time",
		})
	}

	// Create reservation
	response, err := h.service.CreateReservation(&req)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to create reservation " + err.Error(),
		})
	}

	return c.Status(http.StatusCreated).JSON(response)
}
