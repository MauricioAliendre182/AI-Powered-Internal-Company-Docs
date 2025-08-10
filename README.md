# AI-Powered Internal Company Documentation System

A comprehensive AI-powered document management and retrieval system that enables intelligent search and question-answering capabilities for internal company documents. Built with **Go** backend, **Angular** frontend, and **PostgreSQL** with **pgvector** for semantic search.

## 🚀 Features

- **Document Upload & Processing**: Support for PDF, TXT, and DOCX files with automatic text extraction
- **AI-Powered RAG**: Retrieval-Augmented Generation for intelligent document querying
- **Multiple AI Providers**: Factory pattern supporting OpenAI, Google AI (Gemini), and Ollama (local AI)
- **Semantic Search**: Vector-based similarity search using pgvector extension
- **User Management**: JWT-based authentication and user management
- **Responsive UI**: Modern Angular frontend with FontAwesome icons
- **Real-time Processing**: Chunking and embedding generation for uploaded documents

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        Frontend (Angular)                       │
├─────────────────────────────────────────────────────────────────┤
│  • Document Upload & Management  • RAG Query Interface          │
│  • User Authentication          • Responsive Navigation         │
│  • Chunk Pagination            • Document Preview              │
└─────────────────┬───────────────────────────────────────────────┘
                  │ HTTP/REST API
┌─────────────────▼───────────────────────────────────────────────┐
│                      Backend (Go/Gin)                          │
├─────────────────────────────────────────────────────────────────┤
│  • JWT Authentication      • File Processing & Validation       │
│  • Document Management     • RAG Service with Factory Pattern   │
│  • Vector Embeddings       • Rate Limiting & Error Handling    │
└─────────────────┬───────────────────────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────────────────────┐
│                   AI Service Factory                           │
├─────────────────────────────────────────────────────────────────┤
│  OpenAI Service  │  Google AI Service  │  Ollama Service        │
│  ├─ Embeddings   │  ├─ Embeddings      │  ├─ Embeddings        │
│  └─ Chat         │  └─ Chat            │  └─ Chat              │
└─────────────────┬───────────────────────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────────────────────┐
│                 PostgreSQL + pgvector                          │
├─────────────────────────────────────────────────────────────────┤
│  • Documents & Chunks Storage    • Vector Similarity Search     │
│  • User Management              • Embedding Storage            │
│  • UUID Primary Keys            • Transactional Operations     │
└─────────────────────────────────────────────────────────────────┘
```

## 📊 Database Schema

### Entity Relationship Diagram

```
┌─────────────────┐       ┌─────────────────┐       ┌─────────────────┐
│     users       │       │   documents     │       │     chunks      │
├─────────────────┤       ├─────────────────┤       ├─────────────────┤
│ id (UUID) PK    │   ┌───│ id (UUID) PK    │   ┌───│ id (UUID) PK    │
│ email           │   │   │ name            │   │   │ document_id FK  │
│ password_hash   │   │   │ original_filename│   │   │ content         │
│ name            │   │   │ uploaded_at     │   │   │ embedding       │
│ avatar          │   │   │ user_id FK      ├───┘   │ chunk_index     │
│ verified        │   │   └─────────────────┘       │ size            │
│ created_at      │   │                             │ content_type    │
└─────────────────┘   │                             └─────────────────┘
          │           │
          │           │   ┌─────────────────┐
          │           └───│ tokens          │
          │               ├─────────────────┤
          └───────────────│ user_id FK      │
                          │ token_hash      │
                          │ token_type      │
                          │ expires_at      │
                          │ created_at      │
                          └─────────────────┘
```

### Table Descriptions

#### `users`
- **Purpose**: Store user account information
- **Key Features**: UUID primary keys, email verification, avatar support
- **Relationships**: One-to-many with documents and tokens

#### `documents`
- **Purpose**: Store uploaded document metadata
- **Key Features**: Original filename preservation, upload timestamp
- **Relationships**: Belongs to user, has many chunks

#### `chunks`
- **Purpose**: Store processed document chunks with embeddings
- **Key Features**: Vector embeddings for similarity search, content indexing
- **Relationships**: Belongs to document

#### `tokens`
- **Purpose**: Manage authentication and verification tokens
- **Key Features**: JWT tokens, password reset tokens, email verification
- **Relationships**: Belongs to user

## 🤖 AI Factory Pattern Implementation

### Overview
The system uses a factory pattern to support multiple AI providers with a unified interface, allowing easy switching between providers without code changes.

### Provider Selection Logic
```
Configuration Check:
    │
    ├─ USE_LOCAL_AI=true? ────► Ollama Provider (Free, Local)
    │                           │
    ├─ GOOGLE_AI_API_KEY set? ──► Gemini Provider (Cost-effective)
    │                           │
    ├─ OPENAI_API_KEY set? ─────► OpenAI Provider (High-quality)
    │                           │
    └─ No valid config ─────────► Error
