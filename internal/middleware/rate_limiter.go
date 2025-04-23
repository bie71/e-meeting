package middleware

import (
	"sync"
	"time"

	"e_metting/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.Mutex
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) RateLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		rl.mu.Lock()
		defer rl.mu.Unlock()

		now := time.Now()
		windowStart := now.Add(-rl.window)

		// Clean up old requests
		var validRequests []time.Time
		for _, t := range rl.requests[ip] {
			if t.After(windowStart) {
				validRequests = append(validRequests, t)
			}
		}
		rl.requests[ip] = validRequests

		// Check if limit exceeded
		if len(rl.requests[ip]) >= rl.limit {
			log.Warn().Str("ip", ip).Msg("Rate limit exceeded")
			return c.Status(fiber.StatusTooManyRequests).JSON(models.ErrorResponse{
				Error: "Rate limit exceeded",
			})
		}

		// Add new request
		rl.requests[ip] = append(rl.requests[ip], now)
		return c.Next()
	}
}
