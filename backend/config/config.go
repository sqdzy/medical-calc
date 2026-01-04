package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	// Server
	Port        string
	Environment string
	LogLevel    string

	// Database
	DatabaseURL    string
	MigrationsPath string

	// JWT
	JWTSecret        string
	JWTAccessExpiry  time.Duration
	JWTRefreshExpiry time.Duration

	// Security
	EncryptionKey string // 32 bytes for AES-256

	// External APIs
	NCBIApiKey      string
	YandexGPTApiKey string
	YandexIAMToken  string
	YandexFolderID  string
	YandexGPTModel  string

	// CORS
	CORSOrigins string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		Environment:    getEnv("ENVIRONMENT", "development"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		MigrationsPath: getEnv("MIGRATIONS_PATH", "file://migrations"),
		CORSOrigins:    getEnv("CORS_ORIGINS", "*"),
	}

	// Database URL (required)
	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		// Build from components
		cfg.DatabaseURL = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=%s",
			getEnv("DB_USER", "postgres"),
			getEnv("DB_PASSWORD", "postgres"),
			getEnv("DB_HOST", "localhost"),
			getEnv("DB_PORT", "5432"),
			getEnv("DB_NAME", "medical_db"),
			getEnv("DB_SSLMODE", "disable"),
		)
	}

	// JWT configuration
	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	accessExpiry, err := strconv.Atoi(getEnv("JWT_ACCESS_EXPIRY_MINUTES", "15"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_EXPIRY_MINUTES: %w", err)
	}
	cfg.JWTAccessExpiry = time.Duration(accessExpiry) * time.Minute

	refreshExpiry, err := strconv.Atoi(getEnv("JWT_REFRESH_EXPIRY_DAYS", "7"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_EXPIRY_DAYS: %w", err)
	}
	cfg.JWTRefreshExpiry = time.Duration(refreshExpiry) * 24 * time.Hour

	// Encryption key (required for PII)
	cfg.EncryptionKey = os.Getenv("ENCRYPTION_KEY")
	if cfg.EncryptionKey == "" {
		return nil, fmt.Errorf("ENCRYPTION_KEY is required (32 bytes for AES-256)")
	}

	// External APIs (optional for MVP)
	cfg.NCBIApiKey = os.Getenv("NCBI_API_KEY")
	cfg.YandexGPTApiKey = os.Getenv("YANDEX_GPT_API_KEY")
	cfg.YandexIAMToken = os.Getenv("YANDEX_IAM_TOKEN")
	cfg.YandexFolderID = os.Getenv("YANDEX_FOLDER_ID")
	cfg.YandexGPTModel = getEnv("YANDEX_GPT_MODEL", "yandexgpt-lite")

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
