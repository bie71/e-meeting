package services

import (
	"context"
	"e_metting/internal/config"
	"e_metting/internal/models"
	"e_metting/internal/repositories"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type PasswordResetRepository interface {
	CreateToken(ctx context.Context, token *models.PasswordResetToken) error
	GetToken(ctx context.Context, token string) (*models.PasswordResetToken, error)
	DeleteToken(ctx context.Context, token string) error
	DeleteExpiredTokens(ctx context.Context) error
	MarkTokenAsUsed(ctx context.Context, token string) error
}

type PasswordResetService struct {
	userRepo    repositories.UserRepository
	resetRepo   PasswordResetRepository
	emailSender EmailService
	cfg         *config.Config
}

func NewPasswordResetService(
	userRepo repositories.UserRepository,
	resetRepo PasswordResetRepository,
	emailSender EmailService,
	cfg *config.Config,

) *PasswordResetService {
	return &PasswordResetService{
		userRepo:    userRepo,
		resetRepo:   resetRepo,
		emailSender: emailSender,
		cfg:         cfg,
	}
}

func (s *PasswordResetService) RequestReset(ctx context.Context, email string, c *fiber.Ctx) (string, error) {
	// Check if user exists
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user by email")
		// Return success even if user doesn't exist (security through obscurity)
		return "", err
	}

	// Generate reset token
	token := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour) // Token expires in 24 hours

	// Store token
	resetToken := &models.PasswordResetToken{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: expiresAt,
		Used:      false,
	}
	if err := s.resetRepo.CreateToken(ctx, resetToken); err != nil {
		log.Error().Err(err).Msg("Failed to create reset token")
		return "", err
	}

	scheme := "http"
	if c.Protocol() == "https" {
		scheme = "https"
	}
	host := c.Hostname() // tanpa scheme, misalnya: localhost:3000 atau yourdomain.com
	baseURL := fmt.Sprintf("%s://%s", scheme, host)

	resetLink := fmt.Sprintf("%s/api/v1/recover-password?token=%s", baseURL, token)

	go func(email string, resetLink string) {
		if err := s.emailSender.SendPasswordResetEmail(email, resetLink); err != nil {
			log.Error().Err(err).Msg("Failed to send reset email")
		}
	}(user.Email, resetLink)

	return resetLink, nil
}

func (s *PasswordResetService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Verify token
	resetToken, err := s.resetRepo.GetToken(ctx, token)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get reset token")
		return err
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		return err
	}

	// Update user password
	if err := s.userRepo.UpdatePassword(ctx, resetToken.UserID, string(hashedPassword)); err != nil {
		log.Error().Err(err).Msg("Failed to update password")
		return err
	}

	// Mark token as used
	if err := s.resetRepo.MarkTokenAsUsed(ctx, token); err != nil {
		log.Error().Err(err).Msg("Failed to mark token as used")
		return err
	}

	return nil
}
