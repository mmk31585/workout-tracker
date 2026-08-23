package seed

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/mmk31585/workout-tracker/internal/exercise"
	"golang.org/x/crypto/bcrypt"
)

var (
	queryTimeout = time.Second * 5
)

type SeedRepository interface {
	SeedExercises(context.Context, exercise.Exercise) error
	SeedUsers(context.Context) error
}

type PostgresSeedRepository struct {
	db *sql.DB
}

func NewPostgresSeedRepository(db *sql.DB) *PostgresSeedRepository {
	return &PostgresSeedRepository{db: db}
}

func (r *PostgresSeedRepository) SeedExercises(ctx context.Context, ex exercise.Exercise) error {
	query := `
		INSERT INTO exercises (id, name, description, category, muscle_group, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW())
		ON CONFLICT DO NOTHING
	`
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	_, err := r.db.ExecContext(ctx, query, ex.Name, ex.Description, ex.Category, ex.MuscleGroup)
	if err != nil {
		return fmt.Errorf("repository: failed to seed exercise: %w", err)
	}
	return nil
}
func (r *PostgresSeedRepository) SeedUsers(ctx context.Context) error {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("repository: failed to hash password: %w", err)
	}
	query := `
	INSERT INTO users (id, email, display_name, password_hash, created_at, updated_at)
	VALUES (gen_random_uuid(), $1, $2, $3, NOW(), NOW())
	ON CONFLICT (email) DO NOTHING`
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)

	defer cancel()
	_, err = r.db.ExecContext(ctx, query, "admin@example.com", "Admin User", string(hash))
	if err != nil {
		return fmt.Errorf("failed to seed user: %w", err)
	}
	return nil
}
