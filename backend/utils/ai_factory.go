package utils

import (
	"fmt"
)

// AIProvider represents the type of AI provider
type AIProvider string

const (
	OpenAIProvider AIProvider = "openai"
	GeminiProvider AIProvider = "gemini"
	OllamaProvider AIProvider = "ollama"
)

// EmbeddingService interface for embedding generation
// This interface defines methods for generating embeddings
// and getting the provider name
// It allows different AI services to implement their own embedding generation logic
type EmbeddingService interface {
	GenerateEmbedding(text string) (Vector, error)
	GenerateBatchEmbeddings(texts []string) ([]Vector, error)
	GetProviderName() string
}

// ChatService interface for chat completion
// This interface defines methods for generating chat responses
// It allows different AI services to implement their own chat response generation logic
// It also provides methods to get the provider name and model used
type ChatService interface {
	GenerateResponse(question, context string) (string, error)
	GetProviderName() string
	GetModel() string
}

// AIServiceFactory manages different AI providers
// It uses the factory pattern to create instances of EmbeddingService and ChatService
// based on the configuration
type AIServiceFactory struct {
	config *Config
}

// NewAIServiceFactory creates a new AI service factory
// It initializes the factory with the provided configuration
// This allows the factory to create services based on the current configuration
func NewAIServiceFactory(config *Config) *AIServiceFactory {
	return &AIServiceFactory{
		config: config,
	}
}

// CreateEmbeddingService creates an embedding service based on configuration
// It returns an instance of EmbeddingService for the configured provider
// For example, it can return OpenAIEmbeddingService, GeminiEmbeddingService, or OllamaEmbeddingService
func (f *AIServiceFactory) CreateEmbeddingService() (EmbeddingService, error) {
	provider := f.determineProvider()

	switch provider {
	case OpenAIProvider:
		return NewOpenAIEmbeddingService(f.config), nil
	case GeminiProvider:
		return NewGeminiEmbeddingService(f.config), nil
	case OllamaProvider:
		return NewOllamaEmbeddingService(f.config), nil
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", provider)
	}
}

// CreateChatService creates a chat service based on configuration
// It returns an instance of ChatService for the configured provider
// For example, it can return OpenAIChatService, GeminiChatService, or OllamaChatService
func (f *AIServiceFactory) CreateChatService() (ChatService, error) {
	provider := f.determineProvider()

	switch provider {
	case OpenAIProvider:
		return NewOpenAIChatService(f.config), nil
	case GeminiProvider:
		return NewGeminiChatService(f.config), nil
	case OllamaProvider:
		return NewOllamaChatService(f.config), nil
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", provider)
	}
}

// determineProvider determines which AI provider to use based on configuration
// It checks the configuration for the presence of API keys or URLs
// Priority: Local AI (Ollama) > Gemini > OpenAI
func (f *AIServiceFactory) determineProvider() AIProvider {
	// Priority: Local AI (Ollama) > Gemini > OpenAI
	if f.config.UseLocalAI {
		return OllamaProvider
	}

	if f.config.GoogleAIAPIKey != "" {
		return GeminiProvider
	}

	if f.config.OpenAIAPIKey != "" {
		return OpenAIProvider
	}

	// Default fallback (will cause error in factory methods)
	return OpenAIProvider
}

// GetCurrentProvider returns the currently configured provider
// This is useful for logging or debugging purposes
// It returns the AIProvider enum value
// that indicates which provider is currently set
func (f *AIServiceFactory) GetCurrentProvider() AIProvider {
	return f.determineProvider()
}

// ValidateConfiguration validates that the required configuration is present for the determined provider
// It checks for the presence of API keys or URLs based on the provider
// This ensures that the application has the necessary credentials to interact with the AI service
// It returns an error if the configuration is invalid
func (f *AIServiceFactory) ValidateConfiguration() error {
	provider := f.determineProvider()

	switch provider {
	case OpenAIProvider:
		if f.config.OpenAIAPIKey == "" {
			return fmt.Errorf("OPENAI_API_KEY is required for OpenAI provider")
		}
	case GeminiProvider:
		if f.config.GoogleAIAPIKey == "" {
			return fmt.Errorf("GOOGLE_AI_API_KEY is required for Gemini provider")
		}
	case OllamaProvider:
		if f.config.OllamaBaseURL == "" {
			return fmt.Errorf("OLLAMA_BASE_URL is required for Ollama provider")
		}
	}

	return nil
}
