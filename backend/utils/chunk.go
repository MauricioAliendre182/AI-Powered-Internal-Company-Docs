package utils

import (
	"strings"
	"unicode/utf8"
)

// SanitizeUTF8 removes invalid UTF-8 sequences and ensures clean text for database storage
func SanitizeUTF8(text string) string {
	// First, remove null bytes and other control characters that cause issues
	text = strings.ReplaceAll(text, "\x00", "")
	text = strings.ReplaceAll(text, "\ufeff", "") // BOM

	// Remove other problematic bytes
	var cleanText strings.Builder
	for _, r := range text {
		// Keep printable characters, whitespace, and common punctuation
		if utf8.ValidRune(r) && (r >= 32 || r == '\n' || r == '\r' || r == '\t') {
			cleanText.WriteRune(r)
		}
	}

	// Convert to valid UTF-8 by replacing any remaining invalid sequences
	result := strings.ToValidUTF8(cleanText.String(), "")

	// Additional cleanup - remove excessive whitespace
	result = strings.TrimSpace(result)

	return result
}

// splitIntoChunks splits text into chunks of specified size
func SplitIntoChunks(text string, chunkSize int64) []string {
	var chunks []string
	words := strings.Fields(text)

	if len(words) == 0 {
		return chunks
	}

	var currentChunk strings.Builder
	wordCount := 0

	for _, word := range words {
		if wordCount > 0 && currentChunk.Len()+len(word)+1 > int(chunkSize) {
			chunks = append(chunks, currentChunk.String())
			currentChunk.Reset()
			wordCount = 0
		}

		if currentChunk.Len() > 0 {
			currentChunk.WriteString(" ")
		}
		currentChunk.WriteString(word)
		wordCount++
	}

	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	return chunks
}
