package models

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/MauricioAliendre182/backend/utils"
)

// RAGService handles Retrieval-Augmented Generation using the factory pattern
type RAGService struct {
	MaxChunks   int
	chatService utils.ChatService
}

// NewRAGService creates a new RAG service using the factory pattern
func NewRAGService() (*RAGService, error) {
	factory := utils.NewAIServiceFactory(utils.AppConfig)

	// Validate configuration
	if err := factory.ValidateConfiguration(); err != nil {
		return nil, fmt.Errorf("invalid AI configuration: %v", err)
	}

	// Create the chat service
	// This will use the factory to create an instance of ChatService based on the current configuration
	// This allows the RAG service to use the appropriate AI provider for generating responses
	// The factory pattern allows for easy extension if new AI providers are added in the future
	chatService, err := factory.CreateChatService()
	if err != nil {
		return nil, fmt.Errorf("failed to create chat service: %v", err)
	}

	utils.LogInfo("RAG service initialized", "provider", chatService.GetProviderName(), "model", chatService.GetModel())

	// Return a new instance of RAGService with the chat service
	// The MaxChunks field can be adjusted based on your requirements for how many chunks to retrieve
	return &RAGService{
		MaxChunks:   10, // Default value, can be adjusted as needed
		chatService: chatService,
	}, nil
}

// QueryDocuments performs RAG query on document using the factory pattern
// It retrieves relevant chunks based on the question embedding and generates a response using the chat service
// This method encapsulates the logic for querying documents and generating responses
func (r *RAGService) QueryDocuments(question string) (string, error) {
	utils.LogInfo("Starting RAG query", "question", question)

	// Step 1: Get embedding for the question
	questionEmbedding, err := utils.GetEmbedding(question)
	if err != nil {
		return "", fmt.Errorf("failed to get question embedding: %v", err)
	}

	// Clean the embedding to remove any non-float data (timestamps, extra text, etc.)
	cleanedEmbedding := cleanEmbeddingVector(questionEmbedding)

	utils.LogInfo("Generated question embedding", "original_length", len(questionEmbedding), "cleaned_length", len(cleanedEmbedding))

	// Step 2: Find relevant chunks using similarity search
	// This function should be implemented to perform a similarity search
	// It retrieves the most relevant chunks based on the question embedding
	relevantChunks, err := SimilaritySearch(cleanedEmbedding, r.MaxChunks)
	if err != nil {
		utils.LogError("Similarity search failed", err)
		return "", fmt.Errorf("failed to find relevant chunks: %v", err)
	}

	utils.LogInfo("Similarity search completed", "chunks_found", len(relevantChunks), "max_chunks", r.MaxChunks)

	if len(relevantChunks) == 0 {
		utils.LogWarn("No relevant chunks found for question", "question", question)
		return "I couldn't find any relevant information in the documents to answer your question.", nil
	}

	// Step 3: Build context from relevant chunks
	var contextBuilder strings.Builder
	contextBuilder.WriteString("Based on the following information from the documents:\n\n")

	for i, chunk := range relevantChunks {
		utils.LogInfo("Adding chunk to context", "chunk_index", i, "content_length", len(chunk.Content), "document_id", chunk.DocumentID.String())
		contextBuilder.WriteString(fmt.Sprintf("Document %d:\n%s\n\n", i+1, chunk.Content))
	}

	// Step 4: Generate response using the configured AI service with guardrails
	utils.LogInfo("Generating AI response", "context_length", contextBuilder.Len())
	contextText := contextBuilder.String()

	// Count tokens in the context text
	// This helps in understanding the context size and ensuring it fits within the model's limits
	tokens, err := utils.CountTokens(contextText, utils.AppConfig.EmbeddingModel)
	if err != nil {
		utils.LogError("Failed to count tokens", err)
	} else {
		utils.LogInfo("Context token count", "tokens", tokens)
	}

	// Create a safe prompt that includes guardrails
	safePrompt := utils.CreateSafePrompt(question, contextText)
	utils.LogInfo("Created safe prompt", "prompt_length", len(safePrompt))

	// Use the safe prompt as the question parameter and empty context
	// The context is already included in the safe prompt
	return r.chatService.GenerateResponse(safePrompt, "")
}

