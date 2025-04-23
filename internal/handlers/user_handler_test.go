package handlers

import (
	"bytes"
	"e_metting/internal/models"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Register(req models.RegisterRequest) (*models.User, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) Login(req models.LoginRequest) (string, error) {
	args := m.Called(req)
	return args.String(0), args.Error(1)
}

func setupRouter(handler *UserHandler) *fiber.App {
	app := fiber.New()
	app.Post("/register", handler.Register)
	app.Post("/login", handler.Login)
	return app
}

func TestUserHandler_Register(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)
	router := setupRouter(handler)

	tests := []struct {
		name           string
		req            *models.RegisterRequest
		mockSetup      func()
		expectedStatus int
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
				mockService.On("Register", mock.Anything).Return(&models.User{}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Invalid request",
			req: &models.RegisterRequest{
				Username:        "testuser",
				Email:           "invalid-email",
				Password:        "pass",
				ConfirmPassword: "pass",
				Language:        "id",
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			reqBody, _ := json.Marshal(tt.req)
			req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := router.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)
	router := setupRouter(handler)

	tests := []struct {
		name           string
		req            *models.LoginRequest
		mockSetup      func()
		expectedStatus int
	}{
		{
			name: "Success",
			req: &models.LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			mockSetup: func() {
				mockService.On("Login", mock.Anything).Return("token", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid credentials",
			req: &models.LoginRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			mockSetup: func() {
				mockService.On("Login", mock.Anything).Return("", errors.New("invalid credentials"))
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			reqBody, _ := json.Marshal(tt.req)
			req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := router.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			mockService.AssertExpectations(t)
		})
	}
}
