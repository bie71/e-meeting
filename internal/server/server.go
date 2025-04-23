package server

import (
	"e_metting/internal/auth"
	"e_metting/internal/config"
	"e_metting/internal/database"
	"e_metting/internal/handlers"
	"e_metting/internal/middleware"
	"e_metting/internal/repositories"
	"e_metting/internal/services"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Server struct {
	app *fiber.App
	cfg *config.Config
}

func NewServer(cfg *config.Config) *Server {
	// Initialize database
	db, err := database.NewPostgresDB(
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		fmt.Sprintf("%d", cfg.DBPort),
	)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	passwordResetRepo := repositories.NewPasswordResetRepository(db)

	// Initialize JWT config
	jwtConfig := auth.NewJWTConfig(
		cfg.JWT.SecretKey,
		time.Duration(cfg.JWT.TokenDuration)*time.Hour,
	)

	// Initialize services
	emailService := services.NewEmailService(
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		cfg.SMTP.Username,
		cfg.SMTP.Password,
		cfg.SMTP.FromEmail,
	)
	passwordResetService := services.NewPasswordResetService(
		userRepo,
		passwordResetRepo,
		emailService,
	)
	userService := services.NewUserService(userRepo, jwtConfig)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)
	healthHandler := handlers.NewHealthHandler("1.0.0")
	passwordResetHandler := handlers.NewPasswordResetHandler(passwordResetService)

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(100, time.Hour)

	// Setup router
	router := SetupRouter(
		userHandler,
		healthHandler,
		passwordResetHandler,
		rateLimiter,
		jwtConfig,
	)

	return &Server{
		app: router,
		cfg: cfg,
	}
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	return s.app.Listen(addr)
}

func (s *Server) Shutdown() error {
	return s.app.Shutdown()
}
