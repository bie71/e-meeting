package models

import "time"

type RoomStats struct {
	RoomID        string  `json:"room_id"`
	RoomName      string  `json:"room_name"`
	TotalBookings int     `json:"total_bookings"`
	TotalHours    float64 `json:"total_hours"`
	Occupancy     float64 `json:"occupancy_rate"` // Percentage of time room was occupied
	Revenue       float64 `json:"revenue"`
}

type DashboardResponse struct {
	StartDate    time.Time   `json:"start_date"`
	EndDate      time.Time   `json:"end_date"`
	TotalOmzet   float64     `json:"total_omzet"`
	Reservations int         `json:"total_reservations"`
	Visitors     int         `json:"total_visitors"`
	TotalRooms   int         `json:"total_rooms"`
	RoomStats    []RoomStats `json:"room_stats"`
}

type DashboardQuery struct {
	StartDate string `query:"start_date"` // Format: YYYY-MM-DD
	EndDate   string `query:"end_date"`   // Format: YYYY-MM-DD
}
