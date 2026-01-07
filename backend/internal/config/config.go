package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Session  SessionConfig
	Admin    AdminConfig
	Agora    AgoraConfig
	AWS      AWSConfig
	Email    EmailConfig
}

type EmailConfig struct {
	SMTPHost    string
	SMTPPort    string
	SenderEmail string
	SenderName  string
	AppPassword string
}

type ServerConfig struct {
	Port        string
	GinMode     string
	FrontendURL string
	BaseURL     string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type SessionConfig struct {
	Secret   string
	MaxAge   int
	Secure   bool
	SameSite string
}

type AdminConfig struct {
	Username string
	Password string
}

type AgoraConfig struct {
	AppID          string
	AppCertificate string
	EncodedKey     string
}

type AWSConfig struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	S3BucketName    string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	maxAge, _ := strconv.Atoi(getEnv("SESSION_MAX_AGE", "86400"))
	secure, _ := strconv.ParseBool(getEnv("SESSION_SECURE", "false"))

	config := &Config{
		Server: ServerConfig{
			Port:        getEnv("PORT", "8080"),
			GinMode:     getEnv("GIN_MODE", "debug"),
			FrontendURL: getEnv("FRONTEND_URL", "http://localhost:5173"),
			BaseURL:     getEnv("BASE_URL", "http://localhost:8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "uatsecuredlinq_db"),
		},
		Session: SessionConfig{
			MaxAge:   maxAge,
			Secure:   secure,
			SameSite: getEnv("SESSION_SAME_SITE", "lax"),
		},
		Admin: AdminConfig{
			Username: getEnv("ADMIN_USERNAME", "admin"),
			Password: getEnv("ADMIN_PASSWORD", "secure123"),
		},
		Agora: AgoraConfig{
			AppID:          getEnv("AGORA_APP_ID", ""),
			AppCertificate: getEnv("AGORA_APP_CERTIFICATE", ""),
			EncodedKey:     getEnv("AGORA_ENCODED_KEY", ""),
		},
		AWS: AWSConfig{
			AccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
			Region:          getEnv("AWS_REGION", "us-east-1"),
			S3BucketName:    getEnv("AWS_S3_BUCKET_NAME", ""),
		},
		Email: EmailConfig{
			SMTPHost:    getEnv("SMTP_HOST", "smtp.gmail.com"),
			SMTPPort:    getEnv("SMTP_PORT", "587"),
			SenderEmail: getEnv("SENDER_EMAIL", ""),
			SenderName:  getEnv("SENDER_NAME", "SecuredLinQ"),
			AppPassword: getEnv("EMAIL_APP_PASSWORD", ""),
		},
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
