package oauth

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/d33d33/viper" // FIXME https://github.com/spf13/viper/pull/285
	jwt "github.com/dgrijalva/jwt-go"

	"github.com/runabove/metronome/src/api/models"
)

type AuthClaims struct {
	Roles []string `json:"roles"`
	jwt.StandardClaims
}

func GenerateToken(userID string, roles []string) models.Token {
	claims := AuthClaims{
		roles,
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

	return models.Token{tokenString}
}

func GetToken(tokenString string) *jwt.Token {
	token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		} else {
			key, err := hex.DecodeString(viper.GetString("token.key"))
			if err != nil {
				panic(err)
			}
			return key, nil
		}
	})

	if err == nil && token.Valid {
		return token
	}
	return nil
}

func UserID(token *jwt.Token) string {
	claims := token.Claims.(*AuthClaims)
	return claims.Subject
}

func Roles(token *jwt.Token) []string {
	claims := token.Claims.(*AuthClaims)
	return claims.Roles
}
