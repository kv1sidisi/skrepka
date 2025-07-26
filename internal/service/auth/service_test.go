package auth

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/kv1sidisi/skrepka/internal/models"
	"github.com/kv1sidisi/skrepka/internal/storage"
	"github.com/stretchr/testify/require"
	"log/slog"
	"testing"
	"time"
)

// mockUserResolver is mock for UserResolver interface
// to simulate database behavior in tests.
type mockUserResolver struct {
	user *models.User
	err  error
}

func (m *mockUserResolver) ResolveUserByProvider(ctx context.Context, params *storage.ResolveUserParams) (*models.User, error) {
	return m.user, m.err
}

// mockProviderAuthenticator is mock for ProviderAuthenticator interface
// to simulate external provider (like Google) behavior.
type mockProviderAuthenticator struct {
	claims *ProviderClaims
	err    error
}

func (m *mockProviderAuthenticator) Validate(ctx context.Context, token string) (*ProviderClaims, error) {
	return m.claims, m.err
}

func TestAuthService_Authenticate(t *testing.T) {
	// Create dummy user for tests
	mockUser := &models.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Name:  "Test User",
	}

	// Create dummy claims for tests
	mockClaims := &ProviderClaims{
		Email:          "test@example.com",
		ProviderUserID: "12345",
	}

	testCases := []struct {
		name          string
		mockUserRes   *mockUserResolver
		mockProvider  *mockProviderAuthenticator
		expectError   bool
		errorContains string
	}{
		{
			name:          "Success: Happy path",
			mockUserRes:   &mockUserResolver{user: mockUser, err: nil},
			mockProvider:  &mockProviderAuthenticator{claims: mockClaims, err: nil},
			expectError:   false,
			errorContains: "",
		},
		{
			name:          "Failure: Provider validation fails",
			mockUserRes:   &mockUserResolver{user: mockUser, err: nil},
			mockProvider:  &mockProviderAuthenticator{claims: nil, err: models.ErrProvider},
			expectError:   true,
			errorContains: models.ErrProvider.Error(),
		},
		{
			name:          "Failure: User resolver fails",
			mockUserRes:   &mockUserResolver{user: nil, err: fmt.Errorf("database error")},
			mockProvider:  &mockProviderAuthenticator{claims: mockClaims, err: nil},
			expectError:   true,
			errorContains: "database error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create service with mocks
			authService := &Service{
				userResolver: tc.mockUserRes,
				log:          slog.Default(),
				tokenTTl:     time.Hour,
				jwtSecret:    "test-secret",
				providers: map[models.Provider]ProviderAuthenticator{
					models.ProviderGoogle: tc.mockProvider,
				},
			}

			// Call method we want to test
			token, err := authService.Authenticate(context.Background(), models.ProviderGoogle, "any-token")

			// Check results
			if tc.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorContains)
				require.Empty(t, token)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, token)
			}
		})
	}
}

func TestNewAuthService_EmptySecret(t *testing.T) {
	providerRegistry := map[models.Provider]ProviderAuthenticator{
		models.ProviderGoogle: nil,
	}
	_, err := NewAuthService(nil, slog.Default(), time.Hour, "", providerRegistry)
	require.Error(t, err)
	require.Contains(t, err.Error(), "jwt secret cannot be empty")
}
