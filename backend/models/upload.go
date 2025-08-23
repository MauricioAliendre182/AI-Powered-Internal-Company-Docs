package models

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/MauricioAliendre182/backend/db"
	"github.com/MauricioAliendre182/backend/utils"
	"github.com/google/uuid"
)

// Document represents a document in the documents table
type Document struct {
	UploadedAt       time.Time `json:"uploaded_at"`
	Name             string    `json:"name"`
	OriginalFilename string    `json:"original_filename"`
	ID               uuid.UUID `json:"id"`
}

// Chunk represents a chunk in the chunks table
type Chunk struct {
	ContentType string       `json:"content_type"`
	Content     string       `json:"content"`
	Embedding   utils.Vector `json:"embedding"`
	Size        int64        `json:"size"`
	ChunkIndex  int          `json:"chunk_index"`
	ID          uuid.UUID    `json:"id"`
	DocumentID  uuid.UUID    `json:"document_id"`
}

// DocumentResponse for API responses
type DocumentResponse struct {
	UploadedAt       time.Time `json:"uploaded_at"`
	Name             string    `json:"name"`
	OriginalFilename string    `json:"original_filename"`
	ID               uuid.UUID `json:"id"`
}

// ReadFromUpload reads the uploaded file and populates the Document struct
func (d *Document) ReadFromUpload(fileHeader *multipart.FileHeader) error {
	// Set file properties
	d.OriginalFilename = fileHeader.Filename
	d.Name = d.generateFileName()
	d.UploadedAt = time.Now()

	return nil
}

// generateFileName generates a unique filename for the uploaded file
func (d *Document) generateFileName() string {
	// filepath.Ext returns the file extension
	// We use the current timestamp to ensure uniqueness
	ext := filepath.Ext(d.OriginalFilename)
	timestamp := time.Now().Unix()
	return fmt.Sprintf("doc_%d_%s%s", timestamp, d.OriginalFilename[:len(d.OriginalFilename)-len(ext)], ext)
}

// Save saves the document to the database
func (d *Document) Save() error {
	query := `
	INSERT INTO documents (id, name, original_filename, uploaded_at)
	VALUES ($1, $2, $3, $4)
	RETURNING id
	`

	// Generate a new UUID for the document ID
	if d.ID == uuid.Nil {
		// If ID is not set, generate a new UUID
		// This ensures that each document has a unique ID
		d.ID = uuid.New()
	}

	// Prepare the SQL statement
	// Using a prepared statement to prevent SQL injection
	stmt, err := db.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Set the uploaded at time to the current time
	// This is the time when the document was uploaded
	d.UploadedAt = time.Now()
	err = stmt.QueryRow(d.ID, d.Name, d.OriginalFilename, d.UploadedAt).Scan(&d.ID)
	if err != nil {
		return err
	}

	return nil
}

// SaveWithTx saves the document to the database using a transaction
// a transaction allows for atomic operations, ensuring that either all changes are committed or none are applied
func (d *Document) SaveWithTx(tx *sql.Tx) error {
	query := `
	INSERT INTO documents (id, name, original_filename, uploaded_at)
	VALUES ($1, $2, $3, $4)
	RETURNING id
	`

	// Generate a new UUID for the document ID
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}

	// Set the uploaded at time to the current time
	d.UploadedAt = time.Now()

	// Use the transaction instead of the global DB
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	err = stmt.QueryRow(d.ID, d.Name, d.OriginalFilename, d.UploadedAt).Scan(&d.ID)
	if err != nil {
		return err
	}

	return nil
}

