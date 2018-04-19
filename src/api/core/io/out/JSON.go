package out

import (
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/unrolled/render"
)

// JSON perform an http response with a JSON payload.
// status defined the http status code.
func JSON(w http.ResponseWriter, status int, v interface{}) { // nolint: interfacer
	if err := render.New().JSON(w, status, v); err != nil {
		log.WithError(err).Error("Could not awnser to the request")
	}
}
