package routers

import (
	tasksCtrl "github.com/runabove/metronome/src/api/controllers/tasks"
)

var TasksRoutes = Routes{
	Route{"Get tasks", "GET", "/", tasksCtrl.All},
}
