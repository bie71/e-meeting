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
	}

	return app
}
