package storage

import (
	"context"
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

}
