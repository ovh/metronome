package out

import (
	"net/http"
)

// Success perform a 200 success HTTP response.
func Success(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}
