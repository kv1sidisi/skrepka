package service

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kv1sidisi/skrepka/internal/models"
	"github.com/kv1sidisi/skrepka/internal/storage"
	"google.golang.org/api/idtoken"
	"log/slog"
	"time"
)

type AuthClaims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID `json:"user_id"`
}

type UserResolver interface {
	ResolveUserByProvider(ctx context.Context, params *storage.ResolveUserParams) (*models.User, error)
}

type AuthService struct {
	storage        UserResolver
	log            *slog.Logger
	tokenTTl       time.Duration
	jwtSecret      string
	googleClientID string
}

func NewAuthService(storage UserResolver, log *slog.Logger, tokenTTL time.Duration, jwtSecret string, googleClientID string) *AuthService {
	return &AuthService{
		storage:        storage,
		log:            log,
		tokenTTl:       tokenTTL,
		jwtSecret:      jwtSecret,
		googleClientID: googleClientID,
	}
}

func (a *AuthService) AuthByGoogleToken(ctx context.Context, idToken string) (string, error) {
	const op = "AuthService.AuthByGoogleToken"
	log := a.log.With(slog.String("op", op))

	log.Info("attempting authentication")

	payload, err := a.validateAndParseGoogleToken(ctx, idToken)
	if err != nil {
		log.Error("token validation failed", "error", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}

	user, err := a.resolveUser(ctx, payload)
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

// Validates the Google-provided ID token and extracts its payload.
func (a *AuthService) validateAndParseGoogleToken(ctx context.Context, idToken string) (*idtoken.Payload, error) {
	payload, err := idtoken.Validate(ctx, idToken, a.googleClientID)
	if err != nil {
		return nil, fmt.Errorf("google token validation failed: %w", err)
	}
	return payload, nil
}

// Resolves a user against the database, creating a new user if one doesn't exist.
func (a *AuthService) resolveUser(ctx context.Context, payload *idtoken.Payload) (*models.User, error) {
	userEmail, ok := payload.Claims["email"].(string)
	if !ok {
		return nil, fmt.Errorf("email claim is missing or not a string")
	}
	providerID, ok := payload.Claims["sub"].(string)
	if !ok {
		return nil, fmt.Errorf("sub claim (provider ID) is missing or not a string")
	}

	userName, _ := payload.Claims["name"].(string)
	userAvatar, _ := payload.Claims["picture"].(string)

	userParams := storage.ResolveUserParams{
		ProviderName: models.ProviderGoogle,
		ProviderID:   providerID,
		Email:        userEmail,
		Name:         userName,
		AvatarURL:    userAvatar,
	}

	user, err := a.storage.ResolveUserByProvider(ctx, &userParams)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve user by provider: %w", err)
	}
	return user, nil
}

// Creates and signs a new JWT for the given user.
func (a *AuthService) createJWT(user *models.User) (string, error) {
	claims := AuthClaims{
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
