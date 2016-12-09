package out

import (
	"net/http"

	"github.com/runabove/metronome/src/api/models"
)

func BadGateway(w http.ResponseWriter) {
	JSON(w, 502, models.Error{"Bad gateway"})
}
