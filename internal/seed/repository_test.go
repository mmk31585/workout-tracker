package seed

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mmk31585/workout-tracker/internal/exercise"
)



func TestPostgresSeedRepository_SeedExercises(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewPostgresSeedRepository(db)

	// FIX: The query has exactly 4 placeholders ($1, $2, $3, $4). 
	// Provide exactly 4 arguments to match ExecContext.
	mock.ExpectExec(`^\s*INSERT INTO exercises`).
		WithArgs("Push-ups", "Standard push-up", "Strength", "Chest, Triceps, Shoulders").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Call the method
	ex := exercise.Exercise{
		Name:        "Push-ups",
		Description: ptr("Standard push-up"), // Use string literal (or your ptr() helper if fields are *string)
		Category:    ptr("Strength"),
		MuscleGroup: ptr("Chest, Triceps, Shoulders"),
	}
	
	err = repo.SeedExercises(context.Background(), ex)
	if err != nil {
		t.Fatalf("SeedExercises returned unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPostgresSeedRepository_SeedExercises_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewPostgresSeedRepository(db)

	// FIX: Exactly 4 arguments expected to match the repository's ExecContext call
	mock.ExpectExec(`^\s*INSERT INTO exercises`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("database error"))

	ex := exercise.Exercise{
		Name:        "Test Exercise",
		Description: ptr("Test"),
		Category:    ptr("Test"),
		MuscleGroup: ptr("Test"),
	}
	
	err = repo.SeedExercises(context.Background(), ex)
	if err == nil {
		t.Fatalf("SeedExercises expected error but got nil")
	}
	
	// FIX: Use strings.Contains instead of errors.Is with a newly instantiated error
	if !strings.Contains(err.Error(), "database error") {
		t.Errorf("SeedExercises returned unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPostgresSeedRepository_SeedUsers(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewPostgresSeedRepository(db)

	// FIX: The query has exactly 3 placeholders ($1, $2, $3). 
	// Provide exactly 3 arguments. The hash is dynamic, so we use sqlmock.AnyArg() for it.
	mock.ExpectExec(`^\s*INSERT INTO users`).
		WithArgs("admin@example.com", "Admin User", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.SeedUsers(context.Background())
	if err != nil {
		t.Fatalf("SeedUsers returned unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPostgresSeedRepository_SeedUsers_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewPostgresSeedRepository(db)

	// FIX: Exactly 3 arguments expected to match the repository's ExecContext call
	mock.ExpectExec(`^\s*INSERT INTO users`).
		WithArgs("admin@example.com", "Admin User", sqlmock.AnyArg()).
		WillReturnError(errors.New("database error"))

	err = repo.SeedUsers(context.Background())
	if err == nil {
		t.Fatalf("SeedUsers expected error but got nil")
	}
	
	// FIX: Use strings.Contains to verify the wrapped error message
	if !strings.Contains(err.Error(), "database error") {
		t.Errorf("SeedUsers returned unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}