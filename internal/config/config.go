package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	Port        string
	Environment string

	// Database configuration
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Twitch API configuration
	TwitchClientID     string
	TwitchClientSecret string

	// Discord configuration
	DiscordBotToken string

	// Twitter configuration
	TwitterAPIKey       string
	TwitterAPISecret    string
	TwitterAccessToken  string
	TwitterAccessSecret string
}

// LoadConfig loads the configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Create config struct
	cfg := &Config{
		// Server configuration
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),

		// Database configuration
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", ""),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", ""),

		// Twitch API configuration
		TwitchClientID:     getEnv("TWITCH_CLIENT_ID", ""),
		TwitchClientSecret: getEnv("TWITCH_CLIENT_SECRET", ""),

		// Discord configuration
		DiscordBotToken: getEnv("DISCORD_BOT_TOKEN", ""),

		// Twitter configuration
		TwitterAPIKey:       getEnv("TWITTER_API_KEY", ""),
		TwitterAPISecret:    getEnv("TWITTER_API_SECRET", ""),
		TwitterAccessToken:  getEnv("TWITTER_ACCESS_TOKEN", ""),
		TwitterAccessSecret: getEnv("TWITTER_ACCESS_SECRET", ""),
	}

	// Validate required configuration
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate checks if all required configuration is provided
func (c *Config) validate() error {
	// Database configuration is required
	if c.DBUser == "" || c.DBPassword == "" || c.DBName == "" {
		return errors.New("database configuration is required")
	}

	// Twitch API configuration is required
	if c.TwitchClientID == "" || c.TwitchClientSecret == "" {
		return errors.New("Twitch API configuration is required")
	}

	// At least one notification method is required
	hasDiscord := c.DiscordBotToken != ""
	hasTwitter := c.TwitterAPIKey != "" && c.TwitterAPISecret != "" && 
		c.TwitterAccessToken != "" && c.TwitterAccessSecret != ""

	if !hasDiscord && !hasTwitter {
		return errors.New("at least one notification method (Discord or Twitter) is required")
	}

	return nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}