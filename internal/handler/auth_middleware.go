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

// JwtAuthMiddleware creates a new middleware for JWT authentication.
// It returns a function that takes an http.Handler and returns a new http.Handler.
// This new handler will check for a valid JWT before calling the next handler in the chain.
func JwtAuthMiddleware(log *slog.Logger, jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const op = "handler.JwtAuthMiddleware"
			log := log.With(slog.String("op", op))

			header := r.Header
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
				return []byte(jwtSecret), nil
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

			ctx := context.WithValue(r.Context(), models.UserIDKey, userID)
			reqWithCtx := r.WithContext(ctx)

			next.ServeHTTP(w, reqWithCtx)
		})
	}
}
