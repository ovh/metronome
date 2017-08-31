package routers

import (
	wsCtrl "github.com/ovh/metronome/src/api/controllers/ws"
)

// WsRoutes defined websockets endpoints.
var WsRoutes = Routes{
	Route{"Websocket", "GET", "/", wsCtrl.Join},
}
