package health

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/mmk31585/workout-tracker/internal/config"
	"github.com/mmk31585/workout-tracker/internal/storage"
)

type response struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}



// healthCtxKey is the context key for storing storage.Storage.
type healthCtxKey struct{}

// WithStorage adds the storage.Storage to the context.
func WithStorage(ctx context.Context, s *storage.Storage) context.Context {
	return context.WithValue(ctx, healthCtxKey{}, s)
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) error {
	return writeJSON(w, status, response{Error: msg})
}

func writeJSONData(w http.ResponseWriter, status int, data any) error {
	return writeJSON(w, status, response{Data: data})
}

// Health returns the basic health status of the service.
func Health(w http.ResponseWriter, r *http.Request) {
	writeJSONData(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"version": config.Version,
	})
}

// Live returns the liveness status of the service.
func Live(w http.ResponseWriter, r *http.Request) {
	writeJSONData(w, http.StatusOK, map[string]string{
		"status": "live",
	})
}

// Ready returns the readiness status after checking dependencies.
func Ready(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	s, ok := ctx.Value(healthCtxKey{}).(*storage.Storage)
	if !ok || s == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "storage not configured")
		return
	}

	if err := s.DB().PingContext(ctx); err != nil {
		writeJSONError(w, http.StatusServiceUnavailable, err.Error())
		return
	}

	writeJSONData(w, http.StatusOK, map[string]string{
		"status": "ready",
	})
}