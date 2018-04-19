package userctrl

import (
	"errors"
	"net/http"

	"github.com/ovh/metronome/src/api/core"
	"github.com/ovh/metronome/src/api/core/io/in"
	"github.com/ovh/metronome/src/api/core/io/out"
	"github.com/ovh/metronome/src/api/factories"
	"github.com/ovh/metronome/src/api/models"
	authSrv "github.com/ovh/metronome/src/api/services/auth"
	userSrv "github.com/ovh/metronome/src/api/services/user"
)

// Create endpoint handle the user account creation.
func Create(w http.ResponseWriter, r *http.Request) {
	var user models.User

	body, err := in.JSON(r, &user)
	if err != nil {
		out.JSON(w, http.StatusBadRequest, err)
		return
	}

	result, err := core.ValidateJSON("user", "create", string(body))
	if err != nil {
		out.JSON(w, http.StatusInternalServerError, factories.Error(err))
		return
	}

	if !result.Valid {
		out.JSON(w, http.StatusUnprocessableEntity, result.Errors)
		return
	}

	duplicated, err := userSrv.Create(&user)
	if err != nil {
		out.JSON(w, http.StatusInternalServerError, factories.Error(err))
		return
	}

	if duplicated {
		var errs []core.JSONSchemaErr
		errs = append(errs, core.JSONSchemaErr{
			Field:       "name",
			Type:        "duplicated",
			Description: "name is duplicated",
		})

		out.JSON(w, http.StatusUnprocessableEntity, errs)
		return
	}

	out.JSON(w, http.StatusOK, user)
}

// Edit endoint handle user edit.
func Edit(w http.ResponseWriter, r *http.Request) {
	token, err := authSrv.GetToken(r.Header.Get("Authorization"))
	if err != nil {
		out.JSON(w, http.StatusInternalServerError, factories.Error(err))
		return
	}

	if token == nil {
		out.JSON(w, http.StatusUnauthorized, factories.Error(errors.New("Unauthorized")))
		return
	}

	var user models.User

	body, err := in.JSON(r, &user)
	if err != nil {
		out.JSON(w, http.StatusBadRequest, factories.Error(err))
		return
	}

	result, err := core.ValidateJSON("user", "edit", string(body))
	if err != nil {
		out.JSON(w, http.StatusInternalServerError, factories.Error(err))
		return
	}

	if !result.Valid {
		out.JSON(w, http.StatusUnprocessableEntity, result.Errors)
		return
	}

	duplicated, err := userSrv.Edit(authSrv.UserID(token), &user)
	if err != nil {
		out.JSON(w, http.StatusInternalServerError, factories.Error(err))
		return
	}

	if duplicated {
		var errs []core.JSONSchemaErr
		errs = append(errs, core.JSONSchemaErr{
			Field:       "name",
			Type:        "duplicated",
			Description: "name is duplicated",
		})

		out.JSON(w, http.StatusUnprocessableEntity, errs)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Current endoint return the user bind to the token.
func Current(w http.ResponseWriter, r *http.Request) {
	token, err := authSrv.GetToken(r.Header.Get("Authorization"))
	if err != nil {
		out.JSON(w, http.StatusInternalServerError, factories.Error(err))
		return
	}

	if token == nil {
		out.JSON(w, http.StatusUnauthorized, factories.Error(errors.New("Unauthorized")))
		return
	}

	user, err := userSrv.Get(authSrv.UserID(token))
	if err != nil {
		out.JSON(w, http.StatusInternalServerError, err)
		return
	}

	if user == nil {
		out.JSON(w, http.StatusNotFound, factories.Error(errors.New("Not found")))
		return
	}

	out.JSON(w, http.StatusOK, user)
}
