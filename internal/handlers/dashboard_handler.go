package handlers

import (
	"e_metting/internal/models"
	"e_metting/internal/services"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type DashboardHandler struct {
	dashboardService *services.DashboardService
}

func NewDashboardHandler(dashboardService *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
	}
}

// GetDashboardStats godoc
// @Summary Get dashboard statistics
// @Description Get statistics about reservations, visitors, rooms, and revenue
// @Accept json
// @Produce json
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Security BearerAuth
// @Success 200 {object} models.DashboardResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /dashboard [get]
func (h *DashboardHandler) GetDashboardStats(c *fiber.Ctx) error {
	var query models.DashboardQuery
	if err := c.QueryParser(&query); err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid query parameters",
		})
	}

	stats, err := h.dashboardService.GetDashboardStats(&query)
	if err != nil {
		if err.Error() == "invalid start_date format" || err.Error() == "invalid end_date format" {
			return c.Status(http.StatusBadRequest).JSON(models.ErrorResponse{
				Error: err.Error(),
			})
		}

		return c.Status(http.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to fetch dashboard statistics",
		})
	}

	return c.JSON(stats)
}
