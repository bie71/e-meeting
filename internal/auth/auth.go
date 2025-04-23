package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTConfig struct {
	SecretKey     string
	TokenDuration time.Duration
}

func NewJWTConfig(secretKey string, tokenDuration time.Duration) *JWTConfig {
	return &JWTConfig{
		SecretKey:     secretKey,
		TokenDuration: tokenDuration,
	}
}

func (c *JWTConfig) GenerateToken(userID, username, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"role":     role,
		"exp":      time.Now().Add(c.TokenDuration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(c.SecretKey))
}

func (c *JWTConfig) ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(c.SecretKey), nil
	})
}
