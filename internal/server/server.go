package server

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/drmaq/streamnotification/internal/config"
	"github.com/drmaq/streamnotification/internal/db"
	"github.com/drmaq/streamnotification/internal/logger"
	"github.com/drmaq/streamnotification/internal/models"
	"github.com/drmaq/streamnotification/internal/twitch"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Server represents the HTTP server
type Server struct {
	Config       *config.Config
	Logger       *logger.Logger
	DB           *db.Database
	TwitchClient *twitch.Client
	Router       *mux.Router
	templates    *template.Template
	upgrader     websocket.Upgrader
}

// NewServer creates a new HTTP server
func NewServer(cfg *config.Config, logger *logger.Logger, database *db.Database, twitchClient *twitch.Client) *Server {
	s := &Server{
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

	// Load templates
	s.loadTemplates()

	// Set up routes
	s.setupRoutes()

	return s
}

// loadTemplates loads HTML templates
func (s *Server) loadTemplates() {
	// Get the path to the templates directory
	templatesDir := filepath.Join("web", "templates")
	s.templates = template.Must(template.ParseGlob(filepath.Join(templatesDir, "*.html")))
}

// setupRoutes sets up the HTTP routes
func (s *Server) setupRoutes() {
	// Static files
	fs := http.FileServer(http.Dir("./web/static"))
	s.Router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// Web interface routes
	s.Router.HandleFunc("/", s.handleIndex).Methods("GET")
	s.Router.HandleFunc("/streamers", s.handleStreamers).Methods("GET")
	s.Router.HandleFunc("/notifications", s.handleNotifications).Methods("GET")
	s.Router.HandleFunc("/logs", s.handleLogs).Methods("GET")

	// API routes
	api := s.Router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/streamers", s.handleGetStreamers).Methods("GET")
	api.HandleFunc("/streamers", s.handleAddStreamer).Methods("POST")
	api.HandleFunc("/streamers/{id:[0-9]+}", s.handleDeleteStreamer).Methods("DELETE")
	api.HandleFunc("/notifications", s.handleGetNotifications).Methods("GET")
	api.HandleFunc("/notifications", s.handleAddNotification).Methods("POST")
	api.HandleFunc("/notifications/{id:[0-9]+}", s.handleUpdateNotification).Methods("PUT")
	api.HandleFunc("/notifications/{id:[0-9]+}", s.handleDeleteNotification).Methods("DELETE")

	// WebSocket route for live logs
	s.Router.HandleFunc("/ws/logs", s.handleLogWebSocket)
}

// handleIndex handles the index page
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	// Get streamers
	streamers, err := s.DB.GetStreamers()
	if err != nil {
		s.Logger.Error("Failed to get streamers: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Count live streamers
	liveCount := 0
	for _, streamer := range streamers {
		if streamer.IsLive {
			liveCount++
		}
	}

	// Get notification settings
	notifications, err := s.DB.GetNotificationSettings()
	if err != nil {
		s.Logger.Error("Failed to get notification settings: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Count enabled notifications
	discordCount := 0
	twitterCount := 0
	for _, notification := range notifications {
		if notification.Enabled {
			if notification.Type == models.NotificationTypeDiscord {
				discordCount++
			} else if notification.Type == models.NotificationTypeTwitter {
				twitterCount++
			}
		}
	}

	// Render template
	data := map[string]interface{}{
		"StreamerCount":  len(streamers),
		"LiveCount":      liveCount,
		"DiscordCount":   discordCount,
		"TwitterCount":   twitterCount,
		"LastUpdateTime": time.Now().Format("2006-01-02 15:04:05"),
	}

	s.templates.ExecuteTemplate(w, "index.html", data)
}

// handleStreamers handles the streamers page
func (s *Server) handleStreamers(w http.ResponseWriter, r *http.Request) {
	// Get streamers
	streamers, err := s.DB.GetStreamers()
	if err != nil {
		s.Logger.Error("Failed to get streamers: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Render template
	data := map[string]interface{}{
		"Streamers": streamers,
	}

	s.templates.ExecuteTemplate(w, "streamers.html", data)
}

// handleNotifications handles the notifications page
func (s *Server) handleNotifications(w http.ResponseWriter, r *http.Request) {
	// Get notification settings
	notifications, err := s.DB.GetNotificationSettings()
	if err != nil {
		s.Logger.Error("Failed to get notification settings: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Render template
	data := map[string]interface{}{
		"Notifications": notifications,
	}

	s.templates.ExecuteTemplate(w, "notifications.html", data)
}

// handleLogs handles the logs page
func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	// Get logs
	logs := s.Logger.GetEntries()

	// Render template
	data := map[string]interface{}{
		"Logs": logs,
	}

	s.templates.ExecuteTemplate(w, "logs.html", data)
}

// handleLogWebSocket handles WebSocket connections for live logs
func (s *Server) handleLogWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.Logger.Error("Failed to upgrade WebSocket connection: %v", err)
		return
	}
	defer conn.Close()

	// Subscribe to log events
	logCh := s.Logger.Subscribe()
	defer s.Logger.Unsubscribe(logCh)

	// Send initial logs
	logs := s.Logger.GetEntries()
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

// API Handlers

// handleGetStreamers handles GET /api/streamers
func (s *Server) handleGetStreamers(w http.ResponseWriter, r *http.Request) {
	// Get streamers
	streamers, err := s.DB.GetStreamers()
	if err != nil {
		s.Logger.Error("Failed to get streamers: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(streamers)
}

// handleAddStreamer handles POST /api/streamers
func (s *Server) handleAddStreamer(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req struct {
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.Logger.Error("Failed to parse request: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Get streamer info from Twitch
	streamer, err := s.TwitchClient.GetStreamerInfo(req.Username)
	if err != nil {
		s.Logger.Error("Failed to get streamer info: %v", err)
		http.Error(w, "Streamer not found", http.StatusNotFound)
		return
	}

	// Add streamer to database
	if err := s.DB.AddStreamer(streamer); err != nil {
		s.Logger.Error("Failed to add streamer: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Log success
	s.Logger.Info("Added streamer: %s", streamer.DisplayName)

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(streamer)
}

// handleDeleteStreamer handles DELETE /api/streamers/{id}
func (s *Server) handleDeleteStreamer(w http.ResponseWriter, r *http.Request) {
	// Get streamer ID from URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		s.Logger.Error("Invalid streamer ID: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Delete streamer from database
	if err := s.DB.DeleteStreamer(id); err != nil {
		s.Logger.Error("Failed to delete streamer: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Log success
	s.Logger.Info("Deleted streamer with ID: %d", id)

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// handleGetNotifications handles GET /api/notifications
func (s *Server) handleGetNotifications(w http.ResponseWriter, r *http.Request) {
	// Get notification settings
	notifications, err := s.DB.GetNotificationSettings()
	if err != nil {
		s.Logger.Error("Failed to get notification settings: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}

// handleAddNotification handles POST /api/notifications
func (s *Server) handleAddNotification(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var notification models.NotificationSetting
	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		s.Logger.Error("Failed to parse request: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Add notification to database
	if err := s.DB.AddNotificationSetting(&notification); err != nil {
		s.Logger.Error("Failed to add notification: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Log success
	s.Logger.Info("Added notification setting: %s to %s", notification.Type, notification.Destination)

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(notification)
}

// handleUpdateNotification handles PUT /api/notifications/{id}
func (s *Server) handleUpdateNotification(w http.ResponseWriter, r *http.Request) {
	// Get notification ID from URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		s.Logger.Error("Invalid notification ID: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Parse request
	var notification models.NotificationSetting
	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		s.Logger.Error("Failed to parse request: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Set ID from URL
	notification.ID = id

	// Update notification in database
	if err := s.DB.UpdateNotificationSetting(&notification); err != nil {
		s.Logger.Error("Failed to update notification: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Log success
	s.Logger.Info("Updated notification setting: %s to %s", notification.Type, notification.Destination)

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notification)
}

// handleDeleteNotification handles DELETE /api/notifications/{id}
func (s *Server) handleDeleteNotification(w http.ResponseWriter, r *http.Request) {
	// Get notification ID from URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		s.Logger.Error("Invalid notification ID: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Delete notification from database
	if err := s.DB.DeleteNotificationSetting(id); err != nil {
		s.Logger.Error("Failed to delete notification: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Log success
	s.Logger.Info("Deleted notification setting with ID: %d", id)

	// Return success
	w.WriteHeader(http.StatusNoContent)
}