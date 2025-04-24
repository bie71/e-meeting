package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// SeedUsers creates initial user data in the database
func SeedUsers(db *sql.DB) error {
	// Default users to be seeded
	defaultUsers := []struct {
		username string
		email    string
		password string
		role     string
		status   bool
	}{
		{
			username: "admin",
			email:    "admin@example.com",
			password: "admin123",
			role:     "admin",
			status:   true,
		},
		{
			username: "user1",
			email:    "user1@example.com",
			password: "password123",
			role:     "user",
			status:   true,
		},
	}

	if db == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Delete related records first
	_, err = tx.Exec("DELETE FROM password_reset_tokens")
	if err != nil {
		return fmt.Errorf("failed to clean password_reset_tokens table: %v", err)
	}

	// Then delete users
	_, err = tx.Exec("DELETE FROM users")
	if err != nil {
		return fmt.Errorf("failed to clean users table: %v", err)
	}

	// Insert users
	for _, user := range defaultUsers {
		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %v", err)
		}

		// Insert user
		_, err = tx.Exec(`
			INSERT INTO users (
				id, 
				username, 
				email, 
				password, 
				role, 
				status, 
				created_at, 
				updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $7)`,
			uuid.New(),
			user.username,
			user.email,
			hashedPassword,
			user.role,
			user.status,
			time.Now(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert user %s: %v", user.username, err)
		}
		log.Printf("Seeded user: %s with role: %s", user.username, user.role)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	log.Println("Successfully seeded users table")
	return nil
}
