package util

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateRandomState() (string, error) {
	b := make([]byte, 32) // 256 bits of entropy
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
