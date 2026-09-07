package servermiddleware

import (
	"context"
	"net/http"
	"strings"

	"encoding/json"
)

type TokenValidator interface {
	ValidateToken(token string) (string, error)
}
type AuthMiddleware struct {
	Validator TokenValidator
}
type envelope struct {
	Error string `json:"error,omitempty"`
}

func NewAuthMiddlware(validator TokenValidator) *AuthMiddleware{
	return &AuthMiddleware{
		Validator: validator,
	}
}

func writeJSONError(w http.ResponseWriter, status int, errMsg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(envelope{Error: errMsg})
}

func (m *AuthMiddleware) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		token := parts[1]
		userID, err := m.Validator.ValidateToken(token)
		if err != nil {
			writeJSONError(w, http.StatusUnauthorized, err.Error())
			return
		}
		ctx := context.WithValue(r.Context(), "user_id", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