```

### Supported AI Providers

#### 1. OpenAI
- **Models**: GPT-4, GPT-3.5-turbo, text-embedding-3-small
- **Pros**: Highest quality, well-documented
- **Cons**: Most expensive, requires API key
- **Use Case**: Production environments requiring best quality

#### 2. Google AI (Gemini)
- **Models**: Gemini-1.5-flash, models/embedding-001
- **Pros**: Cost-effective, good performance
- **Cons**: Requires API key, newer ecosystem
- **Use Case**: Production environments with cost optimization

#### 3. Ollama (Local AI)
- **Models**: Llama 3.1, nomic-embed-text
- **Pros**: Completely free, privacy-focused, no API limits
- **Cons**: Requires powerful hardware, slower responses
- **Use Case**: Development, privacy-sensitive environments

### Configuration Examples

#### OpenAI Configuration
```env
OPENAI_API_KEY=sk-proj-your-openai-api-key-here
EMBEDDING_MODEL=text-embedding-3-small
CHAT_MODEL=gpt-3.5-turbo
```

#### Google AI Configuration
```env
GOOGLE_AI_API_KEY=AIzaSyC-your-google-ai-key-here
EMBEDDING_MODEL=models/embedding-001
CHAT_MODEL=models/gemini-1.5-flash
```

#### Ollama Configuration
```env
USE_LOCAL_AI=true
OLLAMA_BASE_URL=http://localhost:11434
EMBEDDING_MODEL=nomic-embed-text
CHAT_MODEL=llama3.1:8b
```

## 🔑 AI Provider Setup & Configuration

### 1. OpenAI Setup

#### Getting OpenAI API Key:
1. **Create Account**: Go to [https://platform.openai.com/](https://platform.openai.com/)
2. **Sign Up/Login**: Create account or sign in
3. **Navigate to API Keys**: Go to [https://platform.openai.com/api-keys](https://platform.openai.com/api-keys)
4. **Create New Key**: Click "Create new secret key"
5. **Copy Key**: Save the key securely (starts with `sk-proj-`)

#### Available Models:

**Embedding Models:**
- `text-embedding-3-small` (1536 dimensions, $0.02/1M tokens) - **Recommended**
- `text-embedding-3-large` (3072 dimensions, $0.13/1M tokens)
- `text-embedding-ada-002` (1536 dimensions, $0.10/1M tokens) - Legacy

**Chat Models:**
- `gpt-4-turbo` (Latest GPT-4, $10/1M input tokens) - **Recommended for quality**
- `gpt-3.5-turbo` (Fast and cost-effective, $0.50/1M input tokens) - **Recommended for cost**
- `gpt-4o` (GPT-4 Omni, $15/1M input tokens)
- `gpt-4o-mini` (Smaller GPT-4 Omni, $0.15/1M input tokens)

#### OpenAI Configuration:
```env
OPENAI_API_KEY=sk-proj-your-actual-openai-key-here
EMBEDDING_MODEL=text-embedding-3-small
CHAT_MODEL=gpt-3.5-turbo
```

### 2. Google AI (Gemini) Setup

#### Getting Google AI API Key:
1. **Go to Google AI Studio**: [https://makersuite.google.com/app/apikey](https://makersuite.google.com/app/apikey)
2. **Sign in**: Use your Google account
3. **Create API Key**: Click "Create API Key"
4. **Select Project**: Choose existing project or create new one
5. **Copy Key**: Save the key securely (starts with `AIzaSy`)

#### Available Models:

**Embedding Models:**
- `models/embedding-001` (768 dimensions, generous free tier) - **Recommended**
- `models/text-embedding-004` (768 dimensions, latest version)

**Chat Models:**
- `models/gemini-1.5-flash` (Fast and efficient, $0.075/1M input tokens) - **Recommended**
- `models/gemini-1.5-pro` (Best quality, $3.50/1M input tokens)
- `models/gemini-1.0-pro` (Standard model, $0.50/1M input tokens)

#### Google AI Configuration:
```env
GOOGLE_AI_API_KEY=AIzaSyC-your-actual-google-ai-key-here
EMBEDDING_MODEL=models/embedding-001
CHAT_MODEL=models/gemini-1.5-flash
```

### 3. Ollama (Local AI) Setup

#### Installing Ollama:

**Windows:**
```powershell
# Download from https://ollama.ai/download/windows
# Or use winget (if available)
winget install Ollama.Ollama
```

**macOS:**
```bash
# Download from https://ollama.ai/download/mac
# Or use Homebrew
brew install ollama
```

**Linux:**
```bash
# Official installation script
curl -fsSL https://ollama.ai/install.sh | sh
```

#### Starting Ollama Service:
```bash
# Start Ollama (runs on localhost:11434 by default)
ollama serve
```

#### Downloading Models:

**Embedding Models:**
```bash
# Recommended for embeddings (274MB)
ollama pull nomic-embed-text

