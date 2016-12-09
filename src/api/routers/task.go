package routers

import (
	taskCtrl "github.com/runabove/metronome/src/api/controllers/task"
)

var TaskRoutes = Routes{
	Route{"Create task", "POST", "/", taskCtrl.Create},
	Route{"Delete task", "DELETE", "/{id:\\S{1,256}}", taskCtrl.Delete},
}
