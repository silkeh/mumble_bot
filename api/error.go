package api

import (
	"encoding/json"
	"net/http"
)

const MethodNotAllowedError = "method not allowed"

type Error struct {
	Error string
}

func WriteError(w http.ResponseWriter, code int, error string) error {
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(&Error{Error: error})
}

func WriteMethodNotAllowed(w http.ResponseWriter) error {
	return WriteError(w, http.StatusMethodNotAllowed, MethodNotAllowedError)
}
