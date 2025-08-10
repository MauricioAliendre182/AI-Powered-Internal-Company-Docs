package middlewares

import (
	"fmt"
	"net/http"

	"github.com/MauricioAliendre182/backend/utils"
	"github.com/gin-gonic/gin"
)

// ErrorHandler is a middleware that handles panics and errors
// gin.HandlerFunc is a function that takes a gin context and recovers from panics
// It logs the error and returns a JSON response with a 500 status code
// If a request handler panics, this middleware will catch it
// and log the error details, including the request path, method, and client IP
func ErrorHandler() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			utils.LogError("Panic recovered",
				fmt.Errorf("panic: %s", err),
				"path", c.Request.URL.Path,
				"method", c.Request.Method,
				"ip", c.ClientIP())
		}

		// c.AbortWithStatusJSON refers to aborting the current request and sending a JSON response
		// with a 500 status code
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "An unexpected error occurred",
		})
	})
}

// RequestLogger logs incoming requests
// This is a middleware that logs details of each request
// It uses gin.LoggerWithFormatter to format the log output
func RequestLogger() gin.HandlerFunc {
	// gin.LogFormatterParams is a struct that contains parameters for logging
	// It includes method, path, status code, latency, client IP, and request object
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		utils.LogInfo("Request processed",
			"method", param.Method,
			"path", param.Path,
			"status", param.StatusCode,
			"latency", param.Latency,
			"ip", param.ClientIP,
			"user_agent", param.Request.UserAgent(),
		)
		return ""
	})
}
