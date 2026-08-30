package exercise

import (
	"context"
	"database/sql"
	stderrors "errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	apperrors "github.com/mmk31585/workout-tracker/internal/app_errors"
	"github.com/mmk31585/workout-tracker/internal/config"
)

var errDB = stderrors.New("db error")

func newMockRepo(t *testing.T) (*PostgresExerciseRepository, *sql.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	sqlxDB := sqlx.NewDb(db, "postgres")

	return &PostgresExerciseRepository{db: sqlxDB}, db, mock
}

func setupTestConfig(t *testing.T) {
	t.Helper()
	config.SetForTest(&config.Config{
		DB: config.DBConfig{QueryTimeout: 5},
	})
	t.Cleanup(config.ResetForTest)
}

func TestPostgresExerciseRepository_GetAll(t *testing.T) {
	fixedTime := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		name    string
		mock    func(mock sqlmock.Sqlmock)
		want    []Exercise
		wantErr error
	}{
		{
			name: "success",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "description", "category", "muscle_group", "created_at",
				}).
					AddRow("1", "Squat", "desc", "Legs", "Quads", fixedTime).
					AddRow("2", "Bench Press", "desc", "Chest", "Chest", fixedTime)

				mock.ExpectQuery(`SELECT\s+id\s*,\s*name\s*,\s*description\s*,\s*category\s*,\s*muscle_group\s*,\s*created_at\s+FROM\s+exercises\s+ORDER\s+BY\s+name\s+ASC\s*$`).
					WillReturnRows(rows)
			},
			want: []Exercise{
				{ID: "1", Name: "Squat", Description: new("desc"), Category: new("Legs"), MuscleGroup: new("Quads"), CreatedAt: fixedTime},
				{ID: "2", Name: "Bench Press", Description: new("desc"), Category: new("Chest"), MuscleGroup: new("Chest"), CreatedAt: fixedTime},
			},
		},
		{
			name: "empty result",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "description", "category", "muscle_group", "created_at",
				})

				mock.ExpectQuery(`SELECT\s+id\s*,\s*name\s*,\s*description\s*,\s*category\s*,\s*muscle_group\s*,\s*created_at\s+FROM\s+exercises\s+ORDER\s+BY\s+name\s+ASC\s*$`).
					WillReturnRows(rows)
			},
			want: []Exercise{},
		},
		{
			name: "database error",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`.*`).
					WillReturnError(errDB)
			},
			wantErr: errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestConfig(t)

			repo, _, mock := newMockRepo(t)
			tt.mock(mock)

			got, err := repo.GetAll(context.Background())

			if tt.wantErr != nil {
				if !stderrors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(got) != len(tt.want) {
					t.Fatalf("expected %d exercises, got %d", len(tt.want), len(got))
				}
				for i := range got {
					if got[i].ID != tt.want[i].ID ||
						got[i].Name != tt.want[i].Name ||
						*got[i].Description != *tt.want[i].Description ||
						*got[i].Category != *tt.want[i].Category ||
						*got[i].MuscleGroup != *tt.want[i].MuscleGroup ||
						!got[i].CreatedAt.Equal(tt.want[i].CreatedAt) {
						t.Fatalf("expected %+v, got %+v", tt.want[i], got[i])
					}
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPostgresExerciseRepository_GetByID(t *testing.T) {
	fixedTime := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		name    string
		id      string
		mock    func(mock sqlmock.Sqlmock)
		want    Exercise
		wantErr error
	}{
		{
			name: "success",
			id:   "1",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "description", "category", "muscle_group", "created_at",
				}).AddRow("1", "Squat", "desc", "Legs", "Quads", fixedTime)

				mock.ExpectQuery(`SELECT\s+id\s*,\s*name\s*,\s*description\s*,\s*category\s*,\s*muscle_group\s*,\s*created_at\s+FROM\s+exercises\s+WHERE\s+id\s*=\s*\$1`).
					WithArgs("1").
					WillReturnRows(rows)
			},
			want: Exercise{
				ID:          "1",
				Name:        "Squat",
				Description: new("desc"),
				Category:    new("Legs"),
				MuscleGroup: new("Quads"),
				CreatedAt:   fixedTime,
			},
		},
		{
			name: "not found",
			id:   "missing",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT\s+id\s*,\s*name\s*,\s*description\s*,\s*category\s*,\s*muscle_group\s*,\s*created_at\s+FROM\s+exercises\s+WHERE\s+id\s*=\s*\$1`).
					WithArgs("missing").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: apperrors.ErrNotFound,
		},
		{
			name: "database error",
			id:   "1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT\s+id\s*,\s*name\s*,\s*description\s*,\s*category\s*,\s*muscle_group\s*,\s*created_at\s+FROM\s+exercises\s+WHERE\s+id\s*=\s*\$1`).
					WithArgs("1").
					WillReturnError(errDB)
			},
			wantErr: errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestConfig(t)

			repo, _, mock := newMockRepo(t)
			tt.mock(mock)

			got, err := repo.GetByID(context.Background(), tt.id)

			if tt.wantErr != nil {
				if !stderrors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.ID != tt.want.ID ||
					got.Name != tt.want.Name ||
					*got.Description != *tt.want.Description ||
					*got.Category != *tt.want.Category ||
					*got.MuscleGroup != *tt.want.MuscleGroup ||
					!got.CreatedAt.Equal(tt.want.CreatedAt) {
					t.Fatalf("expected %+v, got %+v", tt.want, got)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPostgresExerciseRepository_UpdateByID(t *testing.T) {
	fixedTime := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		name    string
		id      string
		input   Exercise
		mock    func(mock sqlmock.Sqlmock)
		want    Exercise
		wantErr error
	}{
		{
			name: "success",
			id:   "1",
			input: Exercise{
				Name:        "Squat",
				Description: new("updated desc"),
				Category:    new("Legs"),
				MuscleGroup: new("Quads"),
			},
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "description", "category", "muscle_group", "created_at",
				}).AddRow("1", "Squat", "updated desc", "Legs", "Quads", fixedTime)

				mock.ExpectQuery(`UPDATE\s+exercises\s+SET`).
					WithArgs("1", "Squat", "updated desc", "Legs", "Quads").
					WillReturnRows(rows)
			},
			want: Exercise{
				ID:          "1",
				Name:        "Squat",
				Description: new("updated desc"),
				Category:    new("Legs"),
				MuscleGroup: new("Quads"),
				CreatedAt:   fixedTime,
			},
		},
		{
			name:  "not found",
			id:    "missing",
			input: Exercise{Name: "Squat"},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE\s+exercises\s+SET`).
					WithArgs("missing", "Squat", nil, nil, nil).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: apperrors.ErrNotFound,
		},
		{
			name:  "database error",
			id:    "1",
			input: Exercise{Name: "Squat"},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE\s+exercises\s+SET`).
					WithArgs("1", "Squat", nil, nil, nil).
					WillReturnError(errDB)
			},
			wantErr: errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestConfig(t)

			repo, _, mock := newMockRepo(t)
			tt.mock(mock)

			got, err := repo.UpdateByID(context.Background(), tt.id, tt.input)

			if tt.wantErr != nil {
				if !stderrors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.ID != tt.want.ID ||
					got.Name != tt.want.Name ||
					*got.Description != *tt.want.Description ||
					*got.Category != *tt.want.Category ||
					*got.MuscleGroup != *tt.want.MuscleGroup ||
					!got.CreatedAt.Equal(tt.want.CreatedAt) {
					t.Fatalf("expected %+v, got %+v", tt.want, got)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPostgresExerciseRepository_DeleteByID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		mock    func(mock sqlmock.Sqlmock)
		wantErr error
	}{
		{
			name: "success",
			id:   "1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE\s+FROM\s+exercises\s+WHERE\s+id\s*=\s*\$1`).
					WithArgs("1").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "not found",
			id:   "missing",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE\s+FROM\s+exercises\s+WHERE\s+id\s*=\s*\$1`).
					WithArgs("missing").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: apperrors.ErrNotFound,
		},
		{
			name: "database error",
			id:   "1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE\s+FROM\s+exercises\s+WHERE\s+id\s*=\s*\$1`).
					WithArgs("1").
					WillReturnError(errDB)
			},
			wantErr: errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestConfig(t)

			repo, _, mock := newMockRepo(t)
			tt.mock(mock)

			err := repo.DeleteByID(context.Background(), tt.id)

			if tt.wantErr != nil {
				if !stderrors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}
