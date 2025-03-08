package db

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/drmaq/streamnotification/internal/config"
	"github.com/drmaq/streamnotification/internal/models"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/lib/pq"
	//"github.com/golang-migrate/migrate/v4/source/file"
)

// Database represents a database connection
type Database struct {
	db *sql.DB
}

// NewDatabase creates a new database connection
func NewDatabase(cfg *config.Config) (*Database, error) {
	// Create connection string
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Check connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// Migrate runs database migrations
func (d *Database) Migrate() error {
	// Get the path to the migrations directory
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	migrationsPath := filepath.Join(basepath, "../..", "migrations")

	// Create a new migrate instance
	driver, err := postgres.WithInstance(d.db, &postgres.Config{})
	if err != nil {
		return err
	}

	// Create a new migrate instance
	sourceURL := fmt.Sprintf("file://%s", migrationsPath)
	m, err := migrate.NewWithDatabaseInstance(sourceURL, "postgres", driver)
	if err != nil {
		return err
	}

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

// GetStreamers returns all streamers from the database
func (d *Database) GetStreamers() ([]models.Streamer, error) {
	rows, err := d.db.Query("SELECT id, username, display_name, is_live, last_stream_start, last_notification_sent FROM streamers")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var streamers []models.Streamer
	for rows.Next() {
		var s models.Streamer
		if err := rows.Scan(&s.ID, &s.Username, &s.DisplayName, &s.IsLive, &s.LastStreamStart, &s.LastNotificationSent); err != nil {
			return nil, err
		}
		streamers = append(streamers, s)
	}

	return streamers, nil
}

// AddStreamer adds a new streamer to the database
func (d *Database) AddStreamer(streamer *models.Streamer) error {
	query := `
		INSERT INTO streamers (username, display_name, is_live, last_stream_start, last_notification_sent)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	return d.db.QueryRow(
		query,
		streamer.Username,
		streamer.DisplayName,
		streamer.IsLive,
		streamer.LastStreamStart,
		streamer.LastNotificationSent,
	).Scan(&streamer.ID)
}

// UpdateStreamer updates a streamer in the database
func (d *Database) UpdateStreamer(streamer *models.Streamer) error {
	query := `
		UPDATE streamers
		SET username = $1, display_name = $2, is_live = $3, last_stream_start = $4, last_notification_sent = $5
		WHERE id = $6
	`

	_, err := d.db.Exec(
		query,
		streamer.Username,
		streamer.DisplayName,
		streamer.IsLive,
		streamer.LastStreamStart,
		streamer.LastNotificationSent,
		streamer.ID,
	)

	return err
}

// DeleteStreamer deletes a streamer from the database
func (d *Database) DeleteStreamer(id int) error {
	_, err := d.db.Exec("DELETE FROM streamers WHERE id = $1", id)
	return err
}

// GetNotificationSettings returns all notification settings from the database
func (d *Database) GetNotificationSettings() ([]models.NotificationSetting, error) {
	rows, err := d.db.Query("SELECT id, type, destination, enabled FROM notification_settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []models.NotificationSetting
	for rows.Next() {
		var s models.NotificationSetting
		if err := rows.Scan(&s.ID, &s.Type, &s.Destination, &s.Enabled); err != nil {
			return nil, err
		}
		settings = append(settings, s)
	}

	return settings, nil
}

// AddNotificationSetting adds a new notification setting to the database
func (d *Database) AddNotificationSetting(setting *models.NotificationSetting) error {
	query := `
		INSERT INTO notification_settings (type, destination, enabled)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	return d.db.QueryRow(
		query,
		setting.Type,
		setting.Destination,
		setting.Enabled,
	).Scan(&setting.ID)
}

// UpdateNotificationSetting updates a notification setting in the database
func (d *Database) UpdateNotificationSetting(setting *models.NotificationSetting) error {
	query := `
		UPDATE notification_settings
		SET type = $1, destination = $2, enabled = $3
		WHERE id = $4
	`

	_, err := d.db.Exec(
		query,
		setting.Type,
		setting.Destination,
		setting.Enabled,
		setting.ID,
	)

	return err
}

// DeleteNotificationSetting deletes a notification setting from the database
func (d *Database) DeleteNotificationSetting(id int) error {
	_, err := d.db.Exec("DELETE FROM notification_settings WHERE id = $1", id)
	return err
}
