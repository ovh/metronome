package userCtrl

import (
	"net/http"

	"github.com/ovh/metronome/src/api/core"
	"github.com/ovh/metronome/src/api/core/io/in"
	"github.com/ovh/metronome/src/api/core/io/out"
	"github.com/ovh/metronome/src/api/models"
	authSrv "github.com/ovh/metronome/src/api/services/auth"
	userSrv "github.com/ovh/metronome/src/api/services/user"
)

// Create endpoint handle the user account creation.
func Create(w http.ResponseWriter, r *http.Request) {
	var user models.User

	body, err := in.JSON(r, &user)
	if err != nil {
		out.JSON(w, 400, err)
		return
	}

	result := core.ValidateJSON("user", "create", string(body))
	if !result.Valid {
		out.JSON(w, 422, result.Errors)
		return
	}

	duplicated := userSrv.Create(&user)
	if duplicated {
		var errs []core.JSONSchemaErr
		errs = append(errs, core.JSONSchemaErr{
			"name",
			"duplicated",
			"name is duplicated",
		})
		out.JSON(w, 422, errs)
		return
	}

	out.JSON(w, 200, user)
}

// Edit endoint handle user edit.
func Edit(w http.ResponseWriter, r *http.Request) {
	token := authSrv.GetToken(r.Header.Get("Authorization"))
	if token == nil {
		out.Unauthorized(w)
		return
	}

	var user models.User

	body, err := in.JSON(r, &user)
	if err != nil {
		out.JSON(w, 400, err)
		return
	}

	result := core.ValidateJSON("user", "edit", string(body))
	if !result.Valid {
		out.JSON(w, 422, result.Errors)
		return
	}

	duplicated := userSrv.Edit(authSrv.UserID(token), &user)
	if duplicated {
		var errs []core.JSONSchemaErr
		errs = append(errs, core.JSONSchemaErr{
			"name",
			"duplicated",
			"name is duplicated",
		})
		out.JSON(w, 422, errs)
		return
	}

	out.Success(w)
}

// Current endoint return the user bind to the token.
func Current(w http.ResponseWriter, r *http.Request) {
	token := authSrv.GetToken(r.Header.Get("Authorization"))
	if token == nil {
		out.Unauthorized(w)
		return
	}

	user := userSrv.Get(authSrv.UserID(token))
	if user == nil {
		out.NotFound(w)
		return
	}

	out.JSON(w, 200, user)
}
