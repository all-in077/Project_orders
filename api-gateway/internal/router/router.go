package router

import (
	"api-gateway/internal/config"
	"api-gateway/internal/grpcclient"
	"api-gateway/internal/middleware"
	"api-gateway/internal/proxy"
	"time"

	"github.com/gin-gonic/gin"
)

// New builds and returns the configured Gin router.
//
// Public routes (no auth required):
//   - POST /auth/login
//   - POST /auth/register
//   - POST /auth/refresh
//
// Protected routes (Bearer token required):
//   - ANY /orders/*path  → proxied to order-service
func New(authClient *grpcclient.AuthClient, cfg *config.Config) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	rl := middleware.NewRateLimiter(100, time.Minute)
	r.Use(rl.Middleware())

	r.POST("/auth/login", proxy.NewReverseProxy(cfg.AuthSvcURL))
	r.POST("/auth/register", proxy.NewReverseProxy(cfg.AuthSvcURL))
	r.POST("/auth/refresh", proxy.NewReverseProxy(cfg.AuthSvcURL))

	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware(authClient))
	{
		protected.Any("/orders/*path", proxy.NewReverseProxy(cfg.OrderSvcURL))
	}

	return r
}
