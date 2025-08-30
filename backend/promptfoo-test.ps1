#!/usr/bin/env powershell
# PromptFoo Test Runner for AI-Powered Internal Company Docs
# This script provides convenient commands for running AI evaluation tests
# Ensures proper environment setup and handles different test scenarios

param(
    [Parameter(Position=0)]
    [ValidateSet("install", "basic", "guardrails", "all", "report", "clean", "help")]
    [string]$Command = "help",
    
    [Parameter()]
    [string]$Provider = "all",
    
    [Parameter()]
    [switch]$Verbose = $false,
    
    [Parameter()]
    [switch]$DryRun = $false
)

# Set error handling
$ErrorActionPreference = "Stop"

# Define color functions for better output visibility
function Write-Success { param($Message) Write-Host "✅ $Message" -ForegroundColor Green }
function Write-Warning { param($Message) Write-Host "⚠️  $Message" -ForegroundColor Yellow }
function Write-Error { param($Message) Write-Host "❌ $Message" -ForegroundColor Red }
function Write-Info { param($Message) Write-Host "ℹ️  $Message" -ForegroundColor Cyan }

# Function to check if PromptFoo is installed
function Test-PromptFooInstalled {
    try {
        $version = npm list -g promptfoo --depth=0 2>$null
        if ($version -match "promptfoo@") {
            return $true
        }
    }
    catch {
        return $false
    }
    return $false
}

# Function to install PromptFoo if not present
function Install-PromptFoo {
    Write-Info "Checking PromptFoo installation..."
    
    if (Test-PromptFooInstalled) {
        Write-Success "PromptFoo is already installed"
        return
    }
    
    Write-Info "Installing PromptFoo..."
    try {
        # Check if npm is available
        $npmVersion = npm --version 2>$null
        if (-not $npmVersion) {
            Write-Error "npm is not installed. Please install Node.js first."
            exit 1
        }
        
        # Install PromptFoo globally
        npm install -g promptfoo
        
        if (Test-PromptFooInstalled) {
            Write-Success "PromptFoo installed successfully"
        } else {
            Write-Error "Failed to install PromptFoo"
            exit 1
        }
    }
    catch {
        Write-Error "Failed to install PromptFoo: $($_.Exception.Message)"
        exit 1
    }
}

# Function to validate environment variables
function Test-Environment {
    Write-Info "Validating environment configuration..."
    
    # Check for AI provider API keys based on configuration
    $missingKeys = @()
    
    # Check OpenAI key if OpenAI provider is configured
    if (-not $env:OPENAI_API_KEY -and $Provider -in @("all", "openai")) {
        $missingKeys += "OPENAI_API_KEY"
    }
    
    # Check Google AI key if Google provider is configured
    if (-not $env:GOOGLE_AI_API_KEY -and $Provider -in @("all", "google")) {
        $missingKeys += "GOOGLE_AI_API_KEY"
    }
    
    # Ollama doesn't require API key but should be running
    if ($Provider -in @("all", "ollama")) {
        try {
            $ollamaCheck = Invoke-RestMethod -Uri "http://localhost:11434/api/version" -TimeoutSec 5 -ErrorAction SilentlyContinue
            if (-not $ollamaCheck) {
                Write-Warning "Ollama service may not be running at localhost:11434"
            }
        }
        catch {
            Write-Warning "Cannot connect to Ollama service. Tests may fail for Ollama provider."
        }
    }
    
    if ($missingKeys.Count -gt 0) {
        Write-Warning "Missing environment variables: $($missingKeys -join ', ')"
        Write-Info "Set these variables before running tests with corresponding providers"
    } else {
        Write-Success "Environment validation passed"
    }
}

