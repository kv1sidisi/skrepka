package pact

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/kv1sidisi/skrepka/internal/models"
	"github.com/kv1sidisi/skrepka/internal/storage"
	"log/slog"
	"strings"
)

type PactRepository interface {
	CreatePact(ctx context.Context, params *storage.PactParams) (*models.Pact, error)
}

type Service struct {
	pactRepo PactRepository
	log      *slog.Logger
}

func NewService(log *slog.Logger, pactRepo PactRepository) *Service {
	return &Service{
		log:      log,
		pactRepo: pactRepo,
	}
}

func (s *Service) CreatePact(ctx context.Context, title, description string, creatorID uuid.UUID) (*models.Pact, error) {
	const op = "PactService.CreatePact"
	log := s.log.With(slog.String("op", op), slog.String("CreatorID", creatorID.String()))

	log.Info("attempting to create pact")

	title = strings.TrimSpace(title)
	if title == "" {
		log.Warn("validation failed: title is empty")
		return nil, models.ErrValidation
	}

	description = strings.TrimSpace(description)
	if description == "" {
		log.Warn("validation failed: description is empty")
		return nil, models.ErrValidation
	}

	params := &storage.PactParams{
		Title:       title,
		Description: description,
		CreatorID:   creatorID,
		Status:      "draft",
	}

	log.Info("calling repository to create pact")
	pact, err := s.pactRepo.CreatePact(ctx, params)
	if err != nil {
		log.Error("failed to create pact in repository", "error", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("pact created successfully", slog.String("pactID", pact.ID.String()))

	return pact, nil
}
