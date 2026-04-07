package handler

import (
	"auth-service/internal/domain"
	"auth-service/internal/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles HTTP requests for auth endpoints.
type AuthHandler struct {
	svc *service.AuthService
}

// NewAuthHandler creates an AuthHandler backed by the given AuthService.
func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// registerRequest is the expected JSON body for POST /auth/register.
type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// loginRequest is the expected JSON body for POST /auth/login.
type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// refreshRequest is the expected JSON body for POST /auth/refresh.
type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Register godoc
// POST /auth/register
// Creates a new user account.
// Returns 201 with user_id on success.
// Returns 409 if the email is already taken.
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.svc.Register(c.Request.Context(), service.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "email already taken"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user_id": id})
}

// Login godoc
// POST /auth/login
// Validates credentials and returns an access + refresh token pair.
// Returns 401 if credentials are invalid.
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pair, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
	})
}

// Refresh godoc
// POST /auth/refresh
// Issues a new token pair in exchange for a valid refresh token.
// Returns 401 if the token is invalid or expired.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pair, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidToken) || errors.Is(err, domain.ErrTokenExpired) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
	})
}
