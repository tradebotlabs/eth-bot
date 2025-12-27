package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/eth-trading/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// SessionRepository implements session data access
type SessionRepository struct {
	db *sqlx.DB
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *sqlx.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create creates a new session
func (r *SessionRepository) Create(session *models.Session) error {
	query := `
		INSERT INTO sessions (
			id, user_id, refresh_token, user_agent, ip_address,
			expires_at, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
	`

	_, err := r.db.Exec(
		query,
		session.ID,
		session.UserID,
		session.RefreshToken,
		session.UserAgent,
		session.IPAddress,
		session.ExpiresAt,
		session.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}

	return nil
}

// GetByRefreshToken retrieves a session by refresh token
func (r *SessionRepository) GetByRefreshToken(token string) (*models.Session, error) {
	query := `
		SELECT id, user_id, refresh_token, user_agent, ip_address,
		       expires_at, created_at
		FROM sessions
		WHERE refresh_token = $1
	`

	var session models.Session
	err := r.db.Get(&session, query, token)
	if err == sql.ErrNoRows {
		return nil, models.ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get session by refresh token: %w", err)
	}

	return &session, nil
}

// DeleteByUserID deletes all sessions for a user
func (r *SessionRepository) DeleteByUserID(userID uuid.UUID) error {
	query := `DELETE FROM sessions WHERE user_id = $1`

	_, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("delete sessions by user id: %w", err)
	}

	return nil
}

// Delete deletes a session by ID
func (r *SessionRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM sessions WHERE id = $1`

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	return nil
}

// DeleteExpired deletes all expired sessions
func (r *SessionRepository) DeleteExpired() error {
	query := `DELETE FROM sessions WHERE expires_at < $1`

	_, err := r.db.Exec(query, time.Now())
	if err != nil {
		return fmt.Errorf("delete expired sessions: %w", err)
	}

	return nil
}

// GetByUserID retrieves all sessions for a user
func (r *SessionRepository) GetByUserID(userID uuid.UUID) ([]*models.Session, error) {
	query := `
		SELECT id, user_id, refresh_token, user_agent, ip_address,
		       expires_at, created_at
		FROM sessions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	var sessions []*models.Session
	err := r.db.Select(&sessions, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get sessions by user id: %w", err)
	}

	return sessions, nil
}
