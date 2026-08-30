package scheduledworkout

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	apperrors "github.com/mmk31585/workout-tracker/internal/app_errors"
	"github.com/mmk31585/workout-tracker/internal/db"
)

var ErrNotImplemented = errors.New("repository: method not implemented")

type ScheduledWorkoutRepository interface {
	Create(context.Context, ScheduledWorkout) (ScheduledWorkout, error)
	GetByID(context.Context, string, string) (ScheduledWorkout, error)
	GetByUserID(context.Context, string) ([]ScheduledWorkout, error)
	Update(context.Context, ScheduledWorkout, string) (ScheduledWorkout, error)
	Delete(context.Context, string, string) error
	Transition(context.Context, string, string, ScheduleStatus, ScheduleStatus) (ScheduledWorkout, error)
}

type PostgresScheduledWorkoutRepository struct {
	db *sqlx.DB
}

func NewScheduledWorkoutRepository(db *sqlx.DB) *PostgresScheduledWorkoutRepository {
	return &PostgresScheduledWorkoutRepository{
		db: db,
	}
}

var _ ScheduledWorkoutRepository = (*PostgresScheduledWorkoutRepository)(nil)

func (r *PostgresScheduledWorkoutRepository) Create(ctx context.Context, sw ScheduledWorkout) (ScheduledWorkout, error) {
	var created ScheduledWorkout
	query := `
		INSERT INTO scheduled_workouts
		(
			id,
			user_id,
			workout_plan_id,
			scheduled_at,
			status,
			notes,
			created_at,
			updated_at
		)
		VALUES (gen_random_uuid(),$1,$2,$3,$4,$5,NOW(),NOW())
		RETURNING
			id,
			user_id,
			workout_plan_id,
			scheduled_at,
			status,
			notes,
			created_at,
			updated_at
	`
	ctx, cancel := db.QueryTimeoutContext(ctx)
	defer cancel()
	err := r.db.QueryRowContext(
		ctx,
		query,
		sw.UserID,
		sw.WorkoutPlanID,
		sw.ScheduledAt,
		sw.Status,
		sw.Notes,
	).Scan(
		&created.ID,
		&created.UserID,
		&created.WorkoutPlanID,
		&created.ScheduledAt,
		&created.Status,
		&created.Notes,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		if apperrors.IsUniqueRow(err) {
			return ScheduledWorkout{}, apperrors.ErrConflict
		}
		return ScheduledWorkout{}, fmt.Errorf("repository: failed to create scheduled workout: %w", err)
	}
	return created, nil
}

func (r *PostgresScheduledWorkoutRepository) GetByID(ctx context.Context, id string, userID string) (ScheduledWorkout, error) {
	var sw ScheduledWorkout
	query := `
		SELECT
			id,
			user_id,
			workout_plan_id,
			scheduled_at,
			status,
			notes,
			created_at,
			updated_at
		FROM scheduled_workouts
		WHERE id = $1 AND user_id = $2
	`
	ctx, cancel := db.QueryTimeoutContext(ctx)
	defer cancel()
	err := r.db.GetContext(ctx, &sw, query, id, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return ScheduledWorkout{}, apperrors.ErrNotFound
	}
	if err != nil {
		return ScheduledWorkout{}, fmt.Errorf("repository: failed to get scheduled workout: %w", err)
	}
	return sw, nil
}

func (r *PostgresScheduledWorkoutRepository) GetByUserID(ctx context.Context, userID string) ([]ScheduledWorkout, error) {
	var sws []ScheduledWorkout
	query := `
		SELECT
			id,
			user_id,
			workout_plan_id,
			scheduled_at,
			status,
			notes,
			created_at,
			updated_at
		FROM scheduled_workouts
		WHERE user_id = $1
		ORDER BY scheduled_at ASC
	`
	ctx, cancel := db.QueryTimeoutContext(ctx)
	defer cancel()
	err := r.db.SelectContext(ctx, &sws, query, userID)
	if err != nil {
		return []ScheduledWorkout{}, fmt.Errorf("repository: failed to get scheduled workouts: %w", err)
	}
	if len(sws) == 0 {
		return []ScheduledWorkout{}, nil
	}
	return sws, nil
}

func (r *PostgresScheduledWorkoutRepository) Update(ctx context.Context, sw ScheduledWorkout, userID string) (ScheduledWorkout, error) {
	var updated ScheduledWorkout
	query := `
		UPDATE scheduled_workouts
		SET scheduled_at = $2, notes = $3, updated_at = NOW()
		WHERE id = $1 AND user_id = $4
		RETURNING
			id,
			user_id,
			workout_plan_id,
			scheduled_at,
			status,
			notes,
			created_at,
			updated_at
	`
	ctx, cancel := db.QueryTimeoutContext(ctx)
	defer cancel()
	err := r.db.QueryRowContext(
		ctx,
		query,
		sw.ID,
		sw.ScheduledAt,
		sw.Notes,
		userID,
	).Scan(
		&updated.ID,
		&updated.UserID,
		&updated.WorkoutPlanID,
		&updated.ScheduledAt,
		&updated.Status,
		&updated.Notes,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return ScheduledWorkout{}, apperrors.ErrNotFound
	}
	if err != nil {
		return ScheduledWorkout{}, fmt.Errorf("repository: failed to update scheduled workout: %w", err)
	}
	return updated, nil
}

func (r *PostgresScheduledWorkoutRepository) Delete(ctx context.Context, id string, userID string) error {
	query := `
		DELETE FROM scheduled_workouts
		WHERE id = $1 AND user_id = $2
	`
	ctx, cancel := db.QueryTimeoutContext(ctx)
	defer cancel()

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("repository: failed to delete scheduled workout: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository: failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *PostgresScheduledWorkoutRepository) Transition(ctx context.Context, id string, userID string, from, to ScheduleStatus) (ScheduledWorkout, error) {
	query := `
		UPDATE scheduled_workouts
		SET status = $2, updated_at = NOW()
		WHERE id = $1 AND user_id = $3 AND status = $4
		RETURNING
			id, user_id, workout_plan_id, scheduled_at,
			status, notes, created_at, updated_at
	`
	ctx, cancel := db.QueryTimeoutContext(ctx)
	defer cancel()

	var sw ScheduledWorkout
	err := r.db.QueryRowContext(ctx, query, id, string(to), userID, string(from)).Scan(
		&sw.ID,
		&sw.UserID,
		&sw.WorkoutPlanID,
		&sw.ScheduledAt,
		&sw.Status,
		&sw.Notes,
		&sw.CreatedAt,
		&sw.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return ScheduledWorkout{}, apperrors.ErrNotFound
	}
	if err != nil {
		return ScheduledWorkout{}, fmt.Errorf("repository: failed to transition scheduled workout: %w", err)
	}
	return sw, nil
}
