package tasksctrl

import (
	"errors"
	"net/http"

	"github.com/ovh/metronome/src/api/core/io/out"
	"github.com/ovh/metronome/src/api/factories"
	authSrv "github.com/ovh/metronome/src/api/services/auth"
	tasksSrv "github.com/ovh/metronome/src/api/services/tasks"
)

// All endoint return the user tasks.
func All(w http.ResponseWriter, r *http.Request) {
	token, err := authSrv.GetToken(r.Header.Get("Authorization"))
	if err != nil {
		out.JSON(w, http.StatusInternalServerError, factories.Error(err))
		return
	}

	if token == nil {
		out.JSON(w, http.StatusUnauthorized, factories.Error(errors.New("Unauthorized")))
		return
	}

	tasks, err := tasksSrv.All(authSrv.UserID(token))
	if err != nil {
		out.JSON(w, http.StatusInternalServerError, factories.Error(err))
		return
	}

	out.JSON(w, http.StatusOK, tasks)
}
