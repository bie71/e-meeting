package models

import (
	"time"

	"github.com/google/uuid"
)

type Room struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name" validate:"required"`
	Capacity     int       `json:"capacity" validate:"required,min=1"`
	PricePerHour float64   `json:"price_per_hour" validate:"required,min=0"`
	Status       string    `json:"status" validate:"required,oneof=active inactive"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type RoomFilter struct {
	Search      *string    `json:"search,omitempty"` // Search by name
	RoomTypeID  *uuid.UUID `json:"room_type_id,omitempty"`
	MinCapacity *int       `json:"min_capacity,omitempty"`
	MaxCapacity *int       `json:"max_capacity,omitempty"`
	Status      *string    `json:"status,omitempty"` // active, inactive
}

type PaginationQuery struct {
	Page     int `form:"page,default=1" validate:"min=1"`
	PageSize int `form:"page_size,default=10" validate:"min=1,max=100"`
}

type RoomListResponse struct {
	Rooms      []Room `json:"rooms"`
	TotalCount int    `json:"total_count"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	TotalPages int    `json:"total_pages"`
}

type CreateRoomRequest struct {
	Name         string  `json:"name" validate:"required"`
	Capacity     int     `json:"capacity" validate:"required,min=1"`
	PricePerHour float64 `json:"price_per_hour" validate:"required,min=0"`
	Status       string  `json:"status" validate:"required,oneof=active inactive"`
}

type UpdateRoomRequest struct {
	Name         *string  `json:"name,omitempty"`
	Capacity     *int     `json:"capacity,omitempty" validate:"omitempty,min=1"`
	PricePerHour *float64 `json:"price_per_hour,omitempty" validate:"omitempty,min=0"`
	Status       *string  `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
}

type RoomScheduleQuery struct {
	StartDateTime time.Time `query:"start_datetime" validate:"required" time_format:"2006-01-02T15:04:05Z07:00"`
	EndDateTime   time.Time `query:"end_datetime" validate:"required,gtfield=StartDateTime" time_format:"2006-01-02T15:04:05Z07:00"`
}

type RoomScheduleBlock struct {
	ReservationID uuid.UUID `json:"reservation_id"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Status        string    `json:"status"`
	VisitorCount  int       `json:"visitor_count"`
}

type RoomScheduleResponse struct {
	RoomID    uuid.UUID           `json:"room_id"`
	Schedules []RoomScheduleBlock `json:"schedules"`
	StartTime time.Time           `json:"start_time"`
	EndTime   time.Time           `json:"end_time"`
}
