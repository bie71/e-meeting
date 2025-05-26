package services

import (
	"database/sql"
	"e_metting/internal/models"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

type SnackService struct {
	db *sql.DB
}

func NewSnackService(db *sql.DB) *SnackService {
	return &SnackService{
		db: db,
	}
}
func (s *SnackService) GetSnacks(filter *models.SnackFilter, pagination *models.PaginationQuery) (*models.SnackListResponse, error) {
	if pagination.Page < 1 {
		pagination.Page = 1
	}

	if pagination.PageSize < 1 {
		pagination.PageSize = 10
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Build dynamic filter
	var conditions []string
	var args []interface{}
	argCount := 1

	if filter != nil {
		if filter.Search != nil && *filter.Search != "" {
			conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argCount))
			args = append(args, "%"+*filter.Search+"%")
			argCount++
		}
		if filter.Category != nil && *filter.Category != "" {
			conditions = append(conditions, fmt.Sprintf("category = $%d", argCount))
			args = append(args, *filter.Category)
			argCount++
		}
		if filter.MinPrice != nil {
			conditions = append(conditions, fmt.Sprintf("price >= $%d", argCount))
			args = append(args, *filter.MinPrice)
			argCount++
		}
		if filter.MaxPrice != nil {
			conditions = append(conditions, fmt.Sprintf("price <= $%d", argCount))
			args = append(args, *filter.MaxPrice)
			argCount++
		}
	}

	countQuery := "SELECT COUNT(*) FROM snacks"
	if len(conditions) > 0 {
		countQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	var totalCount int
	err = tx.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("error getting total count: %v", err)
	}

	// Tambah limit dan offset untuk pagination
	offset := (pagination.Page - 1) * pagination.PageSize
	query := `
		SELECT id, name, category, price, created_at, updated_at
		FROM snacks
	`
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += fmt.Sprintf(" ORDER BY category, name LIMIT $%d OFFSET $%d", argCount, argCount+1)

	// Tambah param limit dan offset
	args = append(args, pagination.PageSize, offset)

	log.Printf("Query: %s\nArgs: %v", query, args)

	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying snacks: %v", err)
	}
	defer rows.Close()

	var snacks []models.Snack
	for rows.Next() {
		var snack models.Snack
		err := rows.Scan(
			&snack.ID,
			&snack.Name,
			&snack.Category,
			&snack.Price,
			&snack.CreatedAt,
			&snack.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning snack: %v", err)
		}
		snacks = append(snacks, snack)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating snacks: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	if len(snacks) == 0 {
		return nil, fmt.Errorf("snacks not found")
	}

	return &models.SnackListResponse{
		Snacks:     snacks,
		TotalCount: totalCount,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: (totalCount + pagination.PageSize - 1) / pagination.PageSize,
	}, nil
}

func (s *SnackService) CreateSnack(req *models.CreateSnackRequest) (*models.CreateSnackResponse, error) {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Generate new UUID for the snack
	snackID := uuid.New()
	createdAt := time.Now()

	// Insert new snack
	_, err = tx.Exec(`
		INSERT INTO snacks (id, name, category, price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
	`, snackID, req.Name, req.Category, req.Price, createdAt)

	if err != nil {
		return nil, fmt.Errorf("error creating snack: %v", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	return &models.CreateSnackResponse{
		ID:        snackID,
		Name:      req.Name,
		Category:  req.Category,
		Price:     req.Price,
		CreatedAt: createdAt,
	}, nil
}

func (s *SnackService) DeleteSnack(id uuid.UUID) error {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Delete snack
	result, err := tx.Exec(`DELETE FROM snacks WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("error deleting snack: %v", err)
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
		return fmt.Errorf("snack not found")
	}

	return nil
}

func (s *SnackService) UpdateSnack(id uuid.UUID, req *models.Snack) error {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// First, check if snack exists
	var exists bool
	err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM snacks WHERE id = $1)`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error checking snack existence: %v", err)
	}
	if !exists {
		return fmt.Errorf("snack not found")
	}

	// Update snack
	_, err = tx.Exec(`
		UPDATE snacks
		SET name = $2, category = $3, price = $4, updated_at = NOW()
		WHERE id = $1
	`, id, req.Name, req.Category, req.Price)
	if err != nil {
		return fmt.Errorf("error updating snack: %v", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}

func (s *SnackService) GetSnackByID(id uuid.UUID) (*models.Snack, error) {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Get snack details
	var snack models.Snack
	err = tx.QueryRow(`
		SELECT id, name, category, price, created_at, updated_at
		FROM snacks
		WHERE id = $1
	`, id).Scan(
		&snack.ID, &snack.Name, &snack.Category, &snack.Price, &snack.CreatedAt, &snack.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error querying snack: %v", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	if snack.ID == uuid.Nil {
		return nil, fmt.Errorf("snack not found")
	}

	return &snack, nil
}
