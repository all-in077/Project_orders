package postgres

import (
	"auth-service/internal/domain"
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

// UserRepo implements repository.UserRepository and repository.TokenRepository
// using a PostgreSQL database.
type UserRepo struct {
	db *sqlx.DB
}

// NewUserRepo creates a UserRepo backed by the given sqlx connection.
func NewUserRepo(db *sqlx.DB) *UserRepo {
	return &UserRepo{db: db}
}

// CreateUser inserts a new user row and returns the generated ID.
// Returns domain.ErrUserAlreadyExists if the email is already taken.
func (r *UserRepo) CreateUser(ctx context.Context, user *domain.User) (int64, error) {
	query := `
		INSERT INTO users (email, password_hash, created_at)
		VALUES ($1, $2, $3)
		RETURNING id`

	var id int64
	err := r.db.QueryRowContext(ctx, query,
		user.Email,
		user.PasswordHash,
		time.Now(),
	).Scan(&id)
	if err != nil {
		if isUniqueViolation(err) {
			return 0, domain.ErrUserAlreadyExists
		}
		return 0, err
	}
	return id, nil
}

// GetUserByEmail returns the user with the given email.
// Returns domain.ErrUserNotFound if no such user exists.
func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE email = $1`

	var u domain.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

// GetUserByID returns the user with the given ID.
// Returns domain.ErrUserNotFound if no such user exists.
func (r *UserRepo) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE id = $1`

	var u domain.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

// SaveRefreshToken inserts a refresh token row for the given user.
func (r *UserRepo) SaveRefreshToken(ctx context.Context, token *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4)`

	_, err := r.db.ExecContext(ctx, query,
		token.UserID,
		token.Token,
		token.ExpiresAt,
		time.Now(),
	)
	return err
}

// GetRefreshToken returns the stored token matching the given string.
// Returns domain.ErrInvalidToken if not found.
func (r *UserRepo) GetRefreshToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM refresh_tokens
		WHERE token = $1`

	var t domain.RefreshToken
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&t.ID,
		&t.UserID,
		&t.Token,
		&t.ExpiresAt,
		&t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvalidToken
		}
		return nil, err
	}
	return &t, nil
}

// DeleteRefreshToken removes the token row matching the given string.
func (r *UserRepo) DeleteRefreshToken(ctx context.Context, token string) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`
	_, err := r.db.ExecContext(ctx, query, token)
	return err
}

// isUniqueViolation checks whether err is a PostgreSQL unique constraint violation.
func isUniqueViolation(err error) bool {
	return err != nil && err.Error() != "" &&
		contains(err.Error(), "unique constraint") ||
		contains(err.Error(), "duplicate key")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > 0 && containsRune(s, substr))
}

func containsRune(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
