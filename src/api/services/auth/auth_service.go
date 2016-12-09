package authSrv

import (
	jwt "github.com/dgrijalva/jwt-go"

	"github.com/runabove/metronome/src/api/core/oauth"
	"github.com/runabove/metronome/src/api/models"
)

func GenerateToken(userID string, roles []string) models.Token {
	return oauth.GenerateToken(userID, roles)
}

func GetToken(tokenString string) *jwt.Token {
	return oauth.GetToken(tokenString)
}

func UserId(token *jwt.Token) string {
	return oauth.UserID(token)
}

func Roles(token *jwt.Token) []string {
	return oauth.Roles(token)
}

func AsRole(role string, token *jwt.Token) bool {
	for _, r := range Roles(token) {
		if r == role {
			return true
		}
	}
	return false
}
