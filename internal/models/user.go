package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Username  string         `gorm:"size:50;not null;unique" json:"username" validate:"required,min=3,max=50,alphanum"`
	Email     string         `gorm:"size:100;not null;unique" json:"email" validate:"required,email"`
	Password  string         `gorm:"size:255;not null" json:"-" validate:"required,min=6"`
	Role      string         `gorm:"size:20;not null;default:'user'" json:"role" validate:"required,oneof=user admin"`
	ProfPic   *string        `gorm:"size:255" json:"prof_pic" validate:"omitempty,url"`
	Language  string         `gorm:"size:10;not null;default:'id'" json:"language" validate:"required,oneof=id en"`
	Status    bool           `gorm:"default:true" json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type RegisterRequest struct {
	Username        string `json:"username" validate:"required,min=3,max=50,alphanum"`
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
	Language        string `json:"language"`
}

type RegisterResponse struct {
	Message string    `json:"message"`
	UserID  uuid.UUID `json:"user_id"`
}

type UserProfileResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	ProfPic   *string   `json:"prof_pic"`
	Language  string    `json:"language"`
	Status    bool      `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	jwt.RegisteredClaims
}

type UpdateProfileRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=3,max=50"`
	Language string `json:"language"`
	Password string `json:"password" validate:"omitempty,min=6"`
}
