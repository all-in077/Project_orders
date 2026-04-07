package cmd

import (
	"auth-service/internal/config"
	httphandler "auth-service/internal/handler"
	"auth-service/internal/handler/grpc"
	"auth-service/internal/repository/postgres"
	"auth-service/internal/service"
	"auth-service/internal/token"
	gen "auth-service/proto/gen"
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	googlegrpc "google.golang.org/grpc"
)

func main() {
	cfg := config.Load()

	// connect to PostgreSQL
	db, err := sqlx.Connect("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer db.Close()

	// wire up layers
	userRepo := postgres.NewUserRepo(db)
	tokenManager, err := token.NewManager(cfg.JWTSecret, cfg.AccessTTL, cfg.RefreshTTL)
	if err != nil {
		log.Fatal("failed to create token manager:", err)
	}
	authSvc := service.NewAuthService(userRepo, userRepo, tokenManager)

	// HTTP server
	httpServer := buildHTTPServer(cfg, authSvc)

	// gRPC server
	grpcServer, lis := buildGRPCServer(cfg, authSvc)

	// start HTTP in background
	go func() {
		log.Println("http server started on", cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("http server error:", err)
		}
	}()

	// start gRPC in background
	go func() {
		log.Println("grpc server started on", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("grpc server error:", err)
		}
	}()

	// block until SIGINT or SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("shutting down auth-service...")

	// give in-flight HTTP requests up to 5 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatal("http forced shutdown:", err)
	}

	// gRPC graceful stop waits for active RPCs to finish
	grpcServer.GracefulStop()

	log.Println("auth-service stopped")
}

func buildHTTPServer(cfg *config.Config, svc *service.AuthService) *http.Server {
	h := httphandler.NewAuthHandler(svc)

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.POST("/auth/register", h.Register)
	r.POST("/auth/login", h.Login)
	r.POST("/auth/refresh", h.Refresh)

	return &http.Server{
		Addr:    cfg.HTTPPort,
		Handler: r,
	}
}

func buildGRPCServer(cfg *config.Config, svc *service.AuthService) (*googlegrpc.Server, net.Listener) {
	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		log.Fatal("failed to listen grpc port:", err)
	}

	srv := googlegrpc.NewServer()
	gen.RegisterAuthServiceServer(srv, grpc.NewAuthGRPCHandler(svc))

	return srv, lis
}
