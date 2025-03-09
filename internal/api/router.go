package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/drmaq/streamnotification/internal/config"
	"github.com/drmaq/streamnotification/internal/db"
	"github.com/drmaq/streamnotification/internal/logger"
	"github.com/drmaq/streamnotification/internal/models"
	"github.com/drmaq/streamnotification/internal/twitch"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Router represents the API router
type Router struct {
	Config       *config.Config
	Logger       *logger.Logger
	DB           *db.Database
	TwitchClient *twitch.Client
	Router       *mux.Router
	upgrader     websocket.Upgrader
}

// NewRouter creates a new API router
func NewRouter(cfg *config.Config, logger *logger.Logger, database *db.Database, twitchClient *twitch.Client) *Router {
	r := &Router{
		Config:       cfg,
		Logger:       logger,
		DB:           database,
		TwitchClient: twitchClient,
		Router:       mux.NewRouter(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
		},
	}

	// Set up routes
	r.setupRoutes()

	return r
}

// setupRoutes sets up the HTTP routes for the API
func (r *Router) setupRoutes() {
	// API routes
	r.Router.HandleFunc("/api/streamers", r.handleGetStreamers).Methods("GET")
	r.Router.HandleFunc("/api/streamers", r.handleAddStreamer).Methods("POST")
	r.Router.HandleFunc("/api/streamers/{id:[0-9]+}", r.handleDeleteStreamer).Methods("DELETE")
	r.Router.HandleFunc("/api/notifications", r.handleGetNotifications).Methods("GET")
	r.Router.HandleFunc("/api/notifications", r.handleAddNotification).Methods("POST")
	r.Router.HandleFunc("/api/notifications/{id:[0-9]+}", r.handleUpdateNotification).Methods("PUT")
	r.Router.HandleFunc("/api/notifications/{id:[0-9]+}", r.handleDeleteNotification).Methods("DELETE")
	r.Router.HandleFunc("/api/logs", r.handleGetLogs).Methods("GET")

	// WebSocket route for live logs
	r.Router.HandleFunc("/ws/logs", r.handleLogWebSocket)
}

// handleGetStreamers handles GET /api/streamers
func (r *Router) handleGetStreamers(w http.ResponseWriter, req *http.Request) {
	// Get streamers
	streamers, err := r.DB.GetStreamers()
	if err != nil {
		r.Logger.Error("Failed to get streamers: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(streamers)
}

// handleAddStreamer handles POST /api/streamers
func (r *Router) handleAddStreamer(w http.ResponseWriter, req *http.Request) {
	// Parse request
	var reqBody struct {
		Username string `json:"username"`
	}

	if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
		r.Logger.Error("Failed to parse request: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Get streamer info from Twitch
	streamer, err := r.TwitchClient.GetStreamerInfo(reqBody.Username)
	if err != nil {
		r.Logger.Error("Failed to get streamer info: %v", err)
		http.Error(w, "Streamer not found", http.StatusNotFound)
		return
	}

	// Add streamer to database
	if err := r.DB.AddStreamer(streamer); err != nil {
		r.Logger.Error("Failed to add streamer: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Log success
	r.Logger.Info("Added streamer: %s", streamer.DisplayName)

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(streamer)
}

// handleDeleteStreamer handles DELETE /api/streamers/{id}
func (r *Router) handleDeleteStreamer(w http.ResponseWriter, req *http.Request) {
	// Get streamer ID from URL
	vars := mux.Vars(req)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		r.Logger.Error("Invalid streamer ID: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Delete streamer from database
	if err := r.DB.DeleteStreamer(id); err != nil {
		r.Logger.Error("Failed to delete streamer: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Log success
	r.Logger.Info("Deleted streamer with ID: %d", id)

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// handleGetNotifications handles GET /api/notifications
func (r *Router) handleGetNotifications(w http.ResponseWriter, req *http.Request) {
	// Get notification settings
	notifications, err := r.DB.GetNotificationSettings()
	if err != nil {
		r.Logger.Error("Failed to get notification settings: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}

// handleAddNotification handles POST /api/notifications
func (r *Router) handleAddNotification(w http.ResponseWriter, req *http.Request) {
	// Parse request
	var notification models.NotificationSetting
	if err := json.NewDecoder(req.Body).Decode(&notification); err != nil {
		r.Logger.Error("Failed to parse request: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Add notification to database
	if err := r.DB.AddNotificationSetting(&notification); err != nil {
		r.Logger.Error("Failed to add notification: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Log success
	r.Logger.Info("Added notification setting: %s to %s", notification.Type, notification.Destination)

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(notification)
}

// handleUpdateNotification handles PUT /api/notifications/{id}
func (r *Router) handleUpdateNotification(w http.ResponseWriter, req *http.Request) {
	// Get notification ID from URL
	vars := mux.Vars(req)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		r.Logger.Error("Invalid notification ID: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Parse request
	var notification models.NotificationSetting
	if err := json.NewDecoder(req.Body).Decode(&notification); err != nil {
		r.Logger.Error("Failed to parse request: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Set ID from URL
	notification.ID = id

	// Update notification in database
	if err := r.DB.UpdateNotificationSetting(&notification); err != nil {
		r.Logger.Error("Failed to update notification: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Log success
	r.Logger.Info("Updated notification setting: %s to %s", notification.Type, notification.Destination)

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notification)
}

// handleDeleteNotification handles DELETE /api/notifications/{id}
func (r *Router) handleDeleteNotification(w http.ResponseWriter, req *http.Request) {
	// Get notification ID from URL
	vars := mux.Vars(req)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		r.Logger.Error("Invalid notification ID: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Delete notification from database
	if err := r.DB.DeleteNotificationSetting(id); err != nil {
		r.Logger.Error("Failed to delete notification: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Log success
	r.Logger.Info("Deleted notification setting with ID: %d", id)

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// handleGetLogs handles GET /api/logs
func (r *Router) handleGetLogs(w http.ResponseWriter, req *http.Request) {
	// Get logs
	logs := r.Logger.GetEntries()

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

// handleLogWebSocket handles WebSocket connections for live logs
func (r *Router) handleLogWebSocket(w http.ResponseWriter, req *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := r.upgrader.Upgrade(w, req, nil)
	if err != nil {
		r.Logger.Error("Failed to upgrade WebSocket connection: %v", err)
		return
	}
	defer conn.Close()

	// Subscribe to log events
	logCh := r.Logger.Subscribe()
	defer r.Logger.Unsubscribe(logCh)

	// Send initial logs
	logs := r.Logger.GetEntries()
	for _, log := range logs {
		if err := conn.WriteJSON(log); err != nil {
			break
		}
	}

	// Listen for new logs
	for log := range logCh {
		if err := conn.WriteJSON(log); err != nil {
			break
		}
	}
}