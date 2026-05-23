package auth

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/google/uuid"
)

func GenerateRefreshToken() (raw string, hash string) {
	raw = uuid.NewString() + uuid.NewString()
	hash = HashRefreshToken(raw)
	return raw, hash
}

func HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
