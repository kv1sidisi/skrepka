package storage

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/kv1sidisi/skrepka/internal/models"
)

// PactRepository handles all database operations related to pacts.
type PactRepository struct {
	Db DBConnection
}

// PactParams defines input parameters for resolving user.
// It is used by PactRepository.
type PactParams struct {
	Title       string
	Description string
	CreatorID   uuid.UUID
	Status      string
}

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
