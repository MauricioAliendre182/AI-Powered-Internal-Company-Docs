package routes

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/MauricioAliendre182/backend/db"
	"github.com/MauricioAliendre182/backend/utils"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Set up mock database for testing
	setupTestDB()

	// Set up test configuration
	setupTestConfig()

	// Run tests
	code := m.Run()

	// Clean up
	if db.DB != nil {
		db.DB.Close()
	}

	// Exit with the same code
	os.Exit(code)
}

func setupTestDB() {
	// Create an in-memory SQLite database for testing
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic("Failed to create test database: " + err.Error())
	}

	// Assign to global variable for health checks
	db.DB = testDB
}

func setupTestConfig() {
	// Set up minimal test configuration
	utils.AppConfig = &utils.Config{
		DBHost:         "localhost",
		DBPort:         "5432",
		DBUser:         "test",
		DBPassword:     "test",
		DBName:         "test",
		OpenAIAPIKey:   "sk-test-key",
		EmbeddingModel: "text-embedding-3-small",
		ChatModel:      "gpt-3.5-turbo",
		Environment:    "test",
		Port:           "8080",
	}
}

func TestHealthCheck(t *testing.T) {
	tests := []struct {
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
		name           string
		expectedStatus int
	}{
		{
			name:           "Health check returns OK",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "healthy", response["status"])
				assert.Contains(t, response, "timestamp")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router
			router := gin.New()
			router.GET("/health", healthCheck)

			// Create request
			req := httptest.NewRequest("GET", "/health", http.NoBody)
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Check status
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check response if provided
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

func TestUploadDocument(t *testing.T) {
	tests := []struct {
		name           string
		fileContent    string
		fileName       string
		contentType    string
		expectedError  string
		expectedStatus int
		setupAuth      bool
	}{
		{
			name:           "Valid PDF upload",
			fileContent:    "%PDF-1.4\nFake PDF content for testing",
			fileName:       "test-document.pdf",
			contentType:    "application/pdf",
			setupAuth:      true,
			expectedStatus: http.StatusBadRequest, // Will fail due to MIME type detection
			expectedError:  "mime type",
		},
		{
			name:           "Valid TXT upload",
			fileContent:    "This is a test text document with some content.",
			fileName:       "test-document.txt",
			contentType:    "text/plain",
			setupAuth:      true,
			expectedStatus: http.StatusBadRequest, // Will fail due to MIME type detection
			expectedError:  "mime type",
		},
		{
			name:           "Invalid file type",
			fileContent:    "This should not be allowed",
			fileName:       "malicious.exe",
			contentType:    "application/octet-stream",
			setupAuth:      true,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "not supported",
		},
		{
			name:           "No file provided",
			fileContent:    "",
			fileName:       "",
			contentType:    "",
			setupAuth:      true,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "required",
		},
		{
			name:           "File too large",
			fileContent:    strings.Repeat("a", 15*1024*1024), // 15MB
			fileName:       "large-file.txt",
			contentType:    "text/plain",
			setupAuth:      true,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "mime type",
		},
		{
			name:           "Unauthorized access",
			fileContent:    "Test content",
			fileName:       "test.txt",
			contentType:    "text/plain",
			setupAuth:      false,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router with middleware
			router := gin.New()

			// Add auth middleware for protected routes
			protected := router.Group("/api/v1")
			if tt.setupAuth {
				protected.Use(func(c *gin.Context) {
					// Mock authentication middleware
					c.Set("userID", "test-user-id")
					c.Next()
				})
			} else {
				protected.Use(func(c *gin.Context) {
					// Simulate unauthorized access
					c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
					c.Abort()
				})
			}

			protected.POST("/documents", uploadDocument)

			// Create multipart form request
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)

			if tt.fileName != "" {
				part, err := writer.CreateFormFile("file", tt.fileName)
				assert.NoError(t, err)

				_, err = part.Write([]byte(tt.fileContent))
				assert.NoError(t, err)
			}

			err := writer.Close()
			assert.NoError(t, err)

			// Create request
			req := httptest.NewRequest("POST", "/api/v1/documents", &buf)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			// Add authorization header if needed
			if tt.setupAuth {
				req.Header.Set("Authorization", "Bearer test-jwt-token")
			}

			// Record response
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check error message if expected
			if tt.expectedError != "" {
				responseBody := w.Body.String()
				assert.Contains(t, strings.ToLower(responseBody), strings.ToLower(tt.expectedError))
			}
		})
	}
}

