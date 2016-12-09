package out

import (
	"net/http"
)

func Success(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}
