package utils

import (
	"database/sql"
	"fmt"

	"github.com/MauricioAliendre182/backend/db"
)

// A transaction is a sequence of operations performed as a single logical unit of work
// Transactions ensure that either all operations are completed successfully or none are applied,
// WithTransaction executes a function within a database transaction
// If the function returns an error, the transaction is rolled back
// If the function completes successfully, the transaction is committed
// sql.Tx refers to a transaction in the database
// This function ensures that the transaction is properly handled, including rollback on panic
func WithTransaction(fn func(tx *sql.Tx) error) error {
	// Begin transaction
	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	// Ensure transaction is closed
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // Re-panic after rollback
		}
	}()

	// Execute the function
	if err := fn(tx); err != nil {
		// Rollback on error
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("failed to rollback transaction: %v (original error: %v)", rollbackErr, err)
		}
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}
