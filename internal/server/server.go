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

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type Server struct {
	app *fiber.App
	cfg *config.Config
}

func NewServer(cfg *config.Config) *Server {
	// Initialize database

	db := database.New(cfg)

	err := database.SeedUsers(db.DB())
	if err != nil {
		log.Fatalf("Failed to seed users: %v", err)
	}
	// Initialize repositories
	userRepo := repositories.NewUserRepository(db.GormDB())
	passwordResetRepo := repositories.NewPasswordResetRepository(db.GormDB())

	// Initialize JWT config
	fmt.Println(cfg.JWT.TokenDuration, "======================")
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
		cfg,
	)
	userService := services.NewUserService(userRepo, jwtConfig)
	dashboardDb := services.NewDashboardService(db.DB())
	reservationService := services.NewReservationService(db.DB())
	roomService := services.NewRoomService(db.DB())
	snackService := services.NewSnackService(db.DB())

	validator := validator.New()

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)
	healthHandler := handlers.NewHealthHandler("1.0.0")
	passwordResetHandler := handlers.NewPasswordResetHandler(passwordResetService)
	dashboardHandler := handlers.NewDashboardHandler(dashboardDb)
	reservationHandler := handlers.NewReservationHandler(reservationService)
	roomHandler := handlers.NewRoomHandler(roomService)
	snackHandler := handlers.NewSnackHandler(snackService, validator)

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(100, time.Hour)

	// Setup router
	router := SetupRouter(
		userHandler,
		healthHandler,
		passwordResetHandler,
		rateLimiter,
		jwtConfig,
		dashboardHandler,
		reservationHandler,
		roomHandler,
		snackHandler,
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
