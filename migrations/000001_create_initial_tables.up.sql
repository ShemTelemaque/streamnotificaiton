-- Create streamers table
CREATE TABLE IF NOT EXISTS streamers (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL,
    is_live BOOLEAN NOT NULL DEFAULT false,
    last_stream_start TIMESTAMP,
    last_notification_sent TIMESTAMP
);

-- Create notification_settings table
CREATE TABLE IF NOT EXISTS notification_settings (
    id SERIAL PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    destination VARCHAR(255) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true
);

-- Create indexes
CREATE INDEX idx_streamers_username ON streamers(username);
CREATE INDEX idx_notification_settings_type ON notification_settings(type);