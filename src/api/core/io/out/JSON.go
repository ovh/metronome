package out

import (
	"net/http"

	"github.com/unrolled/render"
)

func JSON(w http.ResponseWriter, status int, v interface{}) {
	if err := render.New().JSON(w, status, v); err != nil {
		panic(err)
	}
}
