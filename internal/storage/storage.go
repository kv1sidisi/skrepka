package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kv1sidisi/skrepka/internal/config"
	"log/slog"
)

// DBConnection is interface for database operations.
// It helps to use real database or mock for tests.
type DBConnection interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Close()
}

// Storage manages database connection pool.
type Storage struct {
	pool *pgxpool.Pool
}

// New creates new connection pool to PostgreSQL database.
// Returns initialized Storage instance.
func New(ctx context.Context, cfg *config.Config) (*Storage, error) {
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s", cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.SSLMode)
	safeDSN := fmt.Sprintf("postgresql://%s:***@%s:%s/%s?sslmode=%s", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.SSLMode)
	slog.Debug("connecting to database", slog.String("dsn", safeDSN))

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}
	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	slog.Info("successfully connected to the database")

	return &Storage{
		pool: pool,
	}, nil
}

// Close closes database connection pool.
func (s *Storage) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

// UserRepository returns new repository for user-related database operations.
func (s *Storage) UserRepository() *UserRepository {
	return &UserRepository{Db: s.pool}
}
