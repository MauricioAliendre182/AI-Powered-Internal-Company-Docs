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

// OpenAIEmbeddingService implements EmbeddingService for OpenAI
type OpenAIEmbeddingService struct {
	config *Config
	apiKey string
}

// OpenAI API structures for embeddings
type openAIEmbeddingRequest struct {
	Input          []string `json:"input"`
	Model          string   `json:"model"`
	EncodingFormat string   `json:"encoding_format,omitempty"`
}

// openAIEmbeddingResponse represents the response structure from OpenAI for embeddings
type openAIEmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// NewOpenAIEmbeddingService creates a new OpenAI embedding service
// It initializes the service with the provided configuration
// This allows the service to use the OpenAI API for generating embeddings
func NewOpenAIEmbeddingService(config *Config) *OpenAIEmbeddingService {
	return &OpenAIEmbeddingService{
		config: config,
		apiKey: config.OpenAIAPIKey,
	}
}

// GenerateEmbedding generates embeddings using OpenAI API
func (s *OpenAIEmbeddingService) GenerateEmbedding(text string) (Vector, error) {
	// Trim whitespace from the input text
	// If the text is empty after trimming, return an error
	cleanedText := strings.TrimSpace(text)
	if cleanedText == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	// Rate limiting
	// Check if the rate limiter allows the request
	// If the rate limit is exceeded, log a warning and return an error
	if !OpenAIRateLimiter.Allow() {
		LogWarn("Rate limit exceeded for OpenAI API call")
		return nil, fmt.Errorf("rate limit exceeded, please try again later")
	}

	// Prepare the request for OpenAI embedding
	// This includes the model and encoding format
	var embedding pq.Float32Array
	// Use a retry mechanism to handle transient errors
	// This allows the service to retry the request in case of temporary issues
	retryConfig := DefaultRetryConfig()

	// Retry the embedding request with backoff
	// This helps to handle transient errors and ensures reliability
	err := RetryWithBackoff(retryConfig, func() error {
		// Make the actual request to OpenAI API
		// This function will handle the HTTP request and response parsing
		// It will populate the embedding variable with the result
		// *embedding is a pointer to pq.Float32Array
		return s.makeEmbeddingRequest(cleanedText, &embedding)
	})

	if err != nil {
		LogError("Failed to get OpenAI embedding after retries", err, "text_length", len(cleanedText))
		return nil, err
	}

	// Convert pq.Float32Array to Vector
	// Vector() is a type conversion that converts pq.Float32Array to Vector
	return Vector(embedding), nil
}

// GenerateBatchEmbeddings generates embeddings for multiple texts
// This is necessary for processing multiple inputs in a single API call (Open AI supports batch embeddings)
func (s *OpenAIEmbeddingService) GenerateBatchEmbeddings(texts []string) ([]Vector, error) {
	// Check if the input texts are empty
	// If the texts slice is empty, return an error
	if len(texts) == 0 {
		return nil, fmt.Errorf("texts cannot be empty")
	}

	// OpenAI supports batch embeddings
	// Clean the input texts by trimming whitespace and removing empty strings
	// This ensures that only valid texts are processed
	cleanedTexts := make([]string, 0, len(texts))
	for _, text := range texts {
		if cleaned := strings.TrimSpace(text); cleaned != "" {
			cleanedTexts = append(cleanedTexts, cleaned)
		}
	}

	// If no valid texts remain after cleaning, return an error
	// This prevents unnecessary API calls with empty inputs
	if len(cleanedTexts) == 0 {
		return nil, fmt.Errorf("no valid texts provided")
	}

	// Rate limiting
	// Check if the rate limiter allows the request
	// If the rate limit is exceeded, log a warning and return an error
	if !OpenAIRateLimiter.Allow() {
		LogWarn("Rate limit exceeded for OpenAI batch embedding call")
		return nil, fmt.Errorf("rate limit exceeded, please try again later")
	}

	// Prepare the request for OpenAI batch embedding
	var embeddings []pq.Float32Array
	// Use a retry mechanism to handle transient errors
	// This allows the service to retry the request in case of temporary issues
	retryConfig := DefaultRetryConfig()

	// Retry the batch embedding request with backoff
	// This helps to handle transient errors and ensures reliability
	err := RetryWithBackoff(retryConfig, func() error {
		// Make the actual request to OpenAI API
		// This function will handle the HTTP request and response parsing
		// It will populate the embeddings variable with the result
		// *embeddings is a pointer to []pq.Float32Array
		return s.makeBatchEmbeddingRequest(cleanedTexts, &embeddings)
	})

	if err != nil {
		LogError("Failed to get OpenAI batch embeddings after retries", err, "text_count", len(cleanedTexts))
		return nil, err
	}

	LogInfo("Successfully generated OpenAI batch embeddings", "text_count", len(cleanedTexts))

	// Convert []pq.Float32Array to []Vector
	result := make([]Vector, len(embeddings))
	for i, embedding := range embeddings {
		result[i] = Vector(embedding)
	}

	return result, nil
}

