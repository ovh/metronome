package out

import (
	"net/http"

	"github.com/unrolled/render"
)

// JSON perform an http response with a JSON payload.
// status defined the http status code.
func JSON(w http.ResponseWriter, status int, v interface{}) {
	if err := render.New().JSON(w, status, v); err != nil {
		panic(err)
	}
}
