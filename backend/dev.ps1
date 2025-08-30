# Windows PowerShell Script for Development Tasks
# Usage: .\dev.ps1 [command]
# Example: .\dev.ps1 lint

param(
    [Parameter(Position=0)]
    [string]$Command = "help"
)

# Set error action preference
$ErrorActionPreference = "Stop"

# Colors for output
function Write-Success { param($Message) Write-Host $Message -ForegroundColor Green }
function Write-Error { param($Message) Write-Host $Message -ForegroundColor Red }
function Write-Warning { param($Message) Write-Host $Message -ForegroundColor Yellow }
function Write-Info { param($Message) Write-Host $Message -ForegroundColor Cyan }

# PromptFoo AI Testing Functions
function Install-PromptFoo {
    Write-Info "Installing PromptFoo for AI testing..."
    
    # Check if npm is available
    try {
        $npmVersion = npm --version
        Write-Info "Found npm version: $npmVersion"
    }
    catch {
        Write-Error "npm is not installed. Please install Node.js first."
        Write-Info "Download Node.js from: https://nodejs.org/"
        return $false
    }
    
    # Check if PromptFoo is already installed
    try {
        $promptfooVersion = promptfoo --version
        Write-Success "PromptFoo is already installed: $promptfooVersion"
        return $true
    }
    catch {
        Write-Info "PromptFoo not found, installing..."
    }
    
    # Install PromptFoo globally
    try {
        Write-Info "Installing PromptFoo globally..."
        npm install -g promptfoo
        
        # Verify installation
        $promptfooVersion = promptfoo --version
        Write-Success "PromptFoo installed successfully: $promptfooVersion"
        return $true
    }
    catch {
        Write-Error "Failed to install PromptFoo: $_"
        return $false
    }
}

function Test-PromptFooEnvironment {
    Write-Info "Checking PromptFoo environment..."
    
    # Check for configuration file
    if (-not (Test-Path "promptfoo-config.yaml")) {
        Write-Error "PromptFoo configuration file not found: promptfoo-config.yaml"
        return $false
    }
    
    # Check for test data directory
    if (-not (Test-Path "test-data")) {
        Write-Error "Test data directory not found: test-data"
        return $false
    }
    
    # Check API keys (at least one should be available)
    $hasOpenAI = $env:OPENAI_API_KEY -ne $null -and $env:OPENAI_API_KEY -ne ""
    $hasGoogleAI = $env:GOOGLE_AI_API_KEY -ne $null -and $env:GOOGLE_AI_API_KEY -ne ""
    $hasOllama = Test-OllamaConnection
    
    if (-not ($hasOpenAI -or $hasGoogleAI -or $hasOllama)) {
        Write-Warning "No AI provider API keys found or Ollama not running"
        Write-Info "Set OPENAI_API_KEY, GOOGLE_AI_API_KEY, or start Ollama service"
        Write-Info "Tests may fail without proper AI provider configuration"
    }
    else {
        Write-Success "AI provider environment validated"
        if ($hasOpenAI) { Write-Info "[OK] OpenAI API key found" }
        if ($hasGoogleAI) { Write-Info "[OK] Google AI API key found" }
        if ($hasOllama) { Write-Info "[OK] Ollama service available" }
    }
    
    return $true
}

function Test-OllamaConnection {
    try {
        # Test if Ollama service is running
        $response = Invoke-RestMethod -Uri "http://localhost:11434/api/version" -TimeoutSec 3 -ErrorAction SilentlyContinue
        return $response -ne $null
    }
    catch {
        return $false
    }
}

function Invoke-PromptFooBasic {
    Write-Info "Running basic PromptFoo AI tests..."
    
    if (-not (Install-PromptFoo)) {
        Write-Error "Failed to install PromptFoo"
        return
    }
    
    if (-not (Test-PromptFooEnvironment)) {
        Write-Error "PromptFoo environment validation failed"
        return
    }
    
    try {
        Write-Info "Executing PromptFoo evaluation..."
        promptfoo eval --config promptfoo-config.yaml
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "PromptFoo tests completed successfully"
        }
        else {
            Write-Warning "PromptFoo tests completed with warnings (exit code: $LASTEXITCODE)"
        }
    }
    catch {
        Write-Error "Failed to run PromptFoo tests: $_"
    }
}

function Invoke-PromptFooReport {
    Write-Info "Opening PromptFoo test results..."
    
    if (-not (Test-Path "promptfoo-results")) {
        Write-Warning "No PromptFoo results found. Run 'promptfoo-basic' first."
        return
    }
    
    try {
        promptfoo view
    }
    catch {
        Write-Error "Failed to open PromptFoo results: $_"
        Write-Info "Try opening the results directory manually: promptfoo-results/"
    }
}

