package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kv1sidisi/skrepka/internal/config"
	"github.com/kv1sidisi/skrepka/internal/models"
	"log/slog"
)

// Connection defines the interface for database operations required by Storage.
// It is satisfied by both a real *pgxpool.Pool and the pgxmock.PgxPoolIface.
type Connection interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Close()
}

type Storage struct {
	pool Connection
}

// NewStorage creates a connection pool to postgres and initializes Storage.
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

// Close closes the database connection pool.
func (s *Storage) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

// ResolveUserParams defines the input parameters for resolving a user.
type ResolveUserParams struct {
	ProviderName string
	ProviderID   string
	Email        string
	Name         string
	AvatarURL    string
}

// ResolveUserByProvider finds an existing user or creates a new one based on provider information.
// The method is structured as a sequence of steps, using early returns and switch statements for clarity.
func (s *Storage) ResolveUserByProvider(ctx context.Context, params *ResolveUserParams) (*models.User, error) {
	// Step 1: Attempt to find the user directly via the auth provider.
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

	switch err {
	case nil:
		// Case 1: Auth provider found. Fetch the associated user and return.
		return s.findUserByID(ctx, authProvider.UserID)
	case pgx.ErrNoRows:
		// Case 2 & 3: Auth provider not found. Proceed to check by email.
	default:
		// An unexpected database error occurred.
		return nil, fmt.Errorf("failed to find auth provider: %w", err)
	}

	// Step 2: Attempt to find the user by email.
	user, err := s.findUserByEmail(ctx, params.Email)
	switch err {
	case nil:
		// Case 2: User found. Create the new auth provider for this existing user.
		err = s.createAuthProvider(ctx, user.ID, params)
		if err != nil {
			return nil, err
		}
		return user, nil
	case pgx.ErrNoRows:
		// Case 3: User not found by email either. Proceed to create a new user.
	default:
		// An unexpected database error occurred.
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	// Step 3: Create a new user and a new auth provider.
	newUser, err := s.createUser(ctx, params)
	if err != nil {
		return nil, err
	}

	err = s.createAuthProvider(ctx, newUser.ID, params)
	if err != nil {
		// In a real-world scenario, this two-step write (user and provider) should be wrapped in a transaction.
		// This will be addressed in Stage 5, as per the development plan.
		// For now, we just return the error.
		return nil, err
	}

	return newUser, nil
}

// findUserByID retrieves a user by their primary key.
func (s *Storage) findUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
        SELECT id, email, name, avatar_url, created_at, updated_at
        FROM users
        WHERE id = $1`
	var user models.User
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by id %s: %w", id, err)
	}
	return &user, nil
}

// findUserByEmail retrieves a user by their email address.
func (s *Storage) findUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
        SELECT id, email, name, avatar_url, created_at, updated_at
        FROM users
        WHERE email = $1`
	var user models.User
	err := s.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	// We return the raw pgx.ErrNoRows to the caller, as it's part of the control flow.
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// createUser inserts a new user into the database.
func (s *Storage) createUser(ctx context.Context, params *ResolveUserParams) (*models.User, error) {
	query := `
        INSERT INTO users (email, name, avatar_url)
        VALUES ($1, $2, $3)
        RETURNING id, created_at, updated_at`
	var newUser models.User
	err := s.pool.QueryRow(ctx, query, params.Email, params.Name, params.AvatarURL).Scan(
		&newUser.ID,
		&newUser.CreatedAt,
		&newUser.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create new user: %w", err)
	}
	// Populate the rest of the user model from the input params.
	newUser.Email = params.Email
	newUser.Name = params.Name
	newUser.AvatarURL = params.AvatarURL
	return &newUser, nil
}

// createAuthProvider inserts a new auth_provider record, linking a user to a provider ID.
func (s *Storage) createAuthProvider(ctx context.Context, userID uuid.UUID, params *ResolveUserParams) error {
	query := `
        INSERT INTO auth_providers (user_id, provider_name, provider_id)
        VALUES ($1, $2, $3)`
	_, err := s.pool.Exec(ctx, query, userID, params.ProviderName, params.ProviderID)
	if err != nil {
		return fmt.Errorf("failed to create auth provider: %w", err)
	}
	return nil
}
