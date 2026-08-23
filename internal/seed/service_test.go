package seed

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/mmk31585/workout-tracker/internal/exercise"
)

// mockSeedRepository is a mock implementation of SeedRepository for testing
type mockSeedRepository struct {
	seedExercisesFunc func(ctx context.Context, ex exercise.Exercise) error
	seedUsersFunc     func(ctx context.Context) error
}

func (m *mockSeedRepository) SeedExercises(ctx context.Context, ex exercise.Exercise) error {
	if m.seedExercisesFunc != nil {
		return m.seedExercisesFunc(ctx, ex)
	}
	return nil
}

func (m *mockSeedRepository) SeedUsers(ctx context.Context) error {
	if m.seedUsersFunc != nil {
		return m.seedUsersFunc(ctx)
	}
	return nil
}

func TestSeederService_Run_Success(t *testing.T) {
	repo := &mockSeedRepository{}
	log := slog.New(slog.DiscardHandler)
	svc := NewSeederService(repo, log)

	err := svc.Run(context.Background())
	if err != nil {
		t.Fatalf("Run returned unexpected error: %v", err)
	}
}

func TestSeederService_Run_SeedExercisesError(t *testing.T) {
	repo := &mockSeedRepository{
		seedExercisesFunc: func(ctx context.Context, ex exercise.Exercise) error {
			return errors.New("seed exercises error")
		},
	}
	log := slog.New(slog.DiscardHandler)
	svc := NewSeederService(repo, log)

	err := svc.Run(context.Background())
	if err == nil {
		t.Fatalf("Run expected error but got nil")
	}
	if !strings.Contains(err.Error(), "seed exercises error") {
		t.Errorf("Run returned unexpected error: %v", err)
	}
}

func TestSeederService_Run_SeedUsersError(t *testing.T) {
	repo := &mockSeedRepository{
		seedUsersFunc: func(ctx context.Context) error {
			return errors.New("seed users error")
		},
	}
	log := slog.New(slog.DiscardHandler)
	svc := NewSeederService(repo, log)

	err := svc.Run(context.Background())
	if err == nil {
		t.Fatalf("Run expected error but got nil")
	}
	if !strings.Contains(err.Error(), "seed users error") {
		t.Errorf("Run returned unexpected error: %v", err)
	}
}

func TestSeederService_SeedExercises(t *testing.T) {
	repo := &mockSeedRepository{}
	log := slog.New(slog.DiscardHandler)
	svc := NewSeederService(repo, log)

	err := svc.SeedExercises(context.Background(), exercises)
	if err != nil {
		t.Fatalf("SeedExercises returned unexpected error: %v", err)
	}
}

func TestSeederService_SeedExercises_Error(t *testing.T) {
	repo := &mockSeedRepository{
		seedExercisesFunc: func(ctx context.Context, ex exercise.Exercise) error {
			return errors.New("seed exercise error")
		},
	}
	log := slog.New(slog.DiscardHandler)
	svc := NewSeederService(repo, log)

	err := svc.SeedExercises(context.Background(), []exercise.Exercise{
		{Name: "Test Exercise"},
	})
	if err == nil {
		t.Fatalf("SeedExercises expected error but got nil")
	}
	if !strings.Contains(err.Error(), "seed exercise error") {
		t.Errorf("SeedExercises returned unexpected error: %v", err)
	}
}

func TestSeederService_SeedUsers(t *testing.T) {
	repo := &mockSeedRepository{}
	log := slog.New(slog.DiscardHandler)
	svc := NewSeederService(repo, log)

	err := svc.SeedUsers(context.Background())
	if err != nil {
		t.Fatalf("SeedUsers returned unexpected error: %v", err)
	}
}

func TestSeederService_SeedUsers_Error(t *testing.T) {
	repo := &mockSeedRepository{
		seedUsersFunc: func(ctx context.Context) error {
			return errors.New("seed user error")
		},
	}
	log := slog.New(slog.DiscardHandler)
	svc := NewSeederService(repo, log)

	err := svc.SeedUsers(context.Background())
	if err == nil {
		t.Fatalf("SeedUsers expected error but got nil")
	}
	if !strings.Contains(err.Error(), "seed user error") {
		t.Errorf("SeedUsers returned unexpected error: %v", err)
	}
}

func TestNewSeederService(t *testing.T) {
	repo := &mockSeedRepository{}
	log := slog.New(slog.DiscardHandler)
	svc := NewSeederService(repo, log)

	if svc == nil {
		t.Fatal("NewSeederService returned nil")
	}
	if svc.repo == nil {
		t.Fatal("NewSeederService did not set repo")
	}
	if svc.log == nil {
		t.Fatal("NewSeederService did not set log")
	}
}