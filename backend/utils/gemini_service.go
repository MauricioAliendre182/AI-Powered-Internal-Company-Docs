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

// GeminiEmbeddingService implements EmbeddingService for Google AI (Gemini)
type GeminiEmbeddingService struct {
	config *Config
	apiKey string
}

// Gemini API structures for embeddings
type geminiEmbeddingRequest struct {
	Model   string `json:"model"`
	Content struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"content"`
}

type geminiEmbeddingResponse struct {
	Embedding struct {
		Values []float32 `json:"values"`
	} `json:"embedding"`
}

// NewGeminiEmbeddingService creates a new Gemini embedding service
// It initializes the service with the provided configuration
// This allows the service to use the correct API key and model for embedding generation
func NewGeminiEmbeddingService(config *Config) *GeminiEmbeddingService {
	return &GeminiEmbeddingService{
		config: config,
		apiKey: config.GoogleAIAPIKey,
	}
}

// GenerateEmbedding generates embeddings using Gemini API
// It returns a Vector containing the embedding values
// It handles rate limiting and retries using the configured retry strategy
func (s *GeminiEmbeddingService) GenerateEmbedding(text string) (Vector, error) {
	// Trim whitespace and check for empty text
	cleanedText := strings.TrimSpace(text)
	if cleanedText == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	// Rate limiting
	// Check if the rate limiter allows the request
	// This prevents exceeding the API rate limits
	if !OpenAIRateLimiter.Allow() {
		LogWarn("Rate limit exceeded for Gemini API call")
		return nil, fmt.Errorf("rate limit exceeded, please try again later")
	}

	// Prepare the embedding request
	var embedding pq.Float32Array
	// Use the default retry configuration for API calls
	retryConfig := DefaultRetryConfig()

	// Retry the embedding request with backoff
	// This allows the service to handle transient errors gracefully
	err := RetryWithBackoff(retryConfig, func() error {
		// If the request fails, it will retry according to the retry configuration
		// Make the actual API request to generate the embedding
		return s.makeEmbeddingRequest(cleanedText, &embedding)
	})

	if err != nil {
		LogError("Failed to get Gemini embedding after retries", err, "text_length", len(cleanedText))
		return nil, err
	}

	return Vector(embedding), nil
}

// GenerateBatchEmbeddings generates embeddings for multiple texts (Gemini doesn't support batch, so we call individually)
func (s *GeminiEmbeddingService) GenerateBatchEmbeddings(texts []string) ([]Vector, error) {
	// Check for empty input
	if len(texts) == 0 {
		return nil, fmt.Errorf("texts cannot be empty")
	}

	// Prepare a slice to hold the embeddings
	// This will hold the embeddings for each text in the input slice
	embeddings := make([]Vector, len(texts))
	for i, text := range texts {
		// Generate embedding for each text
		// This will hold the embeddings for each text in the input slice
		embedding, err := s.GenerateEmbedding(text)
		if err != nil {
			return nil, fmt.Errorf("failed to get embedding for text %d: %v", i, err)
		}
		// Store the embedding in the slice
		embeddings[i] = embedding
	}

	LogInfo("Successfully generated Gemini batch embeddings", "text_count", len(texts))
	return embeddings, nil
}

// GetProviderName returns the provider name
// This is used to identify the AI service provider
// It allows the system to know which AI service is being used for embedding generation
func (s *GeminiEmbeddingService) GetProviderName() string {
	return "Gemini"
}

// makeEmbeddingRequest makes an embedding request to Gemini
func (s *GeminiEmbeddingService) makeEmbeddingRequest(text string, embedding *pq.Float32Array) error {
	// Create request
	// This request structure is specific to Gemini's embedding API
	request := geminiEmbeddingRequest{
		Model: s.config.EmbeddingModel,
	}

	// Set the content parts with the text to be embedded
	// This is the text that will be processed by the Gemini API to generate embeddings
	request.Content.Parts = []struct {
		Text string `json:"text"`
	}{{Text: text}}

	// Marshal the request to JSON
	// This converts the request structure into a format that can be sent over HTTP
	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	// Gemini API endpoint
	// This is the URL for the Gemini embedding API
	// It includes the model name and API key for authentication
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/%s:embedContent?key=%s", s.config.EmbeddingModel, s.apiKey)

	// Create a new HTTP request
	// This request will be sent to the Gemini API to generate the embedding
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set the content type header
	// This tells the API that we are sending JSON data
	req.Header.Set("Content-Type", "application/json")

	// Send the request to the Gemini API
	// Do() is used to execute the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		LogError("Failed to make Gemini API request", err, "text_length", len(text))
		return fmt.Errorf("failed to make request: %v", err)
	}

	// Ensure the response body is closed after reading
	defer resp.Body.Close()

	// Check the response status code
	// If the status code is not OK, log the error and return
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		LogError("Gemini API error", fmt.Errorf("status: %s", resp.Status), "response_body", string(body))
		return fmt.Errorf("Gemini API error: %s - %s", resp.Status, string(body))
	}

	// Decode the response body into the geminiEmbeddingResponse structure
	// This extracts the embedding values from the API response
	var response geminiEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		LogError("Failed to decode Gemini response", err)
		return fmt.Errorf("failed to decode response: %v", err)
	}

	// Check if the response contains embedding values
	// If the embedding values are empty, log an error and return
	if len(response.Embedding.Values) == 0 {
		LogError("No embedding data received from Gemini", fmt.Errorf("empty response"))
		return fmt.Errorf("no embedding data received")
	}

	// Store the embedding values in the provided pq.Float32Array
	// This will be used by the caller to access the generated embeddings
	// *embedding is to dereference the pointer and assign the values
	*embedding = pq.Float32Array(response.Embedding.Values)
	LogInfo("Successfully generated Gemini embedding", "text_length", len(text), "embedding_size", len(*embedding))
	return nil
}

// GeminiChatService implements ChatService for Google AI (Gemini)
type GeminiChatService struct {
	config *Config
	apiKey string
	model  string
}

// Gemini API structures for chat
type geminiChatRequest struct {
	Contents         []geminiContent `json:"contents"`
	GenerationConfig struct {
		Temperature     float32 `json:"temperature,omitempty"`
		MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	} `json:"generationConfig,omitempty"`
}

// Gemini content structure for chat requests
type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

// Gemini part structure for chat responses
type geminiPart struct {
	Text string `json:"text"`
}

// Gemini chat response structure
// This structure is used to parse the response from the Gemini chat API
type geminiChatResponse struct {
	Candidates []struct {
		FinishReason string `json:"finishReason"`
		Content      struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// NewGeminiChatService creates a new Gemini chat service
// It initializes the service with the provided configuration
// This allows the service to use the correct API key and model for chat completion
func NewGeminiChatService(config *Config) *GeminiChatService {
	return &GeminiChatService{
		config: config,
		apiKey: config.GoogleAIAPIKey,
		model:  config.ChatModel,
	}
}

// GenerateResponse generates a response using Gemini chat completion
func (s *GeminiChatService) GenerateResponse(question, context string) (string, error) {
	// Rate limiting
	// Check if the rate limiter allows the request
	// This prevents exceeding the API rate limits
	if !OpenAIRateLimiter.Allow() {
		LogWarn("Rate limit exceeded for Gemini chat completion")
		return "", fmt.Errorf("rate limit exceeded, please try again later")
	}

	var response string

	// Default retry configuration for API calls
	// This allows the service to handle transient errors gracefully
	retryConfig := DefaultRetryConfig()

	// Retry the chat request with backoff
	// This allows the service to handle transient errors gracefully
	err := RetryWithBackoff(retryConfig, func() error {
		// If the request fails, it will retry according to the retry configuration
		// Make the actual API request to generate the response
		// *string means that the response will be written to the provided string pointer
		return s.makeChatRequest(question, context, &response)
	})

	if err != nil {
		LogError("Failed to generate Gemini response after retries", err, "question", question)
		return "", err
	}

	LogInfo("Successfully generated Gemini response", "question_length", len(question), "response_length", len(response))
	return response, nil
}

// GetProviderName returns the provider name
// This is used to identify the AI service provider
// It allows the system to know which AI service is being used for chat completion
func (s *GeminiChatService) GetProviderName() string {
	return "Gemini"
}

// GetModel returns the model name
// This is used to identify the specific model being used for chat completion
// It allows the system to know which model is being used for generating responses
func (s *GeminiChatService) GetModel() string {
	return s.model
}

// makeChatRequest makes a chat completion request to Gemini
func (s *GeminiChatService) makeChatRequest(question, context string, response *string) error {
	// Create system message with context
	systemPrompt := fmt.Sprintf(`You are a helpful assistant that answers questions based on provided context. 
Use the following context to answer the user's question. If the context doesn't contain enough information to answer the question, say so clearly.

Context:
%s

Question: %s`, context, question)

	// Create request
	request := geminiChatRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{{Text: systemPrompt}},
				Role:  "user",
			},
		},
	}
	// Temperature and max output tokens can be adjusted based on requirements
	// These parameters control the randomness and length of the generated response
	request.GenerationConfig.Temperature = 0.1
	// Max output tokens can be adjusted based on requirements
	// This parameter controls the maximum length of the generated response
	request.GenerationConfig.MaxOutputTokens = 1000

	// Marshal the request to JSON
	// This converts the request structure into a format that can be sent over HTTP
	// This is necessary for the Gemini API to understand the request format
	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	// Gemini API endpoint
	// This is the URL for the Gemini chat API
	// It includes the model name and API key for authentication
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/%s:generateContent?key=%s", s.model, s.apiKey)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set the content type header
	// This tells the API that we are sending JSON data
	req.Header.Set("Content-Type", "application/json")

	// Send the request to the Gemini API
	// Do() is used to execute the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		LogError("Failed to make Gemini chat request", err)
		return fmt.Errorf("failed to make request: %v", err)
	}

	// Ensure the response body is closed after reading
	// This prevents resource leaks by closing the response body after use
	defer resp.Body.Close()

	// Check the response status code
	// If the status code is not OK, log the error and return
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		LogError("Gemini chat API error", fmt.Errorf("status: %s", resp.Status), "response_body", string(body))
		return fmt.Errorf("Gemini API error: %s - %s", resp.Status, string(body))
	}

	// Decode the response body into the geminiChatResponse structure
	// This extracts the response candidates from the API response
	var chatResponse geminiChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResponse); err != nil {
		LogError("Failed to decode Gemini chat response", err)
		return fmt.Errorf("failed to decode response: %v", err)
	}

	// Check if the response contains candidates
	// If the candidates are empty, log an error and return
	if len(chatResponse.Candidates) == 0 || len(chatResponse.Candidates[0].Content.Parts) == 0 {
		LogError("No response candidates received from Gemini", fmt.Errorf("empty candidates"))
		return fmt.Errorf("no response candidates received")
	}

	// Extract the response text from the first candidate
	// This is the text that will be returned as the response to the user's question
	// *response is to dereference the pointer and assign the text
	*response = strings.TrimSpace(chatResponse.Candidates[0].Content.Parts[0].Text)
	return nil
}
