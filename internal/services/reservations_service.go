package services

import (
	"database/sql"
	"e_metting/internal/models"
	"fmt"
	"time"
)

type ReservationService struct {
	db *sql.DB
}

func NewReservationService(db *sql.DB) *ReservationService {
	return &ReservationService{
		db: db,
	}
}

func (s *ReservationService) GetReservationHistory(query *models.ReservationHistoryQuery) (*models.ReservationHistoryResponse, error) {
	// Parse dates with default values (last 7 days if not specified)
	endDatetime := time.Now()
	startDatetime := endDatetime.AddDate(0, 0, -7)
	var err error

	// Parse provided dates if they exist
	if query != nil {
		if query.StartDatetime != "" {
			startDatetime, err = time.Parse("2006-01-02 15:04:05", query.StartDatetime)
			if err != nil {
				return nil, fmt.Errorf("invalid start_datetime format (required: YYYY-MM-DD HH:mm:ss): %v", err)
			}
		}

		if query.EndDatetime != "" {
			endDatetime, err = time.Parse("2006-01-02 15:04:05", query.EndDatetime)
			if err != nil {
				return nil, fmt.Errorf("invalid end_datetime format (required: YYYY-MM-DD HH:mm:ss): %v", err)
			}
		}
	}

	// Validate date range
	if endDatetime.Before(startDatetime) {
		return nil, fmt.Errorf("end_datetime cannot be before start_datetime")
	}

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Query reservations with room and user details
	rows, err := tx.Query(`
		SELECT 
			r.id,
			r.room_new_id,
			rm.name as room_name,
			r.user_id,
			u.username,
			r.start_time,
			r.end_time,
			r.visitor_count,
			r.price,
			r.status,
			rm.capacity,
			rm.price_per_hour
		FROM reservations_new r
		JOIN rooms_new rm ON r.room_new_id = rm.id
		JOIN users u ON r.user_id = u.id
		WHERE r.start_time >= $1 
		AND r.end_time <= $2
		ORDER BY r.start_time ASC, rm.name ASC`,
		startDatetime,
		endDatetime,
	)
	if err != nil {
		return nil, fmt.Errorf("error querying reservations: %v", err)
	}
	defer rows.Close()

	var events []models.ReservationEvent
	for rows.Next() {
		var event models.ReservationEvent
		var roomCapacity int
		var pricePerHour float64

		err := rows.Scan(
			&event.ID,
			&event.RoomID,
			&event.RoomName,
			&event.UserID,
			&event.Username,
			&event.StartTime,
			&event.EndTime,
			&event.VisitorCount,
			&event.Price,
			&event.Status,
			&roomCapacity,
			&pricePerHour,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning reservation: %v", err)
		}

		// Add room details to event
		event.RoomDetails = models.RoomInfo{
			Capacity:     roomCapacity,
			PricePerHour: pricePerHour,
		}

		// Calculate duration in hours
		duration := event.EndTime.Sub(event.StartTime).Hours()
		event.DurationHours = duration

		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reservations: %v", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	return &models.ReservationHistoryResponse{
		StartDatetime: startDatetime,
		EndDatetime:   endDatetime,
		Events:        events,
	}, nil
}

func (s *ReservationService) UpdateReservationStatus(req *models.UpdateReservationStatusRequest) (*models.ReservationEvent, error) {
	// Validate status
	if !req.Status.IsValid() {
		return nil, fmt.Errorf("invalid status: must be one of pending, confirmed, cancelled, or completed")
	}

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Update reservation status
	result, err := tx.Exec(`
		UPDATE reservations_new
		SET status = $1, updated_at = NOW()
		WHERE id = $2`,
		req.Status,
		req.ReservationID,
	)
	if err != nil {
		return nil, fmt.Errorf("error updating reservation status: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("error getting rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return nil, fmt.Errorf("reservation not found with ID: %v", req.ReservationID)
	}

	// Fetch updated reservation with all details
	var event models.ReservationEvent
	var roomCapacity int
	var pricePerHour float64

	err = tx.QueryRow(`
		SELECT 
			r.id,
			r.room_new_id,
			rm.name as room_name,
			r.user_id,
			u.username,
			r.start_time,
			r.end_time,
			r.visitor_count,
			r.price,
			r.status,
			rm.capacity,
			rm.price_per_hour
		FROM reservations r
		JOIN rooms_new rm ON r.room_new_id = rm.id
		JOIN users u ON r.user_id = u.id
		WHERE r.id = $1`,
		req.ReservationID,
	).Scan(
		&event.ID,
		&event.RoomID,
		&event.RoomName,
		&event.UserID,
		&event.Username,
		&event.StartTime,
		&event.EndTime,
		&event.VisitorCount,
		&event.Price,
		&event.Status,
		&roomCapacity,
		&pricePerHour,
	)
	if err != nil {
		return nil, fmt.Errorf("error fetching updated reservation: %v", err)
	}

	// Add room details to event
	event.RoomDetails = models.RoomInfo{
		Capacity:     roomCapacity,
		PricePerHour: pricePerHour,
	}

	// Calculate duration in hours
	duration := event.EndTime.Sub(event.StartTime).Hours()
	event.DurationHours = duration

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	return &event, nil
}
