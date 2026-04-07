package domain

import (
	"errors"
	"time"
)

// User represents a registered user in the system.
type User struct {
	ID           int64
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

// RefreshToken represents a stored refresh token tied to a user.
type RefreshToken struct {
	ID        int64
	UserID    int64
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// Sentinel errors for the auth domain.
// Use errors.Is() to check for these in service and handler layers.
var (
	// ErrUserNotFound is returned when a user lookup yields no result.
	ErrUserNotFound = errors.New("user not found")

	// ErrUserAlreadyExists is returned on registration when the email is taken.
	ErrUserAlreadyExists = errors.New("user already exists")

	// ErrInvalidCredentials is returned when email/password do not match.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrInvalidToken is returned when a token fails validation.
	ErrInvalidToken = errors.New("invalid token")

	// ErrTokenExpired is returned when a token exists but has passed its expiry.
	ErrTokenExpired = errors.New("token expired")
)
