package seed

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/mmk31585/workout-tracker/internal/config"
	"github.com/mmk31585/workout-tracker/internal/exercise"
)

func TestPostgresSeedRepository_SeedExercises(t *testing.T) {
	config.SetForTest(&config.Config{
		DB: config.DBConfig{
			QueryTimeout: 5,
		},
	})
	defer config.ResetForTest()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewPostgresSeedRepository(sqlx.NewDb(db, "postgres"))

	mock.ExpectExec(`^\s*INSERT INTO exercises`).
		WithArgs("Push-ups", "Standard push-up", "Strength", "Chest, Triceps, Shoulders").
		WillReturnResult(sqlmock.NewResult(1, 1))

	
	ex := exercise.Exercise{
		Name:        "Push-ups",
		Description: ptr("Standard push-up"), 
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
	config.SetForTest(&config.Config{
		DB: config.DBConfig{
			QueryTimeout: 5,
		},
	})
	defer config.ResetForTest()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewPostgresSeedRepository(sqlx.NewDb(db, "postgres"))

	
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

	
	if !strings.Contains(err.Error(), "database error") {
		t.Errorf("SeedExercises returned unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPostgresSeedRepository_SeedUsers(t *testing.T) {
	config.SetForTest(&config.Config{
		DB: config.DBConfig{
			QueryTimeout: 5,
		},
	})
	defer config.ResetForTest()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewPostgresSeedRepository(sqlx.NewDb(db, "postgres"))

	
	
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
	config.SetForTest(&config.Config{
		DB: config.DBConfig{
			QueryTimeout: 5,
		},
	})
	defer config.ResetForTest()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewPostgresSeedRepository(sqlx.NewDb(db, "postgres"))

	
	mock.ExpectExec(`^\s*INSERT INTO users`).
		WithArgs("admin@example.com", "Admin User", sqlmock.AnyArg()).
		WillReturnError(errors.New("database error"))

	err = repo.SeedUsers(context.Background())
	if err == nil {
		t.Fatalf("SeedUsers expected error but got nil")
	}

	
	if !strings.Contains(err.Error(), "database error") {
		t.Errorf("SeedUsers returned unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}