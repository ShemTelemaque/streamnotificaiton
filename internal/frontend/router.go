package frontend

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/drmaq/streamnotification/internal/config"
	"github.com/drmaq/streamnotification/internal/logger"
	"github.com/gorilla/mux"
)

// Router handles the web interface routes
type Router struct {
	Config    *config.Config
	Logger    *logger.Logger
	Router    *mux.Router
	templates *template.Template
	API       *APIClient
}

// NewRouter creates a new frontend router
func NewRouter(cfg *config.Config, logger *logger.Logger, apiBaseURL string) *Router {
	r := &Router{
		Config: cfg,
		Logger: logger,
		Router: mux.NewRouter(),
		API:    NewAPIClient(apiBaseURL, logger),
	}

	// Load templates
	r.loadTemplates()

	// Set up routes
	r.setupRoutes()

	return r
}

// loadTemplates loads HTML templates
func (r *Router) loadTemplates() {
	// Get the absolute path to the templates directory
	templatesDir, err := filepath.Abs(filepath.Join("web", "templates"))
	if err != nil {
		r.Logger.Error("Failed to get absolute templates path: %v", err)
		return
	}

	// Parse all templates, including those in subdirectories
	r.templates = template.Must(template.ParseGlob(filepath.Join(templatesDir, "*.html")))
	r.templates = template.Must(r.templates.ParseGlob(filepath.Join(templatesDir, "auth/*.html")))
}

// setupRoutes sets up the HTTP routes for the web interface
func (r *Router) setupRoutes() {
	// Static files
	fs := http.FileServer(http.Dir("./web/static"))
	r.Router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// Auth routes
	r.Router.HandleFunc("/login", r.handleLogin).Methods("GET", "POST")
	r.Router.HandleFunc("/register", r.handleRegister).Methods("GET", "POST")

	// Web interface routes
	r.Router.HandleFunc("/", r.handleIndex).Methods("GET")
	r.Router.HandleFunc("/streamers", r.handleStreamers).Methods("GET")
	r.Router.HandleFunc("/notifications", r.handleNotifications).Methods("GET")
	r.Router.HandleFunc("/logs", r.handleLogs).Methods("GET")

	// WebSocket route for live logs
	r.Router.HandleFunc("/ws/logs", r.handleLogWebSocket)
}
