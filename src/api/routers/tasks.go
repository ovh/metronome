package routers

import (
	tasksCtrl "github.com/runabove/metronome/src/api/controllers/tasks"
)

// TasksRoutes defined tasks endpoints.
var TasksRoutes = Routes{
	Route{"Get tasks", "GET", "/", tasksCtrl.All},
}
