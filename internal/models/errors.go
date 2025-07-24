package models

import "errors"

var (
	// ErrValidation is returned when input data provided by a client is invalid.
	// This could be an invalid token, missing fields, etc.
	// It typically results in a 400 or 401 HTTP status.
	ErrValidation = errors.New("validation error")

	// ErrProvider is returned when an external provider (e.g., Google)
	// cannot validate the token.
	ErrProvider = errors.New("provider error")

	// ErrNotFound is returned when a requested entity is not found.
	ErrNotFound = errors.New("not found")
)
