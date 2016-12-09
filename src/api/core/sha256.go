package core

import (
	"crypto/sha256"
	"encoding/hex"
)

func Sha256(in string) string {
	key := []byte(in)
	hash := sha256.Sum256(key)
	return hex.EncodeToString(hash[:])
}
