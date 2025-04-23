package repositories

import (
	"context"
	"e_metting/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MeetingRepository struct {
	db *gorm.DB
}

func NewMeetingRepository(db *gorm.DB) *MeetingRepository {
	return &MeetingRepository{
		db: db,
	}
}

func (r *MeetingRepository) CreateMeeting(ctx context.Context, meeting *models.Meeting) error {
	return r.db.WithContext(ctx).Create(meeting).Error
}

func (r *MeetingRepository) GetMeeting(ctx context.Context, id uuid.UUID) (*models.Meeting, error) {
	var meeting models.Meeting
	err := r.db.WithContext(ctx).
		Preload("Participants").
		First(&meeting, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &meeting, nil
}

func (r *MeetingRepository) ListMeetings(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) ([]*models.Meeting, error) {
	var meetings []*models.Meeting
	query := r.db.WithContext(ctx).
		Preload("Participants").
		Joins("JOIN meeting_participants ON meeting_participants.meeting_id = meetings.id").
		Where("meeting_participants.user_id = ?", userID)

	if !startTime.IsZero() {
		query = query.Where("meetings.start_time >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("meetings.end_time <= ?", endTime)
	}

	err := query.Order("meetings.start_time ASC").Find(&meetings).Error
	if err != nil {
		return nil, err
	}
	return meetings, nil
}

func (r *MeetingRepository) UpdateMeeting(ctx context.Context, meeting *models.Meeting) error {
	return r.db.WithContext(ctx).Model(meeting).Updates(map[string]interface{}{
		"title":       meeting.Title,
		"description": meeting.Description,
		"start_time":  meeting.StartTime,
		"end_time":    meeting.EndTime,
		"location":    meeting.Location,
	}).Error
}

func (r *MeetingRepository) DeleteMeeting(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND created_by = ?", id, userID).
		Delete(&models.Meeting{}).Error
}

func (r *MeetingRepository) AddParticipant(ctx context.Context, meetingID, userID uuid.UUID) error {
	meeting := &models.Meeting{ID: meetingID}
	user := &models.User{ID: userID}
	return r.db.WithContext(ctx).Model(meeting).Association("Participants").Append(user)
}

func (r *MeetingRepository) RemoveParticipant(ctx context.Context, meetingID, userID uuid.UUID) error {
	meeting := &models.Meeting{ID: meetingID}
	user := &models.User{ID: userID}
	return r.db.WithContext(ctx).Model(meeting).Association("Participants").Delete(user)
}
