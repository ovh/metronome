package oauth

import (
	"encoding/base64"

	"github.com/ovh/metronome/src/api/models"
	uuid "github.com/satori/go.uuid"
)

// GenerateRefreshToken returns an Token
func GenerateRefreshToken(userID string, roles []string) (*models.Token, error) {
	return &models.Token{
		Token:  base64.StdEncoding.EncodeToString(uuid.NewV4().Bytes()),
		UserID: userID,
		Roles:  roles,
		Type:   "refresh",
	}, nil
}
