package models

import (
	"time"

	"github.com/google/uuid"
)

type CreateSnackRequest struct {
	Name     string  `json:"name" validate:"required"`
	Category string  `json:"category" validate:"required"`
	Price    float64 `json:"price" validate:"required,gt=0"`
}

type CreateSnackResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	Price     float64   `json:"price"`
	CreatedAt time.Time `json:"created_at"`
}
