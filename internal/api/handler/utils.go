package handler

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse struct is used to standardize error output
type ErrorResponse struct {
	Error string `json:"error"`
}

// WriteError writes a standardized JSON error response with the given status code.
func WriteError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: msg})
}

// WriteJSON writes a standardized JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
