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

type WorkoutPlanRepository interface {
	Create(context.Context, WorkoutPlan) (WorkoutPlan, error)
	GetByID(context.Context, string, string) (WorkoutPlan, error)
	ListByID(context.Context, string) ([]WorkoutPlan, error)
	UpdateByID(context.Context, WorkoutPlan, string) (WorkoutPlan, error)
	DeleteByID(context.Context, string, string) error
}

type PostgresWorkoutPlanRepository struct {
	db *sqlx.DB
}

func NewWorkoutPlanRepository(db *sqlx.DB) *PostgresWorkoutPlanRepository {
	return &PostgresWorkoutPlanRepository{
		db: db,
	}
}

func (r *PostgresWorkoutPlanRepository) Create(ctx context.Context, wp WorkoutPlan) (WorkoutPlan, error) {
	var workoutPlanTemp WorkoutPlan
	query := `
		INSERT INTO workout_plans
		(
			id,
			user_id,
			title,
			description,
			status,
			created_at,
			updated_at
		)
		VALUES (gen_random_uuid(),$1,$2,$3,$4,NOW(),NOW())
		RETURNING
			id,
			user_id,
			title,
			description,
			status,
			created_at,
			updated_at
	`
	ctx, cancel := db.QueryTimeoutContext(ctx)
	defer cancel()

	err := r.db.QueryRowContext(
		ctx,
		query,
		wp.UserID,
		wp.Title,
		wp.Description,
		wp.Status,
	).Scan(
		&workoutPlanTemp.ID,
		&workoutPlanTemp.UserID,
		&workoutPlanTemp.Title,
		&workoutPlanTemp.Description,
		&workoutPlanTemp.Status,
		&workoutPlanTemp.CreatedAt,
		&workoutPlanTemp.UpdatedAt,
	)
	if err != nil {
		if apperrors.IsUniqueRow(err) {
			return WorkoutPlan{}, apperrors.ErrConflict
		}
		return WorkoutPlan{}, fmt.Errorf("repository: failed to create plan: %w", err)
	}
	return workoutPlanTemp, nil
}

func (r *PostgresWorkoutPlanRepository) GetByID(ctx context.Context, id string, userID string) (WorkoutPlan, error) {
	var plan WorkoutPlan
	query := `
		SELECT
			id,
			user_id,
			title,
			description,
			status,
			created_at,
			updated_at
		FROM workout_plans
		WHERE id = $1 AND user_id = $2
	`
	ctx, cancel := db.QueryTimeoutContext(ctx)
	defer cancel()
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&plan.ID,
		&plan.UserID,
		&plan.Title,
		&plan.Description,
		&plan.Status,
		&plan.CreatedAt,
		&plan.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return WorkoutPlan{}, apperrors.ErrNotFound
	}
	if err != nil {
		return WorkoutPlan{}, fmt.Errorf("repository: failed to get plan : %w", err)
	}
	return plan, nil
}

func (r *PostgresWorkoutPlanRepository) UpdateByID(ctx context.Context, wp WorkoutPlan, userID string) (WorkoutPlan, error) {
	var plan WorkoutPlan
	query := `
		UPDATE workout_plans
		SET user_id = $2, title = $3, description = $4, status = $5, updated_at = NOW()
		WHERE id = $1 AND user_id = $6
		RETURNING
			id,
			user_id,
			title,
			description,
			status,
			created_at,
			updated_at
	`
	ctx, cancel := db.QueryTimeoutContext(ctx)
	defer cancel()
	err := r.db.QueryRowContext(
		ctx,
		query,
		wp.ID,
		wp.UserID,
		wp.Title,
		wp.Description,
		wp.Status,
		userID,
	).Scan(
		&plan.ID,
		&plan.UserID,
		&plan.Title,
		&plan.Description,
		&plan.Status,
		&plan.CreatedAt,
		&plan.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return WorkoutPlan{}, apperrors.ErrNotFound
	}
	if err != nil {
		return WorkoutPlan{}, fmt.Errorf("repository: failed to update plan: %w", err)
	}
	return plan, nil
}

func (r *PostgresWorkoutPlanRepository) DeleteByID(ctx context.Context, id string, userID string) error {
	query := `
		DELETE FROM workout_plans
		WHERE id = $1 AND user_id = $2
	`
	ctx, cancel := db.QueryTimeoutContext(ctx)
	defer cancel()

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("repository: failed to delete plan: %w", err)
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

func (r *PostgresWorkoutPlanRepository) ListByID(ctx context.Context, userID string) ([]WorkoutPlan, error) {
	var plans []WorkoutPlan
	query := `
		SELECT
			id,
			user_id,
			title,
			description,
			status,
			created_at,
			updated_at
		FROM workout_plans
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	ctx, cancel := db.QueryTimeoutContext(ctx)
	defer cancel()
	err := r.db.SelectContext(ctx, &plans, query, userID)
	if err != nil {
		return []WorkoutPlan{}, fmt.Errorf("repository: failed to list plans: %w", err)
	}
	return plans, nil
}
