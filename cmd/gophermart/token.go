package main

import (
	"crypto/rand"
	"crypto/sha256"
)

func generateRandomToken() ([]byte, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	token := sha256.Sum256(b)
	return token[:], nil
}