# Alternative options:
ollama pull mxbai-embed-large    # 669MB, more accurate
ollama pull all-minilm           # 23MB, fastest
```

**Chat Models:**
```bash
# Small models (4-8GB RAM):
ollama pull llama3.1:1b         # 1.3GB, basic quality
ollama pull phi3:mini           # 2.3GB, Microsoft model
ollama pull gemma2:2b           # 1.6GB, Google model

# Medium models (8-16GB RAM) - Recommended:
ollama pull llama3.1:8b         # 4.7GB, good balance
ollama pull mistral:7b          # 4.1GB, efficient
ollama pull codegemma:7b        # 5.0GB, good for code

# Large models (16GB+ RAM):
ollama pull llama3.1:70b        # 40GB, highest quality
ollama pull mixtral:8x7b        # 26GB, mixture of experts
```

#### Ollama Configuration:
```env
USE_LOCAL_AI=true
OLLAMA_BASE_URL=http://localhost:11434
EMBEDDING_MODEL=nomic-embed-text
CHAT_MODEL=llama3.1:8b
```

## 🔧 Environment File Setup

### Creating the `.env` File

The `.env` file is **not included** in the GitHub repository for security reasons. You need to create it manually:

#### Step 1: Navigate to Backend Folder
```bash
cd backend
```

#### Step 2: Create `.env` File
```bash
# Windows (PowerShell)
New-Item -Path ".env" -ItemType File

# Windows (Command Prompt)
type nul > .env

# macOS/Linux
touch .env
```

#### Step 3: Add Configuration

Choose **ONE** of the following configurations based on your preferred AI provider:

#### Option A: OpenAI Configuration (Recommended for Production)
```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=internal_docs_password
DB_NAME=internal_docs

# OpenAI Configuration
OPENAI_API_KEY=sk-proj-your-actual-openai-key-here
EMBEDDING_MODEL=text-embedding-3-small
CHAT_MODEL=gpt-3.5-turbo

# Application Configuration
ENVIRONMENT=development
PORT=8090

# File Upload Configuration
MAX_FILE_SIZE=10485760  # 10MB in bytes
CHUNK_SIZE=500

# Rate Limiting Configuration
RATE_LIMIT_MAX_TOKENS=10
RATE_LIMIT_REFILL_RATE=1

# JWT Configuration
JWT_SECRET=YourSuperSecretJWT_Key_2024!x9P3qR7sT1vW5zX8aB4cD6eF2gH

# Email Configuration (Optional)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email@gmail.com
SMTP_PASSWORD=your_app_password
```

#### Option B: Google AI Configuration (Cost-Effective)
```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=internal_docs_password
DB_NAME=internal_docs

# Google AI Configuration
GOOGLE_AI_API_KEY=AIzaSyC-your-actual-google-ai-key-here
EMBEDDING_MODEL=models/embedding-001
CHAT_MODEL=models/gemini-1.5-flash

# Application Configuration
ENVIRONMENT=development
PORT=8090

# File Upload Configuration
MAX_FILE_SIZE=10485760  # 10MB in bytes
CHUNK_SIZE=500

# Rate Limiting Configuration
RATE_LIMIT_MAX_TOKENS=10
RATE_LIMIT_REFILL_RATE=1

# JWT Configuration
JWT_SECRET=YourSuperSecretJWT_Key_2024!x9P3qR7sT1vW5zX8aB4cD6eF2gH

# Email Configuration (Optional)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email@gmail.com
SMTP_PASSWORD=your_app_password
```

#### Option C: Ollama Configuration (Free & Local)
```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=internal_docs_password
DB_NAME=internal_docs

# Local AI Configuration (Ollama)
USE_LOCAL_AI=true
OLLAMA_BASE_URL=http://localhost:11434
EMBEDDING_MODEL=nomic-embed-text
CHAT_MODEL=llama3.1:8b

# Application Configuration
ENVIRONMENT=development
PORT=8090

# File Upload Configuration
MAX_FILE_SIZE=10485760  # 10MB in bytes
CHUNK_SIZE=500

