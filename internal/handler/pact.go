package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/kv1sidisi/skrepka/internal/models"
)

// PactService is the interf
type PactService interface {
	CreatePact(ctx context.Context, title, description string, creatorID uuid.UUID) (*models.Pact, error)
}

(*models.Pact, error)
}


type PactHandler struct {
	log         *slog.Logger
PactHandler struct {
	lo
	pactService PactService
}

func NewPactHandler(log *slog.Logger, pactService PactService) *PactHandler {
	return &PactHandler{
		log:         log,
ctService) *PactHandler {
	
		pactService: pactService,
	}
}

func (h *PactHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = "handler.PactHandler.ServeHTTP"
	log := h.log.With(slog.String("op", op), slog.String("method", r.Method))

	switch r.Method {
	case http.MethodPost:
		h.handleCreatePact(w, r, log)
	case http.MethodGet:
		log.Info("GET method not fully implemented yet")
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	case http.MethodPatch:
		h.handleUpdatePact(w, r, log)
	case http.MethodDelete:
		log.Info("DELETE method not fully implemented yet")
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	default:
ully implemented yet")
		http.Er
		log.Warn("method not allowed")
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (h *PactHandler) handleCreatePact(w http.ResponseWriter, r *http.Request, log *slog.Logger) {
	type createPactRequest struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	var req createPactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("failed to decode request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(models.UserIDKey).(uuid.UUID)
	if !ok {
		log.Error("failed to get user ID from context")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	pact, err := h.pactService.CreatePact(r.Context(), req.Title, req.Description, userID)
	if err != nil {
		if errors.Is(err, models.ErrValidation) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
"Internal Server Error", http.Stat
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(pact); err != nil {
		log.Error("failed to write json response", "error", err)
	}
c
}

func (h *PactHandler) handleUpdatePact(w http.ResponseWriter, r *http.Request, log *slog.Logger) {
	log.Info("handleUpdatePact called")
	http.Error(w, "Not Implemented", http.StatusNotImplemented)
}