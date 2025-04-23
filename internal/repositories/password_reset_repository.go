package repositories

import (
	"context"
	"e_metting/internal/models"
	"time"

	"gorm.io/gorm"
)

type PasswordResetRepository struct {
	db *gorm.DB
}

func NewPasswordResetRepository(db *gorm.DB) *PasswordResetRepository {
	return &PasswordResetRepository{
		db: db,
	}
}

func (r *PasswordResetRepository) CreateToken(ctx context.Context, token *models.PasswordResetToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *PasswordResetRepository) GetToken(ctx context.Context, token string) (*models.PasswordResetToken, error) {
	var resetToken models.PasswordResetToken
	err := r.db.WithContext(ctx).
		Where("token = ? AND expires_at > ?", token, time.Now()).
		First(&resetToken).Error
	if err != nil {
		return nil, err
	}
	return &resetToken, nil
}

func (r *PasswordResetRepository) DeleteToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).
		Where("token = ?", token).
		Delete(&models.PasswordResetToken{}).Error
}

func (r *PasswordResetRepository) DeleteExpiredTokens(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at <= ?", time.Now()).
		Delete(&models.PasswordResetToken{}).Error
}

func (r *PasswordResetRepository) MarkTokenAsUsed(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).
		Model(&models.PasswordResetToken{}).
		Where("token = ?", token).
		Update("used", true).Error
}
