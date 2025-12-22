package idgen

import (
	"crypto/rand"
	"fmt"
)

// GenerateSecureID generates a cryptographically secure ID with the given prefix and length.
// Uses only alphanumeric characters (0-9, a-z) - no dashes or special characters.
func GenerateSecureID(prefix string, length int) (string, error) {
	// Use larger byte array for better entropy
	bytes := make([]byte, length*2)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Generate alphanumeric string (numbers and lowercase letters only)
	const charset = "0123456789abcdefghijklmnopqrstuvwxyz"
	encoded := make([]byte, length)
	for i := 0; i < length; i++ {
		encoded[i] = charset[bytes[i]%36] // 36 = len(charset)
	}

	return fmt.Sprintf("%s_%s", prefix, string(encoded)), nil
}
