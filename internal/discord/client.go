package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/drmaq/streamnotification/internal/logger"
	"github.com/drmaq/streamnotification/internal/models"
)

// Client represents a Discord webhook client
type Client struct {
	Logger *logger.Logger
}

// NewClient creates a new Discord webhook client
func NewClient(logger *logger.Logger) *Client {
	return &Client{
		Logger: logger,
	}
}

// Embed represents a Discord embed message
type Embed struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	Color       int       `json:"color"`
	Timestamp   time.Time `json:"timestamp"`
	Thumbnail   struct {
		URL string `json:"url"`
	} `json:"thumbnail"`
	Fields []struct {
		Name   string `json:"name"`
		Value  string `json:"value"`
		Inline bool   `json:"inline"`
	} `json:"fields"`
}

// WebhookMessage represents a Discord webhook message
type WebhookMessage struct {
	Content string  `json:"content"`
	Embeds  []Embed `json:"embeds"`
}

// SendNotification sends a notification to a Discord webhook
func (c *Client) SendNotification(webhookURL string, event *models.StreamEvent) error {
	// Create embed message
	embed := Embed{
		Title:       fmt.Sprintf("%s is now live on Twitch!", event.DisplayName),
		Description: event.StreamTitle,
		URL:         fmt.Sprintf("https://twitch.tv/%s", event.Username),
		Color:       0x6441A4, // Twitch purple
		Timestamp:   event.StartedAt,
	}

	// Set thumbnail
	embed.Thumbnail.URL = event.ThumbnailURL

	// Add fields
	embed.Fields = []struct {
		Name   string `json:"name"`
		Value  string `json:"value"`
		Inline bool   `json:"inline"`
	}{
		{
			Name:   "Game",
			Value:  event.GameName,
			Inline: true,
		},
		{
			Name:   "Viewers",
			Value:  fmt.Sprintf("%d", event.ViewerCount),
			Inline: true,
		},
	}

	// Create webhook message
	msg := WebhookMessage{
		Embeds: []Embed{embed},
	}

	// Marshal message to JSON
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal Discord message: %w", err)
	}

	// Send webhook request
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to send Discord webhook: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Discord webhook returned error status: %d", resp.StatusCode)
	}

	c.Logger.Info("Sent Discord notification for %s", event.DisplayName)
	return nil
}