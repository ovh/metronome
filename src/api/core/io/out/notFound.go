package out

import (
	"net/http"

	"github.com/runabove/metronome/src/api/models"
)

// NotFound perform a 404 not found HTTP response.
func NotFound(w http.ResponseWriter) {
	JSON(w, 404, models.Error{"Not found"})
}
