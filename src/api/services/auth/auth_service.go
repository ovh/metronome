// Package authSrv handle authorization token operations.
package authSrv

import (
	jwt "github.com/dgrijalva/jwt-go"

	"github.com/runabove/metronome/src/api/core/oauth"
	"github.com/runabove/metronome/src/api/models"
)

// GenerateToken return a new token.
func GenerateToken(userID string, roles []string) models.Token {
	return oauth.GenerateToken(userID, roles)
}

// GetToken return a token from a token string.
// Return nil if the token string is invalid or if the token as expired.
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
