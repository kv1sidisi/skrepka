package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

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
