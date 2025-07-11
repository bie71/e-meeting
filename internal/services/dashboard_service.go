package services

import (
	"database/sql"
	"e_metting/internal/models"
	"fmt"
	"math"
	"time"
)

type DashboardService struct {
	db *sql.DB
}

func NewDashboardService(db *sql.DB) *DashboardService {
	return &DashboardService{
		db: db,
	}
}

func (s *DashboardService) GetDashboardStats(query *models.DashboardQuery) (*models.DashboardResponse, error) {
	// Parse dates

	endDate := time.Now().Local()
	startDate := endDate.AddDate(0, 0, -30)
	var err error

	if query.StartDate != "" {
		startDate, err = time.Parse("2006-01-02", query.StartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date required format (YYYY-MM-DD): %v", err)
		}
	}

	if query.EndDate != "" {
		endDate, err = time.Parse("2006-01-02", query.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date required format (YYYY-MM-DD): %v", err)
		}
		// Tambahkan waktu ke akhir hari
		endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	}

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Get total statistics
	var totalOmzet float64
	var totalReservations, totalVisitors, totalRooms int

	err = tx.QueryRow(`
		SELECT 
			COALESCE(SUM(r.price), 0) as total_omzet,
			COUNT(DISTINCT r.id) as total_reservations,
			COALESCE(SUM(r.visitor_count), 0) as total_visitors,
			(SELECT COUNT(*) FROM rooms) as total_rooms
		FROM rooms rm
		LEFT JOIN reservations r ON r.room_id = rm.id
		WHERE (r.start_time >= $1 AND r.end_time <= $2 AND r.status IN ('confirmed', 'completed')) OR r.id IS NULL`,
		startDate, endDate,
	).Scan(&totalOmzet, &totalReservations, &totalVisitors, &totalRooms)

	if err != nil {
		fmt.Println(err, " error getting total statistics")
		return nil, fmt.Errorf("error getting total statistics: %v", err)
	}

	// Get per-room statistics
	rows, err := tx.Query(`
		WITH room_bookings AS (
				SELECT 
					rm.id as room_id,
					rm.name as room_name,
					COUNT(r.id) as total_bookings,
					COALESCE(SUM(EXTRACT(EPOCH FROM (r.end_time - r.start_time)) / 3600), 0) as total_hours,
					COALESCE(SUM(r.price), 0) as revenue
				FROM rooms rm
				LEFT JOIN reservations r 
					ON r.room_id = rm.id AND r.status IN ('confirmed', 'completed')
					AND r.start_time >= $1 AND r.end_time <= $2
				GROUP BY rm.id, rm.name
			)

		SELECT 
			room_id,
			room_name,
			total_bookings,
			total_hours,
			CASE 
				WHEN $3 = 0 THEN 0
				ELSE (total_hours / ($3 * 24) * 100)
			END as occupancy_rate,
			revenue
		FROM room_bookings
		ORDER BY revenue DESC`,
		startDate, endDate,
		endDate.Sub(startDate).Hours()/24, // Total days in period
	)
	if err != nil {
		fmt.Println(err, " error getting room statistics 2")
		return nil, fmt.Errorf("error getting room statistics: %v", err)
	}
	defer rows.Close()

	var roomStats []models.RoomStats
	for rows.Next() {
		var stat models.RoomStats
		err := rows.Scan(
			&stat.RoomID,
			&stat.RoomName,
			&stat.TotalBookings,
			&stat.TotalHours,
			&stat.Occupancy,
			&stat.Revenue,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning room statistics: %v", err)
		}
		stat.Occupancy = math.Ceil(stat.Occupancy*100) / 100

		roomStats = append(roomStats, stat)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating room statistics: %v", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	return &models.DashboardResponse{
		StartDate:    startDate,
		EndDate:      endDate,
		TotalOmzet:   totalOmzet,
		Reservations: totalReservations,
		Visitors:     totalVisitors,
		TotalRooms:   totalRooms,
		RoomStats:    roomStats,
	}, nil
}
