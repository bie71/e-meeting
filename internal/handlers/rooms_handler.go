package handlers

import (
	"e_metting/internal/models"
	"e_metting/internal/services"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type RoomHandler struct {
	service *services.RoomService
}

func NewRoomHandler(service *services.RoomService) *RoomHandler {
	return &RoomHandler{
		service: service,
	}
}

func (h *RoomHandler) CreateRoom(c *fiber.Ctx) error {
	var req models.CreateRoomRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "invalid request body " + err.Error(),
		})
	}

	room, err := h.service.CreateRoom(&req)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to create room " + err.Error(),
		})
	}

	return c.Status(http.StatusCreated).JSON(room)
}

func (h *RoomHandler) UpdateRoom(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid room ID " + err.Error(),
		})
	}

	var req models.UpdateRoomRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "invalid request body " + err.Error(),
		})
	}

	room, err := h.service.UpdateRoom(id, &req)
	if err != nil {
		if err.Error() == "room not found" {
			return c.Status(http.StatusNotFound).JSON(models.ErrorResponse{
				Error: "room not found " + err.Error(),
			})
		}

		return c.Status(http.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to update room " + err.Error(),
		})
	}

	return c.JSON(room)
}

func (h *RoomHandler) DeleteRoom(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid room ID " + err.Error(),
		})
	}

	err = h.service.DeleteRoom(id)
	if err != nil {
		switch err.Error() {
		case "room not found":
			return c.Status(http.StatusNotFound).JSON(models.ErrorResponse{
				Error: "room not found " + err.Error(),
			})
		case "cannot delete room with active reservations":
			return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
				Error: "cannot delete room with active reservations " + err.Error(),
			})
		default:
			return c.Status(http.StatusInternalServerError).JSON(models.ErrorResponse{
				Error: "Failed to delete room " + err.Error(),
			})
		}
	}

	return c.Status(http.StatusOK).JSON(models.SuccessResponse{
		Message: "Room deleted successfully",
	})
}

func (h *RoomHandler) GetRooms(c *fiber.Ctx) error {

	// Parse pagination query parameters
	var pagination models.PaginationQuery
	if err := c.QueryParser(&pagination); err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "invalid query " + err.Error(),
		})
	}

	// Parse filter from request body (if provided)
	var filter models.RoomFilter
	if c.Body() != nil && len(c.Body()) > 0 {
		if err := c.BodyParser(&filter); err != nil {
			return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
				Error: "invalid query " + err.Error(),
			})
		}
	}

	// Get rooms with filter and pagination
	response, err := h.service.GetRooms(&filter, &pagination)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to fetch rooms " + err.Error(),
		})
	}

	return c.JSON(response)
}

func (h *RoomHandler) GetRoomSchedule(c *fiber.Ctx) error {
	// Parse room ID from URL
	roomID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "invalid room ID " + err.Error(),
		})
	}

	// Parse and validate query parameters
	var query models.RoomScheduleQuery
	if err := c.QueryParser(&query); err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "invalid query " + err.Error(),
		})
	}

	// Validate time range
	if query.StartDateTime.After(query.EndDateTime) {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "start_datetime cannot be after end_datetime",
		})
	}

	// Get room schedule from service
	response, err := h.service.GetRoomSchedule(roomID, &query)
	if err != nil {
		if err.Error() == "room not found" {
			return c.Status(http.StatusNotFound).JSON(models.ErrorResponse{
				Error: "room not found " + err.Error(),
			})
		}

		return c.Status(http.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to fetch room schedule " + err.Error(),
		})
	}

	return c.JSON(response)
}
