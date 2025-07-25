package models

import "github.com/google/uuid"

// AuthProvider represents authentication method in database.
// It links user in our system to their ID from external provider like Google.
type AuthProvider struct {
	ID           uuid.UUID `db:"id"`
	UserID       uuid.UUID `db:"user_id"`
	ProviderName Provider  `db:"provider_name"`
	ProviderID   string    `db:"provider_id"`
}
