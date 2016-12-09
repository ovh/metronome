package out

import (
	"net/http"

	"github.com/runabove/metronome/src/api/models"
)

func Unauthorized(w http.ResponseWriter) {
	JSON(w, 401, models.Error{"Unauthorized"})
}
