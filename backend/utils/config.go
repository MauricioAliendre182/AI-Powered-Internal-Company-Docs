package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds application configuration
type Config struct {
	OllamaBaseURL       string
	Port                string
	DBUser              string
	DBPassword          string
	DBName              string
	ChatModel           string
	OpenAIAPIKey        string
	GoogleAIAPIKey      string
	DBPort              string
	EmbeddingModel      string
	DBHost              string
	Environment         string
	MaxFileSize         int64
	ChunkSize           int64
	RateLimitMaxTokens  int64
	RateLimitRefillRate int64
	UseLocalAI          bool
}

// LoadConfig loads configuration from environment variables with fallbacks
func LoadConfig() (*Config, error) {
	config := &Config{
		// Database defaults
		DBHost:     getEnvWithDefault("DB_HOST", "localhost"),
		DBPort:     getEnvWithDefault("DB_PORT", "5432"),
		DBUser:     getEnvWithDefault("DB_USER", "postgres"),
		DBPassword: os.Getenv("DB_PASSWORD"), // No default for security
		DBName:     getEnvWithDefault("DB_NAME", "internal_docs"),

		// AI Configuration
		UseLocalAI:     getBoolEnvWithDefault("USE_LOCAL_AI", false),
		OpenAIAPIKey:   os.Getenv("OPENAI_API_KEY"),
		GoogleAIAPIKey: os.Getenv("GOOGLE_AI_API_KEY"),
		OllamaBaseURL:  getEnvWithDefault("OLLAMA_BASE_URL", "http://localhost:11434"),
		EmbeddingModel: getEnvWithDefault("EMBEDDING_MODEL", "text-embedding-3-small"),
		ChatModel:      getEnvWithDefault("CHAT_MODEL", "gpt-3.5-turbo"),

		// Application defaults
		Environment: getEnvWithDefault("ENVIRONMENT", "development"),
		Port:        getEnvWithDefault("PORT", "8090"),

		// File upload defaults
		MaxFileSize: getEnvIntWithDefault("MAX_FILE_SIZE", 10*1024*1024), // 10MB
		ChunkSize:   getEnvIntWithDefault("CHUNK_SIZE", 1000),

		// Rate limiting defaults
		RateLimitMaxTokens:  getEnvIntWithDefault("RATE_LIMIT_MAX_TOKENS", 10),
		RateLimitRefillRate: getEnvIntWithDefault("RATE_LIMIT_REFILL_RATE", 1),
	}

	// Validate configuration
	if config.DBPassword == "" {
		return nil, fmt.Errorf("DB_PASSWORD environment variable is required")
	}

	// Validate AI configuration
	if config.UseLocalAI {
		if config.OllamaBaseURL == "" {
			return nil, fmt.Errorf("OLLAMA_BASE_URL is required when USE_LOCAL_AI=true")
		}
	} else if config.OpenAIAPIKey == "" && config.GoogleAIAPIKey == "" {
		return nil, fmt.Errorf("either OPENAI_API_KEY or GOOGLE_AI_API_KEY is required")
	}

	return config, nil
}

// getEnvWithDefault returns environment variable value or default
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntWithDefault returns environment variable as int or default
func getEnvIntWithDefault(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// getBoolEnvWithDefault returns environment variable as bool or default
func getBoolEnvWithDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// Global config instance
// This variable is used throughout the application to access configuration settings
// It is initialized in the InitConfig function
var AppConfig *Config

// InitConfig initializes the global configuration
func InitConfig() error {
	var err error
	AppConfig, err = LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	LogInfo("Configuration loaded successfully",
		"environment", AppConfig.Environment,
		"port", AppConfig.Port,
		"embedding_model", AppConfig.EmbeddingModel,
		"chat_model", AppConfig.ChatModel)

	return nil
}

// ValidateConfig is a helper function that should be implemented in config.go
func ValidateConfig(config *Config) error {
	if config.DBHost == "" {
		return fmt.Errorf("database host is required")
	}

	if config.DBPort != "" {
		if _, err := strconv.Atoi(config.DBPort); err != nil {
			return fmt.Errorf("invalid database port: %v", err)
		}
	}

	// Validate OpenAI API key format if provided
	if config.OpenAIAPIKey != "" {
		if len(config.OpenAIAPIKey) < 10 || !strings.HasPrefix(config.OpenAIAPIKey, "sk-") {
			return fmt.Errorf("invalid OpenAI API key format")
		}
	}

	return nil
}
