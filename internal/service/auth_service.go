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
	payload, err := idtoken.Validate(ctx, idToken, a.googleClientID)
	if err != nil {
		return "", fmt.Errorf("failed to validate google id token: %w", err)
	}

	// Safely extract mandatory claims
	userEmail, ok := payload.Claims["email"].(string)
	if !ok {
		return "", fmt.Errorf("email claim is missing or not a string")
	}
	providerID, ok := payload.Claims["sub"].(string)
	if !ok {
		return "", fmt.Errorf("sub claim (provider ID) is missing or not a string")
	}

	// Extract optional claims, defaulting to an empty string if they are missing
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
		return "", fmt.Errorf("failed to resolve user by google id token: %w", err)
	}

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
