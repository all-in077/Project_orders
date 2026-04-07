package grpc

import (
	"auth-service/internal/domain"
	"auth-service/internal/service"
	gen "auth-service/proto/gen"
	"context"
	"errors"

	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthGRPCHandler implements the gRPC AuthService interface.
// It is called by api-gateway to validate access tokens on every protected request.
type AuthGRPCHandler struct {
	gen.UnimplementedAuthServiceServer
	svc *service.AuthService
}

// NewAuthGRPCHandler creates an AuthGRPCHandler backed by the given AuthService.
func NewAuthGRPCHandler(svc *service.AuthService) *AuthGRPCHandler {
	return &AuthGRPCHandler{svc: svc}
}

// ValidateToken validates the given access token and returns the associated user ID.
// Returns codes.Unauthenticated if the token is invalid or expired.
func (h *AuthGRPCHandler) ValidateToken(ctx context.Context, req *gen.ValidateTokenRequest) (*gen.ValidateTokenResponse, error) {
	userID, err := h.svc.ValidateAccessToken(ctx, req.Token)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidToken) || errors.Is(err, domain.ErrTokenExpired) {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &gen.ValidateTokenResponse{
		UserId: fmt.Sprintf("%d", userID),
	}, nil
}
