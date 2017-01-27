package oauth

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
)

// AuthClaims add roles to the jwt claims.
type AuthClaims struct {
	Roles        []string `json:"roles"`
	RefreshToken string   `json:"refreshToken"`
	jwt.StandardClaims
}

// GenerateAccessToken return a new token.
func GenerateAccessToken(userID string, roles []string, refreshToken string) string {
	claims := AuthClaims{
		roles,
		refreshToken,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Second * time.Duration(viper.GetInt("token.ttl"))).Unix(),
			IssuedAt:  time.Now().Unix(),
			Subject:   userID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	key, err := hex.DecodeString(viper.GetString("token.key"))
	if err != nil {
		panic(err)
	}

	tokenString, err := token.SignedString(key)
	if err != nil {
		panic(err)
	}

	return tokenString
}

// GetToken return a token from a token string.
// Return nil if the token string is invalid or if the token as expired.
func GetToken(tokenString string) *jwt.Token {
	token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		key, err := hex.DecodeString(viper.GetString("token.key"))
		if err != nil {
			panic(err)
		}
		return key, nil
	})

	if err == nil && token.Valid {
		return token
	}
	return nil
}

// UserID return the user id from a token.
func UserID(token *jwt.Token) string {
	claims := token.Claims.(*AuthClaims)
	return claims.Subject
}

// Roles return the roles from a token.
func Roles(token *jwt.Token) []string {
	claims := token.Claims.(*AuthClaims)
	return claims.Roles
}

// RefreshToken return the refreshToken
func RefreshToken(token *jwt.Token) string {
	claims := token.Claims.(*AuthClaims)
	refreshToken, err := base64.StdEncoding.DecodeString(claims.RefreshToken)
	if err != nil {
		panic("")
	}
	return string(refreshToken)
}
