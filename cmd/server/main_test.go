// main_test.go
package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHealthHandler verifies the behavior of the health check endpoint
// across different HTTP methods.
func TestHealthHandler(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Successful Request (GET)",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectedBody:   "{\"status\":\"ok\"}\n",
		},
		{
			name:           "Method Not Allowed (POST)",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method Not Allowed\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, "/health", nil)
			if err != nil {
				t.Fatalf("Could not create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(health)

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tc.expectedStatus)
			}

			if body := rr.Body.String(); body != tc.expectedBody {
				t.Errorf("handler returned unexpected body: got %q want %q",
					body, tc.expectedBody)
			}
		})
	}
}
