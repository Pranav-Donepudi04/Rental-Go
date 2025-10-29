package config

import (
	"os"
	"strconv"
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
