package repositories

import (
	"context"
	"e_metting/internal/models"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error
	GetProfile(ctx context.Context, userID uuid.UUID) (*models.UserProfileResponse, error)
	UpdateProfile(ctx context.Context, userID string, req *models.UpdateProfileRequest) (*models.UserProfileResponse, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) CreateUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Where("username = ?", username).
		First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Where("email = ?", email).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("password", hashedPassword).Error
}

func (r *userRepository) GetProfile(ctx context.Context, userID uuid.UUID) (*models.UserProfileResponse, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Where("id = ?", userID).
		First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &models.UserProfileResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		ProfPic:   user.ProfPic,
		Language:  user.Language,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (r *userRepository) UpdateProfile(ctx context.Context, userID string, req *models.UpdateProfileRequest) (*models.UserProfileResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %v", err)
	}

	// Start transaction
	tx := r.db.Begin()
	defer tx.Rollback()

	// Check if username is already taken by another user
	var count int
	err = tx.Raw(`
		SELECT COUNT(*) 
		FROM users 
		WHERE username = $1 AND id != $2`,
		req.Username, id,
	).Scan(&count).Error
	if err != nil {
		return nil, fmt.Errorf("error checking username uniqueness: %v", err)
	}
	if count > 0 {
		return nil, errors.New("username already taken")
	}

	// Check if email is already taken by another user
	err = tx.Raw(`
		SELECT COUNT(*) 
		FROM users 
		WHERE email = $1 AND id != $2`,
		req.Email, id,
	).Scan(&count).Error
	if err != nil {
		return nil, fmt.Errorf("error checking email uniqueness: %v", err)
	}
	if count > 0 {
		return nil, errors.New("email already taken")
	}

	// Build update query
	query := `
		UPDATE users 
		SET username = $1, 
			email = $2, 
			language = $3, 
			updated_at = $4`
	args := []interface{}{
		req.Username,
		req.Email,
		req.Language,
		time.Now(),
	}

	argCount := 5

	// Add password update if provided
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("error hashing password: %v", err)
		}
		query += fmt.Sprintf(", password = $%d", argCount)
		args = append(args, hashedPassword)
		argCount++
	}

	// Add WHERE clause
	query += fmt.Sprintf(" WHERE id = $%d RETURNING id, username, email, role, status, language, prof_pic, created_at, updated_at", argCount)
	args = append(args, id)

	// Execute update and scan result
	var profile models.UserProfileResponse
	err = tx.Raw(query, args...).Scan(
		&profile,
	).Error
	if err != nil {

		return nil, fmt.Errorf("error updating user profile: %v", err)
	}

	// Commit transaction
	err = tx.Commit().Error
	if err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}
	return &profile, nil
}
