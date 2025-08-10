package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/lib/pq"
)

// OllamaEmbeddingService implements EmbeddingService for Ollama
type OllamaEmbeddingService struct {
	config  *Config
	baseURL string
}

// Ollama API structures for embeddings
type ollamaEmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// ollamaEmbeddingResponse represents the response structure from Ollama
// It contains the embedding data returned by the Ollama API
type ollamaEmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

// NewOllamaEmbeddingService creates a new Ollama embedding service
// It initializes the service with the configuration and base URL for Ollama
// This allows the service to make requests to the Ollama API for generating embeddings
func NewOllamaEmbeddingService(config *Config) *OllamaEmbeddingService {
	return &OllamaEmbeddingService{
		config:  config,
		baseURL: config.OllamaBaseURL,
	}
}

// GenerateEmbedding generates embeddings using Ollama API
func (s *OllamaEmbeddingService) GenerateEmbedding(text string) (Vector, error) {
	// trim whitespace from the input text
	// If the text is empty after trimming, return an error
	cleanedText := strings.TrimSpace(text)
	if cleanedText == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	// Create a retryable embedding request
	// This allows us to handle transient errors and retry the request
	var embedding pq.Float32Array
	retryConfig := DefaultRetryConfig()

	// Retry the embedding request with exponential backoff
	// This helps to handle temporary network issues or API rate limits
	err := RetryWithBackoff(retryConfig, func() error {
		// Make the actual embedding request to Ollama
		// *embedding is to dereference the pointer and assign the embedding data
		return s.makeEmbeddingRequest(cleanedText, &embedding)
	})

	if err != nil {
		LogError("Failed to get Ollama embedding after retries", err, "text_length", len(cleanedText))
		return nil, err
	}

	return Vector(embedding), nil
}

// GenerateBatchEmbeddings generates embeddings for multiple texts (Ollama doesn't support batch, so we call individually)
func (s *OllamaEmbeddingService) GenerateBatchEmbeddings(texts []string) ([]Vector, error) {
	// Check if the input texts slice is empty
	// If it is empty, return an error
	if len(texts) == 0 {
		return nil, fmt.Errorf("texts cannot be empty")
	}

	// Create a slice to hold the embeddings for each text
	// This will store the embeddings generated for each input text
	embeddings := make([]Vector, len(texts))
	for i, text := range texts {
		// Generate embedding for each text
		// This will call the GenerateEmbedding method for each text
		embedding, err := s.GenerateEmbedding(text)
		if err != nil {
			return nil, fmt.Errorf("failed to get embedding for text %d: %v", i, err)
		}
		// Store the embedding in the slice
		// embeddings[i] is the ith embedding for the ith text
		embeddings[i] = embedding
	}

	LogInfo("Successfully generated Ollama batch embeddings", "text_count", len(texts))
	return embeddings, nil
}

// GetProviderName returns the provider name
// This is used to identify the AI service provider
// In this case, it returns "Ollama" as the provider name
func (s *OllamaEmbeddingService) GetProviderName() string {
	return "Ollama"
}

// makeEmbeddingRequest makes an embedding request to Ollama
func (s *OllamaEmbeddingService) makeEmbeddingRequest(text string, embedding *pq.Float32Array) error {
	// Create request
	request := ollamaEmbeddingRequest{
		Model:  s.config.EmbeddingModel,
		Prompt: text,
	}

	// Marshal the request into JSON
	// This converts the request struct into a JSON format that can be sent to the Ollama
	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	// Make API request to Ollama
	// This sends the JSON data to the Ollama API endpoint for embeddings
	// The response will contain the embedding data
	url := fmt.Sprintf("%s/api/embeddings", s.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set the content type to application/json
	// This tells the Ollama API that we are sending JSON data
	req.Header.Set("Content-Type", "application/json")

	// Make the HTTP request to Ollama
	// This sends the request to the Ollama API and waits for the response
	// Do() is used to execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		LogError("Failed to make Ollama API request", err, "text_length", len(text))
		return fmt.Errorf("failed to make request: %v", err)
	}

	// Ensure the response body is closed after reading
	// This prevents resource leaks by ensuring the response body is properly closed
	defer resp.Body.Close()

	// Check if the response status is OK (200)
	// If not, read the response body and log the error
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		LogError("Ollama API error", fmt.Errorf("status: %s", resp.Status), "response_body", string(body))
		return fmt.Errorf("Ollama API error: %s - %s", resp.Status, string(body))
	}

	// Decode the response body into the ollamaEmbeddingResponse struct
	// This extracts the embedding data from the response
	var response ollamaEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		LogError("Failed to decode Ollama response", err)
		return fmt.Errorf("failed to decode response: %v", err)
	}

	// Check if the embedding data is present in the response
	// If the embedding slice is empty, return an error
	if len(response.Embedding) == 0 {
		LogError("No embedding data received from Ollama", fmt.Errorf("empty response"))
		return fmt.Errorf("no embedding data received")
	}

	// Assign the embedding data to the provided pq.Float32Array pointer
	// This allows the caller to receive the generated embedding
	// *embedding is to dereference the pointer and assign the embedding data
	*embedding = pq.Float32Array(response.Embedding)
	LogInfo("Successfully generated Ollama embedding", "text_length", len(text), "embedding_size", len(*embedding))
	return nil
}

