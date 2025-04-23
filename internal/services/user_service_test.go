package services

import (
	"context"
	"e_metting/internal/auth"
	"e_metting/internal/models"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	args := m.Called(ctx, userID, hashedPassword)
	return args.Error(0)
}

func TestUserService_Register(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtConfig := auth.NewJWTConfig("test-secret", 24*time.Hour)
	service := NewUserService(mockRepo, jwtConfig)

	tests := []struct {
		name          string
		req           *models.RegisterRequest
		mockSetup     func()
		expectedError error
	}{
		{
			name: "Success",
			req: &models.RegisterRequest{
				Username:        "testuser",
				Email:           "test@example.com",
				Password:        "password123",
				ConfirmPassword: "password123",
				Language:        "id",
			},
			mockSetup: func() {
				mockRepo.On("GetUserByUsername", mock.Anything, "testuser").Return(nil, nil)
				mockRepo.On("CreateUser", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "Username already exists",
			req: &models.RegisterRequest{
				Username:        "existinguser",
				Email:           "test@example.com",
				Password:        "password123",
				ConfirmPassword: "password123",
				Language:        "id",
			},
			mockSetup: func() {
				mockRepo.On("GetUserByUsername", mock.Anything, "existinguser").Return(&models.User{}, nil)
			},
			expectedError: errors.New("username already exists"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			_, err := service.Register(*tt.req)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_Login(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtConfig := auth.NewJWTConfig("test-secret", 24*time.Hour)
	service := NewUserService(mockRepo, jwtConfig)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Password: string(hashedPassword),
		Role:     "user",
	}

	tests := []struct {
		name          string
		req           *models.LoginRequest
		mockSetup     func()
		expectedError error
	}{
		{
			name: "Success",
			req: &models.LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			mockSetup: func() {
				mockRepo.On("GetUserByUsername", mock.Anything, "testuser").Return(user, nil)
			},
			expectedError: nil,
		},
		{
			name: "Invalid credentials",
			req: &models.LoginRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			mockSetup: func() {
				mockRepo.On("GetUserByUsername", mock.Anything, "testuser").Return(user, nil)
			},
			expectedError: errors.New("invalid credentials"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			_, err := service.Login(*tt.req)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
