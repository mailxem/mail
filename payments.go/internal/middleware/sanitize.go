package middleware

import (
	"html"
	"regexp"
	"strings"
)

// Sanitizer provides input sanitization functions
type Sanitizer struct{}

var (
	// SQL injection patterns (defense in depth, GORM already protects)
	sqlPattern = regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute|script|javascript|<script|onerror=|onload=)`)

	// XSS patterns
	xssPattern = regexp.MustCompile(`(?i)(<script|javascript:|onerror=|onload=|onclick=|<iframe|<object|<embed)`)

	// Path traversal patterns
	pathTraversalPattern = regexp.MustCompile(`\.\.\/|\.\.\\`)
)

// NewSanitizer creates a new sanitizer
func NewSanitizer() *Sanitizer {
	return &Sanitizer{}
}

// SanitizeString performs basic sanitization on string input
func (s *Sanitizer) SanitizeString(input string) string {
	// Trim whitespace
	input = strings.TrimSpace(input)

	// HTML escape
	input = html.EscapeString(input)

	return input
}

// SanitizeEmail validates and sanitizes email addresses
func (s *Sanitizer) SanitizeEmail(email string) string {
	email = strings.TrimSpace(email)
	email = strings.ToLower(email)
	return email
}

// SanitizeAlphanumeric keeps only alphanumeric characters and common separators
func (s *Sanitizer) SanitizeAlphanumeric(input string) string {
	// Keep only letters, numbers, spaces, hyphens, underscores
	reg := regexp.MustCompile(`[^a-zA-Z0-9\s\-_]`)
	return reg.ReplaceAllString(input, "")
}

// ValidateNoSQLInjection checks for SQL injection patterns (defense in depth)
func (s *Sanitizer) ValidateNoSQLInjection(input string) bool {
	return !sqlPattern.MatchString(input)
}

// ValidateNoXSS checks for XSS patterns
func (s *Sanitizer) ValidateNoXSS(input string) bool {
	return !xssPattern.MatchString(input)
}

// ValidateNoPathTraversal checks for path traversal patterns
func (s *Sanitizer) ValidateNoPathTraversal(input string) bool {
	return !pathTraversalPattern.MatchString(input)
}

// SanitizeID sanitizes UUIDs and IDs
func (s *Sanitizer) SanitizeID(id string) string {
	// Keep only valid UUID characters
	reg := regexp.MustCompile(`[^a-zA-Z0-9\-]`)
	return reg.ReplaceAllString(id, "")
}

// ValidateInput performs comprehensive validation
func (s *Sanitizer) ValidateInput(input string) error {
	if !s.ValidateNoSQLInjection(input) {
		return &ValidationError{Message: "input contains forbidden patterns"}
	}
	if !s.ValidateNoXSS(input) {
		return &ValidationError{Message: "input contains forbidden patterns"}
	}
	if !s.ValidateNoPathTraversal(input) {
		return &ValidationError{Message: "input contains forbidden patterns"}
	}
	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