// cleanEmbeddingVector removes any non-float data from embedding vectors
// This fixes issues where timestamps or other data get mixed into the embedding array
func cleanEmbeddingVector(embedding utils.Vector) utils.Vector {
	if len(embedding) == 0 {
		return embedding
	}

	expectedDimensions := 1536 // For OpenAI text-embedding-3-small

	// If the embedding is much larger than expected, it likely contains corrupted data
	if len(embedding) > expectedDimensions*2 {
		return cleanCorruptedEmbedding(embedding, expectedDimensions)
	}

	// If length is reasonable, just return the original
	utils.LogInfo("Embedding vector length is acceptable", "length", len(embedding))
	return embedding
}

// cleanCorruptedEmbedding handles cleaning of corrupted embedding vectors
func cleanCorruptedEmbedding(embedding utils.Vector, expectedDimensions int) utils.Vector {
	utils.LogWarn("Embedding vector appears corrupted", "length", len(embedding), "expected", expectedDimensions)

	cleaned := make(utils.Vector, 0, expectedDimensions)
	validCount := 0

	for i, value := range embedding {
		if validCount >= expectedDimensions {
			break
		}

		// Check if the value is an integer representation of a float
		if isAnIntegerAndInValidEmbeddingValue(value) {
			utils.LogWarn("Skipping invalid embedding value", "index", i, "value", value)
			continue
		}

		// Attempt to extract a valid float from the potentially corrupted value
		// This handles cases where the value might be a string or corrupted float
		if cleanedValue, ok := extractValidFloat(value, i); ok {
			cleaned = append(cleaned, cleanedValue)
			validCount++
		}
	}

	utils.LogInfo("Cleaned corrupted embedding vector", "original_length", len(embedding), "cleaned_length", len(cleaned))
	return cleaned
}

// extractValidFloat attempts to extract a valid float32 from a potentially corrupted value
func extractValidFloat(value float32, index int) (float32, bool) {
	// Check if the value is already reasonable for embeddings
	if isValidEmbeddingValue(value) {
		return value, true
	}

	// Try to extract the float value from corrupted data using regex
	return extractFloatFromCorruptedString(value, index)
}

// isValidEmbeddingValue checks if a float32 value is reasonable for embeddings
// Value must be between -1.0 and 1.0 for most embedding models
func isValidEmbeddingValue(value float32) bool {
	return value >= -1.0 && value <= 1.0
}

func isAnIntegerAndInValidEmbeddingValue(value float32) bool {
	return value > 1.0 && value < -1.0
}

// extractFloatFromCorruptedString uses regex to extract valid float from corrupted data
func extractFloatFromCorruptedString(value float32, index int) (float32, bool) {
	// Regex to extract float values from corrupted strings like "0.028082025-08-03T20:35:09.301793742Z, 8845"
	floatRegex := regexp.MustCompile(`^-?\d*\.?\d+`)

	// Convert the float32 to string to apply regex
	valueStr := fmt.Sprintf("%v", value)

	// Find the first valid float in the string
	// FindString returns the first match of the regex in the string
	// If no match is found, it returns an empty string
	matches := floatRegex.FindString(valueStr)
	if matches == "" {
		if index < 10 {
			utils.LogWarn("Could not extract float from corrupted value", "index", index, "value", valueStr)
		}
		return 0, false
	}

	// ParseFloat(matches, 32) is used to convert the extracted string to a float32
	// This handles cases where the value might be a string or corrupted float
	parsedFloat, err := strconv.ParseFloat(matches, 32)
	if err != nil {
		if index < 10 {
			utils.LogWarn("Could not parse extracted float", "index", index, "extracted", matches, "error", err)
		}
		return 0, false
	}

	// Ensure the parsed float is within a reasonable range for embeddings
	// This is used to filter out any outliers or corrupted values that might have slipped through
	parsedFloat32 := float32(parsedFloat)
	if !isValidEmbeddingValue(parsedFloat32) {
		if index < 10 {
			utils.LogWarn("Extracted float is out of valid range", "index", index, "value", parsedFloat32)
		}
		return 0, false
	}

	if index < 10 { // Log first few corrections for debugging
		utils.LogInfo("Corrected corrupted embedding value", "index", index, "original", valueStr, "corrected", parsedFloat32)
	}

	return parsedFloat32, true
}
