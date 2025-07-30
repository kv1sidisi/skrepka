package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kv1sidisi/skrepka/internal/models"
	"log/slog"
	"net/http"
	"strings"
)

// JwtMiddleware provides http middleware for JWT authentication.
// It checks for valid JSON Web Token in request header.
// If token is valid, it passes request to next handler.
// If not, it returns 401 Unauthorized error.
type JwtMiddleware struct {
	log       *slog.Logger
	next      http.Handler
	jwtSecret string
}

func NewJwtMiddleware(log *slog.Logger, next http.Handler, jwtSecret string) *JwtMiddleware {
	return &JwtMiddleware{
		log:       log,
		next:      next,
		jwtSecret: jwtSecret,
	}
}

// ServeHTTP extracts token from Authorization header.
// It validates token and if it is valid, extracts user ID.
// User ID is then put into request context.
// Request is passed to next handler in chain.
func (j *JwtMiddleware) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	const op = "handler.JwtMiddleware.ServeHTTP"
	log := j.log.With(slog.String("op", op))

	header := req.Header
	token := header.Get("Authorization")
	if token == "" {
		log.Error("Authorization token is empty")
		http.Error(w, "Authorization token is empty", http.StatusUnauthorized)
		return
	}

	bearer, tokenP, ok := strings.Cut(token, " ")
	if !ok || bearer != "Bearer" {
		log.Error("authorization token must contain Bearer and Token parts")
		http.Error(w, "Authorization header is invalid", http.StatusUnauthorized)
		return
	}

	claims := &models.AppClaims{}
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.jwtSecret), nil
	}
	_, err := jwt.ParseWithClaims(tokenP, claims, keyFunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			log.Warn("token is expired", "error", err)
		} else {
			log.Warn("failed to parse token", "error", err)
		}

		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	userID := claims.UserID
	if userID == "" {
		log.Error("user_id is empty in token claims")
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}

	ctx := context.WithValue(req.Context(), models.UserIDKey, userID)

	reqWithCtx := req.WithContext(ctx)

	j.next.ServeHTTP(w, reqWithCtx)
}
