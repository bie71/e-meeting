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
	reservatonsHanlder *handlers.ReservationHandler,
	roomsHandler *handlers.RoomHandler,
	snacksHandler *handlers.SnackHandler,
) *fiber.App {
	app := fiber.New()

	// Middleware
	app.Use(recover.New())
	app.Use(middleware.LoggerMiddleware())
	app.Use(middleware.ErrorHandlerMiddleware())
	app.Static("/swagger", "./public/swagger")
	app.Static("/", "./public")
	// app.Use(rateLimiter.RateLimit())

	// Public routes
	public := app.Group("/api/v1")
	public.Get("/health", healthHandler.HealthCheck)
	public.Get("/swagger.json", handlers.SwaggerUI)
	public.Post("/auth/register", middleware.ValidateRequest[models.RegisterRequest](), userHandler.Register)
	public.Post("/auth/login", middleware.ValidateRequest[models.LoginRequest](), userHandler.Login)
	public.Post("/password/reset/request", middleware.ValidateRequest[models.ResetPasswordRequest](), passwordResetHandler.RequestReset)
	public.Post("/password/reset", passwordResetHandler.ResetPassword)
	public.Get("/download/collection", handlers.DownloadFile) // Download Postman collection
	public.Get("/recover-password", handlers.RecoverPassword) // Serve the password recovery page
	public.Get("/login", handlers.Login)                      // Serve the login page

	// Protected routes
	protected := app.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(jwtConfig))
	{
		// Add protected routes here
		protected.Get("/profile/:id", userHandler.GetProfile)
		protected.Put("/profile/:id", middleware.ValidateRequest[models.UpdateProfileRequest](), userHandler.UpdateProfile)
		protected.Get("/rooms", roomsHandler.GetRooms)
		protected.Get("/rooms/:id/schedule", middleware.ValidateRequest[models.RoomScheduleQuery](), roomsHandler.GetRoomSchedule)
		protected.Get("/snacks", snacksHandler.GetSnacks)
		protected.Post("/reservation/calculation", middleware.ValidateRequest[models.ReservationCalculationRequest](), reservatonsHanlder.CalculateReservationCost)
		protected.Post("/reservation", middleware.ValidateRequest[models.CreateReservationRequest](), reservatonsHanlder.CreateReservation)
		protected.Get("/reservation/:id", reservatonsHanlder.GetReservationByID)
		protected.Get("/reservations/history", reservatonsHanlder.GetReservationHistory)

	}

	adminOnly := app.Group("/api/v1/admin")
	adminOnly.Use(middleware.AdminOnlyMiddleware(jwtConfig.SecretKey))
	{
		// Add admin-only routes here
		adminOnly.Get("/dashboard", dashboardHandler.GetDashboardStats)
		adminOnly.Get("/reservations/history", reservatonsHanlder.GetReservationHistory)
		adminOnly.Post("/reservation/status", reservatonsHanlder.UpdateReservationStatus)
		adminOnly.Post("/rooms", middleware.ValidateRequest[models.CreateRoomRequest](), roomsHandler.CreateRoom)    // Create room
		adminOnly.Put("/rooms/:id", middleware.ValidateRequest[models.UpdateRoomRequest](), roomsHandler.UpdateRoom) // Update room
		adminOnly.Delete("/rooms/:id", roomsHandler.DeleteRoom)
		// Snack management
		adminOnly.Post("/snacks", snacksHandler.CreateSnack) // Create snack

	}

	return app
}
