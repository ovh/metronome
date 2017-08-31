// Package authSrv handle authorization token operations.
package authSrv

import (
	"encoding/base64"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/ovh/metronome/src/api/core/oauth"
	"github.com/ovh/metronome/src/api/models"
	"github.com/ovh/metronome/src/metronome/core"
	"github.com/ovh/metronome/src/metronome/pg"
)

// BearerTokensFromUser return both new Access and Refresh tokens.
func BearerTokensFromUser(user *models.User) *models.BearerToken {
	db := pg.DB()

	refreshToken := oauth.GenerateRefreshToken(user.ID, user.Roles, "refresh")

	// We need to forward the "plain" refreshToken to the client
	// and store an encrypted version into our DB
	refreshTokenPlain := refreshToken.Token
	refreshToken.Token = core.PBKDF2(refreshToken.Token, core.Sha256(user.ID))

	res, err := db.Model(&refreshToken).Insert()
	if err != nil || res.RowsAffected() == 0 {
		panic(err)
	}

	accessToken := oauth.GenerateAccessToken(user.ID, user.Roles, refreshTokenPlain)

	return &models.BearerToken{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenPlain,
		Type:         "bearer",
	}
}

// GetToken return a token from a accessToken string.
// Return nil if the accessToken string is invalid or if the token as expired.
func GetToken(tokenString string) *jwt.Token {
	return oauth.GetToken(tokenString)
}

// UserID return the user id from a token.
func UserID(token *jwt.Token) string {
	return oauth.UserID(token)
}

// Roles return the roles from a token.
func Roles(token *jwt.Token) []string {
	return oauth.Roles(token)
}

// HasRole check if the token as a role.
func HasRole(role string, token *jwt.Token) bool {
	for _, r := range Roles(token) {
		if r == role {
			return true
		}
	}
	return false
}

// getRefreshToken return a RefreshToken from PG
// Returns nil if empty
func getRefreshTokenFromDB(prettyRefreshToken string) (*models.Token, error) {
	db := pg.DB()

	// We need to decode refreshToken
	refreshToken, err := base64.StdEncoding.DecodeString(prettyRefreshToken)
	if err != nil {
		panic("")
	}

	var token models.Token
	err = db.Model(&token).Where("token = ? AND type = 'refresh'", string(refreshToken)).Select()
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// BearerTokensFromRefresh return a new AccessToken
func BearerTokensFromRefresh(prettyRefreshToken string) (*models.BearerToken, error) {

	token, err := getRefreshTokenFromDB(prettyRefreshToken)

	if err == pg.ErrNoRows {
		return nil, err
	}
	if err != nil {
		panic("")
	}

	return &models.BearerToken{
		AccessToken:  oauth.GenerateAccessToken(token.UserID, token.Roles, prettyRefreshToken),
		Type:         "bearer",
		RefreshToken: prettyRefreshToken,
	}, nil
}

// RevokeRefreshTokenFromAccess remove a RefreshToken from DB from an accessToken
func RevokeRefreshTokenFromAccess(token *jwt.Token) error {
	db := pg.DB()

	refreshTokenPlain := oauth.RefreshToken(token)
	users := models.Users{}
	err := db.Model(&users).Where("user_id = ?", oauth.UserID(token)).Select()
	if err != nil {
		panic(err)
	}

	if len(users) == 0 {
		return nil
	}
	user := users[0]

	// We need to regenerate the salt
	ciphertext := core.PBKDF2(refreshTokenPlain, core.Sha256(user.ID))
	var refreshToken models.Token
	_, err = db.Model(&refreshToken).Where("token = ? AND type = 'refresh'", ciphertext).Delete()

	if err != nil {
		return err
	}
	return nil
}
