package models

import "errors"

var (
	// ErrValidation means client provided invalid data.
	// For example, token is bad or some fields are missing.
	// Usually causes 400 or 401 HTTP error.
	ErrValidation = errors.New("validation error")

	// ErrProvider means external provider like Google cannot validate token.
	ErrProvider = errors.New("provider error")

	// ErrNotFound means requested item was not found in database.
	ErrNotFound = errors.New("not found")
)
