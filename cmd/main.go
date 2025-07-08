package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"endorsement-distribution/internal/api"
	"endorsement-distribution/internal/config"
	"endorsement-distribution/internal/store"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	sugar := logger.Sugar()
	sugar.Info("Starting endorsement-distribution service")

	// Initialize database store
	dbStore, err := store.NewPostgresStore(cfg.Database)
	if err != nil {
		sugar.Fatalw("Failed to initialize database store", "error", err)
	}
	defer dbStore.Close()

	// Initialize endorsement distributor
	distributor := store.NewEndorsementDistributor(dbStore, sugar)

	// Initialize API handler
	handler := api.NewHandler(distributor, sugar)

	// Setup router
	router := api.NewRouter(handler)

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		sugar.Infow("Starting HTTP server", "host", cfg.Server.Host, "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			sugar.Fatalw("Failed to start server", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	sugar.Info("Shutting down server...")

	// Give outstanding requests a deadline for completion
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		sugar.Fatalw("Server forced to shutdown", "error", err)
	}

	sugar.Info("Server exited")
} 