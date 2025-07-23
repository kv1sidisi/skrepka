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
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var buf bytes.Buffer

	err := json.NewEncoder(&buf).Encode(map[string]string{"status": "ok"})
	if err != nil {
		h.log.Error("failed to encode json health response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(buf.Bytes()); err != nil {
		h.log.Error("failed to write json health response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
