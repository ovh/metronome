package routers

import (
	wsCtrl "github.com/runabove/metronome/src/api/controllers/ws"
)

// WsRoutes defined websockets endpoints.
var WsRoutes = Routes{
	Route{"Websocket", "GET", "/", wsCtrl.Join},
}
