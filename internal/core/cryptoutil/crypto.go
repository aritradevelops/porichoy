package cryptoutil

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateHash(length int) (string, error) {
	secretBytes := make([]byte, length)
	_, err := rand.Read(secretBytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(secretBytes), nil
}
