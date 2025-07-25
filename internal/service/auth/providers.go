package auth

import (
	"context"
	"github.com/kv1sidisi/skrepka/internal/models"
)

// ProviderClaims contains user information from external provider.
// This structure helps to standardize user data from different providers like Google or Apple.
type ProviderClaims struct {
	Email          string
	ProviderUserID string
	Name           string
	AvatarURL      string
}

// ProviderAuthenticator is interface for different authentication providers.
// It makes sure every provider has Validate method.
type ProviderAuthenticator interface {
	Validate(ctx context.Context, token string) (*ProviderClaims, error)
}

var providerRegistry = make(map[models.Provider]ProviderAuthenticator)

// RegisterProvider adds new authentication provider to central registry.
// This function is called from init() function of each provider-specific implementation.
func RegisterProvider(provider models.Provider, auth ProviderAuthenticator) {
	providerRegistry[provider] = auth
}
