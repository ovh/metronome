package tasksCtrl

import (
	"net/http"

	"github.com/ovh/metronome/src/api/core/io/out"
	authSrv "github.com/ovh/metronome/src/api/services/auth"
	tasksSrv "github.com/ovh/metronome/src/api/services/tasks"
)

// All endoint return the user tasks.
func All(w http.ResponseWriter, r *http.Request) {
	token := authSrv.GetToken(r.Header.Get("Authorization"))
	if token == nil {
		out.Unauthorized(w)
		return
	}

	tasks := tasksSrv.All(authSrv.UserID(token))
	out.JSON(w, 200, tasks)
}
