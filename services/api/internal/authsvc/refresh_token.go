package authsvc

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

func generateRefreshToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}

// refresh token disimpan sebagai hash di DB (bukan plaintext), mirip password.
func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
