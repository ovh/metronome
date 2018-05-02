// Package authsrv handle authorization token operations.
package authsrv

import (
	"errors"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/ovh/metronome/src/api/core/oauth"
	"github.com/ovh/metronome/src/api/models"
	"github.com/ovh/metronome/src/metronome/core"
	"github.com/ovh/metronome/src/metronome/pg"
)

// BearerTokensFromUser return both new Access and Refresh tokens.
func BearerTokensFromUser(user *models.User) (*models.BearerToken, error) {
	db := pg.DB()

	refreshToken, err := oauth.GenerateRefreshToken(user.ID, user.Roles)
	if err != nil {
		return nil, err
	}

	// We need to forward the refresh token to the client
	// and store an encrypted version into our DB
	res, err := db.Model(refreshToken).Insert()
	if err != nil || res.RowsAffected() == 0 {
		return nil, err
	}

	accessToken, err := oauth.GenerateAccessToken(user.ID, user.Roles, refreshToken.Token)
	if err != nil {
		return nil, err
	}

	return &models.BearerToken{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.Token,
		Type:         "bearer",
	}, nil
}

// GetToken return a token from a accessToken string.
// Return nil if the accessToken string is invalid or if the token as expired.
func GetToken(tokenString string) (*jwt.Token, error) {
	if strings.HasPrefix(tokenString, "Bearer ") {
		return oauth.GetToken(tokenString[7:])
	}

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
func getRefreshTokenFromDB(refreshToken string) (*models.Token, error) {
	db := pg.DB()

	token := new(models.Token)
	err := db.Model(token).Where("token = ? AND type = 'refresh'", refreshToken).Select()
	if err != nil {
		return nil, err
	}

	return token, nil
}

// BearerTokensFromRefresh return a new AccessToken
func BearerTokensFromRefresh(refreshToken string) (*models.BearerToken, error) {
	token, err := getRefreshTokenFromDB(refreshToken)
	if err == pg.ErrNoRows {
		return nil, errors.New("No such refresh token")
	}

	if err != nil {
		return nil, err
	}

	accessToken, err := oauth.GenerateAccessToken(token.UserID, token.Roles, refreshToken)
	if err != nil {
		return nil, err
	}

	return &models.BearerToken{
		AccessToken:  accessToken,
		Type:         "bearer",
		RefreshToken: refreshToken,
	}, nil
}

// RevokeRefreshTokenFromAccess remove a RefreshToken from DB from an accessToken
func RevokeRefreshTokenFromAccess(token *jwt.Token) error {
	db := pg.DB()

	refreshTokenPlain, err := oauth.RefreshToken(token)
	if err != nil {
		return err
	}

	users := models.Users{}
	err = db.Model(&users).Where("user_id = ?", oauth.UserID(token)).Select()
	if err != nil {
		return err
	}

	if len(users) == 0 {
		return nil
	}
	user := users[0]

	// We need to regenerate the salt
	ciphertext := core.PBKDF2(string(refreshTokenPlain), core.Sha256(user.ID))
	var refreshToken models.Token
	_, err = db.Model(&refreshToken).Where("token = ? AND type = 'refresh'", ciphertext).Delete()
	return err
}
