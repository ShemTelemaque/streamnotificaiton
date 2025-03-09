package db

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/drmaq/streamnotification/internal/config"
	"github.com/drmaq/streamnotification/internal/errors"
	"github.com/drmaq/streamnotification/internal/models"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
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
		return nil, errors.NewDatabaseError("Failed to open database connection", err)
	}

	// Check connection
	if err := db.Ping(); err != nil {
		return nil, errors.NewDatabaseError("Failed to connect to database", err)
	}

	return &Database{db: db}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	err := d.db.Close()
	if err != nil {
		return errors.NewDatabaseError("Failed to close database connection", err)
	}
	return nil
}

// Migrate runs database migrations
func (d *Database) Migrate() error {
	// Use relative path for migrations
	migrationsPath := "migrations"

	// Register the file source driver
	source, err := (&file.File{}).Open(fmt.Sprintf("file://./%s", filepath.ToSlash(filepath.Clean(migrationsPath))))
	if err != nil {
		return errors.NewDatabaseError("Failed to create file source driver", err)
	}
	defer source.Close()

	// Create a new migrate instance
	driver, err := postgres.WithInstance(d.db, &postgres.Config{})
	if err != nil {
		return errors.NewDatabaseError("Failed to create migration driver", err)
	}

	// Create a new migrate instance
	sourceURL := fmt.Sprintf("file://./%s", filepath.ToSlash(filepath.Clean(migrationsPath)))
	m, err := migrate.NewWithDatabaseInstance(sourceURL, "postgres", driver)
	if err != nil {
		return errors.NewDatabaseError("Failed to create migration instance", err)
	}

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return errors.NewDatabaseError("Failed to run migrations", err)
	}

	return nil
}

// GetStreamers returns all streamers from the database
func (d *Database) GetStreamers() ([]models.Streamer, error) {
	rows, err := d.db.Query("SELECT id, username, display_name, is_live, last_stream_start, last_notification_sent FROM streamers")
	if err != nil {
		return nil, errors.NewDatabaseError("Failed to query streamers", err)
	}
	defer rows.Close()

	var streamers []models.Streamer
	for rows.Next() {
		var s models.Streamer
		if err := rows.Scan(&s.ID, &s.Username, &s.DisplayName, &s.IsLive, &s.LastStreamStart, &s.LastNotificationSent); err != nil {
			return nil, errors.NewDatabaseError("Failed to scan streamer row", err)
		}
		streamers = append(streamers, s)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.NewDatabaseError("Error iterating streamer rows", err)
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

	err := d.db.QueryRow(
		query,
		streamer.Username,
		streamer.DisplayName,
		streamer.IsLive,
		streamer.LastStreamStart,
		streamer.LastNotificationSent,
	).Scan(&streamer.ID)

	if err != nil {
		return errors.NewDatabaseError("Failed to add streamer", err)
	}

	return nil
}

// UpdateStreamer updates a streamer in the database
func (d *Database) UpdateStreamer(streamer *models.Streamer) error {
	query := `
		UPDATE streamers
		SET username = $1, display_name = $2, is_live = $3, last_stream_start = $4, last_notification_sent = $5
		WHERE id = $6
	`

	result, err := d.db.Exec(
		query,
		streamer.Username,
		streamer.DisplayName,
		streamer.IsLive,
		streamer.LastStreamStart,
		streamer.LastNotificationSent,
		streamer.ID,
	)

	if err != nil {
		return errors.NewDatabaseError("Failed to update streamer", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseError("Failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("Streamer not found", nil)
	}

	return nil
}

// DeleteStreamer deletes a streamer from the database
func (d *Database) DeleteStreamer(id int) error {
	result, err := d.db.Exec("DELETE FROM streamers WHERE id = $1", id)
	if err != nil {
		return errors.NewDatabaseError("Failed to delete streamer", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseError("Failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("Streamer not found", nil)
	}

	return nil
}

// GetNotificationSettings returns all notification settings from the database
func (d *Database) GetNotificationSettings() ([]models.NotificationSetting, error) {
	rows, err := d.db.Query("SELECT id, type, destination, enabled FROM notification_settings")
	if err != nil {
		return nil, errors.NewDatabaseError("Failed to query notification settings", err)
	}
	defer rows.Close()

	var settings []models.NotificationSetting
	for rows.Next() {
		var s models.NotificationSetting
		if err := rows.Scan(&s.ID, &s.Type, &s.Destination, &s.Enabled); err != nil {
			return nil, errors.NewDatabaseError("Failed to scan notification setting row", err)
		}
		settings = append(settings, s)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.NewDatabaseError("Error iterating notification setting rows", err)
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

	err := d.db.QueryRow(
		query,
		setting.Type,
		setting.Destination,
		setting.Enabled,
	).Scan(&setting.ID)

	if err != nil {
		return errors.NewDatabaseError("Failed to add notification setting", err)
	}

	return nil
}

// UpdateNotificationSetting updates a notification setting in the database
func (d *Database) UpdateNotificationSetting(setting *models.NotificationSetting) error {
	query := `
		UPDATE notification_settings
		SET type = $1, destination = $2, enabled = $3
		WHERE id = $4
	`

	result, err := d.db.Exec(
		query,
		setting.Type,
		setting.Destination,
		setting.Enabled,
		setting.ID,
	)

	if err != nil {
		return errors.NewDatabaseError("Failed to update notification setting", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseError("Failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("Notification setting not found", nil)
	}

	return nil
}

// DeleteNotificationSetting deletes a notification setting from the database
func (d *Database) DeleteNotificationSetting(id int) error {
	result, err := d.db.Exec("DELETE FROM notification_settings WHERE id = $1", id)
	if err != nil {
		return errors.NewDatabaseError("Failed to delete notification setting", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseError("Failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("Notification setting not found", nil)
	}

	return nil
}
