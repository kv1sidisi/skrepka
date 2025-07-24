package auth

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kv1sidisi/skrepka/internal/models"
	"github.com/kv1sidisi/skrepka/internal/storage"
	"log/slog"
	"time"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID `json:"user_id"`
}

// UserResolver defines the dependency for resolving a user in the storage layer.
// This interface allows for easier testing and decoupling from the concrete storage implementation.
type UserResolver interface {
	ResolveUserByProvider(ctx context.Context, params *storage.ResolveUserParams) (*models.User, error)
}

type Service struct {
	userResolver UserResolver
	log          *slog.Logger
	tokenTTl     time.Duration
	jwtSecret    string
	providers    map[models.Provider]ProviderAuthenticator
}

func NewAuthService(storage UserResolver, log *slog.Logger, tokenTTL time.Duration, jwtSecret string) *Service {
	return &Service{
		userResolver: storage,
		log:          log,
		tokenTTl:     tokenTTL,
		jwtSecret:    jwtSecret,
		providers:    providerRegistry,
	}
}

// Authenticate orchestrates the entire authentication flow for a given provider.
// It selects the appropriate strategy, validates the external token, resolves the user,
// and issues a new internal JWT.
func (a *Service) Authenticate(ctx context.Context, provider models.Provider, token string) (string, error) {
	const op = "AuthService.Authenticate"
	log := a.log.With(slog.String("op", op))

	log.Info("attempting authentication", slog.String("method", provider.String()))

	authenticator, ok := a.providers[provider]
	if !ok {
		log.Error("authenticator is not supported", slog.String("provider", provider.String()))
	}

	claims, err := authenticator.Validate(ctx, token)
	if err != nil {
		log.Error("token validation failed", "error", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("token validated successfully")

	params := &storage.ResolveUserParams{
		ProviderName: provider,
		ProviderID:   claims.ProviderUserID,
		Email:        claims.Email,
		Name:         claims.Name,
		AvatarURL:    claims.AvatarURL,
	}

		user, err := a.userResolver.ResolveUserByProvider(ctx, params)
	if err != nil {
		log.Error("failed to resolve user", "error", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}
	log.Info("user resolved", slog.String("user_id", user.ID.String()))

	signedToken, err := a.createJWT(user)
	if err != nil {
		log.Error("failed to create token", "error", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("authentication successful")
	return signedToken, nil
}

// Creates and signs a new JWT for the given user.
func (a *Service) createJWT(user *models.User) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.tokenTTl)),
		},
		UserID: user.ID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(a.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return signedToken, nil
}
