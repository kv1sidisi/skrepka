package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kv1sidisi/skrepka/internal/handler"
	"github.com/kv1sidisi/skrepka/internal/models"
	"github.com/kv1sidisi/skrepka/internal/service/auth"
	"github.com/kv1sidisi/skrepka/internal/storage"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// mockProviderAuthenticator is mock for ProviderAuthenticator interface
// to simulate external provider (like Google) behavior.
type mockProviderAuthenticator struct {
	claims *auth.ProviderClaims
	err    error
}

func (m *mockProviderAuthenticator) Validate(ctx context.Context, token string) (*auth.ProviderClaims, error) {
	return m.claims, m.err
}

type testServer struct {
	t           *testing.T
	server      *httptest.Server
	mockPool    pgxmock.PgxPoolIface
	authService *auth.Service
	logger      *slog.Logger
}

func newTestServer(t *testing.T, pool pgxmock.PgxPoolIface, authenticators auth.Authenticators) *testServer {
	t.Helper()

	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Mocking User Database repository
	mockedUserRepo := &storage.UserRepository{
		Db: pool,
	}

	mockedAuthService, err := auth.NewAuthService(mockedUserRepo, discardLogger, time.Hour, "very secret key", authenticators)
	require.NoError(t, err)

	// Mocking mux
	oidcHandler := handler.NewOIDCHandler(discardLogger, mockedAuthService)
	mux := http.NewServeMux()
	mux.Handle("/api/v1/auth/oidc", oidcHandler)

	// Starting test server
	server := httptest.NewServer(mux)

	t.Cleanup(func() {
		server.Close()
	})

	return &testServer{
		t:           t,
		server:      server,
		mockPool:    pool,
		authService: mockedAuthService,
		logger:      discardLogger,
	}
}

func (ts *testServer) sendOIDCRequest(provider models.Provider, token string) *http.Response {
	ts.t.Helper()

	// Mocking HTTP request
	requestBody := fmt.Sprintf(`{"provider": "%s", "id_token": "%s"}`, provider, token)
	mockedRequestBody := bytes.NewBufferString(requestBody)

	// Sending HTTP request
	resp, err := http.Post(ts.server.URL+"/api/v1/auth/oidc", "application/json", mockedRequestBody)
	require.NoError(ts.t, err)

	return resp
}

func TestOIDC_SuccessfulLogin_NewUser(t *testing.T) {
	mockUserID := uuid.New()
	providerID := "google123"
	userEmail := "test@gmail.com"
	userName := "testUsername"
	avatarURL := "test-url.com"

	// Mocking DB Connection pool
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	mockPool.ExpectQuery(`SELECT (.+) FROM auth_providers WHERE provider_name = \$1 AND provider_id = \$2`).
		WithArgs(models.ProviderGoogle, providerID).
		WillReturnError(pgx.ErrNoRows)
	mockPool.ExpectQuery(`SELECT (.+) FROM users WHERE email = \$1`).
		WithArgs(userEmail).
		WillReturnError(pgx.ErrNoRows)

	userInsertRows := pgxmock.NewRows([]string{"id", "created_at", "updated_at"}).
		AddRow(mockUserID, time.Now(), time.Now())
	mockPool.ExpectQuery(`INSERT INTO users`).
		WithArgs(userEmail, userName, avatarURL).
		WillReturnRows(userInsertRows)

	mockPool.ExpectExec(`INSERT INTO auth_providers`).
		WithArgs(mockUserID, models.ProviderGoogle, providerID).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// Mocking authenticator
	mockClaims := &auth.ProviderClaims{
		Email:          userEmail,
		ProviderUserID: providerID,
		Name:           userName,
		AvatarURL:      avatarURL,
	}
	mockedAuthenticator := &mockProviderAuthenticator{
		mockClaims,
		nil,
	}
	providerRegistry := auth.Authenticators{
		models.ProviderGoogle: mockedAuthenticator,
	}

	ts := newTestServer(t, mockPool, providerRegistry)

	resp := ts.sendOIDCRequest(models.ProviderGoogle, "very secret token")

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var authResponse models.AuthResponse

	err = json.NewDecoder(resp.Body).Decode(&authResponse)
	require.NoError(t, err)

	require.NotEmpty(t, authResponse.Token)

	require.NoError(t, mockPool.ExpectationsWereMet())

}

func TestOIDC_SuccessfulLogin_ExistingUser(t *testing.T) {
	mockUserID := uuid.New()
	providerID := "google123"
	userEmail := "test@gmail.com"
	userName := "testUsername"
	avatarURL := "test-url.com"

	// Mocking DB Connection pool
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	authProviderRows := pgxmock.NewRows([]string{"id", "user_id", "provider_name", "provider_id"}).
		AddRow(uuid.New(), mockUserID, models.ProviderGoogle, providerID)
	mockPool.ExpectQuery(`SELECT (.+) FROM auth_providers WHERE provider_name = \$1 AND provider_id = \$2`).
		WithArgs(models.ProviderGoogle, providerID).
		WillReturnRows(authProviderRows)

	userRows := pgxmock.NewRows([]string{"id", "email", "name", "avatar_url", "created_at", "updated_at"}).
		AddRow(mockUserID, userEmail, userName, avatarURL, time.Now(), time.Now())
	mockPool.ExpectQuery(`
        SELECT id, email, name, avatar_url, created_at, updated_at
        FROM users
        WHERE id = \$1`).WithArgs(mockUserID).WillReturnRows(userRows)

	// Mocking authenticator
	mockClaims := &auth.ProviderClaims{
		Email:          userEmail,
		ProviderUserID: providerID,
		Name:           userName,
		AvatarURL:      avatarURL,
	}
	mockedAuthenticator := &mockProviderAuthenticator{
		mockClaims,
		nil,
	}
	providerRegistry := auth.Authenticators{
		models.ProviderGoogle: mockedAuthenticator,
	}

	ts := newTestServer(t, mockPool, providerRegistry)

	resp := ts.sendOIDCRequest(models.ProviderGoogle, "another very secret token")

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var authResponse models.AuthResponse

	err = json.NewDecoder(resp.Body).Decode(&authResponse)
	require.NoError(t, err)

	require.NotEmpty(t, authResponse.Token)

	require.NoError(t, mockPool.ExpectationsWereMet())

}
