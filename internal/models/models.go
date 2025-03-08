package models

import (
	"time"
)

// Streamer represents a Twitch streamer being monitored
type Streamer struct {
	ID                  int        `json:"id"`
	Username            string     `json:"username"`
	DisplayName         string     `json:"display_name"`
	IsLive              bool       `json:"is_live"`
	LastStreamStart     *time.Time `json:"last_stream_start"`
	LastNotificationSent *time.Time `json:"last_notification_sent"`
}

// NotificationType represents the type of notification
type NotificationType string

const (
	// NotificationTypeDiscord represents a Discord notification
	NotificationTypeDiscord NotificationType = "discord"
	// NotificationTypeTwitter represents a Twitter notification
	NotificationTypeTwitter NotificationType = "twitter"
)

// NotificationSetting represents a notification destination
type NotificationSetting struct {
	ID          int              `json:"id"`
	Type        NotificationType `json:"type"`
	Destination string           `json:"destination"` // Discord channel ID or Twitter account
	Enabled     bool             `json:"enabled"`
}

// StreamEvent represents a stream event (going live or offline)
type StreamEvent struct {
	StreamerID   int       `json:"streamer_id"`
	Username     string    `json:"username"`
	DisplayName  string    `json:"display_name"`
	EventType    string    `json:"event_type"` // "live" or "offline"
	StreamTitle  string    `json:"stream_title"`
	GameName     string    `json:"game_name"`
	ThumbnailURL string    `json:"thumbnail_url"`
	ViewerCount  int       `json:"viewer_count"`
	StartedAt    time.Time `json:"started_at"`
}