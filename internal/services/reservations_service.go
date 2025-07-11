package services

import (
	"database/sql"
	"e_metting/internal/models"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
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
	endDatetime := time.Now().Local()
	startDatetime := endDatetime.AddDate(0, 0, -7)
	var err error

	if query != nil {
		if query.StartDatetime != "" {
			startDatetime, err = time.Parse("2006-01-02 15:04", query.StartDatetime)
			if err != nil {
				return nil, fmt.Errorf("invalid start_datetime format (YYYY-MM-DD HH:mm): %v", err)
			}
		}

		if query.EndDatetime != "" {
			endDatetime, err = time.Parse("2006-01-02 15:04", query.EndDatetime)
			if err != nil {
				return nil, fmt.Errorf("invalid end_datetime format (YYYY-MM-DD HH:mm): %v", err)
			}
		}
	}

	log.Printf("Querying history from %s to %s", startDatetime, endDatetime)

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Query string & args dinamis
	queryStr := `
		SELECT 
			r.id,
			r.room_id,
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
		JOIN rooms rm ON r.room_id = rm.id
		JOIN users u ON r.user_id = u.id
		WHERE r.start_time <= $2 AND r.end_time >= $1`
	args := []interface{}{startDatetime, endDatetime}

	// Tambahkan filter user_id jika bukan admin
	if query != nil && !query.IsAdmin {
		queryStr += " AND r.user_id = $3"
		args = append(args, query.UserID)
	}

	queryStr += " ORDER BY r.start_time ASC, rm.name ASC"

	rows, err := tx.Query(queryStr, args...)
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

		event.RoomDetails = models.RoomInfo{
			Capacity:     roomCapacity,
			PricePerHour: pricePerHour,
		}
		event.DurationHours = event.EndTime.Sub(event.StartTime).Hours()
		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reservations: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	if len(events) == 0 {
		log.Println("No reservation found in range", startDatetime, endDatetime)
		if query != nil && !query.IsAdmin {
			log.Println("Filtered by user_id:", query.UserID)
		}
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
		UPDATE reservations
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
			r.room_id,
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
		JOIN rooms rm ON r.room_id = rm.id
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

func (s *ReservationService) CalculateReservationCost(req *models.ReservationCalculationRequest) (*models.ReservationCalculationResponse, error) {

	log.Printf("req.StartTime: %v", req.StartTime)
	log.Printf("req.EndTime: %v", req.EndTime)

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Get room details
	var room struct {
		ID           uuid.UUID
		Name         string
		PricePerHour float64
	}
	err = tx.QueryRow(`
		SELECT id, name, price_per_hour
		FROM rooms
		WHERE id = $1 AND status = 'active'
	`, req.RoomID).Scan(&room.ID, &room.Name, &room.PricePerHour)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("room not found or inactive")
		}
		return nil, fmt.Errorf("error querying room: %v", err)
	}

	// Calculate room cost
	duration := req.EndTime.Sub(req.StartTime)
	hours := duration.Hours()
	roomCost := room.PricePerHour * hours

	// Get snack details and calculate costs
	var snackIDs []uuid.UUID
	for _, snack := range req.Snacks {
		snackIDs = append(snackIDs, snack.SnackID)
	}

	rows, err := tx.Query(`
		SELECT id, name, category, price
		FROM snacks
		WHERE id = ANY($1)
	`, pq.Array(snackIDs))
	if err != nil {
		return nil, fmt.Errorf("error querying snacks: %v", err)
	}
	defer rows.Close()

	var snacks []struct {
		ID       uuid.UUID
		Name     string
		Category string
		Price    float64
		Quantity int
	}

	for rows.Next() {
		var snack struct {
			ID       uuid.UUID
			Name     string
			Category string
			Price    float64
		}
		err := rows.Scan(&snack.ID, &snack.Name, &snack.Category, &snack.Price)
		if err != nil {
			return nil, fmt.Errorf("error scanning snack: %v", err)
		}

		// Find quantity for this snack
		for _, reqSnack := range req.Snacks {
			if reqSnack.SnackID == snack.ID {
				snacks = append(snacks, struct {
					ID       uuid.UUID
					Name     string
					Category string
					Price    float64
					Quantity int
				}{
					ID:       snack.ID,
					Name:     snack.Name,
					Category: snack.Category,
					Price:    snack.Price,
					Quantity: reqSnack.Quantity,
				})
				break
			}
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating snacks: %v", err)
	}

	// Calculate total cost
	response := &models.ReservationCalculationResponse{
		Room: struct {
			ID           uuid.UUID `json:"id"`
			Name         string    `json:"name"`
			PricePerHour float64   `json:"price_per_hour"`
			TotalHours   float64   `json:"total_hours"`
			TotalCost    float64   `json:"total_cost"`
		}{
			ID:           room.ID,
			Name:         room.Name,
			PricePerHour: room.PricePerHour,
			TotalHours:   hours,
			TotalCost:    roomCost,
		},
		TotalCost: roomCost,
	}

	// Calculate snack costs
	for _, snack := range snacks {
		subtotal := snack.Price * float64(snack.Quantity)
		response.Snacks = append(response.Snacks, struct {
			ID       uuid.UUID `json:"id"`
			Name     string    `json:"name"`
			Category string    `json:"category"`
			Price    float64   `json:"price"`
			Quantity int       `json:"quantity"`
			Subtotal float64   `json:"subtotal"`
		}{
			ID:       snack.ID,
			Name:     snack.Name,
			Category: snack.Category,
			Price:    snack.Price,
			Quantity: snack.Quantity,
			Subtotal: subtotal,
		})
		response.TotalCost += subtotal
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	return response, nil
}

func (s *ReservationService) GetReservationByID(id uuid.UUID) (*models.ReservationDetailResponse, error) {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Get reservation details with room and user information
	var reservation models.ReservationDetailResponse
	var createdAt, updatedAt time.Time

	err = tx.QueryRow(`
		SELECT 
			r.id, r.status, r.start_time, r.end_time, r.visitor_count, r.price, r.created_at, r.updated_at,
			rm.id, rm.name, rm.capacity, rm.price_per_hour,
			u.id, u.username
		FROM reservations r
		JOIN rooms rm ON r.room_id = rm.id
		JOIN users u ON r.user_id = u.id
		WHERE r.id = $1
	`, id).Scan(
		&reservation.ID, &reservation.Status, &reservation.StartTime, &reservation.EndTime,
		&reservation.VisitorCount, &reservation.Price, &createdAt, &updatedAt,
		&reservation.Room.ID, &reservation.Room.Name, &reservation.Room.Capacity, &reservation.Room.PricePerHour,
		&reservation.User.ID, &reservation.User.Username,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("reservation not found")
		}
		return nil, fmt.Errorf("error fetching reservation: %v", err)
	}

	reservation.CreatedAt = createdAt
	reservation.UpdatedAt = updatedAt

	// Get snacks for this reservation
	rows, err := tx.Query(`
		SELECT 
			s.id, s.name, s.category, rs.price, rs.quantity
		FROM reservation_snacks rs
		JOIN snacks s ON rs.snack_id = s.id
		WHERE rs.reservation_id = $1
	`, id)

	if err != nil {
		return nil, fmt.Errorf("error fetching reservation snacks: %v", err)
	}
	defer rows.Close()

	var totalSnackCost float64
	for rows.Next() {
		var snack struct {
			ID       uuid.UUID
			Name     string
			Category string
			Price    float64
			Quantity int
		}

		err := rows.Scan(&snack.ID, &snack.Name, &snack.Category, &snack.Price, &snack.Quantity)
		if err != nil {
			return nil, fmt.Errorf("error scanning snack: %v", err)
		}

		subtotal := snack.Price * float64(snack.Quantity)
		totalSnackCost += subtotal

		reservation.Snacks = append(reservation.Snacks, struct {
			ID       uuid.UUID `json:"id"`
			Name     string    `json:"name"`
			Category string    `json:"category"`
			Price    float64   `json:"price"`
			Quantity int       `json:"quantity"`
			Subtotal float64   `json:"subtotal"`
		}{
			ID:       snack.ID,
			Name:     snack.Name,
			Category: snack.Category,
			Price:    snack.Price,
			Quantity: snack.Quantity,
			Subtotal: subtotal,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating snacks: %v", err)
	}

	// Calculate total cost (room cost + snack cost)
	reservation.TotalCost = reservation.Price

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	return &reservation, nil
}

func (s *ReservationService) CreateReservation(req *models.CreateReservationRequest) (*models.CreateReservationResponse, error) {
	now := time.Now()

	log.Println("time now: ", now)
	log.Println("req.StartTime: ", req.StartTime)
	// Ensure start time is in the future
	if req.StartTime.Before(now) {
		return nil, fmt.Errorf("reservation start time must be in the future")
	}

	// Ensure end time is after start time
	if !req.EndTime.After(req.StartTime) {
		return nil, fmt.Errorf("reservation end time must be after start time")
	}

	log.Println("req.StartTime ", req.StartTime)
	log.Println("req.EndTime ", req.EndTime)

	// Validate minimum and maximum duration
	duration := req.EndTime.Sub(req.StartTime)
	if duration < 30*time.Minute {
		return nil, fmt.Errorf("reservation must be at least 30 minutes long")
	}
	if duration > 24*time.Hour {
		return nil, fmt.Errorf("reservation cannot exceed 24 hours")
	}

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Check room availability
	var roomCapacity int
	var pricePerHour float64
	err = tx.QueryRow(`
		SELECT capacity, price_per_hour
		FROM rooms
		WHERE id = $1 AND status = 'active'
	`, req.RoomID).Scan(&roomCapacity, &pricePerHour)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("room not found or inactive")
		}
		return nil, fmt.Errorf("error checking room: %v", err)
	}

	// Validate visitor count against room capacity
	if req.VisitorCount > roomCapacity {
		return nil, fmt.Errorf("visitor count exceeds room capacity of %d", roomCapacity)
	}

	// Check for overlapping reservations
	var overlappingCount int
	err = tx.QueryRow(`
		SELECT COUNT(*)
		FROM reservations
		WHERE room_id = $1
		AND status != 'cancelled'
		AND (
			(start_time <= $2 AND end_time > $2)
			OR (start_time < $3 AND end_time >= $3)
			OR (start_time >= $2 AND end_time <= $3)
		)
	`, req.RoomID, req.StartTime, req.EndTime).Scan(&overlappingCount)
	if err != nil {
		return nil, fmt.Errorf("error checking overlapping reservations: %v", err)
	}
	if overlappingCount > 0 {
		return nil, fmt.Errorf("room is already booked for the selected time period")
	}

	// Calculate room cost
	bookingDuration := req.EndTime.Sub(req.StartTime)
	hours := bookingDuration.Hours()
	roomCost := pricePerHour * hours

	// Get snack details and calculate costs
	var snackIDs []uuid.UUID
	for _, snack := range req.Snacks {
		snackIDs = append(snackIDs, snack.SnackID)
	}

	rows, err := tx.Query(`
		SELECT id, name, price
		FROM snacks
		WHERE id = ANY($1)
	`, pq.Array(snackIDs))
	if err != nil {
		return nil, fmt.Errorf("error querying snacks: %v", err)
	}
	defer rows.Close()

	var snacks []struct {
		ID       uuid.UUID
		Name     string
		Price    float64
		Quantity int
	}
	var totalSnackCost float64

	for rows.Next() {
		var snack struct {
			ID    uuid.UUID
			Name  string
			Price float64
		}
		err := rows.Scan(&snack.ID, &snack.Name, &snack.Price)
		if err != nil {
			return nil, fmt.Errorf("error scanning snack: %v", err)
		}

		// Find quantity for this snack
		for _, reqSnack := range req.Snacks {
			if reqSnack.SnackID == snack.ID {
				subtotal := snack.Price * float64(reqSnack.Quantity)
				totalSnackCost += subtotal
				snacks = append(snacks, struct {
					ID       uuid.UUID
					Name     string
					Price    float64
					Quantity int
				}{
					ID:       snack.ID,
					Name:     snack.Name,
					Price:    snack.Price,
					Quantity: reqSnack.Quantity,
				})
				break
			}
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating snacks: %v", err)
	}

	// Calculate total cost
	totalCost := roomCost + totalSnackCost

	// Create reservation
	var reservationID uuid.UUID
	err = tx.QueryRow(`
		INSERT INTO reservations (
			room_id, user_id, start_time, end_time, visitor_count, price, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, req.RoomID, req.UserID, req.StartTime, req.EndTime, req.VisitorCount, totalCost, "pending").Scan(&reservationID)
	if err != nil {
		return nil, fmt.Errorf("error creating reservation: %v", err)
	}

	// Create snack orders
	for _, snack := range snacks {
		_, err = tx.Exec(`
			INSERT INTO reservation_snacks (
				reservation_id, snack_id, quantity, price
			) VALUES ($1, $2, $3, $4)
		`, reservationID, snack.ID, snack.Quantity, snack.Price)
		if err != nil {
			return nil, fmt.Errorf("error creating snack order: %v", err)
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	return &models.CreateReservationResponse{
		ReservationID: reservationID,
		Status:        "pending",
		TotalCost:     totalCost,
		CreatedAt:     time.Now().Local(),
	}, nil
}

func (s *ReservationService) DeleteReservation(id uuid.UUID) error {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Delete reservation
	result, err := tx.Exec(`DELETE FROM reservations WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("error deleting reservation: %v", err)
	}
	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("reservation not found")
	}

	return nil
}
