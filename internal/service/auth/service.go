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

// Claims contains data for our internal JWT.
// It includes standard JWT claims and our custom UserID.
type Claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID `json:"user_id"`
}

// UserResolver is interface for resolving user in storage layer.
// This interface allows for easier testing and decoupling from concrete storage implementation.
type UserResolver interface {
	ResolveUserByProvider(ctx context.Context, params *storage.ResolveUserParams) (*models.User, error)
}

// Service handles all business logic for authentication.
type Service struct {
	userResolver UserResolver
	log          *slog.Logger
	tokenTTl     time.Duration
	jwtSecret    string
	providers    map[models.Provider]ProviderAuthenticator
}

// NewAuthService creates new authentication service.
// It requires storage, logger, and settings for JWT.
// Returns pointer to new service.
func NewAuthService(storage UserResolver, log *slog.Logger, tokenTTL time.Duration, jwtSecret string) (*Service, error) {
	if jwtSecret == "" {
		return nil, fmt.Errorf("jwt secret cannot be empty")
	}
	return &Service{
		userResolver: storage,
		log:          log,
		tokenTTl:     tokenTTL,
		jwtSecret:    jwtSecret,
		providers:    providerRegistry,
	}, nil
}

// Authenticate manages entire authentication flow for given provider.
// It validates external token, finds or creates user, and issues new internal JWT.
// Returns new JWT as string.
func (a *Service) Authenticate(ctx context.Context, provider models.Provider, token string) (string, error) {
	const op = "AuthService.Authenticate"

	authenticator, ok := a.providers[provider]
	if !ok {
		// This is server configuration error, so it's appropriate to log it here.
		a.log.Error("unsupported provider requested", slog.String("provider", provider.String()))
		return "", fmt.Errorf("%s: unsupported provider", op)
	}

	claims, err := authenticator.Validate(ctx, token)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	params := &storage.ResolveUserParams{
		ProviderName: provider,
		ProviderID:   claims.ProviderUserID,
		Email:        claims.Email,
		Name:         claims.Name,
		AvatarURL:    claims.AvatarURL,
	}

	user, err := a.userResolver.ResolveUserByProvider(ctx, params)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return a.createJWT(user)
}

// createJWT creates and signs new JWT for given user.
// Returns signed JWT as string.
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
