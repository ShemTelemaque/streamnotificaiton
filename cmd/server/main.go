package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/drmaq/streamnotification/internal/api"
	"github.com/drmaq/streamnotification/internal/config"
	"github.com/drmaq/streamnotification/internal/db"
	"github.com/drmaq/streamnotification/internal/discord"
	"github.com/drmaq/streamnotification/internal/frontend"
	"github.com/drmaq/streamnotification/internal/logger"
	"github.com/drmaq/streamnotification/internal/twitch"
	"github.com/drmaq/streamnotification/internal/twitter"
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

	// Initialize Discord client
	discordClient := discord.NewClient(logger)

	// Initialize Twitter client
	twitterClient := twitter.NewClient(
		logger,
		cfg.TwitterAPIKey,
		cfg.TwitterAPISecret,
		cfg.TwitterAccessToken,
		cfg.TwitterAccessSecret,
	)

	// Initialize Twitch client
	twitchClient, err := twitch.NewClient(cfg, logger, discordClient, twitterClient)
	if err != nil {
		logger.Fatal("Failed to initialize Twitch client: %v", err)
	}

	// Create API router
	apiRouter := api.NewRouter(cfg, logger, database, twitchClient)

	// Create frontend router with API base URL
	apiBaseURL := fmt.Sprintf("http://localhost:%s", cfg.Port)
	frontendRouter := frontend.NewRouter(cfg, logger, apiBaseURL)

	// Create a main router that combines API and frontend routes
	mainRouter := http.NewServeMux()

	// Mount API routes
	mainRouter.Handle("/api/", apiRouter.Router)
	mainRouter.Handle("/ws/", apiRouter.Router)

	// Mount frontend routes (everything else)
	mainRouter.Handle("/", frontendRouter.Router)

	// Start the HTTP server
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: mainRouter,
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
