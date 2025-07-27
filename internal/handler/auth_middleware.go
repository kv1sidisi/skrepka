package handler

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"log/slog"
	"net/http"
	"strings"
)

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

type AppClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type contextKey string

const userIDKey = contextKey("userID")

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

	tParts := strings.Fields(token)
	if len(tParts) != 2 {
		log.Error("authorization token must contain Bearer and Token parts")
		http.Error(w, "Authorization header is invalid", http.StatusUnauthorized)
		return
	}

	if tParts[0] != "Bearer" {
		log.Error("authorization token first part must be Bearer")
		http.Error(w, "Authorization header is invalid", http.StatusUnauthorized)
		return
	}

	claims := &AppClaims{}
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.jwtSecret), nil
	}
	pToken, err := jwt.ParseWithClaims(tParts[1], claims, keyFunc)
	if err != nil {
		log.Error("failed to parse token", "error", err)
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}
	if !pToken.Valid {
		log.Error("parsed token is invalid")
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	userID := claims.UserID
	if userID == "" {
		log.Error("user_id is empty in token claims")
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}

	ctx := context.WithValue(req.Context(), userIDKey, userID)

	reqWithCtx := req.WithContext(ctx)

	j.next.ServeHTTP(w, reqWithCtx)
}
