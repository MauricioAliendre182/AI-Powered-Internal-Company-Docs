package utils

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

// InitLogger initializes the structured logger
func InitLogger() {
	// Create a JSON handler for structured logging
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	// Use JSON format in production, text format in development
	env := os.Getenv("ENVIRONMENT")
	if env == "production" {
		// slog.NewJSONHandler is used for structured logging in production
		// This allows for better log parsing and analysis
		handler := slog.NewJSONHandler(os.Stdout, opts)
		Logger = slog.New(handler)
	} else {
		// slog.NewTextHandler is used for human-readable logs in development
		// This is useful for debugging and local development
		handler := slog.NewTextHandler(os.Stdout, opts)
		Logger = slog.New(handler)
	}
}

// LogError logs an error with context
func LogError(msg string, err error, fields ...any) {
	if Logger != nil {
		args := append([]any{"error", err}, fields...)
		Logger.Error(msg, args...)
	}
}

// LogInfo logs an info message with context
func LogInfo(msg string, fields ...any) {
	if Logger != nil {
		Logger.Info(msg, fields...)
	}
}

// LogWarn logs a warning message with context
func LogWarn(msg string, fields ...any) {
	if Logger != nil {
		Logger.Warn(msg, fields...)
	}
}

// LogDebug logs a debug message with context
func LogDebug(msg string, fields ...any) {
	if Logger != nil {
		Logger.Debug(msg, fields...)
	}
}
