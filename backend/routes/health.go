package routes

import (
	"net/http"
	"time"

	"github.com/MauricioAliendre182/backend/db"
	"github.com/MauricioAliendre182/backend/utils"
	"github.com/gin-gonic/gin"
)

// HealthStatus represents the health check response
type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Version   string            `json:"version"`
	Services  map[string]string `json:"services"`
}

// healthCheck handles health check requests
func healthCheck(c *gin.Context) {
	health := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0", // You might want to make this configurable
		Services:  make(map[string]string),
	}

	// Check database connectivity
	// db.DB is the database connection
	// Ping the database to check if it's available
	if err := db.DB.Ping(); err != nil {
		utils.LogError("Database health check failed", err)
		health.Status = "unhealthy"
		health.Services["database"] = "unhealthy"
	} else {
		health.Services["database"] = "healthy"
	}

	// Check AI service configuration
	factory := utils.NewAIServiceFactory(utils.AppConfig)
	currentProvider := factory.GetCurrentProvider()

	if err := factory.ValidateConfiguration(); err != nil {
		health.Services["ai_provider"] = string(currentProvider) + "_not_configured"
		health.Status = "degraded"
	} else {
		health.Services["ai_provider"] = string(currentProvider) + "_configured"
	}

	// Add additional provider-specific health checks
	// Use this section to add health checks for specific AI providers
	// For example, if using Ollama, check if the service is healthy
	switch currentProvider {
	case utils.OllamaProvider:
		if isOllamaHealthy() {
			health.Services["ollama"] = "healthy"
		} else {
			health.Services["ollama"] = "unhealthy"
			health.Status = "degraded"
		}
	case utils.GeminiProvider:
		health.Services["gemini"] = "configured"
	case utils.OpenAIProvider:
		health.Services["openai"] = "configured"
	}

	// Return appropriate status code
	if health.Status == "healthy" {
		c.JSON(http.StatusOK, health)
	} else if health.Status == "degraded" {
		c.JSON(http.StatusOK, health) // Still return 200 for degraded
	} else {
		c.JSON(http.StatusServiceUnavailable, health)
	}
}

// readinessCheck handles readiness probe requests
// This endpoint checks if the application is ready to handle requests
func readinessCheck(c *gin.Context) {
	// Check if database is ready
	if err := db.DB.Ping(); err != nil {
		utils.LogError("Database readiness check failed", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not_ready",
			"reason": "database_unavailable",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

// livenessCheck handles liveness probe requests
// This endpoint checks if the application is alive
func livenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}

// isOllamaHealthy checks if Ollama service is accessible
func isOllamaHealthy() bool {
	// Check if Ollama base URL is configured
	// This function checks if the Ollama service is healthy by making a simple request
	if utils.AppConfig.OllamaBaseURL == "" {
		return false
	}

	// Make a simple request to the Ollama service
	// This checks if the service is reachable and responding
	resp, err := http.Get(utils.AppConfig.OllamaBaseURL + "/api/tags")
	if err != nil {
		return false
	}

	// Check if the response status is OK (200)
	// This indicates that the Ollama service is up and running
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
