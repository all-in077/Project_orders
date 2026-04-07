package grpcclient

import (
	"api-gateway/proto/gen"
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AuthClient is a gRPC client for the auth service.
type AuthClient struct {
	conn   *grpc.ClientConn
	client gen.AuthServiceClient
}

// NewAuthClient creates a real gRPC connection to the auth service at addr.
func NewAuthClient(addr string) (*AuthClient, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &AuthClient{
		conn:   conn,
		client: gen.NewAuthServiceClient(conn),
	}, nil
}

// ValidateToken calls the auth service to validate the given Bearer token.
// Returns the userID on success, or an error if the token is invalid.
func (c *AuthClient) ValidateToken(ctx context.Context, token string) (string, error) {
	resp, err := c.client.ValidateToken(ctx, &gen.ValidateTokenRequest{
		Token: token,
	})
	if err != nil {
		return "", err
	}
	return resp.UserId, nil
}

// Close releases the underlying gRPC connection.
func (c *AuthClient) Close() error {
	return c.conn.Close()
}
