package handler

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/kv1sidisi/skrepka/internal/models"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

const testSecret = "test-secret"

type MockHandler struct {
	userID string
}

func (m *MockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(models.UserIDKey).(string)
	if ok {
		m.userID = userID
	}
	w.WriteHeader(http.StatusOK)
}

func generateJwtToken(userID string, secret string, expiration time.Duration) (string, error) {
	claims := models.AppClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func TestJwtAuthMiddleware(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mockHandler := &MockHandler{}

	jwtAuth := JwtAuthMiddleware(log, testSecret)

	protectedHandler := jwtAuth(mockHandler)

	testUserID := "user-123"

	validToken, err := generateJwtToken(testUserID, testSecret, time.Hour)
	if err != nil {
		t.Fatalf("failed to generate valid token: %v", err)
	}

	expiredToken, err := generateJwtToken(testUserID, testSecret, -time.Hour)
	if err != nil {
		t.Fatalf("failed to generate expired token: %v", err)
	}

	wrongSecretToken, err := generateJwtToken(testUserID, "wrong-secret", time.Hour)
	if err != nil {
		t.Fatalf("failed to generate wrong secret token: %v", err)
	}

	tokenWithoutUserID, err := generateJwtToken("", testSecret, time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token without user id: %v", err)
	}

	testCases := []struct {
		name           string
		header         string
		expectedStatus int
		expectedUserID string
	}{
		{
			name:           "Success",
			header:         "Bearer " + validToken,
			expectedStatus: http.StatusOK,
			expectedUserID: testUserID,
		},
		{
			name:           "NoAuthorizationHeader",
			header:         "",
			expectedStatus: http.StatusUnauthorized,
			expectedUserID: "",
		},
		{
			name:           "MalformedHeader_NoBearer",
			header:         validToken,
			expectedStatus: http.StatusUnauthorized,
			expectedUserID: "",
		},
		{
			name:           "MalformedHeader_WrongBearer",
			header:         "Bear " + validToken,
			expectedStatus: http.StatusUnauthorized,
			expectedUserID: "",
		},
		{
			name:           "InvalidToken_Expired",
			header:         "Bearer " + expiredToken,
			expectedStatus: http.StatusUnauthorized,
			expectedUserID: "",
		},
		{
			name:           "InvalidToken_WrongSignature",
			header:         "Bearer " + wrongSecretToken,
			expectedStatus: http.StatusUnauthorized,
			expectedUserID: "",
		},
		{
			name:           "InvalidToken_NoUserID",
			header:         "Bearer " + tokenWithoutUserID,
			expectedStatus: http.StatusUnauthorized,
			expectedUserID: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockHandler.userID = ""

			req := httptest.NewRequest("GET", "/", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}

			rr := httptest.NewRecorder()

			protectedHandler.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, rr.Code)
			}

			if tc.expectedStatus == http.StatusOK {
				if mockHandler.userID != tc.expectedUserID {
					t.Errorf("expected userID %q, got %q", tc.expectedUserID, mockHandler.userID)
				}
			} else {
				if mockHandler.userID != "" {
					t.Errorf("expected empty userID for failed auth, but got %q", mockHandler.userID)
				}
			}
		})
	}
}