// OllamaChatService implements ChatService for Ollama
type OllamaChatService struct {
	config  *Config
	baseURL string
	model   string
}

// Ollama API structures for chat
type ollamaChatRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// ollamaChatResponse represents the response structure from Ollama chat API
// It contains the generated response and a done flag indicating if the response is complete
type ollamaChatResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// NewOllamaChatService creates a new Ollama chat service
// It initializes the service with the configuration and base URL for Ollama
// This allows the service to make requests to the Ollama API for generating chat responses
func NewOllamaChatService(config *Config) *OllamaChatService {
	return &OllamaChatService{
		config:  config,
		baseURL: config.OllamaBaseURL,
		model:   config.ChatModel,
	}
}

// GenerateResponse generates a response using Ollama chat completion
func (s *OllamaChatService) GenerateResponse(question, context string) (string, error) {
	// Default retry configuration for chat requests
	// This allows us to handle transient errors and retry the request
	var response string
	retryConfig := DefaultRetryConfig()

	// Retry the chat request with exponential backoff
	// This helps to handle temporary network issues or API rate limits
	err := RetryWithBackoff(retryConfig, func() error {
		// Make the actual chat request to Ollama
		// This sends the question and context to the Ollama API for generating a response
		// *response is to dereference the pointer and assign the response data
		// This allows us to modify the response directly without returning it
		return s.makeChatRequest(question, context, &response)
	})

	if err != nil {
		LogError("Failed to generate Ollama response after retries", err, "question", question)
		return "", err
	}

	LogInfo("Successfully generated Ollama response", "question_length", len(question), "response_length", len(response))
	return response, nil
}

// GetProviderName returns the provider name
func (s *OllamaChatService) GetProviderName() string {
	return "Ollama"
}

// GetModel returns the model name
func (s *OllamaChatService) GetModel() string {
	return s.model
}

// makeChatRequest makes a chat completion request to Ollama
// *string means that the response will be written to the provided string pointer
// This allows us to modify the response directly without returning it
func (s *OllamaChatService) makeChatRequest(question, context string, response *string) error {
	// Build prompt with context
	prompt := fmt.Sprintf(`You are a helpful assistant that answers questions based on provided context. 
Use the following context to answer the user's question. If the context doesn't contain enough information to answer the question, say so clearly.

Context:
%s

Question: %s

Answer:`, context, question)

	// Create request
	request := ollamaChatRequest{
		Model:  s.model,
		Prompt: prompt,
		Stream: false,
	}

	// Marshal the request into JSON
	// This converts the request struct into a JSON format that can be sent to the Ollama API
	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	// Make API request to Ollama
	// This sends the JSON data to the Ollama API endpoint for chat completion
	// The response will contain the generated answer
	url := fmt.Sprintf("%s/api/generate", s.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set the content type to application/json
	// This tells the Ollama API that we are sending JSON data
	req.Header.Set("Content-Type", "application/json")

	// Make the HTTP request to Ollama
	// This sends the request to the Ollama API and waits for the response
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		LogError("Failed to make Ollama chat request", err)
		return fmt.Errorf("failed to make request: %v", err)
	}

	// Defer the closing of the response body
	// This prevents resource leaks by ensuring the response body is properly closed
	defer resp.Body.Close()

	// Check if the response status is OK (200)
	// If not, read the response body and log the error
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		LogError("Ollama chat API error", fmt.Errorf("status: %s", resp.Status), "response_body", string(body))
		return fmt.Errorf("Ollama API error: %s - %s", resp.Status, string(body))
	}

	// Decode the response body into the ollamaChatResponse struct
	// This extracts the generated response from the Ollama API
	var chatResponse ollamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResponse); err != nil {
		LogError("Failed to decode Ollama chat response", err)
		return fmt.Errorf("failed to decode response: %v", err)
	}

	// Check if the response is empty
	// If the response is empty, log an error and return an error
	if chatResponse.Response == "" {
		LogError("No response received from Ollama", fmt.Errorf("empty response"))
		return fmt.Errorf("no response received")
	}

	// Assign the response to the provided string pointer
	// This allows the caller to receive the generated response
	// *response is to dereference the pointer and assign the response data
	*response = strings.TrimSpace(chatResponse.Response)
	return nil
}
