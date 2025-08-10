package utils

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/ledongthuc/pdf"
)

// ExtractTextFromPDF extracts readable text from a PDF file
func ExtractTextFromPDF(reader io.ReadSeeker) (string, error) {
	// Get the size of the file
	size, err := reader.Seek(0, io.SeekEnd)
	if err != nil {
		return "", fmt.Errorf("failed to seek to end: %w", err)
	}

	// Reset to beginning
	_, err = reader.Seek(0, io.SeekStart)
	if err != nil {
		return "", fmt.Errorf("failed to seek to start: %w", err)
	}

	// Read the entire file into memory
	content := make([]byte, size)
	_, err = io.ReadFull(reader, content)
	if err != nil {
		return "", fmt.Errorf("failed to read content: %w", err)
	}

	// Create a ReaderAt from the byte slice
	readerAt := bytes.NewReader(content)

	// Open the PDF
	pdfReader, err := pdf.NewReader(readerAt, size)
	if err != nil {
		return "", fmt.Errorf("failed to create PDF reader: %w", err)
	}

	var textBuilder strings.Builder

	// Extract text from each page
	// NumPage returns the number of pages in the PDF
	// Page returns the page object for the given page number
	numPages := pdfReader.NumPage()
	for pageNum := 1; pageNum <= numPages; pageNum++ {
		page := pdfReader.Page(pageNum)
		if page.V.IsNull() {
			continue
		}

		// Extract text from the page
		// The GetPlainText method extracts text from the page
		// fonts map[string]*pdf.Font) argument is optional
		// If you pass nil, it will use the default fonts
		// A example of fonts values would be:
		// fonts := map[string]*pdf.Font{
		// 	"F1": pdf.NewFont("Helvetica", 12),
		// 	"F2": pdf.NewFont("Times-Roman", 12),
		// }
		pageText, err := page.GetPlainText(nil) // Pass nil as font map to use default fonts
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
