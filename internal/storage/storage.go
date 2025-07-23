package storage

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kv1sidisi/skrepka/internal/config"
	"log/slog"
)

type Storage struct {
	pool *pgxpool.Pool
}

func NewStorage() (*Storage, error) {
	cfg := config.MustLoad()

	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s", cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.SSLMode)

	safeDSN := fmt.Sprintf("postgresql://%s:***@%s:%s/%s?sslmode=%s", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.SSLMode)

	slog.Debug("connecting to database with dsn", slog.String("dsn", safeDSN))

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err = pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping connection pool: %w", err)
	}

	slog.Info("successfully connected to the database")

	return &Storage{
		pool: pool,
	}, nil
}

func (s *Storage) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}
