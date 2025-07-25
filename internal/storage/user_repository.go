package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kv1sidisi/skrepka/internal/models"
)

// UserRepository handles all database operations related to users.
type UserRepository struct {
	db DBConnection
}

// ResolveUserByProvider finds existing user or creates new one based on provider information.
// This function handles main logic for user sign-in and sign-up.
// Returns pointer to user model.
func (r *UserRepository) ResolveUserByProvider(ctx context.Context, params *ResolveUserParams) (*models.User, error) {
	// Step 1: Attempt to find user directly via auth provider.
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

	switch {
	case err == nil:
		// Case 1: Auth provider found. Fetch associated user and return.
		return r.findUserByID(ctx, authProvider.UserID)
	case errors.Is(err, pgx.ErrNoRows):
		// Case 2 & 3: Auth provider not found. Proceed to check by email.
	default:
		// An unexpected database error occurred.
		return nil, fmt.Errorf("failed to find auth provider: %w", err)
	}

	// Step 2: Attempt to find user by email.
	user, err := r.findUserByEmail(ctx, params.Email)
	switch {
	case err == nil:
		// Case 2: User found. Create new auth provider for this existing user.
		err = r.createAuthProvider(ctx, user.ID, params)
		if err != nil {
			return nil, err
		}
		return user, nil
	case errors.Is(err, pgx.ErrNoRows):
		// Case 3: User not found by email either. Proceed to create new user.
	default:
		// An unexpected database error occurred.
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	// Step 3: Create new user and new auth provider.
	newUser, err := r.createUser(ctx, params)
	if err != nil {
		return nil, err
	}

	err = r.createAuthProvider(ctx, newUser.ID, params)
	if err != nil {
		// In real-world scenario, this two-step write (user and provider) should be wrapped in transaction.
		return nil, err
	}

	return newUser, nil
}

// findUserByID retrieves user by their primary key.
// Returns pointer to user model.
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

// findUserByEmail retrieves user by their email address.
// Returns pointer to user model.
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

// createUser inserts new user into database.
// Returns pointer to new user model.
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

// createAuthProvider inserts new auth_provider record, linking user to provider ID.
// Returns nil on success.
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
