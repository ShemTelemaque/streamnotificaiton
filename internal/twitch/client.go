package twitch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/drmaq/streamnotification/internal/config"
	"github.com/drmaq/streamnotification/internal/db"
	"github.com/drmaq/streamnotification/internal/errors"
	"github.com/drmaq/streamnotification/internal/logger"
	"github.com/drmaq/streamnotification/internal/models"
)

const (
	twitchAPIBaseURL = "https://api.twitch.tv/helix"
	twitchAuthURL   = "https://id.twitch.tv/oauth2/token"
	monitorInterval = 60 * time.Second // Check every minute
)

// Client represents a Twitch API client
type Client struct {
	clientID     string
	clientSecret string
	accessToken  string
	tokenExpiry  time.Time
	httpClient   *http.Client
	logger       *logger.Logger
	mu           sync.Mutex
}

// NewClient creates a new Twitch API client
func NewClient(cfg *config.Config, logger *logger.Logger) (*Client, error) {
	client := &Client{
		clientID:     cfg.TwitchClientID,
		clientSecret: cfg.TwitchClientSecret,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		logger:       logger,
	}

	// Get initial access token
	if err := client.refreshAccessToken(); err != nil {
		return nil, errors.NewAPIError("Failed to get access token", err)
	}

	return client, nil
}

// refreshAccessToken gets a new access token from Twitch
func (c *Client) refreshAccessToken() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if token is still valid
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
		return nil
	}

	// Prepare request
	url := fmt.Sprintf("%s?client_id=%s&client_secret=%s&grant_type=client_credentials",
		twitchAuthURL, c.clientID, c.clientSecret)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return errors.NewAPIError("Failed to create auth request", err)
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.NewAPIError("Failed to send auth request", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return errors.NewAPIError(
			fmt.Sprintf("Twitch auth failed with status %d", resp.StatusCode),
			fmt.Errorf("unexpected status code: %d", resp.StatusCode),
		)
	}

	// Parse response
	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return errors.NewAPIError("Failed to parse auth response", err)
	}

	// Update token
	c.accessToken = result.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)

	return nil
}

// getAuthenticatedRequest creates a new authenticated request to the Twitch API
func (c *Client) getAuthenticatedRequest(method, endpoint string, body interface{}) (*http.Request, error) {
	// Refresh token if needed
	if err := c.refreshAccessToken(); err != nil {
		return nil, errors.NewAPIError("Failed to refresh access token", err)
	}

	// Create request
	url := twitchAPIBaseURL + endpoint
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, errors.NewAPIError("Failed to create API request", err)
	}

	// Add headers
	req.Header.Add("Client-ID", c.clientID)
	req.Header.Add("Authorization", "Bearer "+c.accessToken)

	return req, nil
}

