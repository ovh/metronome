package core

import (
	"crypto/sha256"
	"encoding/hex"
)

// Sha256 hash a string in a sha256 way.
func Sha256(in string) string {
	key := []byte(in)
	hash := sha256.Sum256(key)
	return hex.EncodeToString(hash[:])
}
