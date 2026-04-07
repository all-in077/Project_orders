package token

import (
	"auth-service/internal/domain"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT payload.
type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

// Manager handles JWT generation and validation.
type Manager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewManager creates a Manager with the given secret and token lifetimes.
func NewManager(secret, accessTTL, refreshTTL string) (*Manager, error) {
	access, err := time.ParseDuration(accessTTL)
	if err != nil {
		return nil, err
	}
	refresh, err := time.ParseDuration(refreshTTL)
	if err != nil {
		return nil, err
	}
	return &Manager{
		secret:     []byte(secret),
		accessTTL:  access,
		refreshTTL: refresh,
	}, nil
}

// NewAccessToken generates a signed JWT access token for the given user.
func (m *Manager) NewAccessToken(userID int64) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// NewRefreshToken generates a signed JWT refresh token for the given user.
// Refresh tokens have a longer TTL and are stored in the database.
func (m *Manager) NewRefreshToken(userID int64) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.refreshTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// Validate parses and validates a token string.
// Returns the claims if the token is valid, or a domain error otherwise.
func (m *Manager) Validate(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, domain.ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, domain.ErrTokenExpired
		}
		return nil, domain.ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, domain.ErrInvalidToken
	}
	return claims, nil
}
