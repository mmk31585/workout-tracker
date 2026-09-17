package scheduledworkout

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	apperrors "github.com/mmk31585/workout-tracker/internal/app_errors"
	"github.com/mmk31585/workout-tracker/internal/db"
	workoutplan "github.com/mmk31585/workout-tracker/internal/workout_plan"
	workoutsession "github.com/mmk31585/workout-tracker/internal/workout_session"
)

type ScheduledWorkoutService struct {
	repo            ScheduledWorkoutRepository
	planRepo        workoutplan.WorkoutPlanRepository
	itemRepo        workoutplan.WorkoutPlanItemRepository
	sessionRepo     workoutsession.WorkoutSessionRepository
	sessionItemRepo workoutsession.WorkoutSessionItemRepository
	db              *sqlx.DB
}

func NewScheduledWorkoutService(
	repo ScheduledWorkoutRepository,
	planRepo workoutplan.WorkoutPlanRepository,
	itemRepo workoutplan.WorkoutPlanItemRepository,
	sessionRepo workoutsession.WorkoutSessionRepository,
	sessionItemRepo workoutsession.WorkoutSessionItemRepository,
	db *sqlx.DB,
) *ScheduledWorkoutService {
	return &ScheduledWorkoutService{
		repo:            repo,
		planRepo:        planRepo,
		itemRepo:        itemRepo,
		sessionRepo:     sessionRepo,
		sessionItemRepo: sessionItemRepo,
		db:              db,
	}
}

func (s *ScheduledWorkoutService) ScheduleWorkout(ctx context.Context, userID, planID string, req ScheduleRequest) (ScheduleResponse, error) {
	_, err := s.planRepo.GetByID(ctx, planID, userID)
	if err != nil {
		return ScheduleResponse{}, err
	}

	if !req.ScheduledAt.After(time.Now().UTC()) {
		return ScheduleResponse{}, apperrors.ErrInvalidInput
	}

	existing, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return ScheduleResponse{}, err
	}
	for _, sw := range existing {
		if sw.WorkoutPlanID == planID && sw.Status == ScheduleScheduled {
			return ScheduleResponse{}, fmt.Errorf("%w: already scheduled for this plan", apperrors.ErrConflict)
		}
	}

	sw, err := s.repo.Create(ctx, ScheduledWorkout{
		UserID:        userID,
		WorkoutPlanID: planID,
		ScheduledAt:   req.ScheduledAt.UTC(),
		Status:        ScheduleScheduled,
		Notes:         req.Notes,
	})
	if err != nil {
		return ScheduleResponse{}, err
	}

	return toScheduleResponse(sw), nil
}

