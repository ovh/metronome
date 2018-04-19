// Package routers defined the api endpoints
package routers

import (
	"net/http"
	"path"

	"github.com/gorilla/mux"
)

// Route defined an http endpoint.
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// Routes defined multiple http endoints.
type Routes []Route

// InitRoutes bind the http endpoints.
func InitRoutes() *mux.Router {
	router := mux.NewRouter()
	bind(router, "/task", TaskRoutes)
	bind(router, "/tasks", TasksRoutes)
	bind(router, "/auth", AuthRoutes)
	bind(router, "/user", UserRoutes)
	bind(router, "/ws", WsRoutes)
	return router
}

func bind(router *mux.Router, base string, routes Routes) {
	for _, route := range routes {
		p := path.Join(base, route.Pattern)

		router.
			Methods(route.Method).
			Path(p).
			Name(route.Name).
			HandlerFunc(route.HandlerFunc)

		if p != "/" {
			router.
				Methods(route.Method).
				Path(p + "/").
				Name(route.Name).
				HandlerFunc(route.HandlerFunc)
		}
	}
}
