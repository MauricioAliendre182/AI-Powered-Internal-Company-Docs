package routes

import (
	"net/http"

	"github.com/MauricioAliendre182/backend/models"
	"github.com/MauricioAliendre182/backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// uploadDocumentWithRAG handles document upload and RAG processing
// func uploadDocumentWithRAG(c *gin.Context) {
// 	// Get the uploaded file
// 	fileHeader, err := c.FormFile("file")
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
// 		return
// 	}

// 	// Create new document
// 	var doc models.Document
// 	err = doc.ReadFromUpload(fileHeader)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Validate document
// 	err = doc.ValidateDocument()
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Save document
// 	err = doc.Save()
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Read file content for processing
// 	opened, err := fileHeader.Open()
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
// 		return
// 	}
// 	defer opened.Close()

// 	contentBytes, err := io.ReadAll(opened)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
// 		return
// 	}

// 	// Process document with RAG
// 	ragService := models.NewRAGService()
// 	err = ragService.ProcessAndStoreDocument(&doc, string(contentBytes), 1000) // 1000 char chunks
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process document: " + err.Error()})
// 		return
// 	}

// 	response := models.DocumentResponse{
// 		ID:               doc.ID,
// 		Name:             doc.Name,
// 		OriginalFilename: doc.OriginalFilename,
// 		UploadedAt:       doc.UploadedAt,
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"message":  "Document uploaded and processed successfully",
// 		"document": response,
// 	})
// }

// queryDocuments handles RAG queries with security guardrails
func queryDocuments(c *gin.Context) {
	// Get query from request
	type QueryRequest struct {
		Question string `json:"question" binding:"required"`
	}

	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Sanitize the question
	sanitizedQuestion := utils.SanitizeQuestion(req.Question)

	// Validate question with guardrails
	violations := utils.ValidateQuestion(sanitizedQuestion, utils.DefaultGuardrailConfig())

	// Check for error-level violations
	for _, violation := range violations {
		if violation.Severity == "error" {
			// Log the violation for security monitoring
			utils.LogGuardrailViolation(violation, getUserID(c), sanitizedQuestion)

			c.JSON(http.StatusBadRequest, gin.H{
				"error":       violation.Message,
				"type":        violation.Type,
				"suggestions": violation.Suggestions,
			})
			return
		}
	}

	// Log warning-level violations but continue processing
	for _, violation := range violations {
		if violation.Severity == "warning" {
			utils.LogGuardrailViolation(violation, getUserID(c), sanitizedQuestion)
		}
	}

	// Perform RAG query
	ragService, err := models.NewRAGService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize RAG service: " + err.Error()})
		return
	}

	answer, err := ragService.QueryDocuments(sanitizedQuestion)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Validate the response
	responseViolations := utils.ValidateResponse(answer)
	if len(responseViolations) > 0 {
		utils.LogWarn("Response validation violations detected",
			"user_id", getUserID(c),
			"question", sanitizedQuestion,
			"violations", len(responseViolations),
		)
	}

	c.JSON(http.StatusOK, gin.H{
		"question": sanitizedQuestion,
		"answer":   answer,
		"warnings": getWarnings(violations),
	})
}

// getUserID extracts user ID from context, returns "anonymous" if not found
func getUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return "anonymous"
}

// getWarnings extracts warning messages from violations
func getWarnings(violations []utils.GuardrailViolation) []string {
	var warnings []string
	for _, violation := range violations {
		if violation.Severity == "warning" {
			warnings = append(warnings, violation.Message)
		}
	}
	return warnings
}

// getDocuments returns all documents
func getDocuments(c *gin.Context) {
	documents, err := models.GetAllDocuments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"documents": documents,
	})
}

func deleteDocument(c *gin.Context) {
	documentID := c.Param("id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document ID is required"})
		return
	}

	// Parse UUID
	docUUID, err := uuid.Parse(documentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	// Delete the chunk
	var chunk models.Chunk
	err = chunk.DeleteChunk(docUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Delete the document
	err = models.DeleteDocument(docUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Document deleted successfully",
	})
}

// getDocumentChunks returns chunks for a specific document
func getDocumentChunks(c *gin.Context) {
	documentID := c.Param("id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document ID is required"})
		return
	}

	// Parse UUID
	docUUID, err := uuid.Parse(documentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	// Get chunks for the document
	// models.GetChunksByDocumentID is a function that retrieves chunks from the database
	// It should return a slice of chunks and an error
	chunks, err := models.GetChunksByDocumentID(docUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"document_id": documentID,
		"chunks":      chunks,
	})
}

// getGuardrailStatus returns the current guardrail configuration
func getGuardrailStatus(c *gin.Context) {
	status := utils.GetGuardrailStatus()
	c.JSON(http.StatusOK, gin.H{
		"status": status,
	})
}
