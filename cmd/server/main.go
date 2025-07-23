package main

import (
	"bytes"
	"encoding/json"
	"github.com/kv1sidisi/skrepka/internal/config"
	"github.com/kv1sidisi/skrepka/internal/logger"
	"github.com/kv1sidisi/skrepka/internal/service"
	"github.com/kv1sidisi/skrepka/internal/storage"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log/slog"
	"net/http"
	"os"
)

// health handles the health check endpoint.
// It's used by external services to verify that the application is running.
func health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var buf bytes.Buffer

	err := json.NewEncoder(&buf).Encode(map[string]string{"status": "ok"})
	if err != nil {
		slog.Error("failed to encode json health response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(buf.Bytes()); err != nil {
		slog.Error("failed to write json health response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func main() {
	cfg := config.MustLoad()

	//Logger setup
	writer, err := logger.SetupWriter(cfg.LogPath)
	if err != nil {
		slog.Error("failed to setup log writer", "error", err)
		os.Exit(1)
	}
	log := logger.SetupLogger(cfg.Env, writer)
	slog.SetDefault(log)

	//Storage setup
	postgres, err := storage.NewStorage()
	if err != nil {
		slog.Error("failed to init storage", "error", err)
		os.Exit(1)
	}
	defer postgres.Close()

	//Authentication service
	authService := service.NewAuthService(postgres, cfg.TokenTTL, cfg.JWTSecret, cfg.GoogleClientID)

	//HTTP Server setup
	mux := http.NewServeMux()
	mux.HandleFunc("/health", health)
	mux.Handle("/metrics", promhttp.Handler())

	slog.Info("starting server", "address", cfg.Address)

	server := &http.Server{
		Addr:         cfg.Address,
		Handler:      mux,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	if err := server.ListenAndServe(); err != nil {
		slog.Error("server failed to start", "error", err)
		os.Exit(1)
	}
}
