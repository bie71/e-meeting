package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Meeting struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Title        string         `gorm:"size:100;not null" json:"title" validate:"required"`
	Description  string         `gorm:"type:text" json:"description" validate:"required"`
	StartTime    time.Time      `gorm:"not null" json:"start_time" validate:"required"`
	EndTime      time.Time      `gorm:"not null" json:"end_time" validate:"required"`
	Location     string         `gorm:"size:255;not null" json:"location" validate:"required"`
	CreatedBy    uuid.UUID      `gorm:"type:uuid;not null" json:"created_by"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Participants []User         `gorm:"many2many:meeting_participants;" json:"participants,omitempty"`
}

type MeetingRequest struct {
	Title       string    `json:"title" validate:"required"`
	Description string    `json:"description" validate:"required"`
	StartTime   time.Time `json:"start_time" validate:"required"`
	EndTime     time.Time `json:"end_time" validate:"required"`
	Location    string    `json:"location" validate:"required"`
}

type MeetingResponse struct {
	ID           uuid.UUID `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Location     string    `json:"location"`
	CreatedBy    uuid.UUID `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Participants []User    `json:"participants,omitempty"`
}
