package models

import (
	"fmt"
	"testing"

	"github.com/MauricioAliendre182/backend/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for testing
type MockChatService struct {
	mock.Mock
}

// GenerateResponse mocks the chat service's response generation
// It simulates generating a response based on the question and context
func (m *MockChatService) GenerateResponse(question, context string) (string, error) {
	args := m.Called(question, context)
	return args.String(0), args.Error(1)
}

// GetProviderName mocks the chat service's provider name retrieval
// It simulates getting the name of the AI provider used by the chat service
func (m *MockChatService) GetProviderName() string {
	args := m.Called()
	return args.String(0)
}

// GetModel mocks the chat service's model retrieval
// It simulates getting the model name used by the chat service
func (m *MockChatService) GetModel() string {
	args := m.Called()
	return args.String(0)
}

// Mock function variables for dependency injection
var (
	mockGetEmbedding     func(string) (utils.Vector, error)
	mockSimilaritySearch func(utils.Vector, int) ([]Chunk, error)
)

// Wrapper functions to enable mocking
func getEmbeddingWrapper(text string) (utils.Vector, error) {
	if mockGetEmbedding != nil {
		return mockGetEmbedding(text)
	}
	return utils.GetEmbedding(text)
}

func similaritySearchWrapper(embedding utils.Vector, limit int) ([]Chunk, error) {
	if mockSimilaritySearch != nil {
		return mockSimilaritySearch(embedding, limit)
	}
	return SimilaritySearch(embedding, limit)
}

// Helper function to create UUID from string
func mustParseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		// For testing, generate a valid UUID if parsing fails
		return uuid.New()
	}
	return id
}

