package service

import (
	"context"
	"fmt"
	"github.com/kv1sidisi/skrepka/internal/config"
	"github.com/kv1sidisi/skrepka/internal/models"
	"google.golang.org/api/idtoken"
)

type GoogleAuthenticator struct {
	googleClientID string
}

func NewGoogleAuthenticator(googleClientID string) *GoogleAuthenticator {
	return &GoogleAuthenticator{
		googleClientID: googleClientID,
	}
}

func init() {
	cfg := config.Get()
	googleAuth := NewGoogleAuthenticator(cfg.GoogleClientID)
	RegisterProvider(models.ProviderGoogle, googleAuth)
}

// Validate checks the validity of a Google ID token and extracts user claims.
// It calls Google's token validation service and maps the result to the standardized ProviderClaims.
func (g *GoogleAuthenticator) Validate(ctx context.Context, token string) (*ProviderClaims, error) {
	payload, err := idtoken.Validate(ctx, token, g.googleClientID)
	if err != nil {
		return nil, fmt.Errorf("google token validation failed: %w", err)
	}

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

	claims := &ProviderClaims{
		Email:          userEmail,
		ProviderUserID: providerID,
		Name:           userName,
		AvatarURL:      userAvatar,
	}

	return claims, nil
}
