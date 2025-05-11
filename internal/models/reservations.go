package models

import (
	"time"

	"github.com/google/uuid"
)

type RoomInfo struct {
	Capacity     int     `json:"capacity"`
	PricePerHour float64 `json:"price_per_hour"`
}

type ReservationEvent struct {
	ID            uuid.UUID `json:"id"`
	RoomID        uuid.UUID `json:"room_id"`
	RoomName      string    `json:"room_name"`
	RoomDetails   RoomInfo  `json:"room_details"`
	UserID        uuid.UUID `json:"user_id"`
	Username      string    `json:"username"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	DurationHours float64   `json:"duration_hours"`
	VisitorCount  int       `json:"visitor_count"`
	Price         float64   `json:"price"`
	Status        string    `json:"status"`
}

type ReservationHistoryQuery struct {
	StartDatetime string    `query:"start_datetime"`
	EndDatetime   string    `query:"end_datetime"`
	IsAdmin       bool      `json:"-"`
	UserID        uuid.UUID `json:"-"`
}

type ReservationHistoryResponse struct {
	StartDatetime time.Time          `json:"start_datetime"`
	EndDatetime   time.Time          `json:"end_datetime"`
	Events        []ReservationEvent `json:"events"`
}

type ReservationStatus string

const (
	ReservationStatusPending   ReservationStatus = "pending"
	ReservationStatusConfirmed ReservationStatus = "confirmed"
	ReservationStatusCancelled ReservationStatus = "cancelled"
	ReservationStatusCompleted ReservationStatus = "completed"
)

func (s ReservationStatus) IsValid() bool {
	switch s {
	case ReservationStatusPending, ReservationStatusConfirmed,
		ReservationStatusCancelled, ReservationStatusCompleted:
		return true
	}
	return false
}

type UpdateReservationStatusRequest struct {
	ReservationID uuid.UUID         `json:"reservation_id" validate:"required"`
	Status        ReservationStatus `json:"status" validate:"required"`
}

type ReservationCalculationRequest struct {
	RoomID uuid.UUID `json:"room_id" validate:"required"`
	Snacks []struct {
		SnackID  uuid.UUID `json:"snack_id" validate:"required"`
		Quantity int       `json:"quantity" validate:"required,min=1"`
	} `json:"snacks" validate:"required"`
	StartTime time.Time `json:"start_time" validate:"required"`
	EndTime   time.Time `json:"end_time" validate:"required"`
}

type ReservationCalculationResponse struct {
	Room struct {
		ID           uuid.UUID `json:"id"`
		Name         string    `json:"name"`
		PricePerHour float64   `json:"price_per_hour"`
		TotalHours   float64   `json:"total_hours"`
		TotalCost    float64   `json:"total_cost"`
	} `json:"room"`
	Snacks []struct {
		ID       uuid.UUID `json:"id"`
		Name     string    `json:"name"`
		Category string    `json:"category"`
		Price    float64   `json:"price"`
		Quantity int       `json:"quantity"`
		Subtotal float64   `json:"subtotal"`
	} `json:"snacks"`
	TotalCost float64 `json:"total_cost"`
}

type CreateReservationRequest struct {
	RoomID       uuid.UUID `json:"room_id" validate:"required"`
	UserID       uuid.UUID `json:"user_id" validate:"required"`
	StartTime    time.Time `json:"start_time" validate:"required"`
	EndTime      time.Time `json:"end_time" validate:"required,gtfield=StartTime"`
	VisitorCount int       `json:"visitor_count" validate:"required,min=1"`
	Snacks       []struct {
		SnackID  uuid.UUID `json:"snack_id" validate:"required"`
		Quantity int       `json:"quantity" validate:"required,min=1"`
	} `json:"snacks" validate:"required,dive"`
}

type CreateReservationResponse struct {
	ReservationID uuid.UUID `json:"reservation_id"`
	Status        string    `json:"status"`
	TotalCost     float64   `json:"total_cost"`
	CreatedAt     time.Time `json:"created_at"`
}
