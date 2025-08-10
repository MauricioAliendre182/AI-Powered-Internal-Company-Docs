package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MauricioAliendre182/backend/db"
	"github.com/MauricioAliendre182/backend/routes"
	"github.com/MauricioAliendre182/backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Initialize structured logging
	utils.InitLogger()

	// Load environment variables at application startup
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file:", err)
		utils.LogWarn("Error loading .env file", "error", err)
		// You can decide whether to continue or exit based on your requirements
		// In production environments, you might set environment variables differently
	}

	// Initialize configuration
	// This function reads the configuration from environment variables or a config file
	// It should handle errors gracefully and log them
	if err := utils.InitConfig(); err != nil {
		utils.LogError("Failed to initialize configuration", err)
		log.Fatalf("Configuration error: %v", err)
	}

	// Initialize rate limiter with config values
	utils.InitRateLimiter()

	// Initialize the database
	db.InitDB(
		utils.AppConfig.DBHost,
		utils.AppConfig.DBPort,
		utils.AppConfig.DBUser,
		utils.AppConfig.DBPassword,
		utils.AppConfig.DBName,
	)
	utils.LogInfo("Database initialized successfully")

	// Initialize AI services
	// This function sets up the AI service factory and creates the embedding service
	// It should validate the configuration and log any errors
	if err := utils.InitEmbeddingService(); err != nil {
		utils.LogError("Failed to initialize embedding service", err)
		log.Fatalf("AI service error: %v", err)
	}
	utils.LogInfo("AI services initialized successfully")

	// Set Gin mode based on environment
	if utils.AppConfig.Environment == "production" {
		// gin.SetMode refers to setting the mode of the Gin framework
		// It determines how Gin behaves in different environments
		gin.SetMode(gin.ReleaseMode)
	}

	// HTTP Server for us
	server := gin.Default()

	// Register the routes
	routes.RegisterRoutes(server)

	// Create HTTP server
	// This server will handle incoming HTTP requests
	// It includes read and write timeouts to prevent slow clients from hogging resources
	// Addr: specifies the port on which the server will listen
	// Handler: specifies the router that will handle incoming requests
	// ReadTimeout: specifies the maximum duration for reading the entire request, including the body
	// WriteTimeout: specifies the maximum duration before timing out writes of the response
	// IdleTimeout: specifies the maximum duration for keeping idle connections open
	srv := &http.Server{
		Addr:         ":" + utils.AppConfig.Port,
		Handler:      server,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	// go is a keyword that allows us to run a function asynchronously
	// This means the server will run in the background, allowing the main function to continue executing
	go func() {
		utils.LogInfo("Starting server", "port", utils.AppConfig.Port, "environment", utils.AppConfig.Environment)
		// srv.ListenAndServe() starts the HTTP server
		// It listens on the specified address and port, and handles incoming requests
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utils.LogError("Failed to start server", err)
			log.Fatalf("Server error: %v", err)
		}
	}()

	// IMPORTANT: A channel is a way to communicate between goroutines
	// 			  It allows us to send and receive messages between different
	//            parts of the program
	// 				Example of a channel:
	// 						ch := make(chan string)
	// 						<- refers to receiving a message from the channel
	// 						signal.Notify is used to listen for OS signals
	// Wait for interrupt signal to gracefully shutdown the server
	// make(chan os.Signal, 1) creates a channel to receive OS signals
	// signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) registers the
	// quit channel to receive notifications of specified signals
	// syscall.SIGINT is the interrupt signal (Ctrl+C)
	// syscall.SIGTERM is the termination signal (kill command)
	// This allows the application to handle graceful shutdowns
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	utils.LogInfo("Server is shutting down...")

	// Give outstanding requests 30 seconds to complete
	// context.WithTimeout creates a new context that will be canceled after the specified duration
	// This is useful for gracefully shutting down the server
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	// Ensure the cancel function is called to release resources
	// defer cancel() ensures that the context is canceled when the function returns
	defer cancel()

	// Shutdown the server gracefully
	// srv.Shutdown(ctx) attempts to gracefully shut down the server by waiting for
	// outstanding requests to complete within the specified timeout
	// If there are any errors during shutdown, they will be logged
	if err := srv.Shutdown(ctx); err != nil {
		utils.LogError("Server forced to shutdown", err)
		log.Fatalf("Server shutdown error: %v", err)
	}

	utils.LogInfo("Server exited")
}
