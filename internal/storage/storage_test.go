package storage

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kv1sidisi/skrepka/internal/models"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/require"
)

func TestStorage_FindOrCreateUserByProvider(t *testing.T) {
	mockUserID := uuid.New()
	mockTime := time.Now().UTC().Truncate(time.Millisecond)

	expectedUser := &models.User{
		ID:        mockUserID,
		Email:     "test@example.com",
		Name:      "Test User",
		AvatarURL: "http://example.com/avatar.png",
		CreatedAt: mockTime,
		UpdatedAt: mockTime,
	}

	inputParams := &ResolveUserParams{
		ProviderName: "google",
		ProviderID:   "12345",
		Email:        "test@example.com",
		Name:         "Test User",
		AvatarURL:    "http://example.com/avatar.png",
	}

	testCases := []struct {
		name          string
		mockSetup     func(mock pgxmock.PgxPoolIface)
		expectedUser  *models.User
		expectedError bool
	}{
		{
			name: "user and auth provider exist",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				authProviderRows := pgxmock.NewRows([]string{"id", "user_id", "provider_name", "provider_id"}).
					AddRow(uuid.New(), mockUserID, "google", "12345")
				mock.ExpectQuery(`SELECT id, user_id, provider_name, provider_id FROM auth_providers WHERE provider_name = \$1 AND provider_id = \$2`).
					WithArgs("google", "12345").
					WillReturnRows(authProviderRows)

				userRows := pgxmock.NewRows([]string{"id", "email", "name", "avatar_url", "created_at", "updated_at"}).
					AddRow(expectedUser.ID, expectedUser.Email, expectedUser.Name, expectedUser.AvatarURL, expectedUser.CreatedAt, expectedUser.UpdatedAt)
				mock.ExpectQuery(`SELECT id, email, name, avatar_url, created_at, updated_at FROM users WHERE id = \$1`).
					WithArgs(mockUserID).
					WillReturnRows(userRows)
			},
			expectedUser:  expectedUser,
			expectedError: false,
		},
		{
			name: "user exists, but auth provider is new",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`SELECT (.+) FROM auth_providers`).
					WithArgs("google", "12345").
					WillReturnError(pgx.ErrNoRows)

				userRows := pgxmock.NewRows([]string{"id", "email", "name", "avatar_url", "created_at", "updated_at"}).
					AddRow(expectedUser.ID, expectedUser.Email, expectedUser.Name, expectedUser.AvatarURL, expectedUser.CreatedAt, expectedUser.UpdatedAt)
				mock.ExpectQuery(`SELECT (.+) FROM users WHERE email = \$1`).
					WithArgs(inputParams.Email).
					WillReturnRows(userRows)

				mock.ExpectExec(`INSERT INTO auth_providers`).
					WithArgs(expectedUser.ID, "google", "12345").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedUser:  expectedUser,
			expectedError: false,
		},
		{
			name: "new user and new auth provider",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`SELECT (.+) FROM auth_providers`).
					WithArgs("google", "12345").
					WillReturnError(pgx.ErrNoRows)

				mock.ExpectQuery(`SELECT (.+) FROM users WHERE email = \$1`).
					WithArgs(inputParams.Email).
					WillReturnError(pgx.ErrNoRows)

				userInsertRows := pgxmock.NewRows([]string{"id", "created_at", "updated_at"}).
					AddRow(expectedUser.ID, expectedUser.CreatedAt, expectedUser.UpdatedAt)
				mock.ExpectQuery(`INSERT INTO users`).
					WithArgs(inputParams.Email, inputParams.Name, inputParams.AvatarURL).
					WillReturnRows(userInsertRows)

				mock.ExpectExec(`INSERT INTO auth_providers`).
					WithArgs(expectedUser.ID, "google", "12345").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedUser:  expectedUser,
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			tc.mockSetup(mockPool)

			storage := &Storage{pool: mockPool}
			user, err := storage.ResolveUserByProvider(context.Background(), inputParams)

			if tc.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedUser, user)
			}

			require.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
