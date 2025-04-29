package middleware

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

// AdminOnlyMiddleware ensures that only users with admin role can access the protected routes
func AdminOnlyMiddleware(secretKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			log.Warn().Msg("No authorization header")
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "authorization header is required",
			})
		}

		// Check if the header has the Bearer prefix
		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || strings.ToLower(bearerToken[0]) != "bearer" {
			log.Warn().Msg("Invalid authorization header format")
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization header format",
			})
		}

		// Parse and validate the token
		token, err := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid {
			log.Warn().Err(err).Msg("Invalid token")
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token",
			})
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token claims",
			})
		}

		// Check role claim
		role, exists := claims["role"].(string)
		if !exists || role != "admin" {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"error": "access forbidden",
			})
		}

		// If everything is ok, proceed
		c.Locals("isAdmin", true)
		return c.Next()
	}
}
