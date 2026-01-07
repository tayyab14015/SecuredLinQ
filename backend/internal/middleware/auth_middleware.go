package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/securedlinq/backend/internal/service"
)

const (
	SessionCookieName = "session_id"
	ContextUserKey    = "user"
)

// AuthMiddleware creates authentication middleware
func AuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get session cookie
		sessionID, err := c.Cookie(SessionCookieName)
		if err != nil || sessionID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// Validate session
		sessionInfo, err := authService.ValidateSession(sessionID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired session"})
			c.Abort()
			return
		}

		// Store user info in context
		c.Set(ContextUserKey, sessionInfo)

		// Optionally refresh session
		_ = authService.RefreshSession(sessionID)

		c.Next()
	}
}

// AdminOnlyMiddleware ensures only admin users can access the route
// This is a convenience wrapper around RequireRole(RoleAdmin)
func AdminOnlyMiddleware() gin.HandlerFunc {
	return RequireRole(RoleAdmin)
}

// DriverOnlyMiddleware ensures only driver users can access the route
// This is a convenience wrapper around RequireRole(RoleDriver)
func DriverOnlyMiddleware() gin.HandlerFunc {
	return RequireRole(RoleDriver)
}

// OptionalAuthMiddleware validates session if present but doesn't require it
func OptionalAuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie(SessionCookieName)
		if err != nil || sessionID == "" {
			c.Next()
			return
		}

		sessionInfo, err := authService.ValidateSession(sessionID)
		if err == nil {
			c.Set(ContextUserKey, sessionInfo)
		}

		c.Next()
	}
}

// GetCurrentUser extracts the current user from context
func GetCurrentUser(c *gin.Context) (*service.SessionInfo, bool) {
	user, exists := c.Get(ContextUserKey)
	if !exists {
		return nil, false
	}
	sessionInfo, ok := user.(*service.SessionInfo)
	return sessionInfo, ok
}
