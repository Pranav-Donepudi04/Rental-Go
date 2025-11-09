package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Logger is the global logger instance
	Logger *zap.Logger
	// SugaredLogger is the sugared logger for convenience
	SugaredLogger *zap.SugaredLogger
)

// InitLogger initializes the global logger based on environment
// logLevel: "debug", "info", "warn", "error" (defaults to "info")
// environment: "development" or "production" (defaults to "development")
func InitLogger(logLevel, environment string) error {
	var config zap.Config

	if environment == "production" {
		// Production: JSON output, no stack traces for info/warn
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		// Development: Console output, colored, with stack traces
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Set log level
	level, err := zapcore.ParseLevel(logLevel)
	if err != nil {
		level = zapcore.InfoLevel
	}
	config.Level = zap.NewAtomicLevelAt(level)

	// Build logger
	Logger, err = config.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel), // Stack traces for errors and above
	)
	if err != nil {
		return err
	}

	SugaredLogger = Logger.Sugar()

	// Replace global logger
	zap.ReplaceGlobals(Logger)

	return nil
}

// Sync flushes any buffered log entries
func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}

// WithRequestID creates a logger with request ID field
func WithRequestID(requestID string) *zap.Logger {
	return Logger.With(zap.String("request_id", requestID))
}

// WithUser creates a logger with user context
func WithUser(userID int, userType string) *zap.Logger {
	return Logger.With(
		zap.Int("user_id", userID),
		zap.String("user_type", userType),
	)
}

// WithError creates a logger with error field
func WithError(err error) *zap.Logger {
	return Logger.With(zap.Error(err))
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

// GetLogger returns the global logger instance
func GetLogger() *zap.Logger {
	return Logger
}

// GetSugaredLogger returns the global sugared logger instance
func GetSugaredLogger() *zap.SugaredLogger {
	return SugaredLogger
}

// NewLogger creates a new logger instance (for testing or special cases)
func NewLogger(logLevel, environment string) (*zap.Logger, error) {
	var config zap.Config

	if environment == "production" {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	level, err := zapcore.ParseLevel(logLevel)
	if err != nil {
		level = zapcore.InfoLevel
	}
	config.Level = zap.NewAtomicLevelAt(level)

	return config.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
}
