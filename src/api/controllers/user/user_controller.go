package userCtrl

import (
	"net/http"

	"github.com/runabove/metronome/src/api/core"
	"github.com/runabove/metronome/src/api/core/io/in"
	"github.com/runabove/metronome/src/api/core/io/out"
	"github.com/runabove/metronome/src/api/models"
	authSrv "github.com/runabove/metronome/src/api/services/auth"
	userSrv "github.com/runabove/metronome/src/api/services/user"
)

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

	duplicated := userSrv.Edit(authSrv.UserId(token), &user)
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

func Current(w http.ResponseWriter, r *http.Request) {
	token := authSrv.GetToken(r.Header.Get("Authorization"))
	if token == nil {
		out.Unauthorized(w)
		return
	}

	user := userSrv.Get(authSrv.UserId(token))
	if user == nil {
		out.NotFound(w)
		return
	}

	out.JSON(w, 200, user)
}