# Function to run basic RAG tests
function Invoke-BasicTests {
    Write-Info "Running basic RAG functionality tests..."
    
    $configFile = "promptfoo-config.yaml"
    if (-not (Test-Path $configFile)) {
        Write-Error "Configuration file $configFile not found"
        exit 1
    }
    
    try {
        # Prepare command arguments
        $args = @("eval", "--config", $configFile)
        
        if ($Verbose) {
            $args += "--verbose"
        }
        
        if ($DryRun) {
            Write-Info "Dry run mode - would execute: promptfoo $($args -join ' ')"
            return
        }
        
        # Execute PromptFoo evaluation
        Write-Info "Executing PromptFoo evaluation..."
        & promptfoo $args
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Basic tests completed successfully"
        } else {
            Write-Error "Basic tests failed with exit code $LASTEXITCODE"
            exit $LASTEXITCODE
        }
    }
    catch {
        Write-Error "Failed to run basic tests: $($_.Exception.Message)"
        exit 1
    }
}

# Function to run guardrail security tests
function Invoke-GuardrailTests {
    Write-Info "Running guardrail security tests..."
    
    # For Phase 1, guardrail tests are included in the main config
    # In future phases, this could run a separate guardrail-specific configuration
    Write-Info "Guardrail tests are integrated in the main test suite for Phase 1"
    Invoke-BasicTests
}

# Function to generate and display test report
function Show-TestReport {
    Write-Info "Generating test report..."
    
    $resultsDir = "promptfoo-results"
    if (-not (Test-Path $resultsDir)) {
        Write-Warning "No test results found. Run tests first."
        return
    }
    
    try {
        # Show summary using PromptFoo's built-in reporting
        & promptfoo view
        
        Write-Success "Test report displayed"
        Write-Info "Detailed results available in: $resultsDir"
    }
    catch {
        Write-Error "Failed to generate report: $($_.Exception.Message)"
    }
}

# Function to clean test results
function Clear-TestResults {
    Write-Info "Cleaning test results..."
    
    $resultsDir = "promptfoo-results"
    if (Test-Path $resultsDir) {
        try {
            Remove-Item $resultsDir -Recurse -Force
            Write-Success "Test results cleaned"
        }
        catch {
            Write-Error "Failed to clean results: $($_.Exception.Message)"
        }
    } else {
        Write-Info "No test results to clean"
    }
}

# Function to display help information
function Show-Help {
    Write-Host @"
🤖 PromptFoo Test Runner for AI-Powered Internal Company Docs

USAGE:
    .\promptfoo-test.ps1 <command> [options]

COMMANDS:
    install     Install PromptFoo if not already present
    basic       Run basic RAG functionality tests
    guardrails  Run guardrail security tests
    all         Run all test suites
    report      Display test results and generate report
    clean       Clean test results and temporary files
    help        Show this help message

OPTIONS:
    -Provider <provider>   Target specific provider (all, openai, google, ollama)
    -Verbose              Enable verbose output for debugging
    -DryRun               Show what would be executed without running

EXAMPLES:
    .\promptfoo-test.ps1 install                    # Install PromptFoo
    .\promptfoo-test.ps1 basic -Verbose             # Run basic tests with verbose output
    .\promptfoo-test.ps1 all -Provider openai       # Test only OpenAI provider
    .\promptfoo-test.ps1 report                     # View test results
    .\promptfoo-test.ps1 clean                      # Clean up test artifacts

ENVIRONMENT SETUP:
    Set these environment variables before running tests:
    - OPENAI_API_KEY (for OpenAI provider)
    - GOOGLE_AI_API_KEY (for Google AI provider)
    - Ensure Ollama is running at localhost:11434 (for Ollama provider)

"@ -ForegroundColor White
}

# Main execution logic
switch ($Command) {
    "install" {
        Install-PromptFoo
    }
    "basic" {
        Install-PromptFoo
        Test-Environment
        Invoke-BasicTests
    }
    "guardrails" {
        Install-PromptFoo
        Test-Environment
        Invoke-GuardrailTests
    }
    "all" {
        Install-PromptFoo
        Test-Environment
        Write-Info "Running complete test suite..."
        Invoke-BasicTests
        Write-Success "All tests completed"
    }
    "report" {
        Show-TestReport
    }
    "clean" {
        Clear-TestResults
    }
    "help" {
        Show-Help
    }
    default {
        Write-Error "Unknown command: $Command"
        Show-Help
        exit 1
    }
}
