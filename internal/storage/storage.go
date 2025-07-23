package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kv1sidisi/skrepka/internal/config"
	"github.com/kv1sidisi/skrepka/internal/models"
	"log/slog"
)

type Connection interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Close()
}

type Storage struct {
	pool Connection
}

// NewStorage creates connection pool to postgres and initializes Storage with created pool.
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

// Close closes connection to postgres.
func (s *Storage) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

// ResolveUserParams input params for user resolving.
type ResolveUserParams struct {
	ProviderName string
	ProviderID   string
	Email        string
	Name         string
	AvatarURL    string
}

// ResolveUserByProvider finds an existing user or creates a new one based on the provider's information.
// It handles three scenarios:
// 1. The user and auth provider already exist.
// 2. The user exists (matched by email), but the auth provider is new.
// 3. Both the user and the auth provider are new.
func (s *Storage) ResolveUserByProvider(ctx context.Context, params *ResolveUserParams) (*models.User, error) {
	// 1. Find the authentication provider.
	query := `
        SELECT id, user_id, provider_name, provider_id
        FROM auth_providers
        WHERE provider_name = $1 AND provider_id = $2`

	var authProvider models.AuthProvider

	err := s.pool.QueryRow(ctx, query, params.ProviderName, params.ProviderID).Scan(
		&authProvider.ID,
		&authProvider.UserID,
		&authProvider.ProviderName,
		&authProvider.ProviderID,
	)

	if err == nil {
		// Case 1: Auth provider found, fetch the associated user.
		var user models.User
		userQuery := `
            SELECT id, email, name, avatar_url, created_at, updated_at
            FROM users
            WHERE id = $1`

		err = s.pool.QueryRow(ctx, userQuery, authProvider.UserID).Scan(
			&user.ID,
			&user.Email,
			&user.Name,
			&user.AvatarURL,
			&user.CreatedAt,
			&user.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to find user by id %s: %w", authProvider.UserID, err)
		}

		return &user, nil

	} else if errors.Is(err, pgx.ErrNoRows) {
		// Case 2 & 3: Auth provider not found, try to find user by email.
		var user models.User
		userQuery := `
            SELECT id, email, name, avatar_url, created_at, updated_at
            FROM users
            WHERE email = $1`

		err = s.pool.QueryRow(ctx, userQuery, params.Email).Scan(
			&user.ID,
			&user.Email,
			&user.Name,
			&user.AvatarURL,
			&user.CreatedAt,
			&user.UpdatedAt,
		)

		if err == nil {
			// Case 2: User found, create a new auth provider for them.
			createProviderQuery := `
                INSERT INTO auth_providers (user_id, provider_name, provider_id)
                VALUES ($1, $2, $3)`

			_, err = s.pool.Exec(ctx, createProviderQuery, user.ID, params.ProviderName, params.ProviderID)
			if err != nil {
				return nil, fmt.Errorf("failed to create auth provider for existing user: %w", err)
			}

			return &user, nil

		} else if err == pgx.ErrNoRows {
			// Case 3: User not found, create a new user and a new auth provider.
			createUserQuery := `
                INSERT INTO users (email, name, avatar_url)
                VALUES ($1, $2, $3)
                RETURNING id, created_at, updated_at`

			var newUser models.User
			err = s.pool.QueryRow(ctx, createUserQuery, params.Email, params.Name, params.AvatarURL).Scan(
				&newUser.ID,
				&newUser.CreatedAt,
				&newUser.UpdatedAt,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to create new user: %w", err)
			}

			newUser.Email = params.Email
			newUser.Name = params.Name
			newUser.AvatarURL = params.AvatarURL

			createProviderQuery := `
                INSERT INTO auth_providers (user_id, provider_name, provider_id)
                VALUES ($1, $2, $3)`

			_, err = s.pool.Exec(ctx, createProviderQuery, newUser.ID, params.ProviderName, params.ProviderID)
			if err != nil {
				return nil, fmt.Errorf("failed to create auth provider for new user: %w", err)
			}

			return &newUser, nil
		} else {
			return nil, fmt.Errorf("failed to find user by email: %w", err)
		}

	} else {
		return nil, fmt.Errorf("failed to find auth provider: %w", err)
	}
}
