package utils

import (
	"strings"

	"github.com/MauricioAliendre182/backend/db"
)

// CheckIfAdmin checks if a user is an admin based on their user ID
// This function queries the database to get the user's email and checks
// if it matches any of the admin emails configured in the environment
func CheckIfAdmin(userID string) bool {
	// Get the user's email from the database
	userEmail, err := getUserEmailByID(userID)
	if err != nil {
		// If we can't get the user's email, they're not an admin
		return false
	}

	// Get admin emails from environment configuration
	adminEmails := getAdminEmails()

	// Check if the user's email is in the admin list
	for _, adminEmail := range adminEmails {
		if strings.EqualFold(userEmail, adminEmail) {
			return true
		}
	}

	return false
}

// getUserEmailByID retrieves a user's email from the database by their ID
func getUserEmailByID(userID string) (string, error) {
	query := `SELECT email FROM users WHERE id = $1`

	stmt, err := db.DB.Prepare(query)
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	var email string
	err = stmt.QueryRow(userID).Scan(&email)
	if err != nil {
		return "", err
	}

	return email, nil
}

// getAdminEmails returns a list of admin emails from environment configuration
// Expected format: ADMIN_EMAILS=admin1@company.com,admin2@company.com,admin3@company.com
func getAdminEmails() []string {
	adminEmailsEnv := getEnvWithDefault("ADMIN_EMAILS", "")

	if adminEmailsEnv == "" {
		return []string{}
	}

	// Split by comma and trim whitespace
	emails := strings.Split(adminEmailsEnv, ",")
	var adminEmails []string

	for _, email := range emails {
		trimmedEmail := strings.TrimSpace(email)
		if trimmedEmail != "" {
			adminEmails = append(adminEmails, trimmedEmail)
		}
	}

	return adminEmails
}
