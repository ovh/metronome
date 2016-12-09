package routers

import (
	authCtrl "github.com/runabove/metronome/src/api/controllers/auth"
)

var AuthRoutes = Routes{
	Route{"Get access token", "POST", "/", authCtrl.AccessToken},
}
