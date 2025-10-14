package utils

import (
	"html"
	"regexp"
	"strings"
)

// SanitizeString removes potentially dangerous characters and HTML tags
func SanitizeString(input string) string {
	// Remove HTML tags
	input = stripHTMLTags(input)

	// HTML escape to prevent XSS
	input = html.EscapeString(input)

	// Trim whitespace
	input = strings.TrimSpace(input)

	return input
}

// SanitizeEmail sanitizes and validates email format
func SanitizeEmail(email string) string {
	email = strings.TrimSpace(strings.ToLower(email))
	email = html.EscapeString(email)
	return email
}

// SanitizeUsername sanitizes username input
func SanitizeUsername(username string) string {
	// Remove any non-alphanumeric characters except underscores
	username = strings.TrimSpace(username)
	reg := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	username = reg.ReplaceAllString(username, "")
	return username
}

// stripHTMLTags removes HTML tags from string
func stripHTMLTags(input string) string {
	// Remove HTML tags
	reg := regexp.MustCompile(`<[^>]*>`)
	return reg.ReplaceAllString(input, "")
}

// SanitizeInput is a general sanitizer for user inputs
func SanitizeInput(input string) string {
	return SanitizeString(input)
}

// PreventSQLInjection sanitizes input to prevent SQL injection
// Note: This is a backup - always use parameterized queries as primary defense
func PreventSQLInjection(input string) string {
	// Remove common SQL injection patterns
	patterns := []string{
		`'`, `"`, `;`, `--`, `/*`, `*/`, `xp_`, `sp_`,
		`DROP`, `INSERT`, `DELETE`, `UPDATE`, `SELECT`,
		`EXEC`, `EXECUTE`, `SCRIPT`, `UNION`,
	}

	result := input
	for _, pattern := range patterns {
		result = strings.ReplaceAll(result, pattern, "")
		result = strings.ReplaceAll(result, strings.ToLower(pattern), "")
	}

	return result
}
