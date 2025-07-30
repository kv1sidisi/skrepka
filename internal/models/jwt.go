package models

import "github.com/golang-jwt/jwt/v5"

type AppClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}
