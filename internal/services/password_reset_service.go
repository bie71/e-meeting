package services

import (
	"context"
	"e_metting/internal/config"
	"e_metting/internal/models"
	"e_metting/internal/repositories"
	"fmt"
	"time"

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

func (s *PasswordResetService) RequestReset(ctx context.Context, email string) (string, error) {
	// Check if user exists
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user by email")
		// Return success even if user doesn't exist (security through obscurity)
		return "", nil
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

	// Send reset email
	resetLink := fmt.Sprintf("%s?token=%s", s.cfg.FrontendURL, token)
	if err := s.emailSender.SendPasswordResetEmail(user.Email, resetLink); err != nil {
		log.Error().Err(err).Msg("Failed to send reset email")
		return "", err
	}

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
