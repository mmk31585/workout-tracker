package workoutplan

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	apperrors "github.com/mmk31585/workout-tracker/internal/app_errors"
	"github.com/mmk31585/workout-tracker/internal/db"
)

type WorkoutPlanItemRepository interface {
	Create(context.Context, []WorkoutPlanItem) ([]WorkoutPlanItem, error)
	GetByID(context.Context, string, string) ([]WorkoutPlanItem, error)
	UpdateByID(context.Context, WorkoutPlanItem, string) (WorkoutPlanItem, error)
	DeleteByID(context.Context, string, string) error
}

type PostgresWorkoutPlanItemRepository struct {
	db *sqlx.DB
}

func NewWorkoutPlanItemRepository(db *sqlx.DB) *PostgresWorkoutPlanItemRepository {
	return &PostgresWorkoutPlanItemRepository{
		db: db,
	}
}

var _ WorkoutPlanItemRepository = (*PostgresWorkoutPlanItemRepository)(nil)

func (r *PostgresWorkoutPlanItemRepository) Create(ctx context.Context, wpi []WorkoutPlanItem) ([]WorkoutPlanItem, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("repository: failed to begin transaction: %w", err)
	}
	// Ensure rollback on error; commit on success
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	query := `
		INSERT INTO workout_plan_items
		(
			id,
			workout_plan_id,
			exercise_id,
			order_index,
			sets,
			reps,
			weight,
			unit,
			created_at
		)
		VALUES (gen_random_uuid(),$1,$2,$3,$4,$5,$6,$7,NOW())
		RETURNING
			id,
			workout_plan_id,
			exercise_id,
			order_index,
			sets,
			reps,
			weight,
			unit,
			created_at
	`

	items := make([]WorkoutPlanItem, 0, len(wpi))
	for _, item := range wpi {
		var created WorkoutPlanItem
		err = tx.QueryRowContext(
			ctx,
			query,
			item.WorkoutPlanID,
			item.ExerciseID,
			item.OrderIndex,
			item.Sets,
			item.Reps,
			item.Weight,
			item.Unit,
		).Scan(
			&created.ID,
			&created.WorkoutPlanID,
			&created.ExerciseID,
			&created.OrderIndex,
			&created.Sets,
			&created.Reps,
			&created.Weight,
			&created.Unit,
			&created.CreatedAt,
		)
		if err != nil {
			if apperrors.IsUniqueRow(err) {
				return nil, apperrors.ErrConflict
			}
			return nil, fmt.Errorf("repository: failed to create plan item: %w", err)
		}
		items = append(items, created)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("repository: failed to commit transaction: %w", err)
	}

	return items, nil
}

func (r *PostgresWorkoutPlanItemRepository) GetByID(ctx context.Context, wpId string, userID string) ([]WorkoutPlanItem, error) {
	var items []WorkoutPlanItem
	query := `
		SELECT
			i.id,
			i.workout_plan_id,
			i.exercise_id,
			i.order_index,
			i.sets,
			i.reps,
			i.weight,
			i.unit,
			i.created_at
		FROM workout_plan_items i
		JOIN workout_plans p ON i.workout_plan_id = p.id
		WHERE i.workout_plan_id = $1 AND p.user_id = $2
		ORDER BY order_index ASC
	`
	ctx, cancel := db.QueryTimeoutContext(ctx)
	defer cancel()
	err := r.db.SelectContext(ctx, &items, query, wpId, userID)
	if err != nil {
		return []WorkoutPlanItem{}, fmt.Errorf("repository: failed to get plan item: %w", err)
	}
	if len(items) == 0 {
		return []WorkoutPlanItem{}, nil
	}
	return items, nil
}

func (r *PostgresWorkoutPlanItemRepository) UpdateByID(ctx context.Context, wpi WorkoutPlanItem, userID string) (WorkoutPlanItem, error) {
	var item WorkoutPlanItem
	query := `
		UPDATE workout_plan_items
		SET workout_plan_id = $2, exercise_id = $3, order_index = $4, sets = $5, reps = $6, weight = $7, unit = $8
		WHERE id = $1 AND workout_plan_id IN (SELECT id FROM workout_plans WHERE user_id = $9)
		RETURNING
			id,
			workout_plan_id,
			exercise_id,
			order_index,
			sets,
			reps,
			weight,
			unit,
			created_at
	`
	ctx, cancel := db.QueryTimeoutContext(ctx)
	defer cancel()
	err := r.db.QueryRowContext(
		ctx,
		query,
		wpi.ID,
		wpi.WorkoutPlanID,
		wpi.ExerciseID,
		wpi.OrderIndex,
		wpi.Sets,
		wpi.Reps,
		wpi.Weight,
		wpi.Unit,
		userID,
	).Scan(
		&item.ID,
		&item.WorkoutPlanID,
		&item.ExerciseID,
		&item.OrderIndex,
		&item.Sets,
		&item.Reps,
		&item.Weight,
		&item.Unit,
		&item.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return WorkoutPlanItem{}, apperrors.ErrNotFound
	}
	if err != nil {
		return WorkoutPlanItem{}, fmt.Errorf("repository: failed to update plan item: %w", err)
	}
	return item, nil
}

func (r *PostgresWorkoutPlanItemRepository) DeleteByID(ctx context.Context, id string, userID string) error {
	query := `
		DELETE FROM workout_plan_items
		WHERE id = $1 AND workout_plan_id IN (SELECT id FROM workout_plans WHERE user_id = $2)
	`
	ctx, cancel := db.QueryTimeoutContext(ctx)
	defer cancel()

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("repository: failed to delete plan item: %w", err)
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
