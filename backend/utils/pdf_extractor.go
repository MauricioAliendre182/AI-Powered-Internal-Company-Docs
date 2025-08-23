package utils

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
)

// ExtractTextFromPDF extracts readable text from a PDF file
func ExtractTextFromPDF(reader io.ReadSeeker) (string, error) {
	// Reset to beginning
	// Seek is to the start of the reader
	// The parameters are (offset int64, whence int) where whence is 0 for SeekStart
	_, err := reader.Seek(0, io.SeekStart)
	if err != nil {
		return "", fmt.Errorf("failed to seek to start: %w", err)
	}

	// Open the PDF using unipdf
	// NewPdfReader creates a new PDF reader
	pdfReader, err := model.NewPdfReader(reader)
	if err != nil {
		return "", fmt.Errorf("failed to create PDF reader: %w", err)
	}

	var textBuilder strings.Builder

	// Get number of pages
	// GetNumPages returns the number of pages in the PDF
	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return "", fmt.Errorf("failed to get number of pages: %w", err)
	}

	// Extract text from each page
	for pageNum := 1; pageNum <= numPages; pageNum++ {
		// GetPage returns the page at the specified index
		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			LogWarn("Failed to get page", "page", pageNum, "error", err)
			continue
		}

		// Create text extractor for the page
		// New creates a new text extractor for the given page
		ex, err := extractor.New(page)
		if err != nil {
			LogWarn("Failed to create extractor for page", "page", pageNum, "error", err)
			continue
		}

		// Extract text from the page
		// ExtractText extracts the text content from the page
		pageText, err := ex.ExtractText()
		if err != nil {
			LogWarn("Failed to extract text from page", "page", pageNum, "error", err)
			continue
		}

		// Add page text to builder
		textBuilder.WriteString(pageText)
		textBuilder.WriteString("\n\n") // Add page separator
	}

	extractedText := textBuilder.String()

	// Clean up the extracted text
	extractedText = strings.TrimSpace(extractedText)

	// Remove excessive whitespace
	extractedText = strings.ReplaceAll(extractedText, "\r\n", "\n")
	extractedText = strings.ReplaceAll(extractedText, "\r", "\n")

	// Replace multiple consecutive newlines with double newlines
	for strings.Contains(extractedText, "\n\n\n") {
		extractedText = strings.ReplaceAll(extractedText, "\n\n\n", "\n\n")
	}

	LogInfo("PDF text extraction completed", "pages", numPages, "text_length", len(extractedText))

	return extractedText, nil
}

// ExtractTextFromPDFBytes extracts text from PDF bytes
func ExtractTextFromPDFBytes(data []byte) (string, error) {
	reader := bytes.NewReader(data)
	return ExtractTextFromPDF(reader)
}
