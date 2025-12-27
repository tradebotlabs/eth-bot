package models

import (
	"time"

	"github.com/google/uuid"
)

// UserRole represents user role types
type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleTrader UserRole = "trader"
	RoleViewer UserRole = "viewer"
)

// User represents a user in the system
type User struct {
	ID                     uuid.UUID  `json:"id" db:"id"`
	Email                  string     `json:"email" db:"email"`
	PasswordHash           string     `json:"-" db:"password_hash"` // Never expose password hash
	FullName               string     `json:"full_name" db:"full_name"`
	Role                   UserRole   `json:"role" db:"role"`
	IsActive               bool       `json:"is_active" db:"is_active"`
	IsEmailVerified        bool       `json:"is_email_verified" db:"is_email_verified"`
	EmailVerificationToken *string    `json:"-" db:"email_verification_token"`
	PasswordResetToken     *string    `json:"-" db:"password_reset_token"`
	PasswordResetExpires   *time.Time `json:"-" db:"password_reset_expires"`
	LastLoginAt            *time.Time `json:"last_login_at" db:"last_login_at"`
	CreatedAt              time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at" db:"updated_at"`
}

// UserCreateRequest represents the request to create a new user
type UserCreateRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	FullName string `json:"full_name" validate:"required,min=2"`
}

// UserUpdateRequest represents the request to update user details
type UserUpdateRequest struct {
	FullName *string `json:"full_name,omitempty"`
	Email    *string `json:"email,omitempty" validate:"omitempty,email"`
}

// PasswordChangeRequest represents a password change request
type PasswordChangeRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

// PasswordResetRequest represents a password reset request
type PasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// PasswordResetConfirm represents password reset confirmation
type PasswordResetConfirm struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// UserResponse is the public user response (no sensitive data)
type UserResponse struct {
	ID              uuid.UUID  `json:"id"`
	Email           string     `json:"email"`
	FullName        string     `json:"full_name"`
	Role            UserRole   `json:"role"`
	IsActive        bool       `json:"is_active"`
	IsEmailVerified bool       `json:"is_email_verified"`
	LastLoginAt     *time.Time `json:"last_login_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:              u.ID,
		Email:           u.Email,
		FullName:        u.FullName,
		Role:            u.Role,
		IsActive:        u.IsActive,
		IsEmailVerified: u.IsEmailVerified,
		LastLoginAt:     u.LastLoginAt,
		CreatedAt:       u.CreatedAt,
	}
}
