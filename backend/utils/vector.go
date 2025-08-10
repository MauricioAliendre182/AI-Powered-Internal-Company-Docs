package utils

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
)

// -------------------------------------------------------------------------------
// FLOW HOW THIS Vector WORKS:
// The Value() method formats the vector as [1.0,2.0,3.0]
// instead of the problematic {1.0,2.0,3.0} format that was causing
// the PostgreSQL error.

// Here's the complete flow:
// 		1. OpenAI API returns: []float32{1.0, 2.0, 3.0} as pq.Float32Array
// 		2. Type conversion: Vector(embedding) converts it to our custom Vector type
// 		3. Database insertion: When PostgreSQL sees the Vector type, it calls Value() method
// 		4. Value() method returns: "[1.0,2.0,3.0]" (correct pgvector format)
// 		5. PostgreSQL stores: The vector in the correct format
// When reading back:
// 		1. PostgreSQL returns: "[1.0,2.0,3.0]" from the database
// 		2. Scan() method: Parses this string back into Vector type
// 		3. Application uses: The Vector as a normal slice of float32
// IMPORTANT: The methods in vector.go (Value() and Scan()) are used implicitly
// by the database driver
// -------------------------------------------------------------------------------

// Vector represents a pgvector embedding
// With Vector() we can convert pq.Float32Array to Vector
// This allows us to use Vector as a type for embeddings in our application
// Vector() uses the methods defined below to handle database interactions
type Vector []float32

// Value implements the driver.Valuer interface for database/sql
// It formats the Vector as a PostgreSQL vector string
// This is used when inserting or updating the vector in the database
// driver.Value is the interface that allows us to convert our Vector type to a format suitable for database storage
// We use it as Vector() to convert pq.Float32Array to Vector
// This allows us to store embeddings in a format compatible with PostgreSQL's vector type
func (v Vector) Value() (driver.Value, error) {
	if len(v) == 0 {
		return nil, nil
	}

	// Format as PostgreSQL vector: [1.0,2.0,3.0]
	parts := make([]string, len(v))
	for i, val := range v {
		parts[i] = strconv.FormatFloat(float64(val), 'f', -1, 32)
	}

	return "[" + strings.Join(parts, ",") + "]", nil
}

// Scan implements the sql.Scanner interface for database/sql
// It parses a PostgreSQL vector string into the Vector type
// This is used when reading the vector from the database
func (v *Vector) Scan(value interface{}) error {
	if value == nil {
		*v = nil
		return nil
	}

	switch s := value.(type) {
	case string:
		return v.parseVector(s)
	case []byte:
		return v.parseVector(string(s))
	default:
		return fmt.Errorf("cannot scan %T into Vector", value)
	}
}

// parseVector parses a vector string like "[1.0,2.0,3.0]" into Vector
// It handles both empty vectors and normal vectors
// Returns an error if the format is invalid
func (v *Vector) parseVector(s string) error {
	s = strings.TrimSpace(s)
	if len(s) < 2 || s[0] != '[' || s[len(s)-1] != ']' {
		return fmt.Errorf("invalid vector format: %s", s)
	}

	// Remove brackets
	s = s[1 : len(s)-1]

	// Handle empty vector
	if strings.TrimSpace(s) == "" {
		*v = Vector{}
		return nil
	}

	// Split by comma and parse each float
	parts := strings.Split(s, ",")
	result := make(Vector, len(parts))

	for i, part := range parts {
		val, err := strconv.ParseFloat(strings.TrimSpace(part), 32)
		if err != nil {
			return fmt.Errorf("invalid float in vector: %s", part)
		}
		result[i] = float32(val)
	}

	*v = result
	return nil
}

// ToFloat32Array converts Vector to []float32 for compatibility
func (v Vector) ToFloat32Array() []float32 {
	return []float32(v)
}

// FromFloat32Array creates a Vector from []float32
func FromFloat32Array(arr []float32) Vector {
	return Vector(arr)
}
