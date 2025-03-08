package logger

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// Logger is a custom logger that supports different log levels and can be displayed on the web interface
type Logger struct {
	stdLogger *log.Logger
	mu        sync.Mutex
	entries   []LogEntry
	listeners []chan LogEntry
}

// LogEntry represents a single log message
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     LogLevel  `json:"level"`
	Message   string    `json:"message"`
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	return &Logger{
		stdLogger: log.New(os.Stdout, "", log.LstdFlags),
		entries:   make([]LogEntry, 0),
		listeners: make([]chan LogEntry, 0),
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, v ...interface{}) {
	l.log(DebugLevel, format, v...)
}

// Info logs an info message
func (l *Logger) Info(format string, v ...interface{}) {
	l.log(InfoLevel, format, v...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, v ...interface{}) {
	l.log(WarnLevel, format, v...)
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	l.log(ErrorLevel, format, v...)
}

// Fatal logs a fatal message and exits the application
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.log(FatalLevel, format, v...)
	os.Exit(1)
}

// log logs a message with the specified level
func (l *Logger) log(level LogLevel, format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
	}

	// Log to standard logger
	l.stdLogger.Printf("[%s] %s", l.levelToString(level), message)

	// Store log entry and notify listeners
	l.mu.Lock()
	defer l.mu.Unlock()

	// Add to entries (limit to last 1000 entries)
	l.entries = append(l.entries, entry)
	if len(l.entries) > 1000 {
		l.entries = l.entries[len(l.entries)-1000:]
	}

	// Notify all listeners
	for _, listener := range l.listeners {
		select {
		case listener <- entry:
			// Successfully sent
		default:
			// Channel is full or closed, skip
		}
	}
}

// Subscribe returns a channel that receives log entries
func (l *Logger) Subscribe() chan LogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Create a buffered channel to avoid blocking
	ch := make(chan LogEntry, 100)
	l.listeners = append(l.listeners, ch)

	return ch
}

// Unsubscribe removes a channel from the listeners
func (l *Logger) Unsubscribe(ch chan LogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for i, listener := range l.listeners {
		if listener == ch {
			// Remove the listener
			l.listeners = append(l.listeners[:i], l.listeners[i+1:]...)
			close(ch)
			break
		}
	}
}

// GetEntries returns all stored log entries
func (l *Logger) GetEntries() []LogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Return a copy of the entries
	entries := make([]LogEntry, len(l.entries))
	copy(entries, l.entries)

	return entries
}

// levelToString converts a LogLevel to its string representation
func (l *Logger) levelToString(level LogLevel) string {
	switch level {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}