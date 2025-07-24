package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/kv1sidisi/skrepka/internal/config"
	"github.com/kv1sidisi/skrepka/internal/handler"
	"github.com/kv1sidisi/skrepka/internal/logger"
	"github.com/kv1sidisi/skrepka/internal/service/auth"
	"github.com/kv1sidisi/skrepka/internal/storage"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	ctx := context.Background()
	cfg := config.Get()

	// Logger setup
	writer, err := logger.SetupWriter(cfg.LogPath)
	if err != nil {
		slog.Error("failed to setup log writer", "error", err)
		os.Exit(1)
	}
	log := logger.SetupLogger(cfg.Env, writer)

	// Storage setup
	db, err := storage.New(ctx, cfg)
	if err != nil {
		log.Error("failed to init storage", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	userRepo := db.UserRepository()

	// Authentication service
	authService := auth.NewAuthService(userRepo, log, cfg.TokenTTL, cfg.JWTSecret)

	// Handlers setup
	healthHandler := handler.NewHealthHandler(log)
	oidcAuthHandler := handler.NewOIDCHandler(log, authService)

	// HTTP Server setup
	mux := http.NewServeMux()
	mux.Handle("/health", healthHandler)
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/api/v1/auth/oidc", oidcAuthHandler.HandleOIDCAuthenticate)

	log.Info("starting server", "address", cfg.Address)

	server := &http.Server{
		Addr:         cfg.Address,
		Handler:      mux,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Error("server failed to start", "error", err)
		os.Exit(1)
	}
}