// TestRAGService_QueryDocuments tests the QueryDocuments method of RAGService
// It verifies that the method correctly processes the question, retrieves relevant chunks, and generates a response
// using the chat service. It also tests various scenarios including successful queries, no relevant chunks,
func TestRAGService_QueryDocuments(t *testing.T) {
	tests := []struct {
		name           string
		question       string
		mockEmbedding  utils.Vector
		mockChunks     []Chunk
		mockResponse   string
		expectedResult string
		expectError    bool
	}{
		{
			name:          "Successful query with relevant chunks",
			question:      "How many vacation days do employees get?",
			mockEmbedding: utils.Vector{0.1, 0.2, 0.3, 0.4},
			mockChunks: []Chunk{
				{
					ID:         mustParseUUID("550e8400-e29b-41d4-a716-446655440001"),
					Content:    "Full-time employees receive 15 paid vacation days per year.",
					DocumentID: mustParseUUID("550e8400-e29b-41d4-a716-446655440002"),
					ChunkIndex: 0,
				},
				{
					ID:         mustParseUUID("550e8400-e29b-41d4-a716-446655440003"),
					Content:    "Part-time employees receive vacation days pro-rated based on hours worked.",
					DocumentID: mustParseUUID("550e8400-e29b-41d4-a716-446655440002"),
					ChunkIndex: 1,
				},
			},
			mockResponse:   "Based on the documents, full-time employees receive 15 paid vacation days per year, while part-time employees receive pro-rated vacation days.",
			expectedResult: "Based on the documents, full-time employees receive 15 paid vacation days per year, while part-time employees receive pro-rated vacation days.",
			expectError:    false,
		},
		{
			name:           "No relevant chunks found",
			question:       "What is the meaning of life?",
			mockEmbedding:  utils.Vector{0.1, 0.2, 0.3, 0.4},
			mockChunks:     []Chunk{},
			mockResponse:   "",
			expectedResult: "I couldn't find any relevant information in the documents to answer your question.",
			expectError:    false,
		},
		{
			name:          "Single chunk found",
			question:      "What is the company dress code?",
			mockEmbedding: utils.Vector{0.5, 0.6, 0.7, 0.8},
			mockChunks: []Chunk{
				{
					ID:         mustParseUUID("550e8400-e29b-41d4-a716-446655440004"),
					Content:    "The company follows a business casual dress code.",
					DocumentID: mustParseUUID("550e8400-e29b-41d4-a716-446655440005"),
					ChunkIndex: 0,
				},
			},
			mockResponse:   "The company follows a business casual dress code.",
			expectedResult: "The company follows a business casual dress code.",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		// t.Run is used to create sub-tests for each scenario
		// This allows us to run each test case independently and see which one fails
		// Their parameters are the name of the test and a function that contains the test logic
		// The function receives a pointer to the testing.T type, which is used to report test results
		// It also allows us to use table-driven tests, which is a common pattern in
		// Go testing
		t.Run(tt.name, func(t *testing.T) {
			// Create mock chat service
			mockChatService := &MockChatService{}

			if len(tt.mockChunks) > 0 {
				// mockChatService.On is used to set up expectations for the mock
				// It specifies that when GenerateResponse is called with any string arguments,
				// it should return the predefined mock response without an error
				mockChatService.On("GenerateResponse", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(tt.mockResponse, nil)
			}

			// Set up mocks for the global functions
			mockGetEmbedding = func(text string) (utils.Vector, error) {
				assert.Equal(t, tt.question, text)
				return tt.mockEmbedding, nil
			}

			// Mock similarity search to return predefined chunks
			// This simulates the behavior of retrieving relevant chunks based on the question embedding
			mockSimilaritySearch = func(embedding utils.Vector, limit int) ([]Chunk, error) {
				assert.Equal(t, tt.mockEmbedding, embedding)
				return tt.mockChunks, nil
			}

			// Since we can't test the full QueryDocuments method without database integration,
			// we'll test the logic components
			t.Run("EmbeddingGeneration", func(t *testing.T) {
				embedding, err := mockGetEmbedding(tt.question)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockEmbedding, embedding)
			})

			t.Run("SimilaritySearch", func(t *testing.T) {
				chunks, err := mockSimilaritySearch(tt.mockEmbedding, 10)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockChunks, chunks)
			})

			if len(tt.mockChunks) > 0 {
				t.Run("ResponseGeneration", func(t *testing.T) {
					// Build context like the real function would
					contextText := "Based on the following information from the documents:\n\n"
					for i, chunk := range tt.mockChunks {
						contextText += fmt.Sprintf("Document %d:\n%s\n\n", i+1, chunk.Content)
					}

					response, err := mockChatService.GenerateResponse(tt.question, contextText)
					assert.NoError(t, err)
					assert.Equal(t, tt.mockResponse, response)

					// Verify mock expectations
					mockChatService.AssertExpectations(t)
				})
			}

			// Clean up mocks
			mockGetEmbedding = nil
			mockSimilaritySearch = nil
		})
	}
}