// GetStreamerInfo gets information about a Twitch streamer
func (c *Client) GetStreamerInfo(username string) (*models.Streamer, error) {
	// Validate input
	if username == "" {
		return nil, errors.NewValidationError("Username cannot be empty", nil)
	}

	// Create request
	endpoint := fmt.Sprintf("/users?login=%s", username)
	req, err := c.getAuthenticatedRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err // Error already wrapped
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.NewAPIError("Failed to send request to Twitch API", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewAPIError(
			fmt.Sprintf("Twitch API request failed with status %d", resp.StatusCode),
			fmt.Errorf("unexpected status code: %d", resp.StatusCode),
		)
	}

	// Parse response
	var result struct {
		Data []struct {
			ID          string `json:"id"`
			Login       string `json:"login"`
			DisplayName string `json:"display_name"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.NewAPIError("Failed to parse Twitch API response", err)
	}

	// Check if user exists
	if len(result.Data) == 0 {
		return nil, errors.NewNotFoundError(fmt.Sprintf("Streamer not found: %s", username), nil)
	}

	// Create streamer
	streamer := &models.Streamer{
		Username:    result.Data[0].Login,
		DisplayName: result.Data[0].DisplayName,
		IsLive:      false,
	}

	return streamer, nil
}

// GetStreamStatus checks if a streamer is currently live
func (c *Client) GetStreamStatus(usernames []string) (map[string]*models.StreamEvent, error) {
	if len(usernames) == 0 {
		return make(map[string]*models.StreamEvent), nil
	}

	// Create query string
	var endpoint string
	for i, username := range usernames {
		if i == 0 {
			endpoint = fmt.Sprintf("/streams?user_login=%s", username)
		} else {
			endpoint += fmt.Sprintf("&user_login=%s", username)
		}
	}

	// Create request
	req, err := c.getAuthenticatedRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err // Error already wrapped
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.NewAPIError("Failed to send request to Twitch API", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewAPIError(
			fmt.Sprintf("Twitch API request failed with status %d", resp.StatusCode),
			fmt.Errorf("unexpected status code: %d", resp.StatusCode),
		)
	}

	// Parse response
	var result struct {
		Data []struct {
			UserID       string    `json:"user_id"`
			UserLogin    string    `json:"user_login"`
			UserName     string    `json:"user_name"`
			GameID       string    `json:"game_id"`
			GameName     string    `json:"game_name"`
			Type         string    `json:"type"`
			Title        string    `json:"title"`
			ViewerCount  int       `json:"viewer_count"`
			StartedAt    time.Time `json:"started_at"`
			ThumbnailURL string    `json:"thumbnail_url"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.NewAPIError("Failed to parse Twitch API response", err)
	}

	// Create map of live streamers
	liveStreamers := make(map[string]*models.StreamEvent)
	for _, stream := range result.Data {
		if stream.Type == "live" {
			liveStreamers[stream.UserLogin] = &models.StreamEvent{
				Username:     stream.UserLogin,
				DisplayName:  stream.UserName,
				EventType:    "live",
				StreamTitle:  stream.Title,
				GameName:     stream.GameName,
				ThumbnailURL: stream.ThumbnailURL,
				ViewerCount:  stream.ViewerCount,
				StartedAt:    stream.StartedAt,
			}
		}
	}

	return liveStreamers, nil
}

// StartMonitoring starts monitoring streamers for live status changes
func (c *Client) StartMonitoring(ctx context.Context, database *db.Database) {
	c.logger.Info("Starting Twitch stream monitor")

	ticker := time.NewTicker(monitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Stopping Twitch stream monitor")
			return
		case <-ticker.C:
			if err := c.checkStreamers(database); err != nil {
				c.logger.Error("Failed to check streamers: %v", err)
			}
		}
	}
}

// checkStreamers checks the live status of all streamers
func (c *Client) checkStreamers(database *db.Database) error {
	// Get all streamers from database
	streamers, err := database.GetStreamers()
	if err != nil {
		return errors.NewInternalError("Failed to get streamers from database", err)
	}

	// If no streamers, nothing to do
	if len(streamers) == 0 {
		return nil
	}

	// Get usernames
	usernames := make([]string, len(streamers))
	usernameToID := make(map[string]int)
	for i, streamer := range streamers {
		usernames[i] = streamer.Username
		usernameToID[streamer.Username] = streamer.ID
	}

	// Get live status
	liveStreamers, err := c.GetStreamStatus(usernames)
	if err != nil {
		return errors.NewAPIError("Failed to get stream status", err)
	}

	// Update streamers
	for i := range streamers {
		// Check if streamer is live
		liveEvent, isLive := liveStreamers[streamers[i].Username]

		// If status changed, update streamer
		if isLive != streamers[i].IsLive {
			oldStatus := streamers[i].IsLive
			streamers[i].IsLive = isLive

			// If went live, update last stream start and send notification
			if isLive && !oldStatus {
				streamers[i].LastStreamStart = &liveEvent.StartedAt
				streamers[i].LastNotificationSent = nil

				// Send notification
				c.logger.Info("%s went live playing %s: %s", 
					streamers[i].DisplayName, liveEvent.GameName, liveEvent.StreamTitle)
				
				// TODO: Send notifications to Discord and Twitter
				// This would be handled by notification services
			} else if !isLive && oldStatus {
				// Went offline
				c.logger.Info("%s went offline", streamers[i].DisplayName)
			}

			// Update streamer in database
			if err := database.UpdateStreamer(&streamers[i]); err != nil {
				return errors.NewInternalError("Failed to update streamer in database", err)
			}
		}
	}

	return nil
}