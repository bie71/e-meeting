package handlers

import (
	"e_metting/internal/models"
	"e_metting/internal/services"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type MeetingHandler struct {
	meetingService *services.MeetingService
}

func NewMeetingHandler(meetingService *services.MeetingService) *MeetingHandler {
	return &MeetingHandler{
		meetingService: meetingService,
	}
}

func (h *MeetingHandler) CreateMeeting(c *fiber.Ctx) error {
	var req models.MeetingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	userID := c.Locals("userID").(uuid.UUID)
	meeting, err := h.meetingService.CreateMeeting(c.Context(), req, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create meeting")
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to create meeting",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(meeting)
}

func (h *MeetingHandler) GetMeeting(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid meeting ID",
		})
	}

	meeting, err := h.meetingService.GetMeeting(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get meeting")
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to get meeting",
		})
	}

	return c.JSON(meeting)
}

func (h *MeetingHandler) ListMeetings(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	var start, end time.Time
	var err error

	if startTime != "" {
		start, err = time.Parse(time.RFC3339, startTime)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
				Error: "Invalid start time format",
			})
		}
	}

	if endTime != "" {
		end, err = time.Parse(time.RFC3339, endTime)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
				Error: "Invalid end time format",
			})
		}
	}

	meetings, err := h.meetingService.ListMeetings(c.Context(), userID, start, end)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list meetings")
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to list meetings",
		})
	}

	return c.JSON(meetings)
}

func (h *MeetingHandler) UpdateMeeting(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid meeting ID",
		})
	}

	var req models.MeetingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	userID := c.Locals("userID").(uuid.UUID)
	err = h.meetingService.UpdateMeeting(c.Context(), id, req, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update meeting")
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to update meeting",
		})
	}

	return c.Status(fiber.StatusOK).JSON(models.SuccessResponse{
		Message: "Meeting updated successfully",
	})
}

func (h *MeetingHandler) DeleteMeeting(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid meeting ID",
		})
	}

	userID := c.Locals("userID").(uuid.UUID)
	err = h.meetingService.DeleteMeeting(c.Context(), id, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete meeting")
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to delete meeting",
		})
	}

	return c.Status(fiber.StatusOK).JSON(models.SuccessResponse{
		Message: "Meeting deleted successfully",
	})
}

func (h *MeetingHandler) AddParticipant(c *fiber.Ctx) error {
	meetingID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid meeting ID",
		})
	}

	userID, err := uuid.Parse(c.Params("user_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid user ID",
		})
	}

	err = h.meetingService.AddParticipant(c.Context(), meetingID, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to add participant")
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to add participant",
		})
	}

	return c.Status(fiber.StatusOK).JSON(models.SuccessResponse{
		Message: "Participant added successfully",
	})
}

func (h *MeetingHandler) RemoveParticipant(c *fiber.Ctx) error {
	meetingID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid meeting ID",
		})
	}

	userID, err := uuid.Parse(c.Params("user_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid user ID",
		})
	}

	err = h.meetingService.RemoveParticipant(c.Context(), meetingID, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to remove participant")
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to remove participant",
		})
	}

	return c.Status(fiber.StatusOK).JSON(models.SuccessResponse{
		Message: "Participant removed successfully",
	})
}
