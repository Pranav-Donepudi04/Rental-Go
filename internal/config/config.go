package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	// Server Configuration
	Port     string
	LogLevel string

	// Database Configuration
	DatabaseURL string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	DBSSLMode   string

	// Production Features
	MaxConnections    int
	ConnectionTimeout int

	// Notification Configuration
	TelegramBotToken string
	OwnerChatID      string

	// Security Configuration
	Environment    string // "development" or "production"
	CookieSecure   bool   // Set Secure flag on cookies (true in production)
	CookieSameSite string // "Strict", "Lax", or "None"
	CookieName     string // Session cookie name

	// Payment Configuration
	DefaultPaymentMethod string // Default payment method (e.g., "UPI")
	DefaultUPIID         string // Default UPI ID for payments

	// Server Timeouts
	ReadTimeout  int // HTTP read timeout in seconds
	WriteTimeout int // HTTP write timeout in seconds
	IdleTimeout  int // HTTP idle timeout in seconds
}

func Load() *Config {
	return &Config{
		// Server settings
		Port:     getEnv("SERVER_PORT", "8080"),
		LogLevel: getEnv("LOG_LEVEL", "info"),

		// Database settings
		DatabaseURL: getEnv("DATABASE_URL", ""),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPassword:  getEnv("DB_PASSWORD", ""),
		DBName:      getEnv("DB_NAME", "formdb"),
		DBSSLMode:   getEnv("DB_SSL_MODE", "require"),

		// Production settings
		MaxConnections:    getEnvAsInt("DB_MAX_CONNECTIONS", 25),
		ConnectionTimeout: getEnvAsInt("DB_CONNECTION_TIMEOUT", 30),

		// Notification settings
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		OwnerChatID:      getEnv("TELEGRAM_OWNER_CHAT_ID", ""),

		// Security settings
		Environment:    getEnv("ENVIRONMENT", "development"),
		CookieSecure:   getEnv("ENVIRONMENT", "development") == "production",
		CookieSameSite: getEnv("COOKIE_SAME_SITE", "Strict"),
		CookieName:     getEnv("COOKIE_NAME", "sid"),

		// Payment settings
		DefaultPaymentMethod: getEnv("DEFAULT_PAYMENT_METHOD", "UPI"),
		DefaultUPIID:         getEnv("DEFAULT_UPI_ID", "9848790200@ybl"),

		// Server timeout settings
		ReadTimeout:  getEnvAsInt("READ_TIMEOUT", 15),
		WriteTimeout: getEnvAsInt("WRITE_TIMEOUT", 15),
		IdleTimeout:  getEnvAsInt("IDLE_TIMEOUT", 60),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// Validate checks that all required configuration values are set
// Returns an error if any required values are missing or invalid
func (c *Config) Validate() error {
	var errors []string

	// Database validation
	if c.DatabaseURL == "" {
		// If DATABASE_URL is not set, check individual components
		if c.DBHost == "" {
			errors = append(errors, "DATABASE_URL or DB_HOST is required")
		}
		if c.DBName == "" {
			errors = append(errors, "DATABASE_URL or DB_NAME is required")
		}
		if c.DBUser == "" {
			errors = append(errors, "DATABASE_URL or DB_USER is required")
		}
	}

	// Server validation
	if c.Port == "" {
		errors = append(errors, "SERVER_PORT cannot be empty")
	}
	if portNum, err := strconv.Atoi(c.Port); err != nil || portNum < 1 || portNum > 65535 {
		errors = append(errors, fmt.Sprintf("SERVER_PORT must be a valid port number (1-65535), got: %s", c.Port))
	}

	// Log level validation
	validLogLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true, "fatal": true,
	}
	if !validLogLevels[strings.ToLower(c.LogLevel)] {
		errors = append(errors, fmt.Sprintf("LOG_LEVEL must be one of: debug, info, warn, error, fatal, got: %s", c.LogLevel))
	}

	// Environment validation
	validEnvironments := map[string]bool{
		"development": true, "production": true, "staging": true, "test": true,
	}
	if !validEnvironments[strings.ToLower(c.Environment)] {
		errors = append(errors, fmt.Sprintf("ENVIRONMENT must be one of: development, production, staging, test, got: %s", c.Environment))
	}

	// Cookie SameSite validation
	validSameSite := map[string]bool{
		"Strict": true, "Lax": true, "None": true,
	}
	if !validSameSite[c.CookieSameSite] {
		errors = append(errors, fmt.Sprintf("COOKIE_SAME_SITE must be one of: Strict, Lax, None, got: %s", c.CookieSameSite))
	}

	// Timeout validation
	if c.ReadTimeout < 1 {
		errors = append(errors, "READ_TIMEOUT must be at least 1 second")
	}
	if c.WriteTimeout < 1 {
		errors = append(errors, "WRITE_TIMEOUT must be at least 1 second")
	}
	if c.IdleTimeout < 1 {
		errors = append(errors, "IDLE_TIMEOUT must be at least 1 second")
	}

	// Connection pool validation
	if c.MaxConnections < 1 {
		errors = append(errors, "DB_MAX_CONNECTIONS must be at least 1")
	}
	if c.ConnectionTimeout < 1 {
		errors = append(errors, "DB_CONNECTION_TIMEOUT must be at least 1 second")
	}

	// Payment validation
	if c.DefaultPaymentMethod == "" {
		errors = append(errors, "DEFAULT_PAYMENT_METHOD cannot be empty")
	}
	if c.DefaultUPIID == "" {
		errors = append(errors, "DEFAULT_UPI_ID cannot be empty")
	}

	// Cookie name validation
	if c.CookieName == "" {
		errors = append(errors, "COOKIE_NAME cannot be empty")
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}
