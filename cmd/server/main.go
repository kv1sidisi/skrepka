package main

import (
	"github.com/kv1sidisi/skrepka/internal/config"
	"github.com/kv1sidisi/skrepka/internal/handler"
	"github.com/kv1sidisi/skrepka/internal/logger"
	"github.com/kv1sidisi/skrepka/internal/service/auth"
	"github.com/kv1sidisi/skrepka/internal/storage"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	cfg := config.Get()

	//Logger setup
	writer, err := logger.SetupWriter(cfg.LogPath)
	if err != nil {
		slog.Error("failed to setup log writer", "error", err)
		os.Exit(1)
	}
	log := logger.SetupLogger(cfg.Env, writer)

	//Storage setup
	postgres, err := storage.NewStorage()
	if err != nil {
		slog.Error("failed to init storage", "error", err)
		os.Exit(1)
	}
	defer postgres.Close()

	//Authentication service
	_ = auth.NewAuthService(postgres, log, cfg.TokenTTL, cfg.JWTSecret)

	// Handlers setup
	healthHandler := handler.handler.NewHealthHandler(log)

	//HTTP Server setup
	mux := http.NewServeMux()
	mux.Handle("/api/v1/health", healthHandler)
	mux.Handle("/api/v1/metrics", promhttp.Handler())
	mux.Handle("/api/v1/oids")

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
