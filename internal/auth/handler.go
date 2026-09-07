package auth

import (
	"errors"
	"net/http"

	apperrors "github.com/mmk31585/workout-tracker/internal/app_errors"
	"github.com/mmk31585/workout-tracker/internal/server"
)


type AuthHandler struct {
	AuthService *AuthService
}

func NewAuthHandler(as *AuthService) *AuthHandler {
	return &AuthHandler{AuthService: as}
}

func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var payload SignupRequest
	if err := server.ReadJSON(w, r, &payload); err != nil {
		server.WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := server.Validate.Struct(payload); err != nil {
		server.WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	token, err := h.AuthService.Signup(r.Context(), payload.Email, payload.Password, payload.DisplayName)
	if err != nil {
		if errors.Is(err, apperrors.ErrConflict) {
			server.WriteJSONError(w, http.StatusConflict, err.Error())
			return
		}
		server.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	result := &AuthResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(h.AuthService.TokenExpiration().Seconds()),
	}
	server.JSON(w, http.StatusCreated, result)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var payload LoginRequest
	if err := server.ReadJSON(w, r, &payload); err != nil {
		server.WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := server.Validate.Struct(payload); err != nil {
		server.WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	token, err := h.AuthService.Login(r.Context(), payload.Email, payload.Password)
	if err != nil {
		server.WriteJSONError(w, http.StatusUnauthorized, err.Error())
		return
	}
	result := &AuthResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(h.AuthService.TokenExpiration().Seconds()),
	}
	server.JSON(w, http.StatusOK, result)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
