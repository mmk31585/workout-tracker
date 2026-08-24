package exercise

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	apperrors "github.com/mmk31585/workout-tracker/internal/app_errors"
	"github.com/mmk31585/workout-tracker/internal/db"
)

type ExerciseRepository interface {
	GetAll(context.Context) ([]Exercise, error)
	GetByID(context.Context, string) (Exercise, error)
	UpdateByID(context.Context, string, Exercise) (Exercise, error)
	DeleteByID(context.Context, string) error
}

type PostgresExerciseRepository struct {
	db *sqlx.DB
}

func NewExerciseRepository(db *sqlx.DB) *PostgresExerciseRepository {
	return &PostgresExerciseRepository{
		db: db,
	}
}

func (r *PostgresExerciseRepository) GetAll(ctx context.Context) ([]Exercise, error) {
	var exs []Exercise
	query := `
		SELECT id, name, description, category, muscle_group, created_at
		FROM exercises
		ORDER BY name ASC
	`
	ctx, cancel := db.QueryTimeoutContext(ctx)
	defer cancel()

	err := r.db.SelectContext(ctx, &exs, query)
	if err != nil {
		return nil, fmt.Errorf("repository: failed to get exercises: %w", err)
	}

	if len(exs) == 0 {
		return []Exercise{}, nil
	}
	return exs, nil
}

func (r *PostgresExerciseRepository) GetByID(ctx context.Context, id string) (Exercise, error) {
	var exe Exercise
	query := `
		SELECT id, name, description, category, muscle_group, created_at
		FROM exercises
		WHERE id = $1
	`
	ctx, cancel := db.QueryTimeoutContext(ctx)
	defer cancel()

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&exe.ID,
		&exe.Name,
		&exe.Description,
		&exe.Category,
		&exe.MuscleGroup,
		&exe.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Exercise{}, apperrors.ErrNotFound
	}
	if err != nil {
		return Exercise{}, fmt.Errorf("repository: failed to get exercise: %w", err)
	}
	return exe, nil
}

func (r *PostgresExerciseRepository) UpdateByID(ctx context.Context, id string, ex Exercise) (Exercise, error) {
	query := `
		UPDATE exercises
		SET name = $2, description = $3, category = $4, muscle_group = $5
		WHERE id = $1
		RETURNING
			id,
			name,
			description,
			category,
			muscle_group,
			created_at
	`
	ctx, cancel := db.QueryTimeoutContext(ctx)
	defer cancel()

	err := r.db.QueryRowContext(
		ctx,
		query,
		id,
		ex.Name,
		ex.Description,
		ex.Category,
		ex.MuscleGroup,
	).Scan(
		&ex.ID,
		&ex.Name,
		&ex.Description,
		&ex.Category,
		&ex.MuscleGroup,
		&ex.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Exercise{}, apperrors.ErrNotFound
	}
	if err != nil {
		return Exercise{}, fmt.Errorf("repository: failed to update exercise: %w", err)
	}
	return ex, nil
}

func (r *PostgresExerciseRepository) DeleteByID(ctx context.Context, id string) error {
	query := `
		DELETE FROM exercises
		WHERE id = $1
	`
	ctx, cancel := db.QueryTimeoutContext(ctx)
	defer cancel()

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("repository: failed to delete exercise: %w", err)
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
