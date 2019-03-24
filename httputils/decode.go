package httputils

import (
	"encoding/json"
	"net/http"
)

func DecodeBody(r *http.Request, req interface{}) (interface{}, error) {
	var err error

	switch r.Method[0] {
	case 'D':
		err = json.NewDecoder(r.Body).Decode(req)
	case 'P':
		err = json.NewDecoder(r.Body).Decode(req)
	}

	return req, err
}
