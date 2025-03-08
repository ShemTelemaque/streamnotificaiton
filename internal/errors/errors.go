package errors

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/drmaq/streamnotification/internal/logger"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// ErrorTypeDatabase represents a database error
	ErrorTypeDatabase ErrorType = "database"
	// ErrorTypeAPI represents an API error
	ErrorTypeAPI ErrorType = "api"
	// ErrorTypeConfig represents a configuration error
	ErrorTypeConfig ErrorType = "config"
	// ErrorTypeInternal represents an internal server error
	ErrorTypeInternal ErrorType = "internal"
	// ErrorTypeValidation represents a validation error
	ErrorTypeValidation ErrorType = "validation"
	// ErrorTypeNotFound represents a not found error
	ErrorTypeNotFound ErrorType = "not_found"
	// ErrorTypeUnauthorized represents an unauthorized error
	ErrorTypeUnauthorized ErrorType = "unauthorized"
)

// AppError represents an application error
type AppError struct {
	Type    ErrorType `json:"type"`
	Message string    `json:"message"`
	Err     error     `json:"error,omitempty"`
	Status  int       `json:"status,omitempty"`
}

// Error returns the error message
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// StatusCode returns the HTTP status code for the error
func (e *AppError) StatusCode() int {
	if e.Status != 0 {
		return e.Status
	}

	// Default status codes based on error type
	switch e.Type {
	case ErrorTypeDatabase, ErrorTypeInternal:
		return http.StatusInternalServerError
	case ErrorTypeAPI:
		return http.StatusBadGateway
	case ErrorTypeConfig:
		return http.StatusInternalServerError
	case ErrorTypeValidation:
		return http.StatusBadRequest
	case ErrorTypeNotFound:
		return http.StatusNotFound
	case ErrorTypeUnauthorized:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

// LogError logs the error with the appropriate level
func LogError(log *logger.Logger, err error) {
	appErr, ok := err.(*AppError)
	if !ok {
		log.Error("Error: %v", err)
		return
	}

	switch appErr.Type {
	case ErrorTypeDatabase:
		log.Error("Database error: %v", appErr)
	case ErrorTypeAPI:
		log.Error("API error: %v", appErr)
	case ErrorTypeConfig:
		log.Error("Configuration error: %v", appErr)
	case ErrorTypeInternal:
		log.Error("Internal error: %v", appErr)
	case ErrorTypeValidation:
		log.Warn("Validation error: %v", appErr)
	case ErrorTypeNotFound:
		log.Warn("Not found error: %v", appErr)
	case ErrorTypeUnauthorized:
		log.Warn("Unauthorized error: %v", appErr)
	default:
		log.Error("Unknown error: %v", appErr)
	}
}

// HandleHTTPError writes the error to the HTTP response
func HandleHTTPError(w http.ResponseWriter, err error, log *logger.Logger) {
	// Log the error
	LogError(log, err)

	// Determine status code and message
	statusCode := http.StatusInternalServerError
	message := "Internal server error"

	appErr, ok := err.(*AppError)
	if ok {
		statusCode = appErr.StatusCode()
		message = appErr.Message
	}

	// Write error response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, `{"error":"%s"}`, message)
}

// NewDatabaseError creates a new database error
func NewDatabaseError(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeDatabase,
		Message: message,
		Err:     err,
		Status:  http.StatusInternalServerError,
	}
}

// NewAPIError creates a new API error
func NewAPIError(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeAPI,
		Message: message,
		Err:     err,
		Status:  http.StatusBadGateway,
	}
}

// NewConfigError creates a new configuration error
func NewConfigError(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeConfig,
		Message: message,
		Err:     err,
		Status:  http.StatusInternalServerError,
	}
}

// NewInternalError creates a new internal error
func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeInternal,
		Message: message,
		Err:     err,
		Status:  http.StatusInternalServerError,
	}
}

// NewValidationError creates a new validation error
func NewValidationError(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeValidation,
		Message: message,
		Err:     err,
		Status:  http.StatusBadRequest,
	}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeNotFound,
		Message: message,
		Err:     err,
		Status:  http.StatusNotFound,
	}
}

// NewUnauthorizedError creates a new unauthorized error
func NewUnauthorizedError(message string, err error) *AppError {
	return &AppError{
		Type:    ErrorTypeUnauthorized,
		Message: message,
		Err:     err,
		Status:  http.StatusUnauthorized,
	}
}

// IsDatabaseError checks if the error is a database error
func IsDatabaseError(err error) bool {
	appErr, ok := err.(*AppError)
	return ok && appErr.Type == ErrorTypeDatabase
}

// IsAPIError checks if the error is an API error
func IsAPIError(err error) bool {
	appErr, ok := err.(*AppError)
	return ok && appErr.Type == ErrorTypeAPI
}

// IsConfigError checks if the error is a configuration error
func IsConfigError(err error) bool {
	appErr, ok := err.(*AppError)
	return ok && appErr.Type == ErrorTypeConfig
}

// IsInternalError checks if the error is an internal error
func IsInternalError(err error) bool {
	appErr, ok := err.(*AppError)
	return ok && appErr.Type == ErrorTypeInternal
}

// IsValidationError checks if the error is a validation error
func IsValidationError(err error) bool {
	appErr, ok := err.(*AppError)
	return ok && appErr.Type == ErrorTypeValidation
}

// IsNotFoundError checks if the error is a not found error
func IsNotFoundError(err error) bool {
	appErr, ok := err.(*AppError)
	return ok && appErr.Type == ErrorTypeNotFound
}

// IsUnauthorizedError checks if the error is an unauthorized error
func IsUnauthorizedError(err error) bool {
	appErr, ok := err.(*AppError)
	return ok && appErr.Type == ErrorTypeUnauthorized
}

// IsErrorType checks if the error is of a specific type
func IsErrorType(err error, errorType ErrorType) bool {
	appErr, ok := err.(*AppError)
	return ok && appErr.Type == errorType
}

// ContainsError checks if the error message contains a specific string
func ContainsError(err error, substr string) bool {
	return err != nil && strings.Contains(err.Error(), substr)
}