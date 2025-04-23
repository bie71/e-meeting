package middleware

import (
	"e_metting/internal/auth"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

func AuthMiddleware(jwtConfig *auth.JWTConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			log.Warn().Msg("No authorization header")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Warn().Msg("Invalid authorization header format")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		tokenString := parts[1]
		token, err := jwtConfig.ValidateToken(tokenString)
		if err != nil {
			log.Warn().Err(err).Msg("Invalid token")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Warn().Msg("Invalid token claims")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token claims",
			})
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			log.Warn().Msg("User ID not found in token")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "User ID not found in token",
			})
		}

		c.Locals("userID", userID)
		return c.Next()
	}
}

// func AuthMiddleware() gin.HandlerFunc {
// 	return func(c gin.Context) {
// 		authHeader := c.GetHeader("Authorization")
// 		if authHeader == "" {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
// 			c.Abort()
// 			return
// 		}

// 		// Check if the header starts with "Bearer "
// 		parts := strings.Split(authHeader, " ")
// 		if len(parts) != 2 || parts[0] != "Bearer" {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
// 			c.Abort()
// 			return
// 		}

// 		tokenString := parts[1]
// 		claims := &models.Claims{}

// 		// Parse and validate the token
// 		token, err := jwt.ParseWithClaims(tokenString, claims, func(tokenjwt.Token) (interface{}, error) {
// 			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
// 			}
// 			return []byte(viper.GetString("JWT_SECRET_KEY")), nil
// 		})

// 		if err != nil {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
// 			c.Abort()
// 			return
// 		}

// 		if !token.Valid {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
// 			c.Abort()
// 			return
// 		}

// 		// Store claims in context
// 		c.Set("userID", claims.UserID)
// 		c.Set("username", claims.Username)
// 		c.Next()
// 	}
// }
