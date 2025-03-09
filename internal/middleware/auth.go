package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/drmaq/streamnotification/internal/logger"
	"github.com/drmaq/streamnotification/internal/models"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
)

// AuthMiddleware handles authentication for protected routes
type AuthMiddleware struct {
	DB     *sql.DB
	Logger *logger.Logger
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(db *sql.DB, logger *logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		DB:     db,
		Logger: logger,
	}
}

// RequireAuth middleware checks if the user is authenticated
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get session token from cookie
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Get session from database

		type Session struct {
			ID        int64
			UserID    int64
			Token     string
			ExpiresAt time.Time
		}

		session := &Session{}
		err = m.DB.QueryRow("SELECT id, user_id, token, expires_at FROM sessions WHERE token = $1", cookie.Value).Scan(&session.ID, &session.UserID, &session.Token, &session.ExpiresAt)
		if err != nil || session == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Check if session is expired
		if time.Now().After(session.ExpiresAt) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Get user from database
		user := &models.User{}
		err = m.DB.QueryRow("SELECT id, username FROM users WHERE id = $1", session.UserID).Scan(&user.ID, &user.Username)
		if err != nil || user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Add user to request context
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole middleware checks if the user has the required role
func (m *AuthMiddleware) RequireRole(role string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(UserContextKey).(*models.User)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}

		// Query the user's role from the database
		var userRole string
		err := m.DB.QueryRow("SELECT role FROM user_roles WHERE user_id = $1", user.ID).Scan(&userRole)
		if err != nil || userRole != role {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetUserFromContext retrieves the user from the request context
func GetUserFromContext(ctx context.Context) *models.User {
	user, ok := ctx.Value(UserContextKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}