# Rate Limiting Configuration
RATE_LIMIT_MAX_TOKENS=10
RATE_LIMIT_REFILL_RATE=1

# JWT Configuration
JWT_SECRET=YourSuperSecretJWT_Key_2024!x9P3qR7sT1vW5zX8aB4cD6eF2gH

# Email Configuration (Optional)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email@gmail.com
SMTP_PASSWORD=your_app_password
```

### Configuration Notes:

1. **Replace Placeholder Values**: 
   - `your-actual-openai-key-here` → Your real OpenAI API key
   - `your-actual-google-ai-key-here` → Your real Google AI API key
   - `internal_docs_password` → Your desired database password

2. **JWT Secret**: 
   - Generate a strong secret key for production
   - Use online generators or: `openssl rand -base64 32`

3. **Database Password**: 
   - Use a strong password for production
   - Match this with your PostgreSQL setup

4. **Email Configuration**: 
   - Optional, only needed for password reset functionality
   - Use app-specific passwords for Gmail

### Verification Steps:

1. **Check File Creation**:
```bash
# Verify the file exists
ls -la .env    # macOS/Linux
dir .env       # Windows
```

2. **Validate Configuration**:
```bash
# Start the application to test configuration
go run main.go
```

3. **Test API Keys**:
```bash
# Test health endpoint
curl http://localhost:8090/health
```

The application will automatically detect which AI provider to use based on the environment variables you've set.

## 🎨 Frontend Architecture (Angular)

### Component Structure
```
src/app/
├── core/                           # Core services and guards
│   ├── guards/                     # Route protection
│   │   ├── auth.guard.ts          # Authentication guard
│   │   └── redirect.guard.ts      # Redirect logic
│   ├── interceptors/              # HTTP interceptors
│   │   └── auth.interceptor.ts    # JWT token injection
│   ├── models/                    # TypeScript interfaces
│   │   ├── auth.model.ts         # Authentication models
│   │   ├── user.model.ts         # User models
│   │   └── request-status.model.ts # API status tracking
│   └── services/                  # Core services
│       ├── auth.service.ts       # Authentication logic
│       ├── user.service.ts       # User management
│       ├── token.service.ts      # Token management
│       └── me.service.ts         # Current user service
├── modules/
│   ├── auth/                      # Authentication module
│   │   └── pages/
│   │       ├── login/            # Login component
│   │       ├── register/         # Registration component
│   │       ├── forgot-password/  # Password reset
│   │       └── recovery/         # Account recovery
│   └── shared/                    # Shared components
│       └── components/
│           ├── navbar/           # Navigation bar
│           ├── footer/           # Footer component
│           └── list-documents/   # Document listing with chunks
└── utils/
    └── validators.ts             # Custom form validators
```

### Key Features

#### 1. Authentication System
- **JWT-based authentication** with automatic token refresh
- **Route guards** protecting authenticated routes
- **Interceptors** for automatic token injection

#### 2. Document Management
- **File upload** with drag-and-drop support
- **Document listing** with metadata display
- **Chunk preview** with "Load More" functionality

#### 3. RAG Query Interface
- **Question input** with real-time processing
- **Response display** with source attribution
- **Context visualization** showing relevant chunks

#### 4. Responsive Design
- **Mobile-first approach** with Bootstrap integration
- **FontAwesome icons** for consistent UI
- **Modern component architecture** with standalone components

## 🛠️ Setup and Installation

### Prerequisites
- **Node.js** 18+ (for frontend)
- **Go** 1.21+ (for backend)
- **PostgreSQL** 14+ with pgvector extension
- **Docker** (optional, for containerized setup)

### Quick Start with Docker

1. **Clone the repository**:
```bash
git clone https://github.com/MauricioAliendre182/AI-Powered-Internal-Company-Docs.git
cd AI-Powered-Internal-Company-Docs
```

2. **Configure environment**:
```bash
# Copy example environment file
cp backend/.env.example backend/.env

# Edit with your AI provider credentials
nano backend/.env
```

3. **Start with Docker Compose**:
```bash
docker-compose up --build
```

4. **Access the application**:
- Frontend: http://localhost
- Backend API: http://localhost:8090
- Database: localhost:5432
- pgAdmin: http://localhost:8081

### Manual Setup

#### Backend Setup
```bash
cd backend

# Install dependencies
go mod download

# Set up environment variables
cp .env.example .env
# Edit .env with your configuration

# Run database migrations
go run main.go migrate

# Start the server
go run main.go
```

#### Frontend Setup
```bash
cd frontend

# Install dependencies
npm install

# Start development server
ng serve

