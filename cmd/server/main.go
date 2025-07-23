package main

import (
	"bytes"
	"encoding/json"
	"github.com/kv1sidisi/skrepka/internal/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
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
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.MustLoad()

	logger = SetupLogger(cfg.Env, cfg.LogPath)
	slog.SetDefault(logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", health)
	mux.Handle("/metrics", promhttp.Handler())

	slog.Info("starting server", "address", ":4000")

	if err := http.ListenAndServe(":4000", mux); err != nil {
		slog.Error("server failed to start", "error", err)
		os.Exit(1)
	}
}

// SetupLogger used to set up logger based on env level.
// returns slog.Logger on success
func SetupLogger(env, logPath string) *slog.Logger {
	var logger *slog.Logger

	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		slog.Error("failed to create log directory", "error", err)
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		slog.Error("failed to open log file", "error", err)
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)

	switch env {
	case "debug":
		logger = slog.New(slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case "local":
		logger = slog.New(slog.NewTextHandler(multiWriter, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case "prod":
		logger = slog.New(slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		logger = slog.New(slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return logger
}