function Clear-PromptFooResults {
    Write-Info "Cleaning PromptFoo test results..."
    
    $resultDirs = @("promptfoo-results", ".promptfoo")
    foreach ($dir in $resultDirs) {
        if (Test-Path $dir) {
            try {
                Remove-Item $dir -Recurse -Force
                Write-Success "Removed: $dir"
            }
            catch {
                Write-Warning "Failed to remove $dir`: $($_.Exception.Message)"
            }
        }
    }
    
    Write-Success "PromptFoo results cleaned"
}

function Invoke-PromptFooTest {
    Write-Info "Running PromptFoo integration tests (Go)..."
    
    try {
        # Run PromptFoo tests with specific build tag
        go test -v -timeout=10m -tags="promptfoo" -run "TestPromptFoo" ./...
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "PromptFoo integration tests passed"
        }
        else {
            Write-Warning "PromptFoo integration tests completed with issues"
        }
    }
    catch {
        Write-Error "Failed to run PromptFoo integration tests: $_"
    }
}

# Check if Go is installed
function Test-GoInstallation {
    try {
        $goVersion = go version
        Write-Success "Go is installed: $goVersion"
        return $true
    }
    catch {
        Write-Error "Go is not installed or not in PATH"
        return $false
    }
}

# Install development tools
function Install-DevTools {
    Write-Info "Installing development tools..."
    
    Write-Info "Installing golangci-lint..."
    try {
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
        Write-Success "golangci-lint installed successfully"
    }
    catch {
        Write-Error "Failed to install golangci-lint: $_"
    }
    
    Write-Info "Installing gosec..."
    try {
        go install github.com/securego/gosec/v2/cmd/gosec@latest
        Write-Success "gosec installed successfully"
    }
    catch {
        Write-Error "Failed to install gosec: $_"
    }
    
    Write-Success "Development tools installation complete!"
}

# Run tests
function Invoke-Tests {
    param([switch]$Verbose, [switch]$Coverage, [switch]$Individual)
    
    Write-Info "Running tests (excluding PromptFoo integration tests)..."
    
    if ($Individual) {
        Write-Info "Testing main package..."
        go test -v -race -timeout=2m -tags="!promptfoo" .
        
        Write-Info "Testing models package..."
        go test -v -race -timeout=2m -tags="!promptfoo" ./models
        
        Write-Info "Testing utils package..."
        go test -v -race -timeout=2m -tags="!promptfoo" ./utils
        
        Write-Info "Testing routes package..."
        go test -v -race -timeout=2m -tags="!promptfoo" ./routes
    }
    elseif ($Coverage) {
        Write-Info "Running tests with coverage (excluding PromptFoo)..."
        go test -race -coverprofile=coverage.out -covermode=atomic -tags="!promptfoo" ./...
        go tool cover -html=coverage.out -o coverage.html
        Write-Success "Coverage report generated: coverage.html"
    }
    elseif ($Verbose) {
        go test -v -race -timeout=2m -tags="!promptfoo" ./...
    }
    else {
        go test -race -timeout=2m -tags="!promptfoo" ./...
    }
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "All tests passed!"
    } else {
        Write-Error "Some tests failed!"
        exit 1
    }
}

# Run linting
function Invoke-Lint {
    param([switch]$Fix)
    
    Write-Info "Running linter..."
    
    # Check if golangci-lint is installed
    try {
        golangci-lint version | Out-Null
    }
    catch {
        Write-Warning "golangci-lint not found. Installing..."
        Install-DevTools
    }
    
    if ($Fix) {
        golangci-lint run --fix --timeout=5m
    } else {
        golangci-lint run --timeout=5m
    }
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "Linting completed successfully!"
    } else {
        Write-Warning "Linting found issues. Run 'dev.ps1 lint-fix' to auto-fix some issues."
    }
}

# Run security scan
function Invoke-Security {
    param([switch]$Report)
    
    Write-Info "Running security scan..."
    
    # Check if gosec is installed
    try {
        gosec -version | Out-Null
    }
    catch {
        Write-Warning "gosec not found. Installing..."
        Install-DevTools
    }
    
    if ($Report) {
        Write-Info "Generating security reports..."
        gosec -fmt sarif -out gosec-results.sarif ./...
        gosec -fmt json -out gosec-results.json ./...
        gosec -fmt html -out gosec-results.html ./...
        Write-Success "Security reports generated:"
        Write-Info "  - gosec-results.sarif"
        Write-Info "  - gosec-results.json"
        Write-Info "  - gosec-results.html"
    } else {
        gosec ./...
    }
}

# Format code
function Invoke-Format {
    Write-Info "Formatting Go code..."
    go fmt ./...
    Write-Success "Code formatted!"
}

