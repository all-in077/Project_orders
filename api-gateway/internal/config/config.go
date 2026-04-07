package config

import "os"

// Config holds the application configuration loaded from environment variables.
type Config struct {

	// HTTPPort is the port the API gateway listens on (e.g ":8080")
	HTTPPort string

	// OrderSvcURL is the base HTTP URL of the order service.
	OrderSvcURL string

	// AuthSvcAddr is the gRPC address of the auth service (host:port).
	AuthSvcAddr string

	// AuthSvcURL is the base HTTP URL of the auth service used for reverse proxying.
	AuthSvcURL string
}

// Load returns a Config populated from environment variables.
// Falls back to sensible defaults for local development if a variable is not set.
func Load() *Config {
	return &Config{
		HTTPPort:    getEnv("HTTP_PORT", ":8080"),
		OrderSvcURL: getEnv("ORDER_SVC_URL", "http://localhost:8081"),
		AuthSvcAddr: getEnv("AUTH_SVC_ADDR", "localhost:50051"),
		AuthSvcURL:  getEnv("AUTH_SVC_URL", "http://localhost:8082"),
	}
}

// getEnv returns the value of the environment variable named by key or fallback if te variable is empty or not set.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
