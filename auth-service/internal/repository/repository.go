package repository

import (
	"auth-service/internal/domain"
	"context"
)

// UserRepository defines persistence operations for users.
type UserRepository interface {
	// CreateUser saves a new user and returns the assigned ID.
	CreateUser(ctx context.Context, user *domain.User) (int64, error)

	// GetUserByEmail returns a user by email, or ErrUserNotFound.
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)

	// GetUserByID returns a user by ID, or ErrUserNotFound.
	GetUserByID(ctx context.Context, id int64) (*domain.User, error)
}

// TokenRepository defines persistence operations for refresh tokens.
type TokenRepository interface {
	// SaveRefreshToken stores a refresh token for a user.
	SaveRefreshToken(ctx context.Context, token *domain.RefreshToken) error

	// GetRefreshToken returns a token by its string value, or ErrInvalidToken.
	GetRefreshToken(ctx context.Context, token string) (*domain.RefreshToken, error)

	// DeleteRefreshToken removes a token — used on logout or rotation.
	DeleteRefreshToken(ctx context.Context, token string) error
}