// GetProviderName returns the provider name
func (s *OpenAIEmbeddingService) GetProviderName() string {
	return "OpenAI"
}

// makeEmbeddingRequest makes a single embedding request to OpenAI
func (s *OpenAIEmbeddingService) makeEmbeddingRequest(text string, embedding *pq.Float32Array) error {
	// Create the request structure for OpenAI embedding
	request := openAIEmbeddingRequest{
		Input:          []string{text},
		Model:          s.config.EmbeddingModel,
		EncodingFormat: "float",
	}

	// Marshal the request to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create the HTTP request to get the embedding
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set the necessary headers for the request
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	// Make the HTTP request to OpenAI API
	// Do() executes the request and returns the response
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		LogError("Failed to make OpenAI API request", err, "text_length", len(text))
		return fmt.Errorf("failed to make request: %v", err)
	}

	// Ensure the response body is closed after reading
	// This prevents resource leaks and allows the response body to be read
	defer resp.Body.Close()

	// Check the response status code
	// If the status code is not 200 OK, log an error and return an error
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		LogError("OpenAI API error", fmt.Errorf("status: %s", resp.Status), "response_body", string(body))
		return fmt.Errorf("OpenAI API error: %s - %s", resp.Status, string(body))
	}

	// Decode the response body into the openAIEmbeddingResponse structure
	// This structure contains the embedding data returned by OpenAI
	var response openAIEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		LogError("Failed to decode OpenAI response", err)
		return fmt.Errorf("failed to decode response: %v", err)
	}

	// Check if the response contains valid embedding data
	if len(response.Data) == 0 || len(response.Data[0].Embedding) == 0 {
		LogError("No embedding data received from OpenAI", fmt.Errorf("empty response"))
		return fmt.Errorf("no embedding data received")
	}

	// Populate the embedding variable with the first embedding from the response
	// This is the expected format from OpenAI's embedding API
	// *embedding is a pointer to pq.Float32Array
	*embedding = pq.Float32Array(response.Data[0].Embedding)
	LogInfo("Successfully generated OpenAI embedding", "text_length", len(text), "embedding_size", len(*embedding))
	return nil
}

// makeBatchEmbeddingRequest makes a batch embedding request to OpenAI
func (s *OpenAIEmbeddingService) makeBatchEmbeddingRequest(texts []string, embeddings *[]pq.Float32Array) error {
	// Create the request structure for OpenAI batch embedding
	request := openAIEmbeddingRequest{
		Input:          texts,
		Model:          s.config.EmbeddingModel,
		EncodingFormat: "float",
	}

	// Marshal the request to JSON
	// This converts the request structure into JSON format for the API call
	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create the HTTP request to get the batch embeddings
	// This includes the model and encoding format
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set the necessary headers for the request
	// This includes the content type and authorization header with the API key
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	// Make the HTTP request to OpenAI API
	// Do() executes the request and returns the response
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		LogError("Failed to make OpenAI batch API request", err, "text_count", len(texts))
		return fmt.Errorf("failed to make request: %v", err)
	}

	// Ensure the response body is closed after reading
	// This prevents resource leaks and allows the response body to be read
	defer resp.Body.Close()

	// Check the response status code
	// If the status code is not 200 OK, log an error and return an error
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		LogError("OpenAI batch API error", fmt.Errorf("status: %s", resp.Status), "response_body", string(body))
		return fmt.Errorf("OpenAI API error: %s - %s", resp.Status, string(body))
	}

	// Decode the response body into the openAIEmbeddingResponse structure
	// This structure contains the embedding data returned by OpenAI
	var response openAIEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		LogError("Failed to decode OpenAI batch response", err)
		return fmt.Errorf("failed to decode response: %v", err)
	}

	// Check if the response contains valid embedding data
	// If the response length does not match the input texts, log an error and return an error
	if len(response.Data) != len(texts) {
		LogError("Mismatch in OpenAI batch response length", fmt.Errorf("expected %d, got %d", len(texts), len(response.Data)))
		return fmt.Errorf("response length mismatch")
	}

	// Populate the embeddings slice with the embeddings from the response
	// This is the expected format from OpenAI's batch embedding API
	result := make([]pq.Float32Array, len(response.Data))
	for i, data := range response.Data {
		if len(data.Embedding) == 0 {
			return fmt.Errorf("empty embedding data for text %d", i)
		}
		// Convert the embedding data to pq.Float32Array
		// This is necessary to match the expected return type
		result[i] = pq.Float32Array(data.Embedding)
	}

	// Ensure the embeddings slice is populated with the results
	// This is the expected format from OpenAI's batch embedding API
	*embeddings = result
	return nil
}

