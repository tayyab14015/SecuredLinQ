package repository

import (
	"time"

	"github.com/securedlinq/backend/internal/models"
	"gorm.io/gorm"
)

// SessionRepository handles session database operations
type SessionRepository struct {
	*Repository
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{
		Repository: NewRepository(db),
	}
}

// GetBySessionID retrieves a session by session ID
func (r *SessionRepository) GetBySessionID(sessionID string) (*models.Session, error) {
	var session models.Session
	err := r.db.Where("session_id = ? AND expires_at > ?", sessionID, time.Now()).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// Create creates a new session
func (r *SessionRepository) Create(session *models.Session) error {
	session.CreatedAt = time.Now()
	session.UpdatedAt = time.Now()
	return r.db.Create(session).Error
}

// Update updates a session
func (r *SessionRepository) Update(session *models.Session) error {
	session.UpdatedAt = time.Now()
	return r.db.Save(session).Error
}

// Delete deletes a session by session ID
func (r *SessionRepository) Delete(sessionID string) error {
	return r.db.Where("session_id = ?", sessionID).Delete(&models.Session{}).Error
}

// DeleteByUserID deletes all sessions for a user
func (r *SessionRepository) DeleteByUserID(userID int, userType string) error {
	return r.db.Where("user_id = ? AND user_type = ?", userID, userType).Delete(&models.Session{}).Error
}

// DeleteExpired deletes all expired sessions
func (r *SessionRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&models.Session{}).Error
}

// ExtendSession extends a session's expiration time
func (r *SessionRepository) ExtendSession(sessionID string, duration time.Duration) error {
	return r.db.Model(&models.Session{}).
		Where("session_id = ?", sessionID).
		Updates(map[string]interface{}{
			"expires_at": time.Now().Add(duration),
			"updated_at": time.Now(),
		}).Error
}

