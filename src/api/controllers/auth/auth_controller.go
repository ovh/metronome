package authctrl

import (
	"errors"
	"net/http"

	"github.com/ovh/metronome/src/api/core"
	"github.com/ovh/metronome/src/api/core/io/in"
	"github.com/ovh/metronome/src/api/core/io/out"
	"github.com/ovh/metronome/src/api/factories"
	authSrv "github.com/ovh/metronome/src/api/services/auth"
	userSrv "github.com/ovh/metronome/src/api/services/user"
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
		out.JSON(w, http.StatusBadRequest, factories.Error(err))
		return
	}

	authQueryResult, err := core.ValidateJSON("auth", "authQuery", string(body))
	if err != nil {
		out.JSON(w, http.StatusInternalServerError, factories.Error(err))
		return
	}

	if !authQueryResult.Valid {
		out.JSON(w, http.StatusUnprocessableEntity, authQueryResult.Errors)
		return
	}

	switch tokenQuery.Type {

	case "bearer":
		loginQueryResult, err := core.ValidateJSON("auth", "loginQuery", string(body))
		if err != nil {
			out.JSON(w, http.StatusInternalServerError, factories.Error(err))
			return
		}

		if !loginQueryResult.Valid {
			out.JSON(w, http.StatusUnprocessableEntity, loginQueryResult.Errors)
			return
		}

		user, err := userSrv.Login(tokenQuery.Username, tokenQuery.Password)
		if err != nil {
			out.JSON(w, http.StatusInternalServerError, factories.Error(err))
			return
		}

		if user == nil {
			out.JSON(w, http.StatusUnauthorized, factories.Error(err))
			return
		}

		token, err := authSrv.BearerTokensFromUser(user)
		if err != nil {
			out.JSON(w, http.StatusInternalServerError, factories.Error(err))
			return
		}

		out.JSON(w, http.StatusOK, token)

	case "access":
		accessQueryResult, err := core.ValidateJSON("auth", "accessQuery", string(body))
		if err != nil {
			out.JSON(w, http.StatusInternalServerError, factories.Error(err))
			return
		}

		if !accessQueryResult.Valid {
			out.JSON(w, http.StatusUnprocessableEntity, accessQueryResult.Errors)
			return
		}

		token, err := authSrv.BearerTokensFromRefresh(tokenQuery.RefreshToken)
		if err != nil {
			out.JSON(w, http.StatusUnauthorized, factories.Error(err))
			return
		}

		out.JSON(w, http.StatusOK, token)
	}
}

// LogoutHandler endoint remove RefreshTokens
func LogoutHandler(w http.ResponseWriter, r *http.Request) {

	token, err := authSrv.GetToken(r.Header.Get("Authorization"))
	if err != nil {
		out.JSON(w, http.StatusInternalServerError, factories.Error(err))
		return
	}

	if token == nil {
		out.JSON(w, http.StatusUnauthorized, factories.Error(errors.New("Unauthorized")))
		return
	}

	err = authSrv.RevokeRefreshTokenFromAccess(token)
	if err != nil {
		out.JSON(w, http.StatusInternalServerError, factories.Error(err))
		return
	}

	out.JSON(w, http.StatusOK, true)
}
