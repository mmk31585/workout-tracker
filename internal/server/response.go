package server

import (
	"encoding/json"
	"net/http"
)

type envelope struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

// WriteJSON writes the given value as a JSON response with the specified status code.
// It sets the Content-Type header to application/json; charset=utf-8.
func WriteJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

// WriteJSONError writes a JSON error response with the given error message.
func WriteJSONError(w http.ResponseWriter, status int, err string) error {
	return WriteJSON(w, status, envelope{Error: err})
}

// ReadJSON reads and validates the JSON request body into the given value.
// It enforces a maximum body size of 1MB and disallows unknown fields.
// Returns an error if the JSON is invalid, cannot be decoded, or validation fails.
func ReadJSON(w http.ResponseWriter, r *http.Request, data any) error {
	const maxBytes = 1_048_576 // 1 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(data)
}

// JSON writes a JSON response with the standard envelope format {data: ...}.
func JSON(w http.ResponseWriter, status int, data any) error {
	return WriteJSON(w, status, envelope{Data: data})
}
