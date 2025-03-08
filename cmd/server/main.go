package main

import (
	"context"
	"fmt"

	//"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/drmaq/streamnotification/internal/config"
	"github.com/drmaq/streamnotification/internal/db"
	"github.com/drmaq/streamnotification/internal/logger"
	"github.com/drmaq/streamnotification/internal/server"
	"github.com/drmaq/streamnotification/internal/twitch"
)

func main() {
	// Initialize logger
	logger := logger.NewLogger()
	logger.Info("Starting Twitch Stream Notification Bot")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration: %v", err)
	}

	// Connect to database
	database, err := db.NewDatabase(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Run database migrations
	if err := database.Migrate(); err != nil {
		logger.Fatal("Failed to run database migrations: %v", err)
	}

	// Initialize Twitch client
	twitchClient, err := twitch.NewClient(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize Twitch client: %v", err)
	}

	// Start the HTTP server
	srv := server.NewServer(cfg, logger, database, twitchClient)
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: srv.Router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Server listening on port %s", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start: %v", err)
		}
	}()

	// Start Twitch stream monitor in a goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go twitchClient.StartMonitoring(ctx, database)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited properly")
}
