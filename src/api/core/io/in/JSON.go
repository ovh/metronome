package in

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

// JSON unmarshal a body HTTP request to an interface.
// It also implement body size limit
func JSON(r *http.Request, v interface{}) ([]byte, error) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}

	if err := json.Unmarshal(body, &v); err != nil {
		return body, err
	}
	return body, nil
}
