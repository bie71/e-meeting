package services

import (
	"context"
	"e_metting/internal/auth"
	"e_metting/internal/models"
	"e_metting/internal/repositories"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(req models.RegisterRequest) (*models.User, error)
	Login(req models.LoginRequest) (string, string, error)
	GetProfile(userID string) (*models.UserProfileResponse, error)
	UpdateProfile(userID string, req *models.UpdateProfileRequest) (*models.UserProfileResponse, error)
}

type userService struct {
	userRepo  repositories.UserRepository
	jwtConfig *auth.JWTConfig
}

func NewUserService(userRepo repositories.UserRepository, jwtConfig *auth.JWTConfig) UserService {
	return &userService{
		userRepo:  userRepo,
		jwtConfig: jwtConfig,
	}
}

func (s *userService) Register(req models.RegisterRequest) (*models.User, error) {
	// Check if username already exists
	existingUser, err := s.userRepo.GetUserByUsername(context.Background(), req.Username)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check username existence")
		return nil, errors.New("failed to register user")
	}
	if existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		return nil, errors.New("failed to register user")
	}

	// Create new user
	if req.Language == "" {
		req.Language = "id"
	}
	user := &models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     "user",
		Language: req.Language,
		Status:   true,
	}

	if err := s.userRepo.CreateUser(context.Background(), user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) Login(req models.LoginRequest) (string, string, error) {
	// Get user by username
	user, err := s.userRepo.GetUserByUsername(context.Background(), req.Username)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user by username")
		return "", "", err
	}
	if user == nil {
		return "", "", errors.New("invalid credentials, user not found")
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		log.Error().Err(err).Msg("Failed to compare password")
		return "", "", errors.New("invalid credentials, password doesn't match")
	}

	// Generate JWT token
	token, err := s.jwtConfig.GenerateToken(user.ID.String(), user.Username, user.Role)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate token")
		return "", "", errors.New("failed to generate token")
	}

	return token, user.ID.String(), nil
}

func (s *userService) GetProfile(userID string) (*models.UserProfileResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %v", err)
	}

	profile, err := s.userRepo.GetProfile(context.Background(), id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %v", err)
	}
	return profile, nil
}

func (s *userService) UpdateProfile(userID string, req *models.UpdateProfileRequest) (*models.UserProfileResponse, error) {
	profile, err := s.userRepo.UpdateProfile(context.Background(), userID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update user profile: %v", err)
	}
	return profile, nil
}
