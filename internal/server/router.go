package server

import (
	"e_metting/internal/auth"
	"e_metting/internal/handlers"
	"e_metting/internal/middleware"
	"e_metting/internal/models"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func SetupCORS() fiber.Handler {
	origins := os.Getenv("CORS_ALLOWED_ORIGINS")
	allowCreds := os.Getenv("ALLOW_CREDENTIALS") == "true"

	var corsConfig cors.Config

	// Dynamic origin handler (support * or multiple origins)
	if origins == "*" {
		// special case for AllowCredentials=true + "*"
		if allowCreds {
			corsConfig.AllowOriginsFunc = func(origin string) bool {
				return true // allow any origin
			}
		} else {
			corsConfig.AllowOrigins = "*"
		}
	} else {
		allowedOrigins := strings.Split(origins, ",")
		corsConfig.AllowOrigins = strings.Join(allowedOrigins, ",")
	}

	corsConfig.AllowMethods = "GET,POST,HEAD,PUT,DELETE,OPTIONS"
	corsConfig.AllowHeaders = "Origin, Content-Type, Accept, Authorization"
	corsConfig.AllowCredentials = allowCreds

	return cors.New(corsConfig)
}

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
	uploadHandler *handlers.UploadHandler,
) *fiber.App {
	app := fiber.New(fiber.Config{
		BodyLimit: 2 * 1024 * 1024, // 2MB
	})
	app.Use(SetupCORS())

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
		protected.Post("/rooms", roomsHandler.GetRooms)
		protected.Get("/rooms/:id/schedule", roomsHandler.GetRoomSchedule)
		protected.Get("/snacks", snacksHandler.GetSnacks)
		protected.Post("/reservation/calculation", middleware.ValidateRequest[models.ReservationCalculationRequest](), reservatonsHanlder.CalculateReservationCost)
		protected.Post("/reservation", middleware.ValidateRequest[models.CreateReservationRequest](), reservatonsHanlder.CreateReservation)
		protected.Get("/reservation/:id", reservatonsHanlder.GetReservationByID)
		protected.Get("/reservations/history", reservatonsHanlder.GetReservationHistory)
		protected.Post("/upload-image", uploadHandler.UploadHandler)

	}

	adminOnly := app.Group("/api/v1/admin")
	adminOnly.Use(middleware.AdminOnlyMiddleware(jwtConfig.SecretKey))
	{
		// Add admin-only routes here
		adminOnly.Get("/dashboard", dashboardHandler.GetDashboardStats)
		adminOnly.Get("/reservations/history", reservatonsHanlder.GetReservationHistory)
		adminOnly.Post("/reservation/status", reservatonsHanlder.UpdateReservationStatus)
		adminOnly.Delete("/reservation/:id", reservatonsHanlder.DeleteReservation)
		adminOnly.Post("/rooms", middleware.ValidateRequest[models.CreateRoomRequest](), roomsHandler.CreateRoom)    // Create room
		adminOnly.Put("/rooms/:id", middleware.ValidateRequest[models.UpdateRoomRequest](), roomsHandler.UpdateRoom) // Update room
		adminOnly.Delete("/rooms/:id", roomsHandler.DeleteRoom)
		adminOnly.Get("/rooms/:id", roomsHandler.GetRoomByID)
		// Snack management
		adminOnly.Post("/snacks", snacksHandler.CreateSnack) // Create snack
		adminOnly.Put("/snacks", snacksHandler.UpdateSnack)  // Update snack
		adminOnly.Delete("/snacks/:id", snacksHandler.DeleteSnack)
		adminOnly.Get("/snacks/:id", snacksHandler.GetSnackByID)
		// user management
		adminOnly.Get("/users", userHandler.GetAllUsers)
		adminOnly.Post("/users", userHandler.GetAllUsers)
		adminOnly.Delete("/users/:id", userHandler.DeleteUser)
		adminOnly.Post("/register/users", middleware.ValidateRequest[models.RegisterRequest](), userHandler.Register)

	}

	return app
}
