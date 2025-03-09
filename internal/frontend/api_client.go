package frontend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/drmaq/streamnotification/internal/logger"
	"github.com/drmaq/streamnotification/internal/models"
)

// APIClient handles communication with the backend API
type APIClient struct {
	BaseURL    string
	HTTPClient *http.Client
	Logger     *logger.Logger
}

// NewAPIClient creates a new API client
func NewAPIClient(baseURL string, logger *logger.Logger) *APIClient {
	return &APIClient{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		Logger:     logger,
	}
}

// GetStreamers fetches all streamers from the API
func (c *APIClient) GetStreamers() ([]models.Streamer, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/api/streamers")
	if err != nil {
		return nil, fmt.Errorf("failed to get streamers: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned error: %s", resp.Status)
	}

	var streamers []models.Streamer
	if err := json.NewDecoder(resp.Body).Decode(&streamers); err != nil {
		return nil, fmt.Errorf("failed to decode streamers: %w", err)
	}

	return streamers, nil
}

// AddStreamer adds a new streamer via the API
func (c *APIClient) AddStreamer(username string) (*models.Streamer, error) {
	reqBody, err := json.Marshal(map[string]string{"username": username})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.HTTPClient.Post(
		c.BaseURL+"/api/streamers",
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add streamer: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API returned error: %s", resp.Status)
	}

	var streamer models.Streamer
	if err := json.NewDecoder(resp.Body).Decode(&streamer); err != nil {
		return nil, fmt.Errorf("failed to decode streamer: %w", err)
	}

	return &streamer, nil
}

// DeleteStreamer deletes a streamer via the API
func (c *APIClient) DeleteStreamer(id int) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/streamers/%d", c.BaseURL, id), nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete streamer: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("API returned error: %s", resp.Status)
	}

	return nil
}

// GetNotificationSettings fetches all notification settings from the API
func (c *APIClient) GetNotificationSettings() ([]models.NotificationSetting, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/api/notifications")
	if err != nil {
		return nil, fmt.Errorf("failed to get notification settings: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned error: %s", resp.Status)
	}

	var notifications []models.NotificationSetting
	if err := json.NewDecoder(resp.Body).Decode(&notifications); err != nil {
		return nil, fmt.Errorf("failed to decode notification settings: %w", err)
	}

	return notifications, nil
}

// AddNotificationSetting adds a new notification setting via the API
func (c *APIClient) AddNotificationSetting(notification *models.NotificationSetting) error {
	reqBody, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.HTTPClient.Post(
		c.BaseURL+"/api/notifications",
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return fmt.Errorf("failed to add notification setting: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("API returned error: %s", resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(notification)
}

// UpdateNotificationSetting updates a notification setting via the API
func (c *APIClient) UpdateNotificationSetting(notification *models.NotificationSetting) error {
	reqBody, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(
		"PUT",
		fmt.Sprintf("%s/api/notifications/%d", c.BaseURL, notification.ID),
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return fmt.Errorf("failed to create update request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update notification setting: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned error: %s", resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(notification)
}

// DeleteNotificationSetting deletes a notification setting via the API
func (c *APIClient) DeleteNotificationSetting(id int) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/notifications/%d", c.BaseURL, id), nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete notification setting: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("API returned error: %s", resp.Status)
	}

	return nil
}

// GetLogs fetches log entries from the API
func (c *APIClient) GetLogs() ([]logger.LogEntry, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/api/logs")
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned error: %s", resp.Status)
	}

	var logs []logger.LogEntry
	if err := json.NewDecoder(resp.Body).Decode(&logs); err != nil {
		return nil, fmt.Errorf("failed to decode logs: %w", err)
	}

	return logs, nil
}

// Login authenticates a user via the API
func (c *APIClient) Login(username, password string) (*models.User, error) {
	reqBody, err := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.HTTPClient.Post(
		c.BaseURL+"/api/auth/login",
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to login: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned error: %s", resp.Status)
	}

	var user models.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user: %w", err)
	}

	return &user, nil
}

// Register creates a new user account via the API
func (c *APIClient) Register(username, password string) (*models.User, error) {
	reqBody, err := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.HTTPClient.Post(
		c.BaseURL+"/api/auth/register",
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API returned error: %s", resp.Status)
	}

	var user models.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user: %w", err)
	}

	return &user, nil
}
