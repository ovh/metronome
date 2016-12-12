package routers

import (
	authCtrl "github.com/runabove/metronome/src/api/controllers/auth"
)

// AuthRoutes defined auth endpoints
var AuthRoutes = Routes{
	Route{"Get access token", "POST", "/", authCtrl.AccessToken},
}
