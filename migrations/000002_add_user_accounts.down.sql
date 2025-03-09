-- Remove indexes
DROP INDEX IF EXISTS idx_notification_settings_user_id;
DROP INDEX IF EXISTS idx_user_sessions_user_id;
DROP INDEX IF EXISTS idx_user_sessions_token;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_username;

-- Remove user_id from notification_settings
ALTER TABLE notification_settings
DROP COLUMN IF EXISTS user_id;

-- Drop user_sessions table
DROP TABLE IF EXISTS user_sessions;

-- Drop users table
DROP TABLE IF EXISTS users;