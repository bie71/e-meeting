package models

import (
	"time"

	"github.com/google/uuid"
)

type Snack struct {
	ID        uuid.UUID `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	Price     float64   `json:"price"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SnackFilter struct {
	Search   *string  `json:"search,omitempty"`
	Category *string  `json:"category,omitempty"`
	MinPrice *float64 `json:"min_price,omitempty"`
	MaxPrice *float64 `json:"max_price,omitempty"`
}

type SnackFilterRaw struct {
	Search   string `query:"search"`
	Category string `query:"category"`
	MinPrice string `query:"min_price"`
	MaxPrice string `query:"max_price"`
}

type SnackListResponse struct {
	Snacks     []Snack `json:"snacks"`
	TotalCount int     `json:"total_count"`
	Page       int     `json:"page"`
	PageSize   int     `json:"page_size"`
	TotalPages int     `json:"total_pages"`
}
