package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// Create a global variable to store the database connection
// uppercase because other parts of the app can use this Database
var DB *sql.DB

func InitDB(dbConfig ...string) {
	// Read environment variables
	dbHost := dbConfig[0]
	dbPort := dbConfig[1]
	dbUser := dbConfig[2]
	dbPassword := dbConfig[3]
	dbName := dbConfig[4]

	// Create connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// Connect to PostgreSQL
	var err error
	// Open needs a driver name and a connection string
	// driver name is the name of the driver we are using to connect to the database
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	// Retry pinging the database to check if it's reachable
	for i := 0; i < 10; i++ {
		err = DB.Ping()
		if err == nil {
			break
		}
		log.Printf("Waiting for database... (%d/10)", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Database not reachable: %v", err)
	}

	// Set the maximum number of open connections to the database
	// Pool of ongoing connections that can be used when needed by different parts of the app
	DB.SetMaxOpenConns(10)
	// Set the maximum number of idle connections to the database
	// Pool of idle connections that can be used when needed by different parts of the app
	// How many connections we want to keep open if no one's using these connections at the moment
	// This is to prevent the database from being overloaded
	DB.SetMaxIdleConns(5)

	fmt.Println("Successfully connected to PostgreSQL!")

	// Create the tables in the database
	createTables()

	fmt.Println("Tables created successfully!")
}

// Global variable to track pgvector availability
var hasPgVector bool

// HasPgVector returns whether pgvector extension is available
func HasPgVector() bool {
	return hasPgVector
}

func createTables() {
	// Create UUID extension (always available)
	_, err := DB.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`)
	if err != nil {
		log.Fatalf("Error enabling uuid-ossp: %v", err)
	}

	// Try to create pgvector extension
	hasPgVector = false
	_, err = DB.Exec(`CREATE EXTENSION IF NOT EXISTS "vector"`)
	if err != nil {
		log.Printf("Warning: pgvector extension not available: %v", err)
		log.Printf("Falling back to standard PostgreSQL without vector search")
		log.Printf("To fix this, install pgvector or use Docker with pgvector/pgvector:pg16")
		hasPgVector = false
	} else {
		hasPgVector = true
		log.Println("pgvector extension enabled successfully")
	}

	// Store pgvector availability for other packages
	// This will be used by models to determine search strategy

	// Create the documents table
	// Each uploaded file becomes a document entry.
	createDocumentsTable := `
	CREATE TABLE IF NOT EXISTS documents (
  		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
 		name TEXT NOT NULL,
  		original_filename TEXT,
  		uploaded_at TIMESTAMP DEFAULT now()
	)
	`
	// Execute this query whenever the app starts
	_, err = DB.Exec(createDocumentsTable)

	if err != nil {
		fmt.Println("Error creating documents table:", err)
		// Crash the app if we cannot create the table
		panic("Could not create documents table.")
	}

	// Chunks table with conditional pgvector support
	// On delete cascade means that if the chunk is deleted, all associated records will be deleted as well
	// This is to prevent orphaned records in the chunks table
	// 	Each document is split into chunks. Each chunk stores:
	// 		Raw text
	// 		A vector(1536) pgvector OR JSON array (fallback)
	// 		A link back to the document
	var createChunksTable string
	if hasPgVector {
		createChunksTable = `
		CREATE TABLE IF NOT EXISTS chunks (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			document_id UUID REFERENCES documents(id) ON DELETE CASCADE,
			size BIGINT NOT NULL,
			content_type TEXT NOT NULL,
			content TEXT NOT NULL,
			embedding vector(1536) NOT NULL,
			chunk_index INT NOT NULL
		)
		`
	} else {
		// Fallback: store embeddings as TEXT (JSON array)
		createChunksTable = `
		CREATE TABLE IF NOT EXISTS chunks (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			document_id UUID REFERENCES documents(id) ON DELETE CASCADE,
			size BIGINT NOT NULL,
			content_type TEXT NOT NULL,
			content TEXT NOT NULL,
			embedding TEXT NOT NULL, -- JSON array of floats
			chunk_index INT NOT NULL
		)
		`
	}
	// Execute this query whenever the app starts
	_, err = DB.Exec(createChunksTable)
	if err != nil {
		fmt.Println("Error creating chunks table:", err)
		// Crash the app if we cannot create the table
		panic("Could not create chunks table.")
	}

	// Create appropriate index based on pgvector availability
	if hasPgVector {
		// Vector index for efficient ANN search
		// This index allows for fast similarity search using vector embeddings
		// It uses the ivfflat algorithm for approximate nearest neighbor search
		_, err = DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_chunks_embedding
		ON chunks USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100)
		`)
		if err != nil {
			log.Printf("Warning: Could not create vector index: %v", err)
		} else {
			log.Println("Vector index created successfully")
		}
	} else {
		// Fallback: create text search index
		// This index allows for full-text search on the content field
		// It uses the gin index type for efficient text search
		_, err = DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_chunks_content
		ON chunks USING gin(to_tsvector('english', content))
		`)
		if err != nil {
			log.Printf("Warning: Could not create text search index: %v", err)
		} else {
			log.Println("Text search index created successfully")
		}
	}
	// Create the users table
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT now(),
		avatar TEXT
	)
	`
	// Execute this query whenever the app starts
	_, err = DB.Exec(createUsersTable)

	if err != nil {
		fmt.Println("Error creating users table:", err)
		// Crash the app if we cannot create the table
		panic("Could not create users table.")
	}

	// Create the reset_tokens table
	createResetTokensTable := `
	CREATE TABLE IF NOT EXISTS reset_tokens (
		token TEXT PRIMARY KEY,
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		expiry TIMESTAMP NOT NULL,
		used BOOLEAN DEFAULT false,
		created_at TIMESTAMP DEFAULT now()
	)
	`
	_, err = DB.Exec(createResetTokensTable)
	if err != nil {
		fmt.Println("Error creating reset_tokens table:", err)
		panic("Could not create reset_tokens table.")
	}

	// Create questions table
	// Track what users ask (great for analytics or costs)
	createQuestionsTable := `
	CREATE TABLE IF NOT EXISTS questions (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		user_id UUID REFERENCES users(id),
		document_id UUID REFERENCES documents(id),
		query TEXT NOT NULL,
		answer TEXT,
		asked_at TIMESTAMP DEFAULT now()
	)
	`
	// Execute this query whenever the app starts
	_, err = DB.Exec(createQuestionsTable)

	// If there is an error creating the table, print the error and crash the app
	if err != nil {
		fmt.Println("Error creating questions table:", err)
		panic("Could not create questions table.")
	}
}