// GetDocumentByID retrieves a document by ID
func GetDocumentByID(id uuid.UUID) (Document, error) {
	var doc Document
	query := `
	SELECT id, name, original_filename, uploaded_at
	FROM documents
	WHERE id = $1
	`

	stmt, err := db.DB.Prepare(query)
	if err != nil {
		return doc, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(id).Scan(&doc.ID, &doc.Name, &doc.OriginalFilename, &doc.UploadedAt)
	if err != nil {
		return doc, err
	}

	return doc, nil
}

// GetAllDocuments retrieves all documents from the database
func GetAllDocuments() ([]Document, error) {
	var documents []Document
	query := `
	SELECT id, name, original_filename, uploaded_at
	FROM documents
	ORDER BY uploaded_at DESC
	`

	stmt, err := db.DB.Prepare(query)
	if err != nil {
		return documents, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return documents, err
	}
	defer rows.Close()

	for rows.Next() {
		var doc Document
		err = rows.Scan(&doc.ID, &doc.Name, &doc.OriginalFilename, &doc.UploadedAt)
		if err != nil {
			return documents, err
		}
		documents = append(documents, doc)
	}

	return documents, nil
}

// Delete removes a document from the database
func DeleteDocument(documentID uuid.UUID) error {
	query := `DELETE FROM documents WHERE id = $1`

	stmt, err := db.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Exec executes the statement with the provided ID
	// If the ID is not set, it will return an error
	_, err = stmt.Exec(documentID)
	if err != nil {
		return err
	}

	return nil
}

// ValidateDocument validates the document before saving
func (d *Document) ValidateDocument() error {
	// Check if file name is not empty
	if len(d.OriginalFilename) == 0 {
		return errors.New("file name cannot be empty")
	}

	// Check if name is not empty
	if len(d.Name) == 0 {
		return errors.New("document name cannot be empty")
	}

	return nil
}

// Save saves the chunk to the database
func (c *Chunk) Save() error {
	// Validate chunk before saving
	if err := c.ValidateChunk(); err != nil {
		return fmt.Errorf("chunk validation failed: %v", err)
	}

	query := `
	INSERT INTO chunks (id, document_id, size, content_type, content, embedding, chunk_index)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id
	`

	// Prepare the SQL statement
	// Using a prepared statement to prevent SQL injection
	stmt, err := db.DB.Prepare(query)
	if err != nil {
		return err
	}

	// Close the statement after execution to free resources
	defer stmt.Close()

	// QueryRow executes the statement and returns a single row
	// If the ID is not set, generate a new UUID
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	err = stmt.QueryRow(c.ID, c.DocumentID, c.Size, c.ContentType, c.Content, c.Embedding, c.ChunkIndex).Scan(&c.ID)
	if err != nil {
		return err
	}

	return nil
}

// SaveWithTx saves the chunk to the database using a transaction
// This allows for atomic operations, ensuring that either all changes are committed or none
// are applied
// tx *sql.Tx is a transaction in the database
func (c *Chunk) SaveWithTx(tx *sql.Tx) error {
	// Validate chunk before saving
	if err := c.ValidateChunk(); err != nil {
		return fmt.Errorf("chunk validation failed: %v", err)
	}

	query := `
	INSERT INTO chunks (id, document_id, size, content_type, content, embedding, chunk_index)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id
	`

	// QueryRow executes the statement and returns a single row
	// If the ID is not set, generate a new UUID
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}

	// Use the transaction instead of the global DB
	// This ensures that the chunk is saved within the context of the transaction
	// This is useful for ensuring data consistency, especially when saving multiple chunks
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	err = stmt.QueryRow(c.ID, c.DocumentID, c.Size, c.ContentType, c.Content, c.Embedding, c.ChunkIndex).Scan(&c.ID)
	if err != nil {
		return err
	}

	return nil
}

