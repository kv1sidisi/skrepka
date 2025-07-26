package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/kv1sidisi/skrepka/internal/models"
)

// OIDCAuthenticator is interface for authentication service.
// Using interface helps to easily change auth service later if needed.
type OIDCAuthenticator interface {
	Authenticate(ctx context.Context, provider models.Provider, token string) (string, error)
}

// OIDCHandler handles HTTP requests for user authentication.
type OIDCHandler struct {
	log     *slog.Logger
	service OIDCAuthenticator
}

// NewOIDCHandler creates new OIDCHandler.
// It needs logger and authentication service to work.
// Returns pointer to new OIDCHandler.
func NewOIDCHandler(log *slog.Logger, service OIDCAuthenticator) *OIDCHandler {
	return &OIDCHandler{
		log:     log,
		service: service,
	}
}

// HandleOIDCAuthenticate processes user authentication request.
// It reads provider's token from request, and asks auth service to check it.
// Returns new JWT for our application.
func (h *OIDCHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = "OIDCHandler.HandleOIDCAuthenticate"
	log := h.log.With(slog.String("op", op))

	if r.Method != http.MethodPost {
		log.Warn("method not allowed", slog.String("method", r.Method))
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.OIDCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("failed to decode request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Info("authentication requested", slog.String("provider", req.Provider.String()))

	jwt, err := h.service.Authenticate(r.Context(), req.Provider, req.IDToken)
	if err != nil {
		// Check if error is client-side validation or provider error.
		if errors.Is(err, models.ErrProvider) || errors.Is(err, models.ErrValidation) {
			log.Warn("authentication failed", "error", err)
			http.Error(w, "Authentication failed: invalid token or provider error", http.StatusUnauthorized)
			return
		}

		// For all other errors, assume it's server-side problem.
		log.Error("internal authentication error", "error", err)
		http.Error(w, "An internal error occurred", http.StatusInternalServerError)
		return
	}

	resp := models.AuthResponse{Token: jwt}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Error("failed to write json response", "error", err)
	}
}