func (s *ScheduledWorkoutService) CompleteWorkout(ctx context.Context, userID, id string, req CompleteRequest) (CompleteResponse, error) {
	var result CompleteResponse

	err := db.RunInTx(ctx, s.db, func(tx *sqlx.Tx) error {
		sw, err := s.repo.GetByID(ctx, id, userID)
		if err != nil {
			return err
		}

		if !sw.Status.CanTransitionTo(ScheduleInProgress) {
			return apperrors.ErrIllegalScheduleTransition
		}

		if req.PerformedAt.IsZero() || req.PerformedAt.After(time.Now().UTC()) {
			return apperrors.ErrInvalidInput
		}

		planItems, err := s.itemRepo.GetByPlanID(ctx, sw.WorkoutPlanID, userID)
		if err != nil {
			return err
		}
		planExerciseIDs := make(map[string]bool, len(planItems))
		for _, item := range planItems {
			planExerciseIDs[item.ExerciseID] = true
		}

		seen := make(map[string]bool, len(req.Items))
		for _, item := range req.Items {
			if !planExerciseIDs[item.ExerciseID] {
				return fmt.Errorf("%w: exercise %s not in workout plan", apperrors.ErrInvalidInput, item.ExerciseID)
			}
			if seen[item.ExerciseID] {
				return fmt.Errorf("%w: duplicate exercise_id %s in items", apperrors.ErrInvalidInput, item.ExerciseID)
			}
			seen[item.ExerciseID] = true
			if item.Sets <= 0 || item.Reps <= 0 || item.Weight < 0 {
				return apperrors.ErrInvalidInput
			}
		}

		updated, err := s.repo.Transition(ctx, id, userID, ScheduleScheduled, ScheduleInProgress)
		if err != nil {
			return err
		}
		_ = updated

		session := workoutsession.WorkoutSession{
			UserID:             userID,
			ScheduledWorkoutID: id,
			PerformedAt:        &req.PerformedAt,
			OverallNotes:       req.OverallNotes,
		}
		createdSession, err := s.sessionRepo.Create(ctx, session)
		if err != nil {
			return err
		}

		sessionItems := make([]workoutsession.WorkoutSessionItem, 0, len(req.Items))
		for _, item := range req.Items {
			sessionItems = append(sessionItems, workoutsession.WorkoutSessionItem{
				WorkoutSessionID: createdSession.ID,
				ExerciseID:       item.ExerciseID,
				Sets:             item.Sets,
				Reps:             item.Reps,
				Weight:           item.Weight,
				Unit:             item.Unit,
				Notes:            item.Notes,
			})
		}

		if len(sessionItems) > 0 {
			_, err = s.sessionItemRepo.Create(ctx, sessionItems)
			if err != nil {
				return err
			}
		}

		_, err = s.repo.Transition(ctx, id, userID, ScheduleInProgress, ScheduleCompleted)
		if err != nil {
			return err
		}

		respItems := make([]WorkoutSessionItemResponse, 0, len(req.Items))
		for _, item := range req.Items {
			respItems = append(respItems, WorkoutSessionItemResponse{
				ExerciseID: item.ExerciseID,
				Sets:       item.Sets,
				Reps:       item.Reps,
				Weight:     item.Weight,
				Unit:       item.Unit,
				Notes:      item.Notes,
			})
		}

		result = CompleteResponse{
			ScheduledWorkoutID: id,
			PerformedAt:        req.PerformedAt,
			OverallNotes:       req.OverallNotes,
			Items:              respItems,
		}
		return nil
	})

	return result, err
}

func (s *ScheduledWorkoutService) CancelWorkout(ctx context.Context, userID, id string) (ScheduleResponse, error) {
	sw, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return ScheduleResponse{}, err
	}

	if !sw.Status.CanTransitionTo(ScheduleCancelled) {
		return ScheduleResponse{}, apperrors.ErrIllegalScheduleTransition
	}

	updated, err := s.repo.Transition(ctx, id, userID, sw.Status, ScheduleCancelled)
	if err != nil {
		return ScheduleResponse{}, err
	}

	return toScheduleResponse(updated), nil
}

func (s *ScheduledWorkoutService) GetScheduledWorkout(ctx context.Context, userID, id string) (ScheduleResponse, error) {
	sw, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return ScheduleResponse{}, err
	}

	return toScheduleResponse(sw), nil
}

func (s *ScheduledWorkoutService) ListScheduledWorkouts(ctx context.Context, userID string) ([]ScheduleResponse, error) {
	sws, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]ScheduleResponse, 0, len(sws))
	for _, sw := range sws {
		responses = append(responses, toScheduleResponse(sw))
	}
	return responses, nil
}

func toScheduleResponse(sw ScheduledWorkout) ScheduleResponse {
	return ScheduleResponse{
		ID:            sw.ID,
		WorkoutPlanID: sw.WorkoutPlanID,
		ScheduledAt:   sw.ScheduledAt,
		Status:        sw.Status,
		Notes:         sw.Notes,
		CreatedAt:     sw.CreatedAt,
		UpdatedAt:     sw.UpdatedAt,
	}
}
