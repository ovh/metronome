package oauth

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/runabove/metronome/src/api/models"
	"github.com/runabove/metronome/src/metronome/core"
)

// GenerateRefreshToken returns an Token
func GenerateRefreshToken(userID string, roles []string, typeToken string) (token *models.Token) {

	var salt string
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic("error reading Rand()")
	}
	// base 16, lower-case, two characters per byte
	salt = fmt.Sprintf("%x", b)

	// plaintext is composed of userID, a random salt and the timestamp
	var plaintext bytes.Buffer
	plaintext.WriteString(userID)
	plaintext.WriteString(salt)
	plaintext.WriteString(string(time.Now().Unix()))

	return &models.Token{
		Token:  core.Sha256(plaintext.String()),
		UserID: userID,
		Roles:  roles,
		Type:   typeToken,
	}
}
