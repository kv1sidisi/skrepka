package models

import "github.com/google/uuid"

// AuthProvider represents database entity about authentication method.
type AuthProvider struct {
	ID           uuid.UUID `db:"id"`
	UserID       uuid.UUID `db:"user_id"`
	ProviderName string    `db:"provider_name"`
	ProviderID   string    `db:"provider_id"`
}
