package service

import (
	"testing"

	"github.com/securedlinq/backend/internal/config"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestValidateAdminCredentials(t *testing.T) {
	tests := []struct {
		name           string
		username       string
		password       string
		configUsername string
		configPassword string
		expectedError  error
	}{
		{
			name:           "Valid credentials",
			username:       "admin",
			password:       "password123",
			configUsername: "admin",
			configPassword: "password123",
			expectedError:  nil,
		},
		{
			name:           "Invalid username",
			username:       "wronguser",
			password:       "password123",
			configUsername: "admin",
			configPassword: "password123",
			expectedError:  ErrUserNotFound,
		},
		{
			name:           "Invalid password",
			username:       "admin",
			password:       "wrongpassword",
			configUsername: "admin",
			configPassword: "password123",
			expectedError:  ErrInvalidPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Admin: config.AdminConfig{
					Username: tt.configUsername,
					Password: tt.configPassword,
				},
			}

			service := &AuthService{
				config: cfg,
			}

			err := service.ValidateAdminCredentials(tt.username, tt.password)

			if tt.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			}
		})
	}
}

func TestHashPassword(t *testing.T) {
	password := "testpassword123"
	hash, err := HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)

	// Verify the hash can be used to check the password
	isValid := CheckPasswordHash(password, hash)
	assert.True(t, isValid)

	// Verify wrong password fails
	isValid = CheckPasswordHash("wrongpassword", hash)
	assert.False(t, isValid)
}

func TestCheckPasswordHash(t *testing.T) {
	password := "testpassword123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	tests := []struct {
		name     string
		password string
		hash     string
		expected bool
	}{
		{
			name:     "Correct password",
			password: password,
			hash:     string(hash),
			expected: true,
		},
		{
			name:     "Wrong password",
			password: "wrongpassword",
			hash:     string(hash),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckPasswordHash(tt.password, tt.hash)
			assert.Equal(t, tt.expected, result)
		})
	}
}
