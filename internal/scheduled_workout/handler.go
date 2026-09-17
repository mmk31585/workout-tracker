package scheduledworkout

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	apperrors "github.com/mmk31585/workout-tracker/internal/app_errors"
	"github.com/mmk31585/workout-tracker/internal/server"
	servermiddleware "github.com/mmk31585/workout-tracker/internal/server/server_middleware"
)

type ScheduledWorkoutHandlers struct {
	Service *ScheduledWorkoutService
}

func NewScheduledWorkoutHandler(s *ScheduledWorkoutService) *ScheduledWorkoutHandlers {
	return &ScheduledWorkoutHandlers{Service: s}
}

func (h *ScheduledWorkoutHandlers) ScheduleWorkout(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	planID := chi.URLParam(r, "planId")
	if planID == "" {
		server.WriteJSONError(w, http.StatusBadRequest, "missing plan id")
		return
	}

	var req ScheduleRequest
	if err := server.ReadJSON(w, r, &req); err != nil {
		server.WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := server.Validate.Struct(req); err != nil {
		server.WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.Service.ScheduleWorkout(r.Context(), userID, planID, req)
	if err != nil {
		handleScheduledWorkoutError(w, err)
		return
	}
	server.JSON(w, http.StatusCreated, resp)
}

func (h *ScheduledWorkoutHandlers) ListScheduledWorkouts(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())

	resp, err := h.Service.ListScheduledWorkouts(r.Context(), userID)
	if err != nil {
		handleScheduledWorkoutError(w, err)
		return
	}
	server.JSON(w, http.StatusOK, map[string]any{
		"scheduled_workouts": resp,
	})
}

func (h *ScheduledWorkoutHandlers) GetScheduledWorkout(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	id := chi.URLParam(r, "id")
	if id == "" {
		server.WriteJSONError(w, http.StatusBadRequest, "missing id")
		return
	}

	resp, err := h.Service.GetScheduledWorkout(r.Context(), userID, id)
	if err != nil {
		handleScheduledWorkoutError(w, err)
		return
	}
	server.JSON(w, http.StatusOK, resp)
}

func (h *ScheduledWorkoutHandlers) CompleteWorkout(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	id := chi.URLParam(r, "id")
	if id == "" {
		server.WriteJSONError(w, http.StatusBadRequest, "missing id")
		return
	}

	var req CompleteRequest
	if err := server.ReadJSON(w, r, &req); err != nil {
		server.WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := server.Validate.Struct(req); err != nil {
		server.WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.Service.CompleteWorkout(r.Context(), userID, id, req)
	if err != nil {
		handleScheduledWorkoutError(w, err)
		return
	}
	server.JSON(w, http.StatusCreated, resp)
}

func (h *ScheduledWorkoutHandlers) CancelWorkout(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	id := chi.URLParam(r, "id")
	if id == "" {
		server.WriteJSONError(w, http.StatusBadRequest, "missing id")
		return
	}

	resp, err := h.Service.CancelWorkout(r.Context(), userID, id)
	if err != nil {
		handleScheduledWorkoutError(w, err)
		return
	}
	server.JSON(w, http.StatusOK, resp)
}

func getUserID(ctx context.Context) string {
	if uid := ctx.Value(servermiddleware.UserCtx); uid != nil {
		if str, ok := uid.(string); ok {
			return str
		}
	}
	return ""
}

func handleScheduledWorkoutError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		server.WriteJSONError(w, http.StatusNotFound, "scheduled workout not found")
	case errors.Is(err, apperrors.ErrIllegalScheduleTransition):
		server.WriteJSONError(w, http.StatusConflict, "invalid workout state transition")
	case errors.Is(err, apperrors.ErrInvalidInput):
		server.WriteJSONError(w, http.StatusBadRequest, "invalid input")
	default:
		if errors.Is(err, apperrors.ErrConflict) {
			server.WriteJSONError(w, http.StatusConflict, err.Error())
			return
		}
		server.WriteJSONError(w, http.StatusInternalServerError, "internal error")
	}
}
