package taskctrl

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/ovh/metronome/src/api/core"
	"github.com/ovh/metronome/src/api/core/io/in"
	"github.com/ovh/metronome/src/api/core/io/out"
	"github.com/ovh/metronome/src/api/factories"
	authSrv "github.com/ovh/metronome/src/api/services/auth"
	taskSrv "github.com/ovh/metronome/src/api/services/task"
	"github.com/ovh/metronome/src/metronome/models"
)

// Create endoint handle task creation.
func Create(w http.ResponseWriter, r *http.Request) {
	token, err := authSrv.GetToken(r.Header.Get("Authorization"))
	if err != nil {
		out.JSON(w, http.StatusInternalServerError, factories.Error(err))
		return
	}

	if token == nil {
		out.JSON(w, http.StatusUnauthorized, factories.Error(errors.New("Unauthorized")))
		return
	}

	var task models.Task
	body, err := in.JSON(r, &task)
	if err != nil {
		out.JSON(w, http.StatusBadRequest, factories.Error(err))
		return
	}

	vErr := validateCreation(string(body))
	if vErr != nil {
		out.JSON(w, http.StatusUnprocessableEntity, vErr)
		return
	}

	task.UserID = authSrv.UserID(token)
	success := taskSrv.Create(&task)
	if !success {
		out.JSON(w, http.StatusBadGateway, factories.Error(errors.New("Bad gateway")))
		return
	}

	out.JSON(w, http.StatusOK, task)
}

// Schedule field regexp creation
// Don't forget to \\ all \ in the definitions.json
func validateCreation(payload string) interface{} {
	// schedule regex: https://regex101.com/r/jo8Kji/1/
	result, err := core.ValidateJSON("task", "create", payload)
	if err != nil {
		return factories.Error(err)
	}

	if !result.Valid {
		return result.Errors
	}

	return nil
}

// Delete endoint handle task deletion.
func Delete(w http.ResponseWriter, r *http.Request) {
	token, err := authSrv.GetToken(r.Header.Get("Authorization"))
	if err != nil {
		out.JSON(w, http.StatusInternalServerError, factories.Error(err))
		return
	}

	if token == nil {
		out.JSON(w, http.StatusUnauthorized, factories.Error(errors.New("Unauthorized")))
		return
	}

	success := taskSrv.Delete(mux.Vars(r)["id"], authSrv.UserID(token))
	if !success {
		out.JSON(w, http.StatusBadGateway, factories.Error(errors.New("Bad gateway")))
		return
	}

	w.WriteHeader(http.StatusOK)
}
