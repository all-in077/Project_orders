package main

import (
	"api-gateway/internal/config"
	"api-gateway/internal/grpcclient"
	"api-gateway/internal/router"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.Load()

	authClient, err := grpcclient.NewAuthClient(cfg.AuthSvcAddr)
	if err != nil {
		log.Fatal("failed to connect to auth service:", err)
	}
	defer authClient.Close()

	r := router.New(authClient, cfg)

	srv := &http.Server{
		Addr:    cfg.HTTPPort,
		Handler: r,
	}

	go func() {
		log.Println("gateway started on", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server error:", err)
		}
	}()

	// block until SIGINT or SIGTERM is receivde.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("shutting down gateway...")

	// give in-flight requests up to 5 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("forced shutdown:", err)
	}

	log.Println("gateway stopped")
}
