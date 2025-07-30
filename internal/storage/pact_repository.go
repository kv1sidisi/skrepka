package storage

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/kv1sidisi/skrepka/internal/models"
)

// PactRepository handles database operations for pacts.
type PactRepository struct {
	Db DBConnection
}

// PactParams contains parameters for creating a pact.
type PactParams struct {
	Title       string
	Description string
	CreatorID   uuid.UUID
	Status      string
}

// CreatePact inserts a new pact into the database.
// Returns the new pact model.
func (p *PactRepository) CreatePact(ctx context.Context, params *PactParams) (*models.Pact, error) {
	query := `
		INSERT INTO pacts (title, description, status, creator_user_id) 
		VALUES ($1, $2, $3, $4)
		RETURNING *`

	var newPact models.Pact
	err := p.Db.QueryRow(ctx, query, params.Title, params.Description, params.Status, params.CreatorID).Scan(
		&newPact.ID,
		&newPact.Title,
		&newPact.Description,
		&newPact.Status,
		&newPact.CreatorID,
		&newPact.CreatedAt,
		&newPact.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create new pact: %w", err)
	}
	return &newPact, nil
}
