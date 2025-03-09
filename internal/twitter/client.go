package twitter

import (
	"fmt"
	"net/http"
	"time"

	"github.com/drmaq/streamnotification/internal/logger"
	"github.com/drmaq/streamnotification/internal/models"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

// Client represents a Twitter API client
type Client struct {
	Logger           *logger.Logger
	ConsumerKey      string
	ConsumerSecret   string
	AccessToken      string
	AccessTokenSecret string
	client           *twitter.Client
	Timeout          time.Duration
	RetryCount       int
	RetryDelay       time.Duration
}

// NewClient creates a new Twitter API client
func NewClient(logger *logger.Logger, consumerKey, consumerSecret, accessToken, accessTokenSecret string) *Client {
	c := &Client{
		Logger:           logger,
		ConsumerKey:      consumerKey,
		ConsumerSecret:   consumerSecret,
		AccessToken:      accessToken,
		AccessTokenSecret: accessTokenSecret,
		Timeout:          10 * time.Second, // Default timeout
		RetryCount:       3,               // Default retry count
		RetryDelay:       2 * time.Second, // Default retry delay
	}

	// Initialize Twitter client if credentials are provided
	if consumerKey != "" && consumerSecret != "" && accessToken != "" && accessTokenSecret != "" {
		c.initClient()
	}

	return c
}

// initClient initializes the Twitter API client
func (c *Client) initClient() {
	// Create OAuth1 config
	config := oauth1.NewConfig(c.ConsumerKey, c.ConsumerSecret)
	token := oauth1.NewToken(c.AccessToken, c.AccessTokenSecret)

	// Create HTTP client with OAuth1 authentication and timeout
	httpClient := &http.Client{
		Timeout: c.Timeout,
		Transport: config.Client(oauth1.NoContext, token).Transport,
	}

	// Create Twitter client
	c.client = twitter.NewClient(httpClient)
}

// SendNotification sends a notification tweet
func (c *Client) SendNotification(event *models.StreamEvent) error {
	// Check if client is initialized
	if c.client == nil {
		return fmt.Errorf("Twitter client not initialized")
	}

	// Create tweet text
	tweetText := fmt.Sprintf("%s is now live on Twitch!\n\n%s\n\nPlaying: %s\n\nhttps://twitch.tv/%s",
		event.DisplayName,
		event.StreamTitle,
		event.GameName,
		event.Username)

	// Implement retry logic with exponential backoff
	var tweet *twitter.Tweet
	var resp *http.Response
	var err error
	
	for attempt := 0; attempt <= c.RetryCount; attempt++ {
		// If this is a retry, wait before attempting again
		if attempt > 0 {
			retryWait := c.RetryDelay * time.Duration(1 << uint(attempt-1)) // Exponential backoff
			c.Logger.Info("Retrying Twitter API call in %v (attempt %d/%d)", retryWait, attempt, c.RetryCount)
			time.Sleep(retryWait)
		}
		
		// Post tweet
		tweet, resp, err = c.client.Statuses.Update(tweetText, nil)
		
		// If successful or not a retriable error, break
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		
		// If this was the last attempt, return the error
		if attempt == c.RetryCount {
			if err != nil {
				return fmt.Errorf("failed to post tweet after %d attempts: %w", c.RetryCount+1, err)
			}
			return fmt.Errorf("Twitter API returned error status after %d attempts: %d", c.RetryCount+1, resp.StatusCode)
		}
	}

	c.Logger.Info("Sent Twitter notification for %s (Tweet ID: %d)", event.DisplayName, tweet.ID)
	return nil
}

// UpdateCredentials updates the Twitter API credentials
func (c *Client) UpdateCredentials(consumerKey, consumerSecret, accessToken, accessTokenSecret string) {
	c.ConsumerKey = consumerKey
	c.ConsumerSecret = consumerSecret
	c.AccessToken = accessToken
	c.AccessTokenSecret = accessTokenSecret

	// Re-initialize client with new credentials
	c.initClient()
}

// SetTimeout sets the timeout for Twitter API requests
func (c *Client) SetTimeout(timeout time.Duration) {
	c.Timeout = timeout
	// Re-initialize client to apply new timeout
	if c.ConsumerKey != "" && c.ConsumerSecret != "" && c.AccessToken != "" && c.AccessTokenSecret != "" {
		c.initClient()
	}
}

// SetRetryOptions configures the retry behavior for failed API calls
func (c *Client) SetRetryOptions(retryCount int, retryDelay time.Duration) {
	c.RetryCount = retryCount
	c.RetryDelay = retryDelay
}