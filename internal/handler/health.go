package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// HealthHandler has endpoint to check if service is working.
type HealthHandler struct {
	log *slog.Logger
}

// NewHealthHandler creates new HealthHandler.
// It needs logger to write information about requests.
// Returns pointer to new HealthHandler.
func NewHealthHandler(log *slog.Logger) *HealthHandler {
	return &HealthHandler{
		log: log,
	}
}

// ServeHTTP handles requests for health check.
// This is useful for other services to see if our application is alive.
// Returns JSON message that says "status": "ok".
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = "handler.HealthHandler.ServeHTTP"
	log := h.log.With(slog.String("op", op))

	if r.Method != http.MethodGet {
		log.Error("method not allowed", slog.String("method", r.Method))
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
