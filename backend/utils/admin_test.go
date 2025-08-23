package utils

import (
	"os"
	"testing"
)

func TestGetAdminEmails(t *testing.T) {
	tests := []struct {
		name           string
		envValue       string
		expectedEmails []string
	}{
		{
			name:           "Empty admin emails",
			envValue:       "",
			expectedEmails: []string{},
		},
		{
			name:           "Single admin email",
			envValue:       "admin@company.com",
			expectedEmails: []string{"admin@company.com"},
		},
		{
			name:           "Multiple admin emails",
			envValue:       "admin1@company.com,admin2@company.com,admin3@company.com",
			expectedEmails: []string{"admin1@company.com", "admin2@company.com", "admin3@company.com"},
		},
		{
			name:           "Admin emails with spaces",
			envValue:       " admin1@company.com , admin2@company.com , admin3@company.com ",
			expectedEmails: []string{"admin1@company.com", "admin2@company.com", "admin3@company.com"},
		},
		{
			name:           "Admin emails with empty values",
			envValue:       "admin1@company.com,,admin3@company.com",
			expectedEmails: []string{"admin1@company.com", "admin3@company.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the environment variable
			if tt.envValue != "" {
				os.Setenv("ADMIN_EMAILS", tt.envValue)
			} else {
				os.Unsetenv("ADMIN_EMAILS")
			}

			// Clean up after test
			defer func() {
				os.Unsetenv("ADMIN_EMAILS")
			}()

			result := getAdminEmails()

			if len(result) != len(tt.expectedEmails) {
				t.Errorf("Expected %d emails, got %d", len(tt.expectedEmails), len(result))
				return
			}

			for i, expected := range tt.expectedEmails {
				if result[i] != expected {
					t.Errorf("Expected email at index %d to be '%s', got '%s'", i, expected, result[i])
				}
			}
		})
	}
}

func TestCheckIfAdmin_NoDatabase(t *testing.T) {
	// This test only checks the logic without database interaction
	// We can't easily test the full function without a test database

	// Set up admin emails
	os.Setenv("ADMIN_EMAILS", "admin@company.com,superuser@company.com")
	defer os.Unsetenv("ADMIN_EMAILS")

	adminEmails := getAdminEmails()
	expectedEmails := []string{"admin@company.com", "superuser@company.com"}

	if len(adminEmails) != len(expectedEmails) {
		t.Errorf("Expected %d admin emails, got %d", len(expectedEmails), len(adminEmails))
	}

	for i, expected := range expectedEmails {
		if adminEmails[i] != expected {
			t.Errorf("Expected admin email at index %d to be '%s', got '%s'", i, expected, adminEmails[i])
		}
	}
}
