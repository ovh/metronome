package core

import (
	"crypto/sha1"
	"encoding/hex"

	"golang.org/x/crypto/pbkdf2"
)

var (
	// passwordSecurityIterations is based on New NIST guidelines  (Update August 2016)
	// https://pages.nist.gov/800-63-3/sp800-63b.html#sec5
	passwordSecurityIterations = 10000
	passwordSecurityKeylen     = 512
)

// PBKDF2 is hashing using pbkdf2 method
func PBKDF2(str string, salt string) string {
	hashedPassword := pbkdf2.Key([]byte(str), []byte(salt), passwordSecurityIterations, passwordSecurityKeylen, sha1.New)
	return hex.EncodeToString(hashedPassword)
}
