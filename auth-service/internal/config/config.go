package config

import "os"

// Config holds all configuration for the auth service, loaded from environment variables.
type Config struct {
	HTTPPort    string
	GRPCPort    string
	DatabaseURL string
	JWTSecret   string
	AccessTTL   string // e.g. "15 m"
	RefreshTTL  string // e.g. "720h (30 days)"
}

// Load returns a Config populated from environment variables.
// Falls back to sensible defaults for local development.
func Load() *Config {
	return &Config{
		HTTPPort:    getEnv("HTTP_PORT", ":8082"),
		GRPCPort:    getEnv("GRPC_PORT", ":50051"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/auth?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "change-me-in-production"),
		AccessTTL:   getEnv("ACCESS_TTL", "15m"),
		RefreshTTL:  getEnv("REFRESH_TTL", "720h"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
