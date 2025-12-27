package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/eth-trading/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// UserRepository implements user data access
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (
			id, email, password_hash, full_name, role,
			is_active, is_email_verified, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`

	_, err := r.db.Exec(
		query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.Role,
		user.IsActive,
		user.IsEmailVerified,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, full_name, role,
		       is_active, is_email_verified, email_verification_token,
		       password_reset_token, password_reset_expires, last_login_at,
		       created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := r.db.Get(&user, query, id)
	if err == sql.ErrNoRows {
		return nil, models.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, full_name, role,
		       is_active, is_email_verified, email_verification_token,
		       password_reset_token, password_reset_expires, last_login_at,
		       created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user models.User
	err := r.db.Get(&user, query, email)
	if err == sql.ErrNoRows {
		return nil, models.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(user *models.User) error {
	query := `
		UPDATE users
		SET email = $2,
		    password_hash = $3,
		    full_name = $4,
		    role = $5,
		    is_active = $6,
		    is_email_verified = $7,
		    email_verification_token = $8,
		    password_reset_token = $9,
		    password_reset_expires = $10,
		    updated_at = $11
		WHERE id = $1
	`

	user.UpdatedAt = time.Now()

	result, err := r.db.Exec(
		query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.Role,
		user.IsActive,
		user.IsEmailVerified,
		user.EmailVerificationToken,
		user.PasswordResetToken,
		user.PasswordResetExpires,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rows == 0 {
		return models.ErrUserNotFound
	}

	return nil
}

// UpdateLastLogin updates the user's last login timestamp
func (r *UserRepository) UpdateLastLogin(userID uuid.UUID) error {
	query := `
		UPDATE users
		SET last_login_at = $2
		WHERE id = $1
	`

	_, err := r.db.Exec(query, userID, time.Now())
	if err != nil {
		return fmt.Errorf("update last login: %w", err)
	}

	return nil
}

// EmailExists checks if an email is already registered
func (r *UserRepository) EmailExists(email string) (bool, error) {
	query := `
		SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)
	`

	var exists bool
	err := r.db.Get(&exists, query, email)
	if err != nil {
		return false, fmt.Errorf("check email exists: %w", err)
	}

	return exists, nil
}

// Delete deletes a user (soft delete by setting is_active = false)
func (r *UserRepository) Delete(userID uuid.UUID) error {
	query := `
		UPDATE users
		SET is_active = false, updated_at = $2
		WHERE id = $1
	`

	_, err := r.db.Exec(query, userID, time.Now())
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	return nil
}

// List retrieves all users with pagination
func (r *UserRepository) List(limit, offset int) ([]*models.User, error) {
	query := `
		SELECT id, email, password_hash, full_name, role,
		       is_active, is_email_verified, last_login_at,
		       created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	var users []*models.User
	err := r.db.Select(&users, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	return users, nil
}

// Count returns the total number of users
func (r *UserRepository) Count() (int, error) {
	query := `SELECT COUNT(*) FROM users`

	var count int
	err := r.db.Get(&count, query)
	if err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}

	return count, nil
}