# Build for production
ng build --prod
```

### Database Setup

#### Using Docker (Recommended)
```yaml
# docker-compose.yml includes pgvector-enabled PostgreSQL
db:
  image: pgvector/pgvector:pg16
  environment:
    POSTGRES_DB: internal_docs
    POSTGRES_PASSWORD: your_password
```

#### Manual PostgreSQL Setup
```bash
# Install pgvector extension
# Ubuntu/Debian:
sudo apt install postgresql-16-pgvector

# macOS:
brew install pgvector

# Enable extension in your database:
psql -d your_database -c "CREATE EXTENSION IF NOT EXISTS vector;"
```

## 🔧 Configuration Options

### Environment Variables

#### Database Configuration
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_secure_password
DB_NAME=internal_docs
```

#### AI Provider Configuration
```env
# Choose one:
OPENAI_API_KEY=sk-proj-...           # For OpenAI
GOOGLE_AI_API_KEY=AIzaSyC-...        # For Google AI
USE_LOCAL_AI=true                    # For Ollama

# Model selection
EMBEDDING_MODEL=text-embedding-3-small
CHAT_MODEL=gpt-3.5-turbo
```

#### Application Settings
```env
ENVIRONMENT=development
PORT=8090
MAX_FILE_SIZE=10485760              # 10MB
CHUNK_SIZE=1000                     # Characters per chunk
JWT_SECRET=your_jwt_secret_key
```

#### Rate Limiting
```env
RATE_LIMIT_MAX_TOKENS=10
RATE_LIMIT_REFILL_RATE=1
```

## 📚 API Documentation

### Authentication Endpoints
```
POST /api/v1/auth/register      # User registration
POST /api/v1/auth/login         # User login
POST /api/v1/auth/logout        # User logout
POST /api/v1/auth/refresh       # Token refresh
POST /api/v1/auth/forgot-password # Password reset
```

### Document Management
```
GET    /api/v1/documents        # List user documents
POST   /api/v1/documents        # Upload document
GET    /api/v1/documents/:id    # Get document details
DELETE /api/v1/documents/:id    # Delete document
GET    /api/v1/documents/:id/chunks # Get document chunks
```

### RAG Query
```
POST /api/v1/query              # Query documents with AI
```

### Health & Monitoring
```
GET /health                     # System health check
GET /readiness                  # Readiness probe
GET /liveness                   # Liveness probe
```

## 🧪 Testing

### Backend Testing
```bash
cd backend

# Run unit tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Test specific package
go test ./models/...
```

### Frontend Testing
```bash
cd frontend

# Run unit tests
ng test

# Run e2e tests
ng e2e

# Run tests with coverage
ng test --code-coverage
```

### Integration Testing
```bash
# Test document upload
curl -X POST -F "file=@test.pdf" \
  -H "Authorization: Bearer YOUR_JWT" \
  http://localhost:8090/api/v1/documents

# Test RAG query
curl -X POST http://localhost:8090/api/v1/query \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT" \
  -d '{"question": "What are the company policies?"}'
```

## 🚀 Deployment

### Docker Production Deployment
```bash
# Build production images
docker-compose -f docker-compose.prod.yml build

# Deploy
docker-compose -f docker-compose.prod.yml up -d
```

### Manual Production Deployment
```bash
# Build frontend
cd frontend
ng build --prod

# Build backend
cd ../backend
go build -o main .

# Set production environment variables
export ENVIRONMENT=production
export DB_HOST=your_prod_db_host

# Run with process manager
./main
```

## 🔒 Security Features

- **JWT Authentication** with secure token management
- **File Type Validation** preventing malicious uploads
- **Rate Limiting** protecting against API abuse
- **Input Sanitization** preventing injection attacks
- **CORS Configuration** for secure cross-origin requests
- **Password Hashing** using bcrypt
- **SQL Injection Protection** with parameterized queries

## 📊 Monitoring & Observability

### Health Checks
- System health endpoints for load balancers
- Database connectivity monitoring
- AI service availability checking

### Logging
- Structured JSON logging in production
- Request/response logging with correlation IDs
- Error tracking with stack traces

### Metrics
- API response times
- Document processing metrics
- AI service performance tracking

## 🤝 Contributing

1. **Fork the repository**
2. **Create feature branch**: `git checkout -b feature/amazing-feature`
3. **Commit changes**: `git commit -m 'Add amazing feature'`
4. **Push to branch**: `git push origin feature/amazing-feature`
5. **Open Pull Request**

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **OpenAI** for GPT and embedding models
- **Google AI** for Gemini models
- **Ollama** for local AI capabilities
- **pgvector** for PostgreSQL vector similarity search
- **Angular** and **Go** communities for excellent frameworks