// OpenAIChatService implements ChatService for OpenAI
type OpenAIChatService struct {
	config *Config
	apiKey string
	model  string
}

// OpenAI API structures for chat
type openAIChatRequest struct {
	Model       string              `json:"model"`
	Messages    []openAIChatMessage `json:"messages"`
	Temperature float32             `json:"temperature,omitempty"`
	MaxTokens   int                 `json:"max_tokens,omitempty"`
}

// openAIChatMessage represents a message in the OpenAI chat request
// It includes the role (system, user, assistant) and content of the message
type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openAIChatResponse represents the response structure from OpenAI for chat completion
// It includes the ID, model, choices, and usage statistics
// The choices contain the generated response from the model
type openAIChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// NewOpenAIChatService creates a new OpenAI chat service
// It initializes the service with the provided configuration
// This allows the service to use the OpenAI API for generating chat responses
func NewOpenAIChatService(config *Config) *OpenAIChatService {
	return &OpenAIChatService{
		config: config,
		apiKey: config.OpenAIAPIKey,
		model:  config.ChatModel,
	}
}

// GenerateResponse generates a response using OpenAI chat completion
func (s *OpenAIChatService) GenerateResponse(question, context string) (string, error) {
	// Rate limiting
	// Check if the rate limiter allows the request
	// If the rate limit is exceeded, log a warning and return an error
	if !OpenAIRateLimiter.Allow() {
		LogWarn("Rate limit exceeded for OpenAI chat completion")
		return "", fmt.Errorf("rate limit exceeded, please try again later")
	}

	var response string
	// Use a retry mechanism to handle transient errors
	// This allows the service to retry the request in case of temporary issues
	retryConfig := DefaultRetryConfig()

	// Retry the chat request with backoff
	// This helps to handle transient errors and ensures reliability
	err := RetryWithBackoff(retryConfig, func() error {
		// Make the actual request to OpenAI API
		// This function will handle the HTTP request and response parsing
		// It will populate the response variable with the result
		// *response is a pointer to string
		return s.makeChatRequest(question, context, &response)
	})

	if err != nil {
		LogError("Failed to generate OpenAI response after retries", err, "question", question)
		return "", err
	}

	LogInfo("Successfully generated OpenAI response", "question_length", len(question), "response_length", len(response))
	return response, nil
}

// GetProviderName returns the provider name
func (s *OpenAIChatService) GetProviderName() string {
	return "OpenAI"
}

// GetModel returns the model name
func (s *OpenAIChatService) GetModel() string {
	return s.model
}

// makeChatRequest makes a chat completion request to OpenAI
func (s *OpenAIChatService) makeChatRequest(question, context string, response *string) error {
	// Create system message with context
	systemMessage := fmt.Sprintf(`You are a helpful assistant that answers questions based on provided context. 
Use the following context to answer the user's question. If the context doesn't contain enough information to answer the question, say so clearly.

Context:
%s`, context)

	request := openAIChatRequest{
		Model: s.model,
		Messages: []openAIChatMessage{
			{Role: "system", Content: systemMessage},
			{Role: "user", Content: question},
		},
		Temperature: 0.1,
		MaxTokens:   2000, // Adjust max tokens as needed
	}

	// Marshal the request to JSON
	// This converts the request structure into JSON format for the API call
	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create the HTTP request to OpenAI chat completion
	// This includes the model, messages, temperature, and max tokens
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set the necessary headers for the request
	// This includes the content type and authorization header with the API key
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	// Make the HTTP request to OpenAI API
	// Do() executes the request and returns the response
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		LogError("Failed to make OpenAI chat request", err)
		return fmt.Errorf("failed to make request: %v", err)
	}

	// Ensure the response body is closed after reading
	// This prevents resource leaks and allows the response body to be read
	defer resp.Body.Close()

	// Check the response status code
	// If the status code is not 200 OK, log an error and return an error
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		LogError("OpenAI chat API error", fmt.Errorf("status: %s", resp.Status), "response_body", string(body))
		return fmt.Errorf("OpenAI API error: %s - %s", resp.Status, string(body))
	}

	// Decode the response body into the openAIChatResponse structure
	// This structure contains the generated response from OpenAI
	var chatResponse openAIChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResponse); err != nil {
		LogError("Failed to decode OpenAI chat response", err)
		return fmt.Errorf("failed to decode response: %v", err)
	}

	// Check if the response contains valid choices
	// If the choices are empty, log an error and return an error
	if len(chatResponse.Choices) == 0 {
		LogError("No response choices received from OpenAI", fmt.Errorf("empty choices"))
		return fmt.Errorf("no response choices received")
	}

	// Populate the response variable with the content of the first choice
	// This is the expected format from OpenAI's chat completion API
	// *response is a pointer to string
	*response = strings.TrimSpace(chatResponse.Choices[0].Message.Content)
	return nil
}
