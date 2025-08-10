package utils

import (
	"fmt"
)

// Global embedding service instance
var embeddingService EmbeddingService

// InitEmbeddingService initializes the global embedding service using the factory
func InitEmbeddingService() error {
	// Create a new AI service factory using the global application configuration
	// This factory will be used to create instances of EmbeddingService and ChatService
	factory := NewAIServiceFactory(AppConfig)

	// Validate configuration
	// This function checks if the AI configuration is valid
	// It ensures that the required environment variables are set correctly
	if err := factory.ValidateConfiguration(); err != nil {
		return fmt.Errorf("invalid AI configuration: %v", err)
	}

	// Create the embedding service
	// This will use the factory to create an instance of EmbeddingService based on the current configuration
	// This allows the application to use the appropriate AI provider for generating embeddings
	service, err := factory.CreateEmbeddingService()
	if err != nil {
		return fmt.Errorf("failed to create embedding service: %v", err)
	}

	embeddingService = service
	LogInfo("Embedding service initialized", "provider", service.GetProviderName())
	return nil
}

// GetEmbedding generates embeddings using the configured AI service
// This function takes a text input and returns its embedding as a Vector
// It uses the global embedding service instance initialized in InitEmbeddingService
func GetEmbedding(text string) (Vector, error) {
	if embeddingService == nil {
		return nil, fmt.Errorf("embedding service not initialized")
	}

	return embeddingService.GenerateEmbedding(text)
}

// GetBatchEmbeddings generates embeddings for multiple texts
// This is useful for processing multiple inputs in a single API call
// It returns a slice of Vector, one for each input text
// It uses the global embedding service instance initialized in InitEmbeddingService
func GetBatchEmbeddings(texts []string) ([]Vector, error) {
	if embeddingService == nil {
		return nil, fmt.Errorf("embedding service not initialized")
	}

	return embeddingService.GenerateBatchEmbeddings(texts)
}