// GetChunksByDocumentID retrieves all chunks for a specific document
func GetChunksByDocumentID(documentID uuid.UUID) ([]Chunk, error) {
	var chunks []Chunk
	query := `
	SELECT id, document_id, size, content_type, content, embedding, chunk_index
	FROM chunks
	WHERE document_id = $1
	ORDER BY chunk_index
	`

	stmt, err := db.DB.Prepare(query)
	if err != nil {
		return chunks, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(documentID)
	if err != nil {
		return chunks, err
	}
	defer rows.Close()

	for rows.Next() {
		var chunk Chunk
		err = rows.Scan(&chunk.ID, &chunk.DocumentID, &chunk.Size, &chunk.ContentType, &chunk.Content, &chunk.Embedding, &chunk.ChunkIndex)
		if err != nil {
			return chunks, err
		}
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// GetChunkByID retrieves a chunk by ID
func GetChunkByID(id uuid.UUID) (Chunk, error) {
	var chunk Chunk
	query := `
	SELECT id, document_id, size, content_type, content, embedding, chunk_index
	FROM chunks
	WHERE id = $1
	`

	stmt, err := db.DB.Prepare(query)
	if err != nil {
		return chunk, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(id).Scan(&chunk.ID, &chunk.DocumentID, &chunk.Size, &chunk.ContentType, &chunk.Content, &chunk.Embedding, &chunk.ChunkIndex)
	if err != nil {
		return chunk, err
	}

	return chunk, nil
}

// Delete removes a chunk from the database
func (c *Chunk) DeleteChunk(documentID uuid.UUID) error {
	query := `DELETE FROM chunks WHERE id = $1 AND document_id = $2`

	stmt, err := db.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(c.ID, documentID)
	if err != nil {
		return err
	}

	return nil
}

// ValidateChunk validates the chunk before saving
func (c *Chunk) ValidateChunk() error {
	// Check if content is not empty
	if len(c.Content) == 0 {
		return errors.New("chunk content cannot be empty")
	}

	// Check if document ID is valid
	if c.DocumentID == uuid.Nil {
		return errors.New("valid document ID is required")
	}

	// Check if chunk index is valid
	if c.ChunkIndex < 0 {
		return errors.New("chunk index must be non-negative")
	}

	return nil
}

// ProcessFileToChunks processes an uploaded file and creates chunks
// *multipart.FileHeader is used to handle file uploads in web applications
// It contains metadata about the uploaded file, such as its name, size, and content type
func ProcessFileToChunks(fileHeader *multipart.FileHeader, documentID uuid.UUID, chunkSize int64) ([]Chunk, error) {
	// Validate inputs
	if fileHeader == nil {
		return nil, fmt.Errorf("fileHeader cannot be nil")
	}
	if documentID == uuid.Nil {
		return nil, fmt.Errorf("documentID cannot be nil")
	}
	if chunkSize <= 0 {
		return nil, fmt.Errorf("chunkSize must be positive")
	}

	// Check file size to prevent memory issues
	maxFileSize := int64(50 * 1024 * 1024) // 50MB limit
	if fileHeader.Size > maxFileSize {
		return nil, fmt.Errorf("file size %d exceeds maximum allowed size of %d bytes", fileHeader.Size, maxFileSize)
	}

	// Open and read the file
	// fileHeader is a *multipart.FileHeader, which contains metadata about the uploaded file
	// fileHeader.Open() returns an io.ReadCloser, which we can use to read the file content
	opened, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer opened.Close()

	// io.ReadAll reads the entire content of the file into memory
	// This is suitable for small files. For larger files, consider streaming or processing in chunks
	contentBytes, err := io.ReadAll(opened)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Convert to string and sanitize UTF-8 to prevent database encoding errors
	var content string

	// For PDF files, we need proper text extraction
	if filepath.Ext(fileHeader.Filename) == ".pdf" {
		// Extract text from PDF using proper PDF parsing
		extractedText, err := utils.ExtractTextFromPDFBytes(contentBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to extract text from PDF: %w", err)
		}
		content = utils.SanitizeUTF8(extractedText)
	} else {
		// For other file types, treat as plain text
		content = utils.SanitizeUTF8(string(contentBytes))
	}
	contentType := fileHeader.Header.Get("Content-Type")

	var chunksList []Chunk

	// Split content into chunks
	// utils.SplitText is a utility function that splits the text into smaller chunks
	// Here, we assume it takes the content and the maximum size of each chunk
	chunks := utils.SplitIntoChunks(content, chunkSize)

	// Get embeddings for all chunks
	// utils.GetBatchEmbeddings is a utility function that takes a slice of strings and returns
	var chunkTexts []string
	chunkTexts = append(chunkTexts, chunks...)

	embeddings, err := utils.GetBatchEmbeddings(chunkTexts)
	if err != nil {
		return nil, fmt.Errorf("failed to get embeddings: %v", err)
	}

	// For each chunk, create a Chunk struct and append it to the chunksList
	// Each chunk will have a unique ID, the document ID it belongs to, its size
	for i, chunkText := range chunks {
		// Sanitize chunk text to ensure valid UTF-8
		sanitizedChunk := utils.SanitizeUTF8(chunkText)

		chunk := Chunk{
			ID:          uuid.New(),
			DocumentID:  documentID,
			Size:        int64(len(sanitizedChunk)),
			ContentType: contentType,
			Content:     sanitizedChunk,
			Embedding:   embeddings[i],
			ChunkIndex:  i,
		}

		chunksList = append(chunksList, chunk)
	}
	return chunksList, nil
}

// SimilaritySearch performs vector similarity search to find relevant chunks
// this function uses the pgvector extension for efficient vector operations
// It takes a query embedding and returns the most similar chunks
// The queryEmbedding is a Vector, which is a slice of float32 values representing the embedding vector
// The limit parameter specifies the maximum number of results to return
func SimilaritySearch(queryEmbedding utils.Vector, limit int) ([]Chunk, error) {
	var chunks []Chunk

	utils.LogInfo("Starting similarity search", "embedding_length", len(queryEmbedding), "limit", limit)

	// This query retrieves chunks ordered by their similarity to the query embedding
	// The <=> operator is used for vector similarity search in pgvector
	// It returns the closest chunks based on the embedding distance
	query := `
	SELECT id, document_id, size, content_type, content, embedding, chunk_index,
		   (embedding <=> $1) as distance
	FROM chunks
	ORDER BY distance DESC
	-- LIMIT $2 limits the number of results returned
	LIMIT $2
	`

	// Prepare the SQL statement
	// Using a prepared statement to prevent SQL injection
	stmt, err := db.DB.Prepare(query)
	if err != nil {
		utils.LogError("Failed to prepare similarity search query", err)
		return chunks, err
	}
	defer stmt.Close()

	// Query() executes the statement with the provided queryEmbedding and limit
	// It returns a *sql.Rows, which we can iterate over to get the results
	rows, err := stmt.Query(queryEmbedding, limit)
	if err != nil {
		utils.LogError("Failed to execute similarity search query", err)
		return chunks, err
	}
	defer rows.Close()

	// Iterate over the rows returned by the query
	// rows.Next() moves to the next row in the result set
	for rows.Next() {
		var chunk Chunk
		var distance float32
		// Scan() reads the values from the current row into the chunk struct
		// The order of the arguments must match the order of the columns in the SELECT statement
		// distance is also scanned to get the similarity score
		// unpack the values into the chunk struct
		err = rows.Scan(&chunk.ID, &chunk.DocumentID, &chunk.Size, &chunk.ContentType,
			&chunk.Content, &chunk.Embedding, &chunk.ChunkIndex, &distance)
		if err != nil {
			utils.LogError("Failed to scan chunk row", err)
			return chunks, err
		}
		utils.LogInfo("Found chunk", "chunk_id", chunk.ID.String(), "distance", distance, "content_preview", func() string {
			if len(chunk.Content) > 100 {
				return chunk.Content[:100] + "..."
			}
			return chunk.Content
		}())
		chunks = append(chunks, chunk)
	}

	utils.LogInfo("Similarity search completed", "total_chunks_found", len(chunks))
	return chunks, nil
}

// GetRelevantChunks finds chunks relevant to a query using embeddings
func GetRelevantChunks(queryText string, limit int) ([]Chunk, error) {
	// Get embedding for the query text
	embedding, err := utils.GetEmbedding(queryText)
	if err != nil {
		return nil, fmt.Errorf("failed to get embedding: %v", err)
	}

	// Perform similarity search using the embedding
	return SimilaritySearch(embedding, limit)
}
