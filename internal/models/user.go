package models

import (
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user account
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	Token    string `json:"token,omitempty"`
}

// CreateUser creates a new user account
func (u *User) CreateUser(db *sql.DB) error {
	query := `
		INSERT INTO users (username, password_hash)
		VALUES ($1, $2)
		RETURNING id`

	return db.QueryRow(
		query,
		u.Username,
		u.Password,
	).Scan(&u.ID)
}

// GetUserByUsername retrieves a user by their username
func GetUserByUsername(db *sql.DB, username string) (*User, error) {
	user := &User{}
	query := `SELECT id, username, password_hash FROM users WHERE username = $1`

	err := db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

// HashPassword creates a bcrypt hash of a password
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword verifies a password against its hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}