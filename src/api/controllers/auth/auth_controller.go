package authCtrl

import (
	"net/http"

	"github.com/runabove/metronome/src/api/core"
	"github.com/runabove/metronome/src/api/core/io/in"
	"github.com/runabove/metronome/src/api/core/io/out"
	authSrv "github.com/runabove/metronome/src/api/services/auth"
	userSrv "github.com/runabove/metronome/src/api/services/user"
)

type accessTokenQuery struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AccessToken endoint handle token requests.
func AccessToken(w http.ResponseWriter, r *http.Request) {
	var accessTokenQuery accessTokenQuery

	body, err := in.JSON(r, &accessTokenQuery)
	if err != nil {
		out.JSON(w, 400, err)
		return
	}

	result := core.ValidateJSON("auth", "accessTokenQuery", string(body))
	if !result.Valid {
		out.JSON(w, 422, result.Errors)
		return
	}

	user := userSrv.Login(accessTokenQuery.Username, accessTokenQuery.Password)
	if user == nil {
		out.Unauthorized(w)
		return
	}
	token := authSrv.GenerateToken(user.ID, user.Roles)
	out.JSON(w, 200, token)
}