# Run go vet
function Invoke-Vet {
    Write-Info "Running go vet..."
    go vet ./...
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "go vet completed successfully!"
    } else {
        Write-Error "go vet found issues!"
        exit 1
    }
}

# Build application
function Invoke-Build {
    param([switch]$Debug)
    
    Write-Info "Building application..."
    
    if ($Debug) {
        go build -v -o main.exe .
    } else {
        go build -v -ldflags="-s -w" -o main.exe .
    }
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "Build completed successfully! Output: main.exe"
    } else {
        Write-Error "Build failed!"
        exit 1
    }
}

# Run application
function Invoke-Run {
    Write-Info "Running application..."
    go run .
}

# Clean build artifacts
function Invoke-Clean {
    Write-Info "Cleaning build artifacts..."
    
    $filesToRemove = @("main.exe", "coverage.out", "coverage.html", "gosec-results.*")
    
    foreach ($pattern in $filesToRemove) {
        $files = Get-ChildItem -Path $pattern -ErrorAction SilentlyContinue
        if ($files) {
            Remove-Item $files -Force
            Write-Info "Removed: $($files.Name -join ', ')"
        }
    }
    
    go clean -cache
    go clean -testcache
    
    Write-Success "Cleanup completed!"
}

# Dependency management
function Invoke-DepsInstall {
    Write-Info "Downloading dependencies..."
    go mod download
    Write-Success "Dependencies downloaded!"
}

function Invoke-DepsVerify {
    Write-Info "Verifying dependencies..."
    go mod verify
    Write-Success "Dependencies verified!"
}

function Invoke-DepsTidy {
    Write-Info "Tidying dependencies..."
    go mod tidy
    Write-Success "Dependencies tidied!"
}

function Invoke-DepsUpdate {
    Write-Info "Updating dependencies..."
    go get -u ./...
    go mod tidy
    Write-Success "Dependencies updated!"
}

# Complete development check
function Invoke-DevCheck {
    Write-Info "Running complete development check..."
    
    Invoke-Format
    Invoke-Vet
    Invoke-Lint
    Invoke-Security
    Invoke-Tests
    
    Write-Success "✅ All development checks passed!"
}

# CI-like check
function Invoke-CiCheck {
    Write-Info "Running CI-like checks..."
    
    Invoke-DepsVerify
    Invoke-Format
    Invoke-Vet
    Invoke-Lint
    Invoke-Security
    Invoke-Tests -Coverage
    
    Write-Success "✅ All CI checks passed!"
}

# Docker operations
function Invoke-DockerBuild {
    Write-Info "Building Docker image..."
    docker build -t ai-docs-backend .
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "Docker image built successfully!"
    } else {
        Write-Error "Docker build failed!"
        exit 1
    }
}

function Invoke-DockerRun {
    Write-Info "Running Docker container..."
    docker run -p 8080:8080 ai-docs-backend
}

# Show help
function Show-Help {
    Write-Info "Available commands:"
    Write-Host ""
    Write-Host "Testing:" -ForegroundColor Yellow
    Write-Host "  test                 Run all tests"
    Write-Host "  test-verbose         Run tests with verbose output"
    Write-Host "  test-coverage        Run tests with coverage report"
    Write-Host "  test-individual      Run tests for each package individually"
    Write-Host "  test-guardrails      Test only guardrails functionality"
    Write-Host ""
    Write-Host "PromptFoo AI Testing:" -ForegroundColor Yellow
    Write-Host "  promptfoo-install    Install PromptFoo CLI globally"
    Write-Host "  promptfoo-basic      Run basic PromptFoo tests (Phase 1)"
    Write-Host "  promptfoo-report     Generate and open PromptFoo HTML report"
    Write-Host "  promptfoo-clean      Clear PromptFoo results cache"
    Write-Host "  promptfoo-test       Run Go integration tests for PromptFoo"
    Write-Host ""
    Write-Host "Code Quality:" -ForegroundColor Yellow
    Write-Host "  lint                 Run linter"
    Write-Host "  lint-fix             Run linter with auto-fix"
    Write-Host "  security             Run security scanner"
    Write-Host "  security-report      Generate detailed security reports"
    Write-Host "  fmt                  Format Go code"
    Write-Host "  vet                  Run go vet"
    Write-Host ""
    Write-Host "Build `& Run:" -ForegroundColor Yellow
    Write-Host "  build                Build application"
    Write-Host "  build-debug          Build with debug information"
    Write-Host "  run                  Run application"
    Write-Host ""
    Write-Host "Dependencies:" -ForegroundColor Yellow
    Write-Host "  deps-install         Download dependencies"
    Write-Host "  deps-verify          Verify dependencies"
    Write-Host "  deps-tidy            Tidy dependencies"
    Write-Host "  deps-update          Update dependencies"
    Write-Host ""
    Write-Host "AI Testing (PromptFoo):" -ForegroundColor Yellow
    Write-Host "  promptfoo-install    Install PromptFoo for AI testing"
    Write-Host "  promptfoo-basic      Run basic RAG and guardrail tests"
    Write-Host "  promptfoo-report     View PromptFoo test results"
    Write-Host "  promptfoo-clean      Clean PromptFoo test results"
    Write-Host "  promptfoo-test       Run PromptFoo integration tests (Go)"
    Write-Host ""
    Write-Host "Docker:" -ForegroundColor Yellow
    Write-Host "  docker-build         Build Docker image"
    Write-Host "  docker-run           Run Docker container"
    Write-Host ""
    Write-Host "Workflows:" -ForegroundColor Yellow
    Write-Host "  dev-setup            Install development tools"
    Write-Host "  dev-check            Run all development checks"
    Write-Host "  ci-check             Run CI-like checks"
    Write-Host "  clean                Clean build artifacts"
    Write-Host ""
    Write-Host "Environment:" -ForegroundColor Yellow
    Write-Host "  env-help             Show environment variable help"
    Write-Host ""
    Write-Host "Examples:" -ForegroundColor Green
    Write-Host "  .\dev.ps1 test"
    Write-Host "  .\dev.ps1 lint-fix"
    Write-Host "  .\dev.ps1 dev-check"
    Write-Host "  .\dev.ps1 security-report"
}

