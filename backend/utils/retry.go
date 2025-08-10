package utils

import (
	"math"
	"math/rand"
	"time"
)

// RetryConfig defines configuration for retry logic
type RetryConfig struct {
	MaxRetries        int
	InitialDelay      time.Duration
	MaxDelay          time.Duration
	BackoffMultiplier float64
	Jitter            bool
}

// DefaultRetryConfig returns a sensible default retry configuration
// MaxRetries: maximum number of retries
// InitialDelay: initial delay before the first retry
// MaxDelay: maximum delay between retries
// BackoffMultiplier: multiplier for exponential backoff
// Jitter: whether to add jitter to the delay
// Jitter helps to prevent thundering herd problems by adding a small random delay
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:        3,
		InitialDelay:      1 * time.Second,
		MaxDelay:          30 * time.Second,
		BackoffMultiplier: 2.0,
		Jitter:            true,
	}
}

// RetryWithBackoff executes a function with exponential backoff retry logic
// Retry logic for OpenAI API calls or other operations that may fail
// config: RetryConfig defines the retry parameters
// fn: function to execute, which returns an error if it fails
// This function will retry the execution of fn up to MaxRetries times
// It will wait for InitialDelay before the first retry, and then apply exponential backoff
func RetryWithBackoff(config RetryConfig, fn func() error) error {
	var lastErr error

	// for loop to handle retries
	// It will attempt to execute the function up to MaxRetries times
	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate delay with exponential backoff
			// backoff starts with InitialDelay and increases exponentially
			// based on BackoffMultiplier for each subsequent attempt
			// backoff refers to the strategy of waiting longer between retries
			// to avoid overwhelming the service and to give it time to recover
			delay := time.Duration(float64(config.InitialDelay) * math.Pow(config.BackoffMultiplier, float64(attempt-1)))

			// Apply maximum delay cap
			// delay works as a cap to prevent excessive waiting
			// This ensures that the delay does not exceed MaxDelay
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}

			// Add jitter to prevent thundering herd
			// jitter refers to a small random delay added to the backoff
			// This helps to spread out the retries across multiple clients
			if config.Jitter {
				jitter := time.Duration(rand.Float64() * float64(delay) * 0.1)
				delay += jitter
			}

			time.Sleep(delay)
		}

		// Execute the function
		// If it returns an error, we will retry
		// err := fn() is the function that we are trying to execute
		// If it returns an error, we will retry
		if err := fn(); err != nil {
			lastErr = err
			continue
		}

		// Success
		return nil
	}

	return lastErr
}
