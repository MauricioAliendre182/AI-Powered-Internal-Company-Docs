package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/MauricioAliendre182/backend/utils"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	// Setup test environment with minimal required values
	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("DB_NAME", "test_internal_docs")
	os.Setenv("LOG_LEVEL", "error")
	os.Setenv("JWT_SECRET", "test-secret-for-testing")

	// Set a default AI provider for tests that need it
	os.Setenv("OPENAI_API_KEY", "sk-test-key-for-testing")

	// Run tests
	code := m.Run()

	// Cleanup
	os.Unsetenv("ENVIRONMENT")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("OPENAI_API_KEY")

	// Exit with the same code
	os.Exit(code)
}

func TestApplicationStartup(t *testing.T) {
	tests := []struct {
		envVars     map[string]string
		name        string
		expectPanic bool
	}{
		{
			name: "Valid configuration",
			envVars: map[string]string{
				"DB_HOST":        "localhost",
				"DB_PORT":        "5432",
				"DB_USER":        "postgres",
				"DB_PASSWORD":    "test_password",
				"DB_NAME":        "test_internal_docs",
				"OPENAI_API_KEY": "sk-test-key",
				"PORT":           "8090",
				"ENVIRONMENT":    "test",
				"JWT_SECRET":     "test-secret",
			},
			expectPanic: false,
		},
		{
			name: "Missing database configuration",
			envVars: map[string]string{
				"OPENAI_API_KEY": "sk-test-key",
				"PORT":           "8090",
				// DB_PASSWORD is missing, which should cause an error
			},
			expectPanic: true, // LoadConfig will return error, which we convert to panic
		},
		{
			name: "Missing AI configuration",
			envVars: map[string]string{
				"DB_HOST":     "localhost",
				"DB_PORT":     "5432",
				"DB_USER":     "postgres",
				"DB_PASSWORD": "test_password",
				"DB_NAME":     "test_internal_docs",
				"PORT":        "8090",
				"ENVIRONMENT": "test",
				"JWT_SECRET":  "test-secret",
			},
			expectPanic: false, // LoadConfig returns error, doesn't panic
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment
			originalEnv := make(map[string]string)

			// Clear and set test environment
			for key := range tt.envVars {
				originalEnv[key] = os.Getenv(key)
				os.Unsetenv(key)
			}

			// Also clear database environment variables for clean test
			dbKeys := []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME"}
			originalDBEnv := make(map[string]string)
			for _, key := range dbKeys {
				if _, exists := tt.envVars[key]; !exists { // Only clear if not in test config
					originalDBEnv[key] = os.Getenv(key)
					os.Unsetenv(key)
				}
			}

			// Also clear AI-related environment variables for clean test
			aiKeys := []string{"OPENAI_API_KEY", "GOOGLE_AI_API_KEY", "USE_LOCAL_AI", "OLLAMA_BASE_URL"}
			originalAIEnv := make(map[string]string)
			for _, key := range aiKeys {
				if _, exists := tt.envVars[key]; !exists { // Only clear if not in test config
					originalAIEnv[key] = os.Getenv(key)
					os.Unsetenv(key)
				}
			}

			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Test configuration loading
			if tt.expectPanic {
				assert.Panics(t, func() {
					// This would normally call initializeApplication()
					// For now, just test config loading
					config, err := utils.LoadConfig()
					if err != nil {
						panic(err)
					}
					_ = config
				})
			} else {
				config, err := utils.LoadConfig()
				if tt.name == "Missing AI configuration" {
					// This specific test should return an error
					assert.Error(t, err)
					assert.Nil(t, config)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, config)
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

			// Restore DB environment
			for key, originalValue := range originalDBEnv {
				if originalValue == "" {
					os.Unsetenv(key)
				} else {
					os.Setenv(key, originalValue)
				}
			}

			// Restore AI environment
			for key, originalValue := range originalAIEnv {
				if originalValue == "" {
					os.Unsetenv(key)
				} else {
					os.Setenv(key, originalValue)
				}
			}
		})
	}
}

