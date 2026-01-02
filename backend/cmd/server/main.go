package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"harmony/internal/config"
	"harmony/internal/database"
	"harmony/internal/handlers"
	"harmony/internal/services"
	"harmony/internal/transcoder"
)

// Version information (set via ldflags)
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Configure logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.SlogLevel(),
	}))
	slog.SetDefault(logger)

	// Log startup information
	slog.Info("harmony server starting",
		"version", Version,
		"build_time", BuildTime,
		"git_commit", GitCommit,
	)

	// Print configuration
	cfg.Print()

	// Initialize database
	db, err := database.New(database.Config{
		Path: cfg.DBPath,
	})
	if err != nil {
		slog.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate(); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Initialize Redis (optional - continue if not available)
	var redis *database.RedisClient
	redis, err = database.NewRedis(database.RedisConfig{
		URL: cfg.RedisURL,
	})
	if err != nil {
		slog.Warn("redis not available, caching disabled", "error", err)
		redis = nil
	}

	// Initialize transcoder
	trans, err := transcoder.New(transcoder.Config{
		CacheDir:   cfg.CachePath,
		MaxCacheGB: 10.0,
	})
	if err != nil {
		slog.Warn("transcoder not available", "error", err)
		trans = nil
	}

	// Create repositories
	trackRepo := database.NewTrackRepository(db.DB)
	albumRepo := database.NewAlbumRepository(db.DB)
	artistRepo := database.NewArtistRepository(db.DB)

	// Initialize library service
	libService := services.NewLibraryService(
		cfg.MediaPath,
		cfg.ArtworkPath,
		trackRepo,
		albumRepo,
		artistRepo,
	)

	// Configure router
	routerCfg := handlers.RouterConfig{
		AllowedOrigins: []string{"*"}, // Allow all in container, restrict via reverse proxy
		MediaRoot:      cfg.MediaPath,
		CacheDir:       cfg.ArtworkPath,
		BaseURL:        fmt.Sprintf("http://localhost:%d", cfg.Port),
	}

	// Create router
	router := handlers.NewRouter(routerCfg, db, redis, trans, libService)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		slog.Info("starting HTTP server", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Auto-scan on startup if enabled
	if cfg.ScanOnStartup {
		slog.Info("starting initial library scan")
		go func() {
			if err := libService.FullScan(context.Background()); err != nil {
				slog.Error("initial scan failed", "error", err)
			}
		}()
	}

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server gracefully
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server shutdown error", "error", err)
	}

	// Close Redis connection if available
	if redis != nil {
		redis.Close()
	}

	slog.Info("server stopped")
}