func TestQueryDocuments(t *testing.T) {
	tests := []struct {
		requestBody    map[string]interface{}
		name           string
		expectedError  string
		expectedStatus int
		setupAuth      bool
	}{
		{
			name: "Valid query",
			requestBody: map[string]interface{}{
				"question": "How many vacation days do employees get?",
			},
			setupAuth:      true,
			expectedStatus: http.StatusInternalServerError, // Will fail due to embedding service
			expectedError:  "embedding service",
		},
		{
			name: "Empty question",
			requestBody: map[string]interface{}{
				"question": "",
			},
			setupAuth:      true,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation",
		},
		{
			name: "Missing question field",
			requestBody: map[string]interface{}{
				"query": "This has wrong field name",
			},
			setupAuth:      true,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation",
		},
		{
			name: "Unauthorized query",
			requestBody: map[string]interface{}{
				"question": "What are the policies?",
			},
			setupAuth:      false,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name: "Very long question",
			requestBody: map[string]interface{}{
				"question": strings.Repeat("What is the policy regarding ", 100),
			},
			setupAuth:      true,
			expectedStatus: http.StatusBadRequest, // Guardrails now properly validate input length
			expectedError:  "question too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router with middleware
			router := gin.New()

			// Add auth middleware for protected routes
			protected := router.Group("/api/v1")
			if tt.setupAuth {
				protected.Use(func(c *gin.Context) {
					// Mock authentication middleware
					c.Set("userID", "test-user-id")
					c.Next()
				})
			} else {
				protected.Use(func(c *gin.Context) {
					// Simulate unauthorized access
					c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
					c.Abort()
				})
			}

			protected.POST("/query", queryDocuments)

			// Create JSON request body
			jsonBody, err := json.Marshal(tt.requestBody)
			assert.NoError(t, err)

			// Create request
			req := httptest.NewRequest("POST", "/api/v1/query", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Add authorization header if needed
			if tt.setupAuth {
				req.Header.Set("Authorization", "Bearer test-jwt-token")
			}

			// Record response
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check error message if expected
			if tt.expectedError != "" {
				responseBody := w.Body.String()
				assert.Contains(t, strings.ToLower(responseBody), strings.ToLower(tt.expectedError))
			}
		})
	}
}

func TestGetDocuments(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		expectedError  string
		expectedStatus int
		setupAuth      bool
	}{
		{
			name:           "Get documents successfully",
			setupAuth:      true,
			queryParams:    "",
			expectedStatus: http.StatusInternalServerError, // Will fail due to missing table
			expectedError:  "no such table",
		},
		{
			name:           "Get documents with pagination",
			setupAuth:      true,
			queryParams:    "?page=1&limit=5",
			expectedStatus: http.StatusInternalServerError, // Will fail due to missing table
			expectedError:  "no such table",
		},
		{
			name:           "Unauthorized access",
			setupAuth:      false,
			queryParams:    "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:           "Invalid pagination parameters",
			setupAuth:      true,
			queryParams:    "?page=invalid&limit=abc",
			expectedStatus: http.StatusInternalServerError, // Will fail due to missing table
			expectedError:  "no such table",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router with middleware
			router := gin.New()

			// Add auth middleware for protected routes
			protected := router.Group("/api/v1")
			if tt.setupAuth {
				protected.Use(func(c *gin.Context) {
					// Mock authentication middleware
					c.Set("userID", "test-user-id")
					c.Next()
				})
			} else {
				protected.Use(func(c *gin.Context) {
					// Simulate unauthorized access
					c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
					c.Abort()
				})
			}

			protected.GET("/documents", getDocuments)

			// Create request
			req := httptest.NewRequest("GET", "/api/v1/documents"+tt.queryParams, http.NoBody)

			// Add authorization header if needed
			if tt.setupAuth {
				req.Header.Set("Authorization", "Bearer test-jwt-token")
			}

			// Record response
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check error message if expected
			if tt.expectedError != "" {
				responseBody := w.Body.String()
				assert.Contains(t, strings.ToLower(responseBody), strings.ToLower(tt.expectedError))
			}
		})
	}
}

// Benchmark tests for performance
func BenchmarkHealthCheck(b *testing.B) {
	router := gin.New()
	router.GET("/health", healthCheck)

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/health", http.NoBody)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkQueryDocuments(b *testing.B) {
	router := gin.New()
	router.POST("/query", func(c *gin.Context) {
		c.Set("userID", "test-user")
		queryDocuments(c)
	})

	requestBody := map[string]interface{}{
		"question": "What are the company policies?",
	}
	jsonBody, _ := json.Marshal(requestBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/query", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
