package pact

import (
	"context"
	"fmt"
	"github.com/kv1sidisi/skrepka/internal/storage"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kv1sidisi/skrepka/internal/models"
	"github.com/stretchr/testify/require"
	"log/slog"
)

type mockPactRepository struct {
	pact *models.Pact
	err  error
}

func (m *mockPactRepository) CreatePact(ctx context.Context, params *storage.PactParams) (*models.Pact, error) {
	if m.err != nil {
		return nil, m.err
	}
	pact := &models.Pact{
		ID:          uuid.New(),
		Title:       params.Title,
		Description: params.Description,
		Status:      params.Status,
		CreatorID:   uuid.New(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	return pact, nil
}

func TestPactService_CreatePact(t *testing.T) {
	mockCreatorID := uuid.New()

	testCases := []struct {
		name          string
		title         string
		description   string
		creatorID     uuid.UUID
		mockRepo      PactRepository
		expectError   bool
		errorContains string
	}{
		{
			name:        "Success: Happy path",
			title:       "My Test Pact",
			description: "A valid description.",
			creatorID:   mockCreatorID,
			mockRepo:    &mockPactRepository{pact: &models.Pact{}, err: nil},
			expectError: false,
		},
		{
			name:          "Failure: Validation error for empty title",
			title:         "   ",
			description:   "A description.",
			creatorID:     mockCreatorID,
			mockRepo:      &mockPactRepository{},
			expectError:   true,
			errorContains: models.ErrValidation.Error(),
		},
		{
			name:          "Failure: Repository returns an error",
			title:         "A Valid Title",
			description:   "A description.",
			creatorID:     mockCreatorID,
			mockRepo:      &mockPactRepository{pact: nil, err: fmt.Errorf("database is down")},
			expectError:   true,
			errorContains: "database is down",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pactService := NewService(slog.Default(), tc.mockRepo)

			pact, err := pactService.CreatePact(context.Background(), tc.title, tc.description, tc.creatorID)

			if tc.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorContains)
				require.Nil(t, pact)
			} else {
				require.NoError(t, err)
				require.NotNil(t, pact)
				require.Equal(t, tc.title, pact.Title)
				require.Equal(t, "draft", pact.Status) // Check that the service set the status correctly
				require.NotEqual(t, uuid.Nil, pact.ID) // Check that the DB (mock) generated an ID
			}
		})
	}
}
