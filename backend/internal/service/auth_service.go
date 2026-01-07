package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/securedlinq/backend/internal/config"
	"github.com/securedlinq/backend/internal/models"
	"github.com/securedlinq/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Custom error types for better error handling
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrAccountDeactivated = errors.New("account is deactivated")
	ErrInvalidSession     = errors.New("invalid or expired session")
)

// AuthService handles authentication business logic
type AuthService struct {
	sessionRepo *repository.SessionRepository
	driverRepo  *repository.DriverRepository
	config      *config.Config
}

// NewAuthService creates a new auth service
func NewAuthService(sessionRepo *repository.SessionRepository, driverRepo *repository.DriverRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		sessionRepo: sessionRepo,
		driverRepo:  driverRepo,
		config:      cfg,
	}
}

// LoginCredentials represents login request data
type LoginCredentials struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// SessionInfo represents session information
type SessionInfo struct {
	SessionID string `json:"session_id"`
	UserID    int    `json:"user_id"`
	UserType  string `json:"user_type"`
	ExpiresAt int64  `json:"expires_at"`
}

// ValidateAdminCredentials validates admin username and password using config-based credentials
func (s *AuthService) ValidateAdminCredentials(username, password string) error {
	// Use config-based authentication
	if username == s.config.Admin.Username && password == s.config.Admin.Password {
		return nil
	}

	// Check if it matches config username but wrong password
	if username == s.config.Admin.Username {
		return ErrInvalidPassword
	}

	return ErrUserNotFound
}

// ValidateDriverCredentials validates driver username and password
func (s *AuthService) ValidateDriverCredentials(username, password string) (*models.Driver, error) {
	if s.driverRepo == nil {
		return nil, errors.New("driver repository not configured")
	}

	driver, err := s.driverRepo.GetByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, ErrUserNotFound
	}

	if !driver.IsActive {
		return nil, ErrAccountDeactivated
	}

	if err := bcrypt.CompareHashAndPassword([]byte(driver.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidPassword
	}

	return driver, nil
}

// GetDriverByID gets driver by ID
func (s *AuthService) GetDriverByID(driverID int) (*models.Driver, error) {
	if s.driverRepo == nil {
		return nil, errors.New("driver repository not configured")
	}
	return s.driverRepo.GetByID(uint(driverID))
}

// CreateSession creates a new session for an authenticated user
func (s *AuthService) CreateSession(userID int, userType string) (*SessionInfo, error) {
	// Generate random session ID
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(time.Duration(s.config.Session.MaxAge) * time.Second)

	session := &models.Session{
		SessionID: sessionID,
		UserID:    userID,
		UserType:  userType,
		ExpiresAt: expiresAt,
	}

	if err := s.sessionRepo.Create(session); err != nil {
		return nil, err
	}

	return &SessionInfo{
		SessionID: sessionID,
		UserID:    userID,
		UserType:  userType,
		ExpiresAt: expiresAt.Unix(),
	}, nil
}

// ValidateSession validates a session and returns session info
func (s *AuthService) ValidateSession(sessionID string) (*SessionInfo, error) {
	session, err := s.sessionRepo.GetBySessionID(sessionID)
	if err != nil {
		return nil, ErrInvalidSession
	}

	if session.ExpiresAt.Before(time.Now()) {
		s.sessionRepo.Delete(sessionID)
		return nil, ErrInvalidSession
	}

	return &SessionInfo{
		SessionID: session.SessionID,
		UserID:    session.UserID,
		UserType:  session.UserType,
		ExpiresAt: session.ExpiresAt.Unix(),
	}, nil
}

// RefreshSession extends a session's expiration time
func (s *AuthService) RefreshSession(sessionID string) error {
	duration := time.Duration(s.config.Session.MaxAge) * time.Second
	return s.sessionRepo.ExtendSession(sessionID, duration)
}

// InvalidateSession invalidates a session (logout)
func (s *AuthService) InvalidateSession(sessionID string) error {
	return s.sessionRepo.Delete(sessionID)
}

// InvalidateAllUserSessions invalidates all sessions for a user
func (s *AuthService) InvalidateAllUserSessions(userID int, userType string) error {
	return s.sessionRepo.DeleteByUserID(userID, userType)
}

// CleanupExpiredSessions removes all expired sessions
func (s *AuthService) CleanupExpiredSessions() error {
	return s.sessionRepo.DeleteExpired()
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash compares a password with a hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// generateSessionID generates a random session ID
func generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
