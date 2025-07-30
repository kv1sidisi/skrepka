package storage

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kv1sidisi/skrepka/internal/models"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/require"
)

func TestPactRepository_CreatePact(t *testing.T) {
	mockPactID := uuid.New()
	mockCreatorID := uuid.New()
	mockTime := time.Now().UTC().Truncate(time.Millisecond)

	inputParams := &PactParams{
		Title:       "Test Pact",
		Description: "A description for the test pact.",
		CreatorID:   mockCreatorID,
		Status:      "draft",
	}

	expectedPact := &models.Pact{
		ID:          mockPactID,
		Title:       inputParams.Title,
		Description: inputParams.Description,
		Status:      inputParams.Status,
		CreatorID:   inputParams.CreatorID,
		CreatedAt:   mockTime,
		UpdatedAt:   mockTime,
	}

	testCases := []struct {
		name          string
		mockSetup     func(mock pgxmock.PgxPoolIface)
		expectedPact  *models.Pact
		expectedError bool
	}{
		{
			name: "Success",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "title", "description", "status", "creator_user_id", "created_at", "updated_at"}).
					AddRow(expectedPact.ID, expectedPact.Title, expectedPact.Description, expectedPact.Status, expectedPact.CreatorID, expectedPact.CreatedAt, expectedPact.UpdatedAt)

				query := regexp.QuoteMeta(`
		INSERT INTO pacts (title, description, status, creator_user_id) 
		VALUES ($1, $2, $3, $4) 
		RETURNING *`)

				mock.ExpectQuery(query).
					WithArgs(inputParams.Title, inputParams.Description, inputParams.Status, inputParams.CreatorID).
					WillReturnRows(rows)
			},
			expectedPact:  expectedPact,
			expectedError: false,
		},
		{
			name: "Database Error",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				query := regexp.QuoteMeta(`
		INSERT INTO pacts (title, description, status, creator_user_id) 
		VALUES ($1, $2, $3, $4) 
		RETURNING *`)

				mock.ExpectQuery(query).
					WithArgs(inputParams.Title, inputParams.Description, inputParams.Status, inputParams.CreatorID).
					WillReturnError(fmt.Errorf("mock db error"))
			},
			expectedPact:  nil,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			tc.mockSetup(mockPool)

			repo := &PactRepository{Db: mockPool}
			pact, err := repo.CreatePact(context.Background(), inputParams)

			if tc.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedPact, pact)
			}

			require.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
