package httputils

import (
	"encoding/json"
	"net/http"
)

// StatusHandler is for any http status code.
type StatusHandler struct {
	Err  error
	Code int
}

// ServeHTTP is http.Handler insterface implementation.
func (s StatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	WriteError(s.Err, s.Code, w)
}

// WriteError writes error on ResponseWriter
func WriteError(err error, status int, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
