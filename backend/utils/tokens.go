package utils

import (
	"fmt"

	tiktoken "github.com/pkoukk/tiktoken-go"
)

// tiktoken is used to count tokens in a string
// It uses OpenAI's encoding to determine the number of tokens in the input text
// CountTokens returns the number of tokens in the input string using OpenAI's encoding
func CountTokens(text, model string) (int, error) {
	// Get the encoding for the specified model
	// This function retrieves the encoding for the given model
	// EncodingForModel returns an encoding object that can be used to encode text
	// If the model is not supported, it returns an error
	enc, err := tiktoken.EncodingForModel(model)
	if err != nil {
		return 0, fmt.Errorf("failed to get encoding: %w", err)
	}

	// Encode the text using the encoding object
	// The Encode method encodes the input text and returns a slice of tokens
	tokens := enc.Encode(text, nil, nil)
	return len(tokens), nil
}
