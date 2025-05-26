package repositories

import (
	"context"
	"e_metting/internal/models"
	"errors"
	"fmt"
	"strings"
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
	GetUsers(ctx context.Context, userFilter *models.UserFilter, pagination *models.PaginationQuery) (*models.UsersResponse, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
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
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("user not found")
		}
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

	// Add status update if provided
	if req.Status != "" {
		var status bool
		if strings.EqualFold(req.Status, "active") {
			status = true
		}
		query += fmt.Sprintf(", status = $%d", argCount)
		args = append(args, status)
		argCount++
	}

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

	// Add prof_pic update if provided
	if req.UrlProfPic != "" {
		query += fmt.Sprintf(", prof_pic = $%d", argCount)
		args = append(args, req.UrlProfPic)
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
func (r *userRepository) GetUsers(ctx context.Context, userFilter *models.UserFilter, pagination *models.PaginationQuery) (*models.UsersResponse, error) {
	var users []models.User
	var totalCount int64

	if pagination.Page < 1 {
		pagination.Page = 1
	}

	if pagination.PageSize < 1 {
		pagination.PageSize = 10
	}

	db := r.db.WithContext(ctx).Model(&models.User{})

	// Apply filters
	if userFilter != nil {
		if userFilter.Status != nil {
			db = db.Where("status = ?", *userFilter.Status)
		}
		if userFilter.UserId != nil {
			db = db.Where("id = ?", *userFilter.UserId)
		}
		if userFilter.Role != nil {
			db = db.Where("role = ?", *userFilter.Role)
		}
		if userFilter.Search != nil {
			searchTerm := "%" + *userFilter.Search + "%"
			db = db.Where("username ILIKE ? OR email ILIKE ?", searchTerm, searchTerm)
		}
	}

	// Count total results (before pagination)
	if err := db.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	// Apply pagination
	page := pagination.Page
	pageSize := pagination.PageSize
	offset := (page - 1) * pageSize

	err := db.Limit(pageSize).Offset(offset).Find(&users).Error
	if err != nil {
		return nil, err
	}

	totalPages := int((totalCount + int64(pageSize) - 1) / int64(pageSize))

	return &models.UsersResponse{
		Users:      users,
		TotalCount: int(totalCount),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *userRepository) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	if err := r.db.Unscoped().Delete(&models.PasswordResetToken{}, "user_id = ?", userID).Error; err != nil {
		return fmt.Errorf("failed to delete password reset tokens for user %s: %v", userID, err)
	}
	return r.db.WithContext(ctx).Unscoped().Delete(&models.User{}, "id = ?", userID).Error
}
