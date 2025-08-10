package routes

import (
	"time"

	"github.com/MauricioAliendre182/backend/middlewares"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(server *gin.Engine) {
	// Global middleware
	// It handles errors and logs requests
	server.Use(middlewares.ErrorHandler())
	server.Use(middlewares.RequestLogger())

	// Configure cors
	server.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:4200", "http://localhost"}, // or your frontend domain
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check endpoints (no authentication required)
	server.GET("/health", healthCheck)
	server.GET("/readiness", readinessCheck)
	server.GET("/liveness", livenessCheck)

	// API versioning
	v1 := server.Group("/api/v1")

	// Authenticated routes
	authenticated := v1.Group("")
	authenticated.Use(middlewares.Authenticate)

	// Non-authenticated routes
	nonAuthenticated := v1.Group("")

	// Authentication routes
	auth := nonAuthenticated.Group("/auth")
	{
		auth.POST("/signup", signup)
		auth.POST("/login", login)
		auth.POST("/refresh-token", refreshToken)
		auth.POST("/is-available", isAvalable)
		auth.POST("/forgot-password", forgotPassword)
		auth.GET("/verify-reset-token/:token", verifyResetToken)
		auth.POST("/reset-password", resetPassword)
	}

	// Profile routes (authenticated)
	profile := authenticated.Group("/auth")
	{
		profile.GET("/profile", getOwnProfile)
	}

	// Alternative profile endpoint
	authenticated.GET("/me/profile", getOwnProfile)

	// Document routes (authenticated)
	docs := authenticated.Group("/documents")
	{
		docs.POST("", uploadDocument)
		docs.GET("", getDocuments)
		docs.GET("/:id/chunks", getDocumentChunks)
		docs.DELETE("/:id", deleteDocument)
	}

	// RAG query endpoint (authenticated)
	authenticated.POST("/query", queryDocuments)
}
