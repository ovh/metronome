package authCtrl

import (
	"net/http"

	"github.com/runabove/metronome/src/api/core"
	"github.com/runabove/metronome/src/api/core/io/in"
	"github.com/runabove/metronome/src/api/core/io/out"
	authSrv "github.com/runabove/metronome/src/api/services/auth"
	userSrv "github.com/runabove/metronome/src/api/services/user"
)

type tokenQuery struct {
	Type         string `json:"type"`
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
	AccessToken  string `json:"accessToken,omitempty"`
}

// AuthHandler endoint handle token requests.
func AuthHandler(w http.ResponseWriter, r *http.Request) {
	var tokenQuery tokenQuery

	body, err := in.JSON(r, &tokenQuery)
	if err != nil {
		out.JSON(w, 400, err)
		return
	}

	authQueryResult := core.ValidateJSON("auth", "authQuery", string(body))
	if !authQueryResult.Valid {
		out.JSON(w, 422, authQueryResult.Errors)
		return
	}

	switch tokenQuery.Type {

	case "bearer":
		loginQueryResult := core.ValidateJSON("auth", "loginQuery", string(body))
		if !loginQueryResult.Valid {
			out.JSON(w, 422, loginQueryResult.Errors)
			return
		}

		user := userSrv.Login(tokenQuery.Username, tokenQuery.Password)
		if user == nil {
			out.Unauthorized(w)
			return
		}
		token := authSrv.BearerTokensFromUser(user)
		out.JSON(w, 200, token)

	case "access":
		accessQueryResult := core.ValidateJSON("auth", "accessQuery", string(body))
		if !accessQueryResult.Valid {
			out.JSON(w, 422, accessQueryResult.Errors)
			return
		}
		token, err := authSrv.BearerTokensFromRefresh(tokenQuery.RefreshToken)
		if err != nil {
			out.Unauthorized(w)
		}
		out.JSON(w, 200, token)
	}
}

// LogoutHandler endoint remove RefreshTokens
func LogoutHandler(w http.ResponseWriter, r *http.Request) {

	token := authSrv.GetToken(r.Header.Get("Authorization"))
	if token == nil {
		out.Unauthorized(w)
		return
	}

	err := authSrv.RevokeRefreshTokenFromAccess(token)
	if err != nil {
		out.JSON(w, 500, err.Error())
		return
	}

	out.JSON(w, 200, true)
}
