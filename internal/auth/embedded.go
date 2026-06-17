package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
)

const (
	// DefaultEmbeddedAPIToken is the development embedded token. Replace in production.
	DefaultEmbeddedAPIToken = "pk_dev_a8f3c2e1b9d74f6a0e5c3b9d2f7a1e4c8b6d0f3a7e2c9b5d1f8a4e6c0b3d7f9"
)

func ValidateBearerToken(tokenString string, tokenIssuer *TokenIssuer, embeddedAPIToken string) error {
	if embeddedAPIToken != "" && subtle.ConstantTimeCompare([]byte(tokenString), []byte(embeddedAPIToken)) == 1 {
		return nil
	}

	return tokenIssuer.ValidateToken(tokenString)
}

func GenerateEmbeddedAPIToken() (string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate embedded api token: %w", err)
	}

	return "pk_" + hex.EncodeToString(randomBytes), nil
}
