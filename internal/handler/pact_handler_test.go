package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kv1sidisi/skrepka/internal/models"
	"github.com/stretchr/testify/require"
	"log/slog"
)

type mockPactService struct {
	pact *models.Pact
	err  error
}

func (m *mockPactService) CreatePact(ctx context.Context, title, description string, creatorID uuid.UUID) (*models.Pact, error) {
	if m.err != nil {
		return nil, m.err
	}
	pactCopy := *m.pact
	return &pactCopy, nil
}

func (m *mockPactService) UpdatePact(ctx context.Context, pactID, userID uuid.UUID, title, description *string) (*models.Pact, error) {
	return nil, nil
}

func TestPactHandler_handleCreatePact(t *testing.T) {
	mockUserID := uuid.New()
	mockPactID := uuid.New()

	expectedPact := &models.Pact{
		ID:          mockPactID,
		Title:       "Test Pact",
		Description: "Test Description",
		Status:      "draft",
		CreatorID:   mockUserID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	testCases := []struct {
		name           string
		requestBody    map[string]interface{}
		mockService    *mockPactService
		expectedStatus int
		expectBody     bool
	}{
		{
			name: "Success",
			requestBody: map[string]interface{}{
				"title":       "Test Pact",
				"description": "Test Description",
			},
			mockService:    &mockPactService{pact: expectedPact, err: nil},
			expectedStatus: http.StatusCreated,
			expectBody:     true,
		},
		{
			name: "Failure: Service validation error",
			requestBody: map[string]interface{}{
				"title": "",
			},
			mockService:    &mockPactService{pact: nil, err: models.ErrValidation},
			expectedStatus: http.StatusBadRequest,
			expectBody:     false,
		},
		{
			name: "Failure: Service internal error",
			requestBody: map[string]interface{}{
				"title": "Good Title",
			},
			mockService:    &mockPactService{pact: nil, err: fmt.Errorf("database is down")},
			expectedStatus: http.StatusInternalServerError,
			expectBody:     false,
		},
		{
			name:           "Failure: Invalid JSON body",
			requestBody:    nil,
			mockService:    &mockPactService{},
			expectedStatus: http.StatusBadRequest,
			expectBody:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pactHandler := NewPactHandler(slog.Default(), tc.mockService)

			var reqBody []byte
			var err error
			if tc.name == "Failure: Invalid JSON body" {
				reqBody = []byte(`{"title": "bad json"`)
			} else {
				reqBody, err = json.Marshal(tc.requestBody)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/pacts/", bytes.NewReader(reqBody))

			ctx := context.WithValue(req.Context(), models.UserIDKey, mockUserID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			pactHandler.ServeHTTP(rr, req)

			require.Equal(t, tc.expectedStatus, rr.Code)

			if tc.expectBody {
				var responsePact models.Pact
				err := json.NewDecoder(rr.Body).Decode(&responsePact)
				require.NoError(t, err)
				require.Equal(t, expectedPact.Title, responsePact.Title)
				require.Equal(t, expectedPact.Description, responsePact.Description)
			}
		})
	}
}
