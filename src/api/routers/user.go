package routers

import (
	userCtrl "github.com/runabove/metronome/src/api/controllers/user"
)

// UserRoutes defined user endpoints.
var UserRoutes = Routes{
	Route{"Create a user", "POST", "/", userCtrl.Create},
	Route{"Edit a user", "PATCH", "/", userCtrl.Edit},
	Route{"Retrive current user", "GET", "/", userCtrl.Current},
}