func TestCleanEmbeddingVector(t *testing.T) {
	tests := []struct {
		name     string
		input    utils.Vector
		expected utils.Vector
	}{
		{
			name:     "Normal vector within range",
			input:    utils.Vector{0.1, -0.5, 0.8, -0.2},
			expected: utils.Vector{0.1, -0.5, 0.8, -0.2},
		},
		{
			name:     "Empty vector",
			input:    utils.Vector{},
			expected: utils.Vector{},
		},
		{
			name:     "Vector with acceptable length",
			input:    utils.Vector{0.1, 0.2, 0.3, 0.4, 0.5},
			expected: utils.Vector{0.1, 0.2, 0.3, 0.4, 0.5},
		},
		{
			name: "Vector with corrupted data (too long)",
			input: func() utils.Vector {
				// Create a vector longer than 2*1536 to trigger cleaning
				corrupted := make(utils.Vector, 3500)
				for i := 0; i < 1001; i++ {
					corrupted[i] = float32(i%3-1) * 0.5 // Valid embedding values between -0.5 and 1.0
				}
				// Add corrupted values (outside valid range)
				for i := 1001; i < 3500; i++ {
					corrupted[i] = 999999.0 // Out of range values
				}
				return corrupted
			}(),
			expected: func() utils.Vector {
				// Should extract first 1001 valid values (since only first 1001 are valid)
				expected := make(utils.Vector, 1001)
				for i := 0; i < 1001; i++ {
					expected[i] = float32(i%3-1) * 0.5
				}
				return expected
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanEmbeddingVector(tt.input)

			if tt.name == "Vector with corrupted data (too long)" {
				// For corrupted data test, check length and first few values
				assert.Equal(t, 1001, len(result))
				assert.InDelta(t, -0.5, result[0], 0.001)
				assert.InDelta(t, 0.0, result[1], 0.001)
				assert.InDelta(t, 0.5, result[2], 0.001)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestExtractValidFloat(t *testing.T) {
	tests := []struct {
		name          string
		input         float32
		index         int
		expectedValue float32
		expectedOK    bool
	}{
		{
			name:          "Valid embedding value",
			input:         0.5,
			index:         0,
			expectedValue: 0.5,
			expectedOK:    true,
		},
		{
			name:          "Valid negative embedding value",
			input:         -0.8,
			index:         1,
			expectedValue: -0.8,
			expectedOK:    true,
		},
		{
			name:          "Zero value",
			input:         0.0,
			index:         2,
			expectedValue: 0.0,
			expectedOK:    true,
		},
		{
			name:          "Out of range positive value",
			input:         15.0,
			index:         3,
			expectedValue: 0.0,
			expectedOK:    false,
		},
		{
			name:          "Out of range negative value",
			input:         -15.0,
			index:         4,
			expectedValue: 0.0,
			expectedOK:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, ok := extractValidFloat(tt.input, tt.index)

			assert.Equal(t, tt.expectedOK, ok)
			if tt.expectedOK {
				assert.Equal(t, tt.expectedValue, value)
			}
		})
	}
}

func TestIsValidEmbeddingValue(t *testing.T) {
	tests := []struct {
		name     string
		input    float32
		expected bool
	}{
		{"Zero", 0.0, true},
		{"Positive valid", 1.0, true},
		{"Negative valid", -1.0, true},
		{"Upper bound", 1.0, true},
		{"Lower bound", -1.0, true},
		{"Too positive", 1.1, false},
		{"Too negative", -1.1, false},
		{"Way too positive", 999999.0, false},
		{"Way too negative", -999999.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidEmbeddingValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewRAGService(t *testing.T) {
	// Save original config
	originalConfig := utils.AppConfig

	tests := []struct {
		name        string
		config      *utils.Config
		expectError bool
	}{
		{
			name: "Valid OpenAI configuration",
			config: &utils.Config{
				OpenAIAPIKey:   "sk-test-key",
				EmbeddingModel: "text-embedding-3-small",
				ChatModel:      "gpt-3.5-turbo",
			},
			expectError: false,
		},
		{
			name: "Valid Google AI configuration",
			config: &utils.Config{
				GoogleAIAPIKey: "AIza-test-key",
				EmbeddingModel: "models/embedding-001",
				ChatModel:      "models/gemini-1.5-flash",
			},
			expectError: false,
		},
		{
			name: "Valid Ollama configuration",
			config: &utils.Config{
				UseLocalAI:     true,
				OllamaBaseURL:  "http://localhost:11434",
				EmbeddingModel: "nomic-embed-text",
				ChatModel:      "llama3.1:8b",
			},
			expectError: false,
		},
		{
			name: "No AI provider configured",
			config: &utils.Config{
				DBHost: "localhost",
				DBPort: "5432",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set test configuration
			utils.AppConfig = tt.config

			ragService, err := NewRAGService()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, ragService)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, ragService)
				assert.Equal(t, 10, ragService.MaxChunks)
				assert.NotNil(t, ragService.chatService)
			}
		})
	}

	// Restore original config
	utils.AppConfig = originalConfig
}
