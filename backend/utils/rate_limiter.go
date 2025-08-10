package utils

import (
	"sync"
	"time"
)

// RateLimiter implements a simple token bucket rate limiter
type RateLimiter struct {
	tokens     int64
	maxTokens  int64
	refillRate int64
	lastRefill time.Time
	mutex      sync.Mutex
}

// NewRateLimiter creates a new rate limiter
// maxTokens: maximum number of tokens in the bucket
// refillRate: number of tokens to add per second
func NewRateLimiter(maxTokens, refillRate int64) *RateLimiter {
	return &RateLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if an operation is allowed (consumes one token)
func (r *RateLimiter) Allow() bool {
	// r.mutex.Lock() is to ensure thread safety
	// This prevents concurrent access issues
	// when multiple goroutines try to access the rate limiter at the same time
	r.mutex.Lock()

	// defer r.mutex.Unlock()
	// This ensures that the mutex is unlocked when the function exits,
	// preventing deadlocks and allowing other goroutines to proceed
	// It is a common practice to use defer for unlocking in Go
	defer r.mutex.Unlock()

	// Refill tokens based on time passed
	// Calculate how many seconds have passed since the last refill
	// and add tokens accordingly
	now := time.Now()
	timePassed := now.Sub(r.lastRefill).Seconds()
	tokensToAdd := int64(timePassed * float64(r.refillRate))

	// If tokensToAdd is negative, it means we are trying to add tokens in the past
	// This can happen if the system clock is adjusted backwards
	if tokensToAdd > 0 {
		r.tokens += tokensToAdd
		if r.tokens > r.maxTokens {
			r.tokens = r.maxTokens
		}
		r.lastRefill = now
	}

	// Check if we have tokens available
	// If tokens are available, consume one token and return true
	// This allows the operation to proceed
	if r.tokens > 0 {
		r.tokens--
		return true
	}

	return false
}

// GetTokens returns the current number of tokens
func (r *RateLimiter) GetTokens() int64 {
	// r.mutex.Lock() is to ensure thread safety
	// This prevents concurrent access issues
	// when multiple goroutines try to access the rate limiter at the same time
	r.mutex.Lock()

	// defer r.mutex.Unlock()
	// This ensures that the mutex is unlocked when the function exits,
	// preventing deadlocks and allowing other goroutines to proceed
	// It is a common practice to use defer for unlocking in Go
	defer r.mutex.Unlock()
	return r.tokens
}

// Global rate limiter for OpenAI API calls
// Initialize with default values, will be updated when config is loaded
var OpenAIRateLimiter = NewRateLimiter(10, 1) // Default: 10 tokens, refill 1 per second

// InitRateLimiter initializes the rate limiter with config values
// This should be called after AppConfig is loaded
func InitRateLimiter() {
	if AppConfig != nil {
		OpenAIRateLimiter = NewRateLimiter(AppConfig.RateLimitMaxTokens, AppConfig.RateLimitRefillRate)
		LogInfo("Rate limiter initialized", "max_tokens", AppConfig.RateLimitMaxTokens, "refill_rate", AppConfig.RateLimitRefillRate)
	}
}
