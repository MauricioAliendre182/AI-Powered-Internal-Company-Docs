package routes

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/MauricioAliendre182/backend/models"
	"github.com/MauricioAliendre182/backend/utils"
	"github.com/gin-gonic/gin"
)

// uploadDocument handles the document upload and processing
// This function is responsible for receiving the uploaded file, validating it,
// saving it as a document, processing the file into chunks, generating embeddings,
// and saving those chunks to the database.
// RAG is used in this function to process the document and generate embeddings for each chunk.
func uploadDocument(c *gin.Context) {
	utils.LogInfo("Starting document upload process")

	// Get the uploaded file
	fileHeader, err := c.FormFile("file")
	if err != nil {
		utils.LogError("Failed to get uploaded file", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	// Validate file type
	if err := utils.ValidateFileType(fileHeader); err != nil {
		utils.LogError("Invalid file type", err, "filename", fileHeader.Filename)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check file size (10MB limit)
	maxFileSize := utils.AppConfig.MaxFileSize
	if !utils.IsValidFileSize(fileHeader.Size, maxFileSize) {
		utils.LogError("File size exceeds limit", fmt.Errorf("file size: %d bytes", fileHeader.Size), "filename", fileHeader.Filename)
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds 10MB limit"})
		return
	}

	// Use transaction to ensure data consistency
	// func(tx *sql.Tx) error is a function type that takes a transaction and returns an error
	// This allows us to perform multiple database operations within a transaction
	err = utils.WithTransaction(func(tx *sql.Tx) error {
		// Create new document
		var doc models.Document
		if err := doc.ReadFromUpload(fileHeader); err != nil {
			return fmt.Errorf("failed to read from upload: %v", err)
		}

		// Validate document
		if err := doc.ValidateDocument(); err != nil {
			return fmt.Errorf("document validation failed: %v", err)
		}

		// Save document (you'll need to modify this to accept a transaction)
		if err := doc.SaveWithTx(tx); err != nil {
			return fmt.Errorf("failed to save document: %v", err)
		}

		// Process file into chunks
		chunkSize := utils.AppConfig.ChunkSize
		if chunkSize <= 0 {
			chunkSize = 1000 // Default fallback
		}

		chunks, err := models.ProcessFileToChunks(fileHeader, doc.ID, chunkSize)
		if err != nil {
			return fmt.Errorf("failed to process file into chunks: %v", err)
		}

		// Save chunks with embeddings
		for _, chunk := range chunks {
			if err := chunk.SaveWithTx(tx); err != nil {
				return fmt.Errorf("failed to save chunk: %v", err)
			}
		}

		// Prepare response
		response := models.DocumentResponse(doc)

		utils.LogInfo("Document uploaded successfully",
			"document_id", doc.ID.String(),
			"filename", doc.OriginalFilename,
			"chunks_created", len(chunks))

		c.JSON(http.StatusOK, gin.H{
			"message":        "Document uploaded successfully",
			"document":       response,
			"chunks_created": len(chunks),
		})

		return nil
	})

	if err != nil {
		utils.LogError("Failed to upload document", err, "filename", fileHeader.Filename)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}
