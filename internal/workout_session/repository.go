package workoutsession

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	apperrors "github.com/mmk31585/workout-tracker/internal/app_errors"
	"github.com/mmk31585/workout-tracker/internal/db"
)

type WorkoutSessionRepository interface {
	Create(context.Context, WorkoutSession) (WorkoutSession, error)
	GetByID(context.Context, string, string) (WorkoutSession, error)
	GetByUserID(context.Context, string) ([]WorkoutSession, error)
	GetByScheduledWorkoutID(context.Context, string, string) (WorkoutSession, error)
	Update(context.Context, WorkoutSession, string) (WorkoutSession, error)
	Delete(context.Context, string, string) error
}

type PostgresWorkoutSessionRepository struct {
	db *sqlx.DB
}

func NewWorkoutSessionRepository(db *sqlx.DB) *PostgresWorkoutSessionRepository {
	return &PostgresWorkoutSessionRepository{
		db: db,
	}
}

var _ WorkoutSessionRepository = (*PostgresWorkoutSessionRepository)(nil)

func (r *PostgresWorkoutSessionRepository) Create(ctx context.Context, ws WorkoutSession) (WorkoutSession, error) {
	var created WorkoutSession
	query := `
		INSERT INTO workout_sessions
		(
			id,
			user_id,
			scheduled_workout_id,
			performed_at,
			overall_notes,
			created_at,
			updated_at
		)
		VALUES (gen_random_uuid(),$1,$2,COALESCE($3,NOW()),$4,NOW(),NOW())
		RETURNING
			id,
			user_id,
			scheduled_workout_id,
			performed_at,
			overall_notes,
			created_at,
			updated_at
	`
	ctx, cancel := db.QueryTimeoutContext(ctx)
	defer cancel()
	err := r.db.QueryRowContext(
		ctx,
		query,
		ws.UserID,
		ws.ScheduledWorkoutID,
		ws.PerformedAt,
		ws.OverallNotes,
	).Scan(
		&created.ID,
		&created.UserID,
		&created.ScheduledWorkoutID,
		&created.PerformedAt,
		&created.OverallNotes,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		if apperrors.IsUniqueRow(err) {
			return WorkoutSession{}, apperrors.ErrConflict
		}
		return WorkoutSession{}, fmt.Errorf("repository: failed to create workout session: %w", err)
	}
	return created, nil
}

func (r *PostgresWorkoutSessionRepository) GetByID(ctx context.Context, id string, userID string) (WorkoutSession, error) {
	var ws WorkoutSession
	query := `
		SELECT
			id,
			user_id,
			scheduled_workout_id,
			performed_at,
			overall_notes,
			created_at,
			updated_at
		FROM workout_sessions
		WHERE id = $1 AND user_id = $2
	`
	ctx, cancel := db.QueryTimeoutContext(ctx)
	defer cancel()
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&ws.ID,
		&ws.UserID,
		&ws.ScheduledWorkoutID,
		&ws.PerformedAt,
		&ws.OverallNotes,
		&ws.CreatedAt,
		&ws.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return WorkoutSession{}, apperrors.ErrNotFound
	}
	if err != nil {
		return WorkoutSession{}, fmt.Errorf("repository: failed to get workout session: %w", err)
	}
	return ws, nil
}

func (r *PostgresWorkoutSessionRepository) GetByUserID(ctx context.Context, userID string) ([]WorkoutSession, error) {
	var sessions []WorkoutSession
	query := `
		SELECT
			id,
			user_id,
			scheduled_workout_id,
			performed_at,
			overall_notes,
			created_at,
			updated_at
		FROM workout_sessions
		WHERE user_id = $1
		ORDER BY performed_at DESC
	`
	ctx, cancel := db.QueryTimeoutContext(ctx)
	defer cancel()
	err := r.db.SelectContext(ctx, &sessions, query, userID)
	if err != nil {
		return []WorkoutSession{}, fmt.Errorf("repository: failed to get workout sessions: %w", err)
	}
	if len(sessions) == 0 {
		return []WorkoutSession{}, nil
	}
	return sessions, nil
}

func (r *PostgresWorkoutSessionRepository) GetByScheduledWorkoutID(ctx context.Context, scheduledWorkoutID string, userID string) (WorkoutSession, error) {
	var ws WorkoutSession
	query := `
		SELECT
			s.id,
			s.user_id,
			s.scheduled_workout_id,
			s.performed_at,
			s.overall_notes,
			s.created_at,
			s.updated_at
		FROM workout_sessions s
		WHERE s.scheduled_workout_id = $1 AND s.user_id = $2
	`
	ctx, cancel := db.QueryTimeoutContext(ctx)
	defer cancel()
	err := r.db.QueryRowContext(ctx, query, scheduledWorkoutID, userID).Scan(
		&ws.ID,
		&ws.UserID,
		&ws.ScheduledWorkoutID,
		&ws.PerformedAt,
		&ws.OverallNotes,
		&ws.CreatedAt,
		&ws.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return WorkoutSession{}, apperrors.ErrNotFound
	}
	if err != nil {
		return WorkoutSession{}, fmt.Errorf("repository: failed to get workout session by scheduled workout: %w", err)
	}
	return ws, nil
}

func (r *PostgresWorkoutSessionRepository) Update(ctx context.Context, ws WorkoutSession, userID string) (WorkoutSession, error) {
	var updated WorkoutSession
	query := `
		UPDATE workout_sessions
		SET performed_at = $2, overall_notes = $3, updated_at = NOW()
		WHERE id = $1 AND user_id = $4
		RETURNING
			id,
			user_id,
			scheduled_workout_id,
			performed_at,
			overall_notes,
			created_at,
			updated_at
	`
	ctx, cancel := db.QueryTimeoutContext(ctx)
	defer cancel()
	err := r.db.QueryRowContext(
		ctx,
		query,
		ws.ID,
		ws.PerformedAt,
		ws.OverallNotes,
		userID,
	).Scan(
		&updated.ID,
		&updated.UserID,
		&updated.ScheduledWorkoutID,
		&updated.PerformedAt,
		&updated.OverallNotes,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return WorkoutSession{}, apperrors.ErrNotFound
	}
	if err != nil {
		return WorkoutSession{}, fmt.Errorf("repository: failed to update workout session: %w", err)
	}
	return updated, nil
}

func (r *PostgresWorkoutSessionRepository) Delete(ctx context.Context, id string, userID string) error {
	query := `
		DELETE FROM workout_sessions
		WHERE id = $1 AND user_id = $2
	`
	ctx, cancel := db.QueryTimeoutContext(ctx)
	defer cancel()

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("repository: failed to delete workout session: %w", err)
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
