package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/securedlinq/backend/internal/service"
)

// Role represents user roles in the system
type Role string

const (
	RoleAdmin  Role = "admin"
	RoleDriver Role = "driver"
)

// HasRole checks if the current user has the specified role
func HasRole(c *gin.Context, role Role) bool {
	user, exists := c.Get(ContextUserKey)
	if !exists {
		return false
	}

	sessionInfo, ok := user.(*service.SessionInfo)
	if !ok {
		return false
	}

	return sessionInfo.UserType == string(role)
}

// RequireRole creates middleware that requires a specific role
func RequireRole(role Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get(ContextUserKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		sessionInfo, ok := user.(*service.SessionInfo)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
			c.Abort()
			return
		}

		if sessionInfo.UserType != string(role) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Insufficient permissions",
				"message": string(role) + " access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole creates middleware that requires any of the specified roles
func RequireAnyRole(roles ...Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get(ContextUserKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		sessionInfo, ok := user.(*service.SessionInfo)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
			c.Abort()
			return
		}

		for _, role := range roles {
			if sessionInfo.UserType == string(role) {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Insufficient permissions",
			"message": "Access denied for this resource",
		})
		c.Abort()
	}
}

// GetUserRole returns the current user's role
func GetUserRole(c *gin.Context) (Role, bool) {
	user, exists := c.Get(ContextUserKey)
	if !exists {
		return "", false
	}

	sessionInfo, ok := user.(*service.SessionInfo)
	if !ok {
		return "", false
	}

	return Role(sessionInfo.UserType), true
}

// IsAdmin checks if the current user is an admin
func IsAdmin(c *gin.Context) bool {
	return HasRole(c, RoleAdmin)
}

// IsDriver checks if the current user is a driver
func IsDriver(c *gin.Context) bool {
	return HasRole(c, RoleDriver)
}
