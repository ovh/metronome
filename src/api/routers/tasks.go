package routers

import (
	tasksCtrl "github.com/ovh/metronome/src/api/controllers/tasks"
)

// TasksRoutes defined tasks endpoints.
var TasksRoutes = Routes{
	Route{"Get tasks", "GET", "/", tasksCtrl.All},
}
