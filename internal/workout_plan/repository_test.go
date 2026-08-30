package workoutplan

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

func newMockRepo(t *testing.T) (*PostgresWorkoutPlanRepository, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	sqlxDB := sqlx.NewDb(db, "postgres")

	return &PostgresWorkoutPlanRepository{db: sqlxDB}, mock
}

func newMockItemRepo(t *testing.T) (*PostgresWorkoutPlanItemRepository, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	sqlxDB := sqlx.NewDb(db, "postgres")

	return &PostgresWorkoutPlanItemRepository{db: sqlxDB}, mock
}

func setupTestConfig(t *testing.T) {
	t.Helper()
	config.SetForTest(&config.Config{
		DB: config.DBConfig{QueryTimeout: 5},
	})
	t.Cleanup(config.ResetForTest)
}

func TestPostgresWorkoutPlanRepository_Create(t *testing.T) {
	fixedTime := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		name    string
		input   WorkoutPlan
		mock    func(mock sqlmock.Sqlmock)
		want    WorkoutPlan
		wantErr bool
	}{
		{
			name: "success",
			input: WorkoutPlan{
				UserID:      "user-1",
				Title:       "Push Day",
				Description: "Chest and triceps",
				Status:      "active",
			},
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "user_id", "title", "description", "status", "created_at", "updated_at",
				}).AddRow("plan-1", "user-1", "Push Day", "Chest and triceps", "active", fixedTime, fixedTime)

				mock.ExpectQuery(`INSERT\s+INTO\s+workout_plans`).
					WithArgs("user-1", "Push Day", "Chest and triceps", "active").
					WillReturnRows(rows)
			},
			want: WorkoutPlan{
				ID:          "plan-1",
				UserID:      "user-1",
				Title:       "Push Day",
				Description: "Chest and triceps",
				Status:      "active",
				CreatedAt:   fixedTime,
				UpdatedAt:   fixedTime,
			},
		},
		{
			name: "database error",
			input: WorkoutPlan{
				UserID: "user-1",
				Title:  "Push Day",
				Status: "active",
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`INSERT\s+INTO\s+workout_plans`).
					WithArgs("user-1", "Push Day", "", "active").
					WillReturnError(errDB)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestConfig(t)

			repo, mock := newMockRepo(t)
			tt.mock(mock)

			got, err := repo.Create(context.Background(), tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !stderrors.Is(err, errDB) {
					t.Fatalf("expected error %v, got %v", errDB, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got != tt.want {
					t.Fatalf("expected %+v, got %+v", tt.want, got)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPostgresWorkoutPlanItemRepository_Create(t *testing.T) {
	fixedTime := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	queryPattern := `INSERT\s+INTO\s+workout_plan_items`

	newRow := func(id string, item WorkoutPlanItem) *sqlmock.Rows {
		return sqlmock.NewRows([]string{
			"id", "workout_plan_id", "exercise_id", "order_index", "sets", "reps", "weight", "unit", "created_at",
		}).AddRow(
			id,
			item.WorkoutPlanID,
			item.ExerciseID,
			item.OrderIndex,
			item.Sets,
			item.Reps,
			item.Weight,
			item.Unit,
			fixedTime,
		)
	}

	tests := []struct {
		name    string
		input   []WorkoutPlanItem
		mock    func(mock sqlmock.Sqlmock)
		want    []WorkoutPlanItem
		wantErr bool
	}{
		{
			name: "success with multiple items",
			input: []WorkoutPlanItem{
				{WorkoutPlanID: "plan-1", ExerciseID: "ex-1", OrderIndex: 0, Sets: 3, Reps: 10, Weight: 60.5, Unit: "kg"},
				{WorkoutPlanID: "plan-1", ExerciseID: "ex-2", OrderIndex: 1, Sets: 4, Reps: 8, Weight: 80, Unit: "kg"},
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(queryPattern).
					WithArgs("plan-1", "ex-1", 0, 3, 10, 60.5, "kg").
					WillReturnRows(newRow("item-1", WorkoutPlanItem{WorkoutPlanID: "plan-1", ExerciseID: "ex-1", OrderIndex: 0, Sets: 3, Reps: 10, Weight: 60.5, Unit: "kg"}))
				mock.ExpectQuery(queryPattern).
					WithArgs("plan-1", "ex-2", 1, 4, 8, 80.0, "kg").
					WillReturnRows(newRow("item-2", WorkoutPlanItem{WorkoutPlanID: "plan-1", ExerciseID: "ex-2", OrderIndex: 1, Sets: 4, Reps: 8, Weight: 80, Unit: "kg"}))
				mock.ExpectCommit()
			},
			want: []WorkoutPlanItem{
				{ID: "item-1", WorkoutPlanID: "plan-1", ExerciseID: "ex-1", OrderIndex: 0, Sets: 3, Reps: 10, Weight: 60.5, Unit: "kg", CreatedAt: fixedTime},
				{ID: "item-2", WorkoutPlanID: "plan-1", ExerciseID: "ex-2", OrderIndex: 1, Sets: 4, Reps: 8, Weight: 80, Unit: "kg", CreatedAt: fixedTime},
			},
		},
		{
			name:  "empty input inserts nothing",
			input: []WorkoutPlanItem{},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectCommit()
			},
			want: []WorkoutPlanItem{},
		},
		{
			name: "database error",
			input: []WorkoutPlanItem{
				{WorkoutPlanID: "plan-1", ExerciseID: "ex-1", OrderIndex: 0, Sets: 3, Reps: 10, Weight: 60, Unit: "kg"},
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(queryPattern).
					WithArgs("plan-1", "ex-1", 0, 3, 10, 60.0, "kg").
					WillReturnError(errDB)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestConfig(t)

			repo, mock := newMockItemRepo(t)
			tt.mock(mock)

			got, err := repo.Create(context.Background(), tt.input)

			if tt.wantErr {
				if !stderrors.Is(err, errDB) {
					t.Fatalf("expected error %v, got %v", errDB, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(got) != len(tt.want) {
					t.Fatalf("expected %d items, got %d", len(tt.want), len(got))
				}
				for i := range got {
					if got[i] != tt.want[i] {
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

func TestNewWorkoutPlanRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewWorkoutPlanRepository(sqlx.NewDb(db, "postgres"))
	if repo == nil {
		t.Fatal("NewExerciseRepository returned nil")
	}
}

func TestPostgresWorkoutPlanRepository_UpdateByID(t *testing.T) {
	fixedTime := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		name    string
		input   WorkoutPlan
		mock    func(mock sqlmock.Sqlmock)
		want    WorkoutPlan
		wantErr error
	}{
		{
			name: "success",
			input: WorkoutPlan{
				ID:          "plan-1",
				UserID:      "user-1",
				Title:       "Pull Day",
				Description: "Back and biceps",
				Status:      "active",
			},
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "user_id", "title", "description", "status", "created_at", "updated_at",
				}).AddRow("plan-1", "user-1", "Pull Day", "Back and biceps", "active", fixedTime, fixedTime)

				mock.ExpectQuery(`UPDATE\s+workout_plans\s+SET`).
					WithArgs("plan-1", "user-1", "Pull Day", "Back and biceps", "active", "user-1").
					WillReturnRows(rows)
			},
			want: WorkoutPlan{
				ID:          "plan-1",
				UserID:      "user-1",
				Title:       "Pull Day",
				Description: "Back and biceps",
				Status:      "active",
				CreatedAt:   fixedTime,
				UpdatedAt:   fixedTime,
			},
		},
		{
			name:  "not found",
			input: WorkoutPlan{ID: "missing", UserID: "user-1", Title: "Pull Day", Status: "active"},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE\s+workout_plans\s+SET`).
					WithArgs("missing", "user-1", "Pull Day", "", "active", "user-1").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: apperrors.ErrNotFound,
		},
		{
			name:  "database error",
			input: WorkoutPlan{ID: "plan-1", UserID: "user-1", Title: "Pull Day", Status: "active"},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE\s+workout_plans\s+SET`).
					WithArgs("plan-1", "user-1", "Pull Day", "", "active", "user-1").
					WillReturnError(errDB)
			},
			wantErr: errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestConfig(t)

			repo, mock := newMockRepo(t)
			tt.mock(mock)

			got, err := repo.UpdateByID(context.Background(), tt.input, "user-1")

			if tt.wantErr != nil {
				if !stderrors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got != tt.want {
					t.Fatalf("expected %+v, got %+v", tt.want, got)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPostgresWorkoutPlanItemRepository_UpdateByID(t *testing.T) {
	fixedTime := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		name    string
		input   WorkoutPlanItem
		mock    func(mock sqlmock.Sqlmock)
		want    WorkoutPlanItem
		wantErr error
	}{
		{
			name: "success",
			input: WorkoutPlanItem{
				ID:            "item-1",
				WorkoutPlanID: "plan-1",
				ExerciseID:    "ex-1",
				OrderIndex:    2,
				Sets:          5,
				Reps:          12,
				Weight:        70.5,
				Unit:          "kg",
			},
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "workout_plan_id", "exercise_id", "order_index", "sets", "reps", "weight", "unit", "created_at",
				}).AddRow("item-1", "plan-1", "ex-1", 2, 5, 12, 70.5, "kg", fixedTime)

				mock.ExpectQuery(`UPDATE\s+workout_plan_items\s+SET`).
					WithArgs("item-1", "plan-1", "ex-1", 2, 5, 12, 70.5, "kg", "user-1").
					WillReturnRows(rows)
			},
			want: WorkoutPlanItem{
				ID:            "item-1",
				WorkoutPlanID: "plan-1",
				ExerciseID:    "ex-1",
				OrderIndex:    2,
				Sets:          5,
				Reps:          12,
				Weight:        70.5,
				Unit:          "kg",
				CreatedAt:     fixedTime,
			},
		},
		{
			name:  "not found",
			input: WorkoutPlanItem{ID: "missing", WorkoutPlanID: "plan-1", ExerciseID: "ex-1"},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE\s+workout_plan_items\s+SET`).
					WithArgs("missing", "plan-1", "ex-1", 0, 0, 0, 0.0, "", "user-1").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: apperrors.ErrNotFound,
		},
		{
			name:  "database error",
			input: WorkoutPlanItem{ID: "item-1", WorkoutPlanID: "plan-1", ExerciseID: "ex-1"},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE\s+workout_plan_items\s+SET`).
					WithArgs("item-1", "plan-1", "ex-1", 0, 0, 0, 0.0, "", "user-1").
					WillReturnError(errDB)
			},
			wantErr: errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestConfig(t)

			repo, mock := newMockItemRepo(t)
			tt.mock(mock)

			got, err := repo.UpdateByID(context.Background(), tt.input, "user-1")

			if tt.wantErr != nil {
				if !stderrors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got != tt.want {
					t.Fatalf("expected %+v, got %+v", tt.want, got)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPostgresWorkoutPlanRepository_DeleteByID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		mock    func(mock sqlmock.Sqlmock)
		wantErr error
	}{
		{
			name: "success",
			id:   "plan-1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE\s+FROM\s+workout_plans\s+WHERE\s+id\s*=\s*\$1`).
					WithArgs("plan-1", "user-1").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "not found",
			id:   "missing",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE\s+FROM\s+workout_plans\s+WHERE\s+id\s*=\s*\$1`).
					WithArgs("missing", "user-1").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: apperrors.ErrNotFound,
		},
		{
			name: "database error",
			id:   "plan-1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE\s+FROM\s+workout_plans\s+WHERE\s+id\s*=\s*\$1`).
					WithArgs("plan-1", "user-1").
					WillReturnError(errDB)
			},
			wantErr: errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestConfig(t)

			repo, mock := newMockRepo(t)
			tt.mock(mock)

			err := repo.DeleteByID(context.Background(), tt.id, "user-1")

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

func TestPostgresWorkoutPlanItemRepository_DeleteByID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		mock    func(mock sqlmock.Sqlmock)
		wantErr error
	}{
		{
			name: "success",
			id:   "item-1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE\s+FROM\s+workout_plan_items\s+WHERE\s+id\s*=\s*\$1`).
					WithArgs("item-1", "user-1").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "not found",
			id:   "missing",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE\s+FROM\s+workout_plan_items\s+WHERE\s+id\s*=\s*\$1`).
					WithArgs("missing", "user-1").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: apperrors.ErrNotFound,
		},
		{
			name: "database error",
			id:   "item-1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE\s+FROM\s+workout_plan_items\s+WHERE\s+id\s*=\s*\$1`).
					WithArgs("item-1", "user-1").
					WillReturnError(errDB)
			},
			wantErr: errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestConfig(t)

			repo, mock := newMockItemRepo(t)
			tt.mock(mock)

			err := repo.DeleteByID(context.Background(), tt.id, "user-1")

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

func TestPostgresWorkoutPlanItemRepository_GetByID(t *testing.T) {
	fixedTime := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		name    string
		id      string
		mock    func(mock sqlmock.Sqlmock)
		want    []WorkoutPlanItem
		wantErr error
	}{
		{
			name: "success",
			id:   "plan-1",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "workout_plan_id", "exercise_id", "order_index", "sets", "reps", "weight", "unit", "created_at",
				}).
					AddRow("item-1", "plan-1", "ex-1", 0, 3, 10, 60.5, "kg", fixedTime).
					AddRow("item-2", "plan-1", "ex-2", 1, 4, 8, 80.0, "kg", fixedTime)

				mock.ExpectQuery(`SELECT\s+i\.id\s*,\s*i\.workout_plan_id.*FROM\s+workout_plan_items`).
					WithArgs("plan-1", "user-1").
					WillReturnRows(rows)
			},
			want: []WorkoutPlanItem{
				{ID: "item-1", WorkoutPlanID: "plan-1", ExerciseID: "ex-1", OrderIndex: 0, Sets: 3, Reps: 10, Weight: 60.5, Unit: "kg", CreatedAt: fixedTime},
				{ID: "item-2", WorkoutPlanID: "plan-1", ExerciseID: "ex-2", OrderIndex: 1, Sets: 4, Reps: 8, Weight: 80.0, Unit: "kg", CreatedAt: fixedTime},
			},
		},
		{
			name: "empty result",
			id:   "plan-missing-items",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "workout_plan_id", "exercise_id", "order_index", "sets", "reps", "weight", "unit", "created_at",
				})

				mock.ExpectQuery(`SELECT\s+i\.id\s*,\s*i\.workout_plan_id.*FROM\s+workout_plan_items`).
					WithArgs("plan-missing-items", "user-1").
					WillReturnRows(rows)
			},
			want: []WorkoutPlanItem{},
		},
		{
			name: "database error",
			id:   "plan-1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT\s+i\.id\s*,\s*i\.workout_plan_id.*FROM\s+workout_plan_items`).
					WithArgs("plan-1", "user-1").
					WillReturnError(errDB)
			},
			wantErr: errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestConfig(t)

			repo, mock := newMockItemRepo(t)
			tt.mock(mock)

			got, err := repo.GetByID(context.Background(), tt.id, "user-1")

			if tt.wantErr != nil {
				if !stderrors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(got) != len(tt.want) {
					t.Fatalf("expected %d items, got %d", len(tt.want), len(got))
				}
				for i := range got {
					if got[i] != tt.want[i] {
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
