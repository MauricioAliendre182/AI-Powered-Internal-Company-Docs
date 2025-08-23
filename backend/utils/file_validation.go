package utils

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
)

// AllowedFileTypes defines the supported file types
var AllowedFileTypes = map[string]bool{
	".txt":  true,
	".md":   true,
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".rtf":  true,
	".odt":  true,
}

// AllowedMimeTypes defines the supported MIME types
var AllowedMimeTypes = map[string]bool{
	"text/plain":         true,
	"text/markdown":      true,
	"application/pdf":    true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/rtf":                         true,
	"application/vnd.oasis.opendocument.text": true,
}

// ValidateFileType validates if the uploaded file type is allowed
func ValidateFileType(fileHeader *multipart.FileHeader) error {
	if fileHeader == nil {
		return fmt.Errorf("file header is nil")
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !AllowedFileTypes[ext] {
		return fmt.Errorf("file type '%s' is not supported. Allowed types: %v", ext, getAllowedExtensions())
	}

	// Check MIME type from header
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType != "" && !AllowedMimeTypes[contentType] {
		return fmt.Errorf("MIME type '%s' is not supported", contentType)
	}

	return nil
}

// getAllowedExtensions returns a slice of allowed file extensions
func getAllowedExtensions() []string {
	var extensions []string
	for ext := range AllowedFileTypes {
		extensions = append(extensions, ext)
	}
	return extensions
}

// GetFileExtension returns the file extension from filename
func GetFileExtension(filename string) string {
	return strings.ToLower(filepath.Ext(filename))
}

// IsValidFileSize checks if file size is within limits
func IsValidFileSize(size int64, maxSize int64) bool {
	return size > 0 && size <= maxSize
}
