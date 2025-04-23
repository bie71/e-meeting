package server

import (
	"e_metting/internal/auth"
	"e_metting/internal/handlers"
	"e_metting/internal/middleware"
	"e_metting/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func SetupRouter(
	userHandler *handlers.UserHandler,
	healthHandler *handlers.HealthHandler,
	passwordResetHandler *handlers.PasswordResetHandler,
	rateLimiter *middleware.RateLimiter,
	jwtConfig *auth.JWTConfig,
	dashboardHandler *handlers.DashboardHandler,
) *fiber.App {
	app := fiber.New()

	// Middleware
	app.Use(recover.New())
	app.Use(middleware.LoggerMiddleware())
	app.Use(middleware.ErrorHandlerMiddleware())
	// app.Use(rateLimiter.RateLimit())

	// Public routes
	public := app.Group("/api/v1")
	public.Get("/health", healthHandler.HealthCheck)
	public.Get("/swagger.json", handlers.SwaggerUI)
	public.Post("/auth/register", middleware.ValidateRequest[models.RegisterRequest](), userHandler.Register)
	public.Post("/auth/login", middleware.ValidateRequest[models.LoginRequest](), userHandler.Login)
	public.Post("/password/reset/request", middleware.ValidateRequest[models.ResetPasswordRequest](), passwordResetHandler.RequestReset)
	public.Post("/password/reset", middleware.ValidateRequest[models.ResetPasswordConfirmRequest](), passwordResetHandler.ResetPassword)

	// Protected routes
	protected := app.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(jwtConfig))
	{
		// Add protected routes here
		protected.Get("/profile/:id", userHandler.GetProfile)
		protected.Put("/profile/:id", middleware.ValidateRequest[models.UpdateProfileRequest](), userHandler.UpdateProfile)
		protected.Get("/dashboard", dashboardHandler.GetDashboardStats)
	}

	return app
}

// {
//         protected.GET("/users/:id", userHandler.GetProfile)
//         protected.POST("/users/:id", userHandler.UpdateProfile)
//         protected.GET("/dashboard", dashboardHandler.GetDashboardStats)
//         protected.GET("/rooms", roomHandler.GetRooms)
//         protected.GET("/rooms/:id/schedule", roomHandler.GetRoomSchedule)
//         protected.GET("/snacks", snackHandler.GetSnacks)
//         protected.POST("/reservation/calculation", reservationHandler.CalculateReservationCost)
//         protected.POST("/reservation", reservationHandler.CreateReservation)
//         protected.GET("/reservation/history", reservationHandler.GetReservationHistory)
//     }
