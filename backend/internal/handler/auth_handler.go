package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/securedlinq/backend/internal/config"
	"github.com/securedlinq/backend/internal/middleware"
	"github.com/securedlinq/backend/internal/service"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService   *service.AuthService
	driverService *service.DriverService
	config        *config.Config
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService, driverService *service.DriverService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authService:   authService,
		driverService: driverService,
		config:        cfg,
	}
}

// LoginRequest represents login request body
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// DriverRegisterRequest represents driver registration request
type DriverRegisterRequest struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required,min=6"`
	PhoneNumber string `json:"phone_number" binding:"required"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
}

// Login handles admin login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
		return
	}

	// Validate credentials
	err := h.authService.ValidateAdminCredentials(req.Username, req.Password)
	if err != nil {
		// Return specific error messages
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin account not found. Please check your username."})
		case errors.Is(err, service.ErrInvalidPassword):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect password. Please try again."})
		default:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		}
		return
	}

	// Create session (userID = 0 for config-based admin)
	sessionInfo, err := h.authService.CreateSession(0, "admin")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	// Set session cookie
	c.SetSameSite(h.getSameSite())
	c.SetCookie(
		middleware.SessionCookieName,
		sessionInfo.SessionID,
		h.config.Session.MaxAge,
		"/",
		"",
		h.config.Session.Secure,
		true, // HttpOnly
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login successful",
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	sessionID, err := c.Cookie(middleware.SessionCookieName)
	if err == nil && sessionID != "" {
		_ = h.authService.InvalidateSession(sessionID)
	}

	// Clear cookie
	c.SetSameSite(h.getSameSite())
	c.SetCookie(
		middleware.SessionCookieName,
		"",
		-1,
		"/",
		"",
		h.config.Session.Secure,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged out successfully",
	})
}

// ValidateSession validates the current session
func (h *AuthHandler) ValidateSession(c *gin.Context) {
	sessionID, err := c.Cookie(middleware.SessionCookieName)
	if err != nil || sessionID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"valid": false, "error": "No session cookie found"})
		return
	}

	sessionInfo, err := h.authService.ValidateSession(sessionID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"valid": false, "error": "Session invalid or expired"})
		return
	}

	response := gin.H{
		"valid":     true,
		"user_type": sessionInfo.UserType,
		"user_id":   sessionInfo.UserID,
		"expires":   time.Unix(sessionInfo.ExpiresAt, 0).Format(time.RFC3339),
	}

	// Include user info based on type
	if sessionInfo.UserType == "driver" {
		driver, err := h.authService.GetDriverByID(sessionInfo.UserID)
		if err == nil {
			response["driver"] = gin.H{
				"id":           driver.ID,
				"username":     driver.Username,
				"phone_number": driver.PhoneNumber,
				"first_name":   driver.FirstName,
				"last_name":    driver.LastName,
			}
		}
	}
	// Admin users are config-based, no additional info to return

	c.JSON(http.StatusOK, response)
}

// DriverRegister handles driver registration
func (h *AuthHandler) DriverRegister(c *gin.Context) {
	var req DriverRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data. Username, password (min 6 chars), and phone number are required."})
		return
	}

	// Validate password length
	if len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 6 characters"})
		return
	}

	driver, err := h.driverService.RegisterDriver(&service.RegisterDriverRequest{
		Username:    req.Username,
		Password:    req.Password,
		PhoneNumber: req.PhoneNumber,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Email:       req.Email,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Registration successful",
		"driver": gin.H{
			"id":           driver.ID,
			"username":     driver.Username,
			"phone_number": driver.PhoneNumber,
			"first_name":   driver.FirstName,
			"last_name":    driver.LastName,
		},
	})
}

// DriverLogin handles driver login
func (h *AuthHandler) DriverLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
		return
	}

	// Validate driver credentials
	driver, err := h.authService.ValidateDriverCredentials(req.Username, req.Password)
	if err != nil {
		// Return specific error messages
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Driver account not found. Please register first or check your username."})
		case errors.Is(err, service.ErrInvalidPassword):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect password. Please try again."})
		case errors.Is(err, service.ErrAccountDeactivated):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Your account has been deactivated. Please contact your administrator."})
		default:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		}
		return
	}

	// Create session for driver
	sessionInfo, err := h.authService.CreateSession(int(driver.ID), "driver")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	// Set session cookie
	c.SetSameSite(h.getSameSite())
	c.SetCookie(
		middleware.SessionCookieName,
		sessionInfo.SessionID,
		h.config.Session.MaxAge,
		"/",
		"",
		h.config.Session.Secure,
		true, // HttpOnly
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login successful",
		"driver": gin.H{
			"id":           driver.ID,
			"username":     driver.Username,
			"phone_number": driver.PhoneNumber,
			"first_name":   driver.FirstName,
			"last_name":    driver.LastName,
		},
	})
}

func (h *AuthHandler) getSameSite() http.SameSite {
	switch h.config.Session.SameSite {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}
