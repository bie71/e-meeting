package services

import (
	"context"
	"errors"
	"time"

	"e_metting/internal/models"
	"e_metting/internal/repositories"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type MeetingService struct {
	meetingRepo repositories.MeetingRepository
}

func NewMeetingService(meetingRepo repositories.MeetingRepository) *MeetingService {
	return &MeetingService{
		meetingRepo: meetingRepo,
	}
}

func (s *MeetingService) CreateMeeting(ctx context.Context, req models.MeetingRequest, userID uuid.UUID) (*models.Meeting, error) {
	meeting := &models.Meeting{
		Title:       req.Title,
		Description: req.Description,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Location:    req.Location,
		CreatedBy:   userID,
	}

	err := s.meetingRepo.CreateMeeting(ctx, meeting)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create meeting")
		return nil, err
	}

	return meeting, nil
}

func (s *MeetingService) GetMeeting(ctx context.Context, id uuid.UUID) (*models.Meeting, error) {
	meeting, err := s.meetingRepo.GetMeeting(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get meeting")
		return nil, err
	}

	return meeting, nil
}

func (s *MeetingService) ListMeetings(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) ([]*models.Meeting, error) {
	meetings, err := s.meetingRepo.ListMeetings(ctx, userID, startTime, endTime)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list meetings")
		return nil, err
	}

	return meetings, nil
}

func (s *MeetingService) UpdateMeeting(ctx context.Context, id uuid.UUID, req models.MeetingRequest, userID uuid.UUID) error {
	meeting, err := s.meetingRepo.GetMeeting(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get meeting for update")
		return err
	}

	if meeting.CreatedBy != userID {
		return errors.New("unauthorized to update meeting")
	}

	meeting.Title = req.Title
	meeting.Description = req.Description
	meeting.StartTime = req.StartTime
	meeting.EndTime = req.EndTime
	meeting.Location = req.Location

	err = s.meetingRepo.UpdateMeeting(ctx, meeting)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update meeting")
		return err
	}

	return nil
}

func (s *MeetingService) DeleteMeeting(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	meeting, err := s.meetingRepo.GetMeeting(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get meeting for delete")
		return err
	}

	if meeting.CreatedBy != userID {
		return errors.New("unauthorized to delete meeting")
	}

	err = s.meetingRepo.DeleteMeeting(ctx, id, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete meeting")
		return err
	}

	return nil
}

func (s *MeetingService) AddParticipant(ctx context.Context, meetingID, userID uuid.UUID) error {
	err := s.meetingRepo.AddParticipant(ctx, meetingID, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to add participant")
		return err
	}

	return nil
}

func (s *MeetingService) RemoveParticipant(ctx context.Context, meetingID, userID uuid.UUID) error {
	err := s.meetingRepo.RemoveParticipant(ctx, meetingID, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to remove participant")
		return err
	}

	return nil
}
