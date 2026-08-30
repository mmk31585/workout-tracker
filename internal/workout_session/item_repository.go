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

type WorkoutSessionItemRepository interface {
	Create(context.Context, []WorkoutSessionItem) ([]WorkoutSessionItem, error)
	GetBySessionID(context.Context, string, string) ([]WorkoutSessionItem, error)
	Update(context.Context, WorkoutSessionItem, string) (WorkoutSessionItem, error)
	Delete(context.Context, string, string) error
}

type PostgresWorkoutSessionItemRepository struct {
	db *sqlx.DB
}

func NewWorkoutSessionItemRepository(db *sqlx.DB) *PostgresWorkoutSessionItemRepository {
	return &PostgresWorkoutSessionItemRepository{
		db: db,
	}
}

var _ WorkoutSessionItemRepository = (*PostgresWorkoutSessionItemRepository)(nil)

func (r *PostgresWorkoutSessionItemRepository) Create(ctx context.Context, wsi []WorkoutSessionItem) ([]WorkoutSessionItem, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("repository: failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	query := `
		INSERT INTO workout_session_items
		(
			id,
			workout_session_id,
			exercise_id,
			sets,
			reps,
			weight,
			unit,
			notes,
			created_at
		)
		VALUES (gen_random_uuid(),$1,$2,$3,$4,$5,$6,$7,NOW())
		RETURNING
			id,
			workout_session_id,
			exercise_id,
			sets,
			reps,
			weight,
			unit,
			notes,
			created_at
	`

	items := make([]WorkoutSessionItem, 0, len(wsi))
	for _, item := range wsi {
		var created WorkoutSessionItem
		err = tx.QueryRowContext(
			ctx,
			query,
			item.WorkoutSessionID,
			item.ExerciseID,
			item.Sets,
			item.Reps,
			item.Weight,
			item.Unit,
			item.Notes,
		).Scan(
			&created.ID,
			&created.WorkoutSessionID,
			&created.ExerciseID,
			&created.Sets,
			&created.Reps,
			&created.Weight,
			&created.Unit,
			&created.Notes,
			&created.CreatedAt,
		)
		if err != nil {
			if apperrors.IsUniqueRow(err) {
				return nil, apperrors.ErrConflict
			}
			return nil, fmt.Errorf("repository: failed to create session item: %w", err)
		}
		items = append(items, created)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("repository: failed to commit transaction: %w", err)
	}

	return items, nil
}

func (r *PostgresWorkoutSessionItemRepository) GetBySessionID(ctx context.Context, sessionID string, userID string) ([]WorkoutSessionItem, error) {
	var items []WorkoutSessionItem
	query := `
		SELECT
			i.id,
			i.workout_session_id,
			i.exercise_id,
			i.sets,
			i.reps,
			i.weight,
			i.unit,
			i.notes,
			i.created_at
		FROM workout_session_items i
		JOIN workout_sessions s ON i.workout_session_id = s.id
		WHERE i.workout_session_id = $1 AND s.user_id = $2
		ORDER BY i.created_at ASC
	`
	ctx, cancel := db.QueryTimeoutContext(ctx)
	defer cancel()
	err := r.db.SelectContext(ctx, &items, query, sessionID, userID)
	if err != nil {
		return []WorkoutSessionItem{}, fmt.Errorf("repository: failed to get session items: %w", err)
	}
	if len(items) == 0 {
		return []WorkoutSessionItem{}, nil
	}
	return items, nil
}

func (r *PostgresWorkoutSessionItemRepository) Update(ctx context.Context, wsi WorkoutSessionItem, userID string) (WorkoutSessionItem, error) {
	var item WorkoutSessionItem
	query := `
		UPDATE workout_session_items
		SET workout_session_id = $2, exercise_id = $3, sets = $4, reps = $5, weight = $6, unit = $7, notes = $8
		WHERE id = $1 AND workout_session_id IN (SELECT id FROM workout_sessions WHERE user_id = $9)
		RETURNING
			id,
			workout_session_id,
			exercise_id,
			sets,
			reps,
			weight,
			unit,
			notes,
			created_at
	`
	ctx, cancel := db.QueryTimeoutContext(ctx)
	defer cancel()
	err := r.db.QueryRowContext(
		ctx,
		query,
		wsi.ID,
		wsi.WorkoutSessionID,
		wsi.ExerciseID,
		wsi.Sets,
		wsi.Reps,
		wsi.Weight,
		wsi.Unit,
		wsi.Notes,
		userID,
	).Scan(
		&item.ID,
		&item.WorkoutSessionID,
		&item.ExerciseID,
		&item.Sets,
		&item.Reps,
		&item.Weight,
		&item.Unit,
		&item.Notes,
		&item.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return WorkoutSessionItem{}, apperrors.ErrNotFound
	}
	if err != nil {
		return WorkoutSessionItem{}, fmt.Errorf("repository: failed to update session item: %w", err)
	}
	return item, nil
}

func (r *PostgresWorkoutSessionItemRepository) Delete(ctx context.Context, id string, userID string) error {
	query := `
		DELETE FROM workout_session_items
		WHERE id = $1 AND workout_session_id IN (SELECT id FROM workout_sessions WHERE user_id = $2)
	`
	ctx, cancel := db.QueryTimeoutContext(ctx)
	defer cancel()

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("repository: failed to delete session item: %w", err)
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
