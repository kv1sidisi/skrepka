package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kv1sidisi/skrepka/internal/models"
)

// UserRepository handles database operations related to users.
type UserRepository struct {
	db DBConnection
}

// ResolveUserByProvider finds an existing user or creates a new one based on provider information.
func (r *UserRepository) ResolveUserByProvider(ctx context.Context, params *ResolveUserParams) (*models.User, error) {
	// Step 1: Attempt to find the user directly via the auth provider.
	query := `
        SELECT id, user_id, provider_name, provider_id
        FROM auth_providers
        WHERE provider_name = $1 AND provider_id = $2`
	var authProvider models.AuthProvider
	err := r.db.QueryRow(ctx, query, params.ProviderName, params.ProviderID).Scan(
		&authProvider.ID,
		&authProvider.UserID,
		&authProvider.ProviderName,
		&authProvider.ProviderID,
	)

	switch err {
	case nil:
		// Case 1: Auth provider found. Fetch the associated user and return.
		return r.findUserByID(ctx, authProvider.UserID)
	case pgx.ErrNoRows:
		// Case 2 & 3: Auth provider not found. Proceed to check by email.
	default:
		// An unexpected database error occurred.
		return nil, fmt.Errorf("failed to find auth provider: %w", err)
	}

	// Step 2: Attempt to find the user by email.
	user, err := r.findUserByEmail(ctx, params.Email)
	switch err {
	case nil:
		// Case 2: User found. Create the new auth provider for this existing user.
		err = r.createAuthProvider(ctx, user.ID, params)
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
	newUser, err := r.createUser(ctx, params)
	if err != nil {
		return nil, err
	}

	err = r.createAuthProvider(ctx, newUser.ID, params)
	if err != nil {
		// In a real-world scenario, this two-step write (user and provider) should be wrapped in a transaction.
		return nil, err
	}

	return newUser, nil
}

// findUserByID retrieves a user by their primary key.
func (r *UserRepository) findUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
        SELECT id, email, name, avatar_url, created_at, updated_at
        FROM users
        WHERE id = $1`
	var user models.User
	err := r.db.QueryRow(ctx, query, id).Scan(
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
func (r *UserRepository) findUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
        SELECT id, email, name, avatar_url, created_at, updated_at
        FROM users
        WHERE email = $1`
	var user models.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// createUser inserts a new user into the database.
func (r *UserRepository) createUser(ctx context.Context, params *ResolveUserParams) (*models.User, error) {
	query := `
        INSERT INTO users (email, name, avatar_url)
        VALUES ($1, $2, $3)
        RETURNING id, created_at, updated_at`
	var newUser models.User
	err := r.db.QueryRow(ctx, query, params.Email, params.Name, params.AvatarURL).Scan(
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
	return &newUser, nil
}

// createAuthProvider inserts a new auth_provider record, linking a user to a provider ID.
func (r *UserRepository) createAuthProvider(ctx context.Context, userID uuid.UUID, params *ResolveUserParams) error {
	query := `
        INSERT INTO auth_providers (user_id, provider_name, provider_id)
        VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, query, userID, params.ProviderName, params.ProviderID)
	if err != nil {
		return fmt.Errorf("failed to create auth provider: %w", err)
	}
	return nil
}
