package routers

import (
	authCtrl "github.com/ovh/metronome/src/api/controllers/auth"
)

// AuthRoutes defined auth endpoints
var AuthRoutes = Routes{
	Route{"Get access token", "POST", "/", authCtrl.AuthHandler},
	Route{"Logoff a user", "POST", "/logout", authCtrl.LogoutHandler},
}
