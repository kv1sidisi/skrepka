package auth

import (
	"context"
	"github.com/kv1sidisi/skrepka/internal/models"
)

type ProviderClaims struct {
	Email          string
	ProviderUserID string
	Name           string
	AvatarURL      string
}

type ProviderAuthenticator interface {
	Validate(ctx context.Context, token string) (*ProviderClaims, error)
}

var providerRegistry = make(map[models.Provider]ProviderAuthenticator)

// RegisterProvider adds a new authentication provider strategy to the central registry.
// This function is called from the init() function of each provider-specific implementation.
func RegisterProvider(provider models.Provider, auth ProviderAuthenticator) {
	providerRegistry[provider] = auth
}
