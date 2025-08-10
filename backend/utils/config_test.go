package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		checkFunc   func(*testing.T, *Config)
	}{
		{
			name: "Valid OpenAI configuration",
			envVars: map[string]string{
				"DB_HOST":         "localhost",
				"DB_PORT":         "5432",
				"DB_USER":         "postgres",
				"DB_PASSWORD":     "test_password",
				"DB_NAME":         "test_db",
				"OPENAI_API_KEY":  "sk-test-key-here",
				"EMBEDDING_MODEL": "text-embedding-3-small",
				"CHAT_MODEL":      "gpt-3.5-turbo",
				"PORT":            "8090",
				"ENVIRONMENT":     "test",
				"JWT_SECRET":      "test-jwt-secret",
			},
			expectError: false,
			checkFunc: func(t *testing.T, config *Config) {
				assert.Equal(t, "localhost", config.DBHost)
				assert.Equal(t, "5432", config.DBPort)
				assert.Equal(t, "sk-test-key-here", config.OpenAIAPIKey)
				assert.Equal(t, "text-embedding-3-small", config.EmbeddingModel)
				assert.Equal(t, "gpt-3.5-turbo", config.ChatModel)
				assert.Equal(t, "8090", config.Port)
				assert.Equal(t, "test", config.Environment)
			},
		},
		{
			name: "Valid Google AI configuration",
			envVars: map[string]string{
				"DB_HOST":           "localhost",
				"DB_PORT":           "5432",
				"DB_USER":           "postgres",
				"DB_PASSWORD":       "test_password",
				"DB_NAME":           "test_db",
				"GOOGLE_AI_API_KEY": "AIza-test-key-here",
				"EMBEDDING_MODEL":   "models/embedding-001",
				"CHAT_MODEL":        "models/gemini-1.5-flash",
				"PORT":              "8090",
				"ENVIRONMENT":       "test",
				"JWT_SECRET":        "test-jwt-secret",
			},
			expectError: false,
			checkFunc: func(t *testing.T, config *Config) {
				assert.Equal(t, "AIza-test-key-here", config.GoogleAIAPIKey)
				assert.Equal(t, "models/embedding-001", config.EmbeddingModel)
				assert.Equal(t, "models/gemini-1.5-flash", config.ChatModel)
			},
		},
		{
			name: "Valid Ollama configuration",
			envVars: map[string]string{
				"DB_HOST":         "localhost",
				"DB_PORT":         "5432",
				"DB_USER":         "postgres",
				"DB_PASSWORD":     "test_password",
				"DB_NAME":         "test_db",
				"USE_LOCAL_AI":    "true",
				"OLLAMA_BASE_URL": "http://localhost:11434",
				"EMBEDDING_MODEL": "nomic-embed-text",
				"CHAT_MODEL":      "llama3.1:8b",
				"PORT":            "8090",
				"ENVIRONMENT":     "test",
				"JWT_SECRET":      "test-jwt-secret",
			},
			expectError: false,
			checkFunc: func(t *testing.T, config *Config) {
				assert.True(t, config.UseLocalAI)
				assert.Equal(t, "http://localhost:11434", config.OllamaBaseURL)
				assert.Equal(t, "nomic-embed-text", config.EmbeddingModel)
				assert.Equal(t, "llama3.1:8b", config.ChatModel)
			},
		},
		{
			name: "Missing database configuration",
			envVars: map[string]string{
				"OPENAI_API_KEY": "sk-test-key",
				"PORT":           "8090",
			},
			expectError: true,
			checkFunc:   nil,
		},
		{
			name: "Missing AI provider configuration",
			envVars: map[string]string{
				"DB_HOST":     "localhost",
				"DB_PORT":     "5432",
				"DB_USER":     "postgres",
				"DB_PASSWORD": "test_password",
				"DB_NAME":     "test_db",
				"PORT":        "8090",
				"ENVIRONMENT": "test",
				"JWT_SECRET":  "test-jwt-secret",
			},
			expectError: true, // Should fail because no AI provider is configured
			checkFunc:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment
			originalEnv := make(map[string]string)
			for key := range tt.envVars {
				originalEnv[key] = os.Getenv(key)
			}

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Test configuration loading
			config, err := LoadConfig()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
				if tt.checkFunc != nil {
					tt.checkFunc(t, config)
				}
			}

			// Restore original environment
			for key, originalValue := range originalEnv {
				if originalValue == "" {
					os.Unsetenv(key)
				} else {
					os.Setenv(key, originalValue)
				}
			}
		})
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid configuration",
			config: &Config{
				DBHost:     "localhost",
				DBPort:     "5432",
				DBUser:     "postgres",
				DBPassword: "password",
				DBName:     "testdb",
				Port:       "8090",
			},
			expectError: false,
		},
		{
			name: "Missing database host",
			config: &Config{
				DBPort:     "5432",
				DBUser:     "postgres",
				DBPassword: "password",
				DBName:     "testdb",
			},
			expectError: true,
			errorMsg:    "database host is required",
		},
		{
			name: "Invalid port format",
			config: &Config{
				DBHost:     "localhost",
				DBPort:     "invalid",
				DBUser:     "postgres",
				DBPassword: "password",
				DBName:     "testdb",
			},
			expectError: true,
			errorMsg:    "invalid database port",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
