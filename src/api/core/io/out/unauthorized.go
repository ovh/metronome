package out

import (
	"net/http"

	"github.com/ovh/metronome/src/api/models"
)

// Unauthorized perform a 401 unauthorized HTTP response.
func Unauthorized(w http.ResponseWriter) {
	JSON(w, 401, models.Error{"Unauthorized"})
}
