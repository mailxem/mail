package utils

import (
	"crypto/rand"
	"fmt"

	"github.com/golang-jwt/jwt/v4"
)

func ValidateRefreshToken(token string, secret string) (*jwt.RegisteredClaims, error) {
	claims := &jwt.RegisteredClaims{}
	_, _, err := new(jwt.Parser).ParseUnverified(token, claims)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// 🎲 GenerateRandomString generates a random string of specified length using crypto/rand
func GenerateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)

	// 🔒 Use crypto/rand for secure random generation
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random string: %w", err)
	}

	// 🔄 Map random bytes to charset
	for i := 0; i < length; i++ {
		b[i] = charset[b[i]%byte(len(charset))]
	}

	return string(b), nil
}
