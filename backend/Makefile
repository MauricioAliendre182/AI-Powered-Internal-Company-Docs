# AI-Powered Internal Company Docs - Backend Makefile
# This Makefile provides common development commands

.PHONY: help test test-verbose test-coverage lint lint-install security security-install fmt vet build clean run deps-install deps-verify deps-tidy docker-build docker-run

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Testing
test: ## Run all tests
	go test -race -timeout=2m ./...

test-verbose: ## Run tests with verbose output
	go test -v -race -timeout=2m ./...

test-coverage: ## Run tests with coverage report
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-individual: ## Run tests for each package individually
	@echo "Testing main package..."
	go test -v -race -timeout=2m .
	@echo "Testing models package..."
	go test -v -race -timeout=2m ./models
	@echo "Testing utils package..."
	go test -v -race -timeout=2m ./utils
	@echo "Testing routes package..."
	go test -v -race -timeout=2m ./routes

# Linting and Code Quality
lint-install: ## Install golangci-lint
	@echo "Installing golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "golangci-lint is already installed"; \
	else \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi

lint: lint-install ## Run golangci-lint
	golangci-lint run --timeout=5m

lint-fix: lint-install ## Run golangci-lint with auto-fix
	golangci-lint run --fix --timeout=5m

# Security Scanning
security-install: ## Install gosec security scanner
	@echo "Installing gosec..."
	@if command -v gosec >/dev/null 2>&1; then \
		echo "gosec is already installed"; \
	else \
		echo "Installing gosec..."; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
	fi

security: security-install ## Run gosec security scanner
	gosec ./...

security-report: security-install ## Run gosec and generate reports
	gosec -fmt sarif -out gosec-results.sarif ./...
	gosec -fmt json -out gosec-results.json ./...
	gosec -fmt html -out gosec-results.html ./...
	@echo "Security reports generated:"
	@echo "  - gosec-results.sarif"
	@echo "  - gosec-results.json"
	@echo "  - gosec-results.html"

# Code Formatting
fmt: ## Format Go code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

# Build and Run
build: ## Build the application
	go build -v -ldflags="-s -w" -o main .

build-debug: ## Build with debug information
	go build -v -o main .

run: ## Run the application
	go run .

# Dependencies
deps-install: ## Download dependencies
	go mod download

deps-verify: ## Verify dependencies
	go mod verify

deps-tidy: ## Tidy dependencies
	go mod tidy

deps-update: ## Update dependencies
	go get -u ./...
	go mod tidy

# Docker
docker-build: ## Build Docker image
	docker build -t ai-docs-backend .

docker-run: ## Run Docker container
	docker run -p 8080:8080 ai-docs-backend

# Cleanup
clean: ## Clean build artifacts and test outputs
	rm -f main
	rm -f coverage.out coverage.html
	rm -f gosec-results.*
	go clean -cache
	go clean -testcache

# Development workflow
dev-setup: deps-install lint-install security-install ## Set up development environment
	@echo "Development environment setup complete!"
	@echo "Available commands:"
	@make help

dev-check: fmt vet lint security test ## Run all checks (format, vet, lint, security, test)
	@echo "✅ All development checks passed!"

ci-check: deps-verify fmt vet lint security test-coverage ## Run CI-like checks locally
	@echo "✅ All CI checks passed!"

# Guardrails specific
test-guardrails: ## Test guardrails functionality specifically
	go test -v -run ".*[Gg]uardrail.*" ./utils

# Quick commands
quick-test: ## Quick test without race detection
	go test ./...

quick-lint: ## Quick lint without installing
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --fast; \
	else \
		echo "golangci-lint not installed. Run 'make lint-install' first."; \
	fi

# PromptFoo AI Testing Commands
promptfoo-install: ## Install PromptFoo for AI testing
	@echo "Installing PromptFoo..."
	@if command -v npm >/dev/null 2>&1; then \
		npm install -g promptfoo; \
		echo "PromptFoo installed successfully"; \
	else \
		echo "❌ npm is not installed. Please install Node.js first."; \
		exit 1; \
	fi

promptfoo-basic: promptfoo-install ## Run basic RAG functionality tests
	@echo "Running basic PromptFoo tests..."
	@if [ -f "promptfoo-config.yaml" ]; then \
		promptfoo eval --config promptfoo-config.yaml; \
	else \
		echo "❌ promptfoo-config.yaml not found"; \
		exit 1; \
	fi

promptfoo-report: ## View PromptFoo test results
	@echo "Opening PromptFoo results..."
	@if command -v promptfoo >/dev/null 2>&1; then \
		promptfoo view; \
	else \
		echo "❌ PromptFoo not installed. Run 'make promptfoo-install' first."; \
	fi

promptfoo-clean: ## Clean PromptFoo test results
	@echo "Cleaning PromptFoo results..."
	@rm -rf promptfoo-results/ .promptfoo/
	@echo "✅ PromptFoo results cleaned"

promptfoo-test: ## Run PromptFoo integration tests (Go)
	@echo "Running PromptFoo integration tests..."
	go test -v -timeout=10m -run TestPromptFoo ./...

# Help for environment setup
env-help: ## Show environment setup help
	@echo "Environment Variables for Local Development:"
	@echo ""
	@echo "Required:"
	@echo "  JWT_SECRET=your-secret-key"
	@echo "  DB_HOST=localhost"
	@echo "  DB_PORT=5432"
	@echo "  DB_USER=postgres"
	@echo "  DB_PASSWORD=your-password"
	@echo "  DB_NAME=your-database"
	@echo ""
	@echo "Optional AI Provider Keys (choose one):"
	@echo "  OPENAI_API_KEY=sk-proj-your-openai-key"
	@echo "  GOOGLE_AI_API_KEY=AIzaSyC-your-google-ai-key"
	@echo "  USE_LOCAL_AI=true (for Ollama)"
	@echo ""
	@echo "Other:"
	@echo "  ENVIRONMENT=development"
	@echo "  ALLOWED_ORIGINS=http://localhost:4200"
	@echo ""
	@echo "Create a .env file in the backend directory with these variables."
