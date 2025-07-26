package handler

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler(t *testing.T) {
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := NewHealthHandler(discardLogger)

	testCases := []struct {
		name               string
		method             string
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:               "OK",
			method:             http.MethodGet,
			expectedStatusCode: http.StatusOK,
			expectedBody:       `{"status":"ok"}`,
		},
		{
			name:               "Method Not Allowed - PUT",
			method:             http.MethodPut,
			expectedStatusCode: http.StatusMethodNotAllowed,
			expectedBody:       "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, "/health", nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)

			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, rr.Body.String())
			}
		})
	}
}
