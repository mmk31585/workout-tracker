package seed

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mmk31585/workout-tracker/internal/exercise"
)

var exercises = []exercise.Exercise{
	{Name: "Push-ups", Description: ptr("Standard push-up"), Category: ptr("Strength"), MuscleGroup: ptr("Chest, Triceps, Shoulders")},
	{Name: "Pull-ups", Description: ptr("Standard pull-up"), Category: ptr("Strength"), MuscleGroup: ptr("Back, Biceps")},
	{Name: "Bodyweight Squats", Description: ptr("Standard squat"), Category: ptr("Strength"), MuscleGroup: ptr("Quadriceps, Glutes")},
	{Name: "Bench Press", Description: ptr("Barbell bench press"), Category: ptr("Strength"), MuscleGroup: ptr("Chest, Triceps, Shoulders")},
	{Name: "Deadlift", Description: ptr("Barbell deadlift"), Category: ptr("Strength"), MuscleGroup: ptr("Back, Legs, Core")},
	{Name: "Overhead Press", Description: ptr("Barbell overhead press"), Category: ptr("Strength"), MuscleGroup: ptr("Shoulders, Triceps")},
	{Name: "Bicep Curls", Description: ptr("Dumbbell bicep curls"), Category: ptr("Strength"), MuscleGroup: ptr("Biceps")},
	{Name: "Tricep Extensions", Description: ptr("Dumbbell tricep extensions"), Category: ptr("Strength"), MuscleGroup: ptr("Triceps")},
	{Name: "Lunges", Description: ptr("Walking lunges"), Category: ptr("Strength"), MuscleGroup: ptr("Quadriceps, Glutes")},
	{Name: "Plank", Description: ptr("Standard plank hold"), Category: ptr("Core"), MuscleGroup: ptr("Abdominals, Lower Back")},
}

type SeederService struct {
	repo SeedRepository
	log  *slog.Logger
}

func NewSeederService(rp SeedRepository, log *slog.Logger) *SeederService {
	return &SeederService{
		repo: rp,
		log:  log,
	}
}
func (s *SeederService) Run(ctx context.Context) error {

	if err := s.SeedExercises(ctx, exercises); err != nil {
		return err
	}
	return s.SeedUsers(ctx)
}
func (s *SeederService) SeedExercises(ctx context.Context, exs []exercise.Exercise) error {

	for _, ex := range exs {
		if err := s.repo.SeedExercises(ctx, ex); err != nil {
			return fmt.Errorf("failed to seed exercise %s: %w", ex.Name, err)
		}
		s.log.Info("seeded exercise", "name", ex.Name)
	}
	return nil
}

func (s *SeederService) SeedUsers(ctx context.Context) error {

	if err := s.repo.SeedUsers(ctx); err != nil {
		return fmt.Errorf("failed to seed user: %w", err)
	}

	s.log.Info("seeded user", "email", "admin@example.com")
	return nil
}

func ptr[T any](v T) *T { return &v }