func TestConfigurationValidation(t *testing.T) {
	t.Run("Database connection validation", func(t *testing.T) {
		// Test database configuration validation
		validConfigs := []map[string]string{
			{
				"DB_HOST":     "localhost",
				"DB_PORT":     "5432",
				"DB_USER":     "postgres",
				"DB_PASSWORD": "password",
				"DB_NAME":     "testdb",
			},
			{
				"DB_HOST":     "127.0.0.1",
				"DB_PORT":     "3306",
				"DB_USER":     "root",
				"DB_PASSWORD": "rootpass",
				"DB_NAME":     "mysql_test",
			},
		}

		for i, config := range validConfigs {
			t.Run(fmt.Sprintf("Valid config %d", i+1), func(t *testing.T) {
				// Clear environment first
				aiKeys := []string{"OPENAI_API_KEY", "GOOGLE_AI_API_KEY", "USE_LOCAL_AI"}
				for _, key := range aiKeys {
					os.Unsetenv(key)
				}

				// Set minimal AI config to satisfy validation
				os.Setenv("OPENAI_API_KEY", "sk-test-key")
				defer os.Unsetenv("OPENAI_API_KEY")

				for key, value := range config {
					os.Setenv(key, value)
					defer os.Unsetenv(key)
				}

				loadedConfig, err := utils.LoadConfig()
				assert.NoError(t, err)
				if loadedConfig != nil {
					assert.Equal(t, config["DB_HOST"], loadedConfig.DBHost)
					assert.Equal(t, config["DB_PORT"], loadedConfig.DBPort)
					assert.Equal(t, config["DB_USER"], loadedConfig.DBUser)
				}
			})
		}
	})

	t.Run("AI provider configuration validation", func(t *testing.T) {
		aiConfigs := []struct {
			config map[string]string
			name   string
			valid  bool
		}{
			{
				name: "OpenAI configuration",
				config: map[string]string{
					"OPENAI_API_KEY":  "sk-test-key",
					"EMBEDDING_MODEL": "text-embedding-3-small",
					"CHAT_MODEL":      "gpt-3.5-turbo",
				},
				valid: true,
			},
			{
				name: "Google AI configuration",
				config: map[string]string{
					"GOOGLE_AI_API_KEY": "AIza-test-key",
					"EMBEDDING_MODEL":   "models/embedding-001",
					"CHAT_MODEL":        "models/gemini-1.5-flash",
				},
				valid: true,
			},
			{
				name: "Ollama configuration",
				config: map[string]string{
					"USE_LOCAL_AI":    "true",
					"OLLAMA_BASE_URL": "http://localhost:11434",
					"EMBEDDING_MODEL": "nomic-embed-text",
					"CHAT_MODEL":      "llama3.1:8b",
				},
				valid: true,
			},
			{
				name: "Invalid OpenAI key format",
				config: map[string]string{
					"OPENAI_API_KEY":  "invalid-key-format",
					"EMBEDDING_MODEL": "text-embedding-3-small",
					"CHAT_MODEL":      "gpt-3.5-turbo",
				},
				valid: false,
			},
		}

		for _, tc := range aiConfigs {
			t.Run(tc.name, func(t *testing.T) {
				// Clear existing AI config
				aiKeys := []string{"OPENAI_API_KEY", "GOOGLE_AI_API_KEY", "USE_LOCAL_AI", "OLLAMA_BASE_URL"}
				for _, key := range aiKeys {
					os.Unsetenv(key)
				}

				// Set required database config for all tests
				os.Setenv("DB_PASSWORD", "test_password")
				os.Setenv("JWT_SECRET", "test-secret")
				defer os.Unsetenv("DB_PASSWORD")
				defer os.Unsetenv("JWT_SECRET")

				// Set test config
				for key, value := range tc.config {
					os.Setenv(key, value)
					defer os.Unsetenv(key)
				}

				config, err := utils.LoadConfig()
				if tc.valid {
					assert.NoError(t, err)
					assert.NotNil(t, config)
				} else {
					// For invalid configs, we expect an error during loading
					// or during validation if loading succeeds
					if err == nil && config != nil {
						// If loading succeeded, try validation
						validationErr := utils.ValidateConfig(config)
						assert.Error(t, validationErr)
					} else {
						// Loading failed as expected for invalid config
						assert.Error(t, err)
					}
				}
			})
		}
	})
}

func TestEnvironmentSpecificBehavior(t *testing.T) {
	environments := []string{"development", "test", "production"}

	for _, env := range environments {
		t.Run(fmt.Sprintf("Environment: %s", env), func(t *testing.T) {
			// Set up required environment variables for config loading
			os.Setenv("ENVIRONMENT", env)
			os.Setenv("DB_PASSWORD", "test_password")
			os.Setenv("JWT_SECRET", "test-secret")
			os.Setenv("OPENAI_API_KEY", "sk-test-key")

			defer func() {
				os.Unsetenv("ENVIRONMENT")
				os.Unsetenv("DB_PASSWORD")
				os.Unsetenv("JWT_SECRET")
				os.Unsetenv("OPENAI_API_KEY")
			}()

			config, err := utils.LoadConfig()
			assert.NoError(t, err)
			assert.NotNil(t, config)

			if config != nil {
				assert.Equal(t, env, config.Environment)

				// Test environment-specific behavior
				switch env {
				case "development":
					// Development might have different defaults
					assert.NotEmpty(t, config.Environment)
				case "test":
					// Test environment should have minimal logging
					assert.Equal(t, "test", config.Environment)
				case "production":
					// Production should have strict validation
					assert.Equal(t, "production", config.Environment)
				}
			}
		})
	}
}
