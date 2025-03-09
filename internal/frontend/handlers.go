package frontend

import (
	"net/http"
	"time"

	"github.com/drmaq/streamnotification/internal/models"
	"github.com/gorilla/websocket"
)

// handleIndex handles the index page
func (r *Router) handleIndex(w http.ResponseWriter, req *http.Request) {
	// Get streamers from API
	streamers, err := r.API.GetStreamers()
	if err != nil {
		r.Logger.Error("Failed to get streamers: %v", err)
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

	// Get notification settings from API
	notifications, err := r.API.GetNotificationSettings()
	if err != nil {
		r.Logger.Error("Failed to get notification settings: %v", err)
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

	r.templates.ExecuteTemplate(w, "index.html", data)
}

// handleStreamers handles the streamers page
func (r *Router) handleStreamers(w http.ResponseWriter, req *http.Request) {
	// Get streamers from API
	streamers, err := r.API.GetStreamers()
	if err != nil {
		r.Logger.Error("Failed to get streamers: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Render template
	data := map[string]interface{}{
		"Streamers": streamers,
	}

	r.templates.ExecuteTemplate(w, "streamers.html", data)
}

// handleNotifications handles the notifications page
func (r *Router) handleNotifications(w http.ResponseWriter, req *http.Request) {
	// Get notification settings from API
	notifications, err := r.API.GetNotificationSettings()
	if err != nil {
		r.Logger.Error("Failed to get notification settings: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Render template
	data := map[string]interface{}{
		"Notifications": notifications,
	}

	r.templates.ExecuteTemplate(w, "notifications.html", data)
}

// handleLogs handles the logs page
func (r *Router) handleLogs(w http.ResponseWriter, req *http.Request) {
	// Get logs from API
	logs, err := r.API.GetLogs()
	if err != nil {
		r.Logger.Error("Failed to get logs: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Render template
	data := map[string]interface{}{
		"Logs": logs,
	}

	r.templates.ExecuteTemplate(w, "logs.html", data)
}

// handleLogWebSocket handles WebSocket connections for live logs
func (r *Router) handleLogWebSocket(w http.ResponseWriter, req *http.Request) {
	// Create a WebSocket upgrader
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins in development
		},
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		r.Logger.Error("Failed to upgrade WebSocket connection: %v", err)
		return
	}
	defer conn.Close()

	// Create a proxy WebSocket connection to the backend
	backendURL := r.API.BaseURL + "/ws/logs"
	backendURL = "ws" + backendURL[4:] // Replace http with ws

	r.Logger.Info("Connecting to backend WebSocket: %s", backendURL)

	// Connect to backend WebSocket
	backendConn, _, err := websocket.DefaultDialer.Dial(backendURL, nil)
	if err != nil {
		r.Logger.Error("Failed to connect to backend WebSocket: %v", err)
		return
	}
	defer backendConn.Close()

	// Forward messages from backend to client
	go func() {
		for {
			// Read message from backend
			messageType, message, err := backendConn.ReadMessage()
			if err != nil {
				r.Logger.Error("Error reading from backend WebSocket: %v", err)
				break
			}

			// Write message to client
			if err := conn.WriteMessage(messageType, message); err != nil {
				r.Logger.Error("Error writing to client WebSocket: %v", err)
				break
			}
		}
	}()

	// Forward messages from client to backend
	for {
		// Read message from client
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			r.Logger.Error("Error reading from client WebSocket: %v", err)
			break
		}

		// Write message to backend
		if err := backendConn.WriteMessage(messageType, message); err != nil {
			r.Logger.Error("Error writing to backend WebSocket: %v", err)
			break
		}
	}
}

// handleLogin handles the login page
func (r *Router) handleLogin(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		r.templates.ExecuteTemplate(w, "auth/login.html", nil)
		return
	}

	// Handle POST request
	if err := req.ParseForm(); err != nil {
		r.Logger.Error("Failed to parse login form: %v", err)
		data := map[string]interface{}{
			"Error": "Invalid form data",
		}
		r.templates.ExecuteTemplate(w, "auth/login.html", data)
		return
	}

	// Call API to authenticate user
	response, err := r.API.Login(req.Form.Get("username"), req.Form.Get("password"))
	if err != nil {
		r.Logger.Error("Failed to login: %v", err)
		data := map[string]interface{}{
			"Error": "Invalid credentials",
		}
		r.templates.ExecuteTemplate(w, "auth/login.html", data)
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    response.Token,
		Path:     "/",
		HttpOnly: true,
	})

	// Redirect to home page
	http.Redirect(w, req, "/", http.StatusSeeOther)
}

// handleRegister handles the register page
func (r *Router) handleRegister(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		r.templates.ExecuteTemplate(w, "auth/register.html", nil)
		return
	}

	// Handle POST request
	if err := req.ParseForm(); err != nil {
		r.Logger.Error("Failed to parse register form: %v", err)
		data := map[string]interface{}{
			"Error": "Invalid form data",
		}
		r.templates.ExecuteTemplate(w, "auth/register.html", data)
		return
	}

	// Call API to register user
	response, err := r.API.Register(req.Form.Get("username"), req.Form.Get("password"))
	if err != nil {
		r.Logger.Error("Failed to register: %v", err)
		data := map[string]interface{}{
			"Error": "Registration failed",
		}
		r.templates.ExecuteTemplate(w, "auth/register.html", data)
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    response.Token,
		Path:     "/",
		HttpOnly: true,
	})

	// Redirect to home page
	http.Redirect(w, req, "/", http.StatusSeeOther)
}
