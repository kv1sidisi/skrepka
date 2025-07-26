package auth

import (
	"context"
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
type ProviderAuthenticator interface {
	Validate(ctx context.Context, token string) (*ProviderClaims, error)
}
