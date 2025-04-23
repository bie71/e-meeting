package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Uptime    string    `json:"uptime"`
}

type HealthHandler struct {
	startTime time.Time
	version   string
}

func NewHealthHandler(version string) *HealthHandler {
	return &HealthHandler{
		startTime: time.Now(),
		version:   version,
	}
}

func (h *HealthHandler) HealthCheck(c *fiber.Ctx) error {
	uptime := time.Since(h.startTime)
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   h.version,
		Uptime:    uptime.String(),
	}

	log.Info().
		Str("status", response.Status).
		Str("version", response.Version).
		Str("uptime", response.Uptime).
		Msg("Health check")

	return c.JSON(response)
}
