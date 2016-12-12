package out

import (
	"net/http"

	"github.com/runabove/metronome/src/api/models"
)

// BadGateway perform a 502 bad gateway HTTP response.
func BadGateway(w http.ResponseWriter) {
	JSON(w, 502, models.Error{"Bad gateway"})
}
