package auth

import (
	"context"
	"fmt"
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

// Validate checks validity of Google ID token.
// It calls Google's validation service and gets user information.
// Returns user claims like email and name.
func (g *GoogleAuthenticator) Validate(ctx context.Context, token string) (*ProviderClaims, error) {
	payload, err := idtoken.Validate(ctx, token, g.googleClientID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", models.ErrProvider, err)
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