function Show-EnvHelp {
    Write-Info "Environment Variables for Local Development:"
    Write-Host ""
    Write-Host "Required:" -ForegroundColor Yellow
    Write-Host "  JWT_SECRET=your-secret-key"
    Write-Host "  DB_HOST=localhost"
    Write-Host "  DB_PORT=5432"
    Write-Host "  DB_USER=postgres"
    Write-Host "  DB_PASSWORD=your-password"
    Write-Host "  DB_NAME=your-database"
    Write-Host ""
    Write-Host "Optional AI Provider Keys (choose one):" -ForegroundColor Yellow
    Write-Host "  OPENAI_API_KEY=sk-proj-your-openai-key"
    Write-Host "  GOOGLE_AI_API_KEY=AIzaSyC-your-google-ai-key"
    Write-Host "  USE_LOCAL_AI=true (for Ollama)"
    Write-Host ""
    Write-Host "Other:" -ForegroundColor Yellow
    Write-Host "  ENVIRONMENT=development"
    Write-Host "  ALLOWED_ORIGINS=http://localhost:4200"
    Write-Host ""
    Write-Host "Create a .env file in the backend directory with these variables." -ForegroundColor Green
}

# Main command dispatcher
if (-not (Test-GoInstallation)) {
    exit 1
}

switch ($Command.ToLower()) {
    "test" { Invoke-Tests }
    "test-verbose" { Invoke-Tests -Verbose }
    "test-coverage" { Invoke-Tests -Coverage }
    "test-individual" { Invoke-Tests -Individual }
    "test-guardrails" { 
        Write-Info "Testing guardrails functionality..."
        go test -v -run ".*[Gg]uardrail.*" ./utils
    }
    "promptfoo-install" { Install-PromptFoo }
    "promptfoo-basic" { Invoke-PromptFooBasic }
    "promptfoo-report" { Invoke-PromptFooReport }
    "promptfoo-clean" { Clear-PromptFooResults }
    "promptfoo-test" { Invoke-PromptFooTest }
    "lint" { Invoke-Lint }
    "lint-fix" { Invoke-Lint -Fix }
    "security" { Invoke-Security }
    "security-report" { Invoke-Security -Report }
    "fmt" { Invoke-Format }
    "vet" { Invoke-Vet }
    "build" { Invoke-Build }
    "build-debug" { Invoke-Build -Debug }
    "run" { Invoke-Run }
    "deps-install" { Invoke-DepsInstall }
    "deps-verify" { Invoke-DepsVerify }
    "deps-tidy" { Invoke-DepsTidy }
    "deps-update" { Invoke-DepsUpdate }
    "docker-build" { Invoke-DockerBuild }
    "docker-run" { Invoke-DockerRun }
    "dev-setup" { Install-DevTools }
    "dev-check" { Invoke-DevCheck }
    "ci-check" { Invoke-CiCheck }
    "clean" { Invoke-Clean }
    "env-help" { Show-EnvHelp }
    "help" { Show-Help }
    default { 
        Write-Error "Unknown command: $Command"
        Show-Help
        exit 1
    }
}
