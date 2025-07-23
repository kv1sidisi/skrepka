package handler

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
)

type HealthHandler struct {
	log *slog.Logger
}

func NewHealthHandler(log *slog.Logger) *HealthHandler {
	return &HealthHandler{
		log: log,
	}
}

// ServeHTTP handles the health check endpoint.
// It's used by external services to verify that the application is running.
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = "handler.HealthHandler.ServeHTTP"
	log := h.log.With(slog.String("op", op))

	if r.Method != http.MethodGet {
		log.Warn("method not allowed", slog.String("method", r.Method))
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Info("health check requested")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		log.Error("failed to write json response", "error", err)
	}
}
