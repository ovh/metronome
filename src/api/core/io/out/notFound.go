package out

import (
	"net/http"

	"github.com/runabove/metronome/src/api/models"
)

func NotFound(w http.ResponseWriter) {
	JSON(w, 404, models.Error{"Not found"})
}
