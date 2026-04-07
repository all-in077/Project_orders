package service

import (
	"auth-service/internal/domain"
	"auth-service/internal/repository"
	"auth-service/internal/token"
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// AuthService implements the core business logic for registration, login, token refresh, and token validation.
type AuthService struct {
	userRepo  repository.UserRepository
	tokenRepo repository.TokenRepository
	tokens    *token.Manager
}

// NewAuthService creates an AuthService wired to the given repositories and token manager.
func NewAuthService(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	tokens *token.Manager,
) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		tokens:    tokens,
	}
}

// RegisterInput holds the data required to register a new user.
type RegisterInput struct {
	Email    string
	Password string
}

// TokenPair holds an access and refresh token returned after login or refresh.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// Register creates a new user account.
// Returns ErrUserAlreadyExists if the email is already taken.
func (s *AuthService) Register(ctx context.Context, input RegisterInput) (int64, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	id, err := s.userRepo.CreateUser(ctx, &domain.User{
		Email:        input.Email,
		PasswordHash: string(hash),
	})
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Login validates credentials and returns a token pair on success.
// Returns ErrInvalidCredentials if the email is not found or the password is wrong.
func (s *AuthService) Login(ctx context.Context, email, password string) (*TokenPair, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		// map ErrUserNotFound to ErrInvalidCredentials
		// so the caller cannot tell whether the email exists
		if err == domain.ErrUserNotFound {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return s.issueTokenPair(ctx, user.ID)
}

// Refresh validates a refresh token and issues a new token pair.
// The old refresh token is deleted (token rotation).
// Returns ErrInvalidToken or ErrTokenExpired if the token is not usable.
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	stored, err := s.tokenRepo.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	if time.Now().After(stored.ExpiresAt) {
		// token is in the DB but has expired — clean it up
		_ = s.tokenRepo.DeleteRefreshToken(ctx, refreshToken)
		return nil, domain.ErrTokenExpired
	}

	// rotate: delete old token before issuing new pair
	if err := s.tokenRepo.DeleteRefreshToken(ctx, refreshToken); err != nil {
		return nil, err
	}

	return s.issueTokenPair(ctx, stored.UserID)
}

// ValidateAccessToken validates an access token and returns the userID from its claims.
// Called by the gRPC handler which is used by api-gateway on every protected request.
func (s *AuthService) ValidateAccessToken(ctx context.Context, accessToken string) (int64, error) {
	claims, err := s.tokens.Validate(accessToken)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}

// issueTokenPair generates a new access+refresh token pair and persists the refresh token.
func (s *AuthService) issueTokenPair(ctx context.Context, userID int64) (*TokenPair, error) {
	accessToken, err := s.tokens.NewAccessToken(userID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.tokens.NewRefreshToken(userID)
	if err != nil {
		return nil, err
	}

	claims, err := s.tokens.Validate(refreshToken)
	if err != nil {
		return nil, err
	}

	err = s.tokenRepo.SaveRefreshToken(ctx, &domain.RefreshToken{
		UserID:    userID,
		Token:     refreshToken,
		ExpiresAt: claims.RegisteredClaims.ExpiresAt.Time,
	})
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
