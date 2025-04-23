package main

import (
	"context"
	"e_metting/internal/config"
	"e_metting/internal/server"
	"log"
	"os/signal"
	"syscall"
	"time"
)

func gracefulShutdown(app *server.Server, done chan bool) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")
	stop() // Allow Ctrl+C to force shutdown

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.Shutdown(); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")

	// Notify the main goroutine that the shutdown is complete
	done <- true
}

func main() {
	// Load config
	cfg := config.NewConfig()

	// Create server
	app := server.NewServer(cfg)

	// Channel to wait for graceful shutdown
	done := make(chan bool, 1)
	go gracefulShutdown(app, done)

	// Start server
	if err := app.Start(); err != nil {
		log.Printf("Server stopped: %v", err)
	}

	// Wait for graceful shutdown to complete
	<-done
}
