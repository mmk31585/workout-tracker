package workoutplan

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	apperrors "github.com/mmk31585/workout-tracker/internal/app_errors"
	"github.com/mmk31585/workout-tracker/internal/server"
	servermiddleware "github.com/mmk31585/workout-tracker/internal/server/server_middleware"
)

type WorkoutPlanHandler struct {
	Service *WorkoutPlanService
}

func NewWorkoutPlanHandler(s *WorkoutPlanService) *WorkoutPlanHandler {
	return &WorkoutPlanHandler{Service: s}
}

func (h *WorkoutPlanHandler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	var req CreateWorkoutPlanRequest
	if err := server.ReadJSON(w, r, &req); err != nil {
		server.WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := server.Validate.Struct(req); err != nil {
		server.WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.Service.CreatePlan(r.Context(), userID, req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	server.JSON(w, http.StatusCreated, resp)
}

func (h *WorkoutPlanHandler) GetPlan(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	planID := chi.URLParam(r, "id")
	if planID == "" {
		server.WriteJSONError(w, http.StatusBadRequest, "missing plan id")
		return
	}

	resp, err := h.Service.GetPlan(r.Context(), userID, planID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	server.JSON(w, http.StatusOK, resp)
}

func (h *WorkoutPlanHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	// no plan ID needed

	resp, err := h.Service.ListPlans(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	server.JSON(w, http.StatusOK, resp)
}

func (h *WorkoutPlanHandler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	planID := chi.URLParam(r, "id")
	if planID == "" {
		server.WriteJSONError(w, http.StatusBadRequest, "missing plan id")
		return
	}

	var req UpdateWorkoutPlanRequest
	if err := server.ReadJSON(w, r, &req); err != nil {
		server.WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := server.Validate.Struct(req); err != nil {
		server.WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.Service.UpdatePlan(r.Context(), userID, planID, req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	server.JSON(w, http.StatusOK, resp)
}

func (h *WorkoutPlanHandler) DeletePlan(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	planID := chi.URLParam(r, "id")
	if planID == "" {
		server.WriteJSONError(w, http.StatusBadRequest, "missing plan id")
		return
	}

	err := h.Service.DeletePlan(r.Context(), userID, planID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func getUserID(ctx context.Context) string {
	if uid := ctx.Value(servermiddleware.UserCtx); uid != nil {
		if str, ok := uid.(string); ok {
			return str
		}
	}
	return ""
}

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		server.WriteJSONError(w, http.StatusNotFound, "resource not found")
	case errors.Is(err, apperrors.ErrConflict):
		server.WriteJSONError(w, http.StatusConflict, "resource conflict")
	case errors.Is(err, apperrors.ErrInvalidInput):
		server.WriteJSONError(w, http.StatusBadRequest, "invalid input")
	default:
		server.WriteJSONError(w, http.StatusInternalServerError, "internal error")
	}
}
