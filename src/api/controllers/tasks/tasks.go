package tasksCtrl

import (
	"net/http"

	"github.com/runabove/metronome/src/api/core/io/out"
	authSrv "github.com/runabove/metronome/src/api/services/auth"
	tasksSrv "github.com/runabove/metronome/src/api/services/tasks"
)

func All(w http.ResponseWriter, r *http.Request) {
	token := authSrv.GetToken(r.Header.Get("Authorization"))
	if token == nil {
		out.Unauthorized(w)
		return
	}

	tasks := tasksSrv.All(authSrv.UserId(token))
	out.JSON(w, 200, tasks)
}
