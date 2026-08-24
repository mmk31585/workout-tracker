package workoutsession

import (
	"context"
	"database/sql"
	stderrors "errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	apperrors "github.com/mmk31585/workout-tracker/internal/app_errors"
)

func TestPostgresWorkoutSessionItemRepository_Create(t *testing.T) {
	fixedTime := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	queryPattern := `INSERT\s+INTO\s+workout_session_items`
	notes := "hard set"

	newRow := func(id string, item WorkoutSessionItem) *sqlmock.Rows {
		return sqlmock.NewRows([]string{
			"id", "workout_session_id", "exercise_id", "sets", "reps", "weight", "unit", "notes", "created_at",
		}).AddRow(
			id,
			item.WorkoutSessionID,
			item.ExerciseID,
			item.Sets,
			item.Reps,
			item.Weight,
			item.Unit,
			item.Notes,
			fixedTime,
		)
	}

	tests := []struct {
		name    string
		input   []WorkoutSessionItem
		mock    func(mock sqlmock.Sqlmock)
		want    []WorkoutSessionItem
		wantErr bool
	}{
		{
			name: "success with multiple items",
			input: []WorkoutSessionItem{
				{WorkoutSessionID: "sess-1", ExerciseID: "ex-1", Sets: 3, Reps: 10, Weight: 60.5, Unit: "kg", Notes: &notes},
				{WorkoutSessionID: "sess-1", ExerciseID: "ex-2", Sets: 4, Reps: 8, Weight: 80, Unit: "kg"},
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(queryPattern).
					WithArgs("sess-1", "ex-1", 3, 10, 60.5, "kg", &notes).
					WillReturnRows(newRow("item-1", WorkoutSessionItem{WorkoutSessionID: "sess-1", ExerciseID: "ex-1", Sets: 3, Reps: 10, Weight: 60.5, Unit: "kg", Notes: &notes}))
				mock.ExpectQuery(queryPattern).
					WithArgs("sess-1", "ex-2", 4, 8, 80.0, "kg", (*string)(nil)).
					WillReturnRows(newRow("item-2", WorkoutSessionItem{WorkoutSessionID: "sess-1", ExerciseID: "ex-2", Sets: 4, Reps: 8, Weight: 80, Unit: "kg"}))
				mock.ExpectCommit()
			},
			want: []WorkoutSessionItem{
				{ID: "item-1", WorkoutSessionID: "sess-1", ExerciseID: "ex-1", Sets: 3, Reps: 10, Weight: 60.5, Unit: "kg", Notes: &notes, CreatedAt: fixedTime},
				{ID: "item-2", WorkoutSessionID: "sess-1", ExerciseID: "ex-2", Sets: 4, Reps: 8, Weight: 80, Unit: "kg", CreatedAt: fixedTime},
			},
		},
		{
			name:  "empty input inserts nothing",
			input: []WorkoutSessionItem{},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectCommit()
			},
			want: []WorkoutSessionItem{},
		},
		{
			name: "database error",
			input: []WorkoutSessionItem{
				{WorkoutSessionID: "sess-1", ExerciseID: "ex-1", Sets: 3, Reps: 10, Weight: 60, Unit: "kg"},
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(queryPattern).
					WithArgs("sess-1", "ex-1", 3, 10, 60.0, "kg", (*string)(nil)).
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
					if got[i].ID != tt.want[i].ID || got[i].WorkoutSessionID != tt.want[i].WorkoutSessionID || got[i].ExerciseID != tt.want[i].ExerciseID {
						t.Fatalf("expected %+v, got %+v", tt.want[i], got[i])
					}
					if got[i].Sets != tt.want[i].Sets || got[i].Reps != tt.want[i].Reps || got[i].Weight != tt.want[i].Weight || got[i].Unit != tt.want[i].Unit {
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

func TestPostgresWorkoutSessionItemRepository_GetBySessionID(t *testing.T) {
	fixedTime := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		name      string
		sessionID string
		userID    string
		mock      func(mock sqlmock.Sqlmock)
		want      []WorkoutSessionItem
		wantErr   error
	}{
		{
			name:      "success",
			sessionID: "sess-1",
			userID:    "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "workout_session_id", "exercise_id", "sets", "reps", "weight", "unit", "notes", "created_at",
				}).
					AddRow("item-1", "sess-1", "ex-1", 3, 10, 60.5, "kg", "good", fixedTime).
					AddRow("item-2", "sess-1", "ex-2", 4, 8, 80.0, "kg", nil, fixedTime)

				mock.ExpectQuery(`SELECT\s+i\.id\s*,\s*i\.workout_session_id.*FROM\s+workout_session_items`).
					WithArgs("sess-1", "user-1").
					WillReturnRows(rows)
			},
			want: []WorkoutSessionItem{
				{ID: "item-1", WorkoutSessionID: "sess-1", ExerciseID: "ex-1", Sets: 3, Reps: 10, Weight: 60.5, Unit: "kg", Notes: new("good"), CreatedAt: fixedTime},
				{ID: "item-2", WorkoutSessionID: "sess-1", ExerciseID: "ex-2", Sets: 4, Reps: 8, Weight: 80.0, Unit: "kg", CreatedAt: fixedTime},
			},
		},
		{
			name:      "empty result",
			sessionID: "sess-empty",
			userID:    "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "workout_session_id", "exercise_id", "sets", "reps", "weight", "unit", "notes", "created_at",
				})

				mock.ExpectQuery(`SELECT\s+i\.id\s*,\s*i\.workout_session_id.*FROM\s+workout_session_items`).
					WithArgs("sess-empty", "user-1").
					WillReturnRows(rows)
			},
			want: []WorkoutSessionItem{},
		},
		{
			name:      "database error",
			sessionID: "sess-1",
			userID:    "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT\s+i\.id\s*,\s*i\.workout_session_id.*FROM\s+workout_session_items`).
					WithArgs("sess-1", "user-1").
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

			got, err := repo.GetBySessionID(context.Background(), tt.sessionID, tt.userID)

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
					if got[i].ID != tt.want[i].ID || got[i].Sets != tt.want[i].Sets {
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

func TestPostgresWorkoutSessionItemRepository_Update(t *testing.T) {
	fixedTime := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	notes := "updated"

	tests := []struct {
		name    string
		input   WorkoutSessionItem
		userID  string
		mock    func(mock sqlmock.Sqlmock)
		want    WorkoutSessionItem
		wantErr error
	}{
		{
			name: "success",
			input: WorkoutSessionItem{
				ID:               "item-1",
				WorkoutSessionID: "sess-1",
				ExerciseID:       "ex-1",
				Sets:             5,
				Reps:             12,
				Weight:           70.5,
				Unit:             "kg",
				Notes:            &notes,
			},
			userID: "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "workout_session_id", "exercise_id", "sets", "reps", "weight", "unit", "notes", "created_at",
				}).AddRow("item-1", "sess-1", "ex-1", 5, 12, 70.5, "kg", notes, fixedTime)

				mock.ExpectQuery(`UPDATE\s+workout_session_items\s+SET`).
					WithArgs("item-1", "sess-1", "ex-1", 5, 12, 70.5, "kg", &notes, "user-1").
					WillReturnRows(rows)
			},
			want: WorkoutSessionItem{
				ID:               "item-1",
				WorkoutSessionID: "sess-1",
				ExerciseID:       "ex-1",
				Sets:             5,
				Reps:             12,
				Weight:           70.5,
				Unit:             "kg",
				Notes:            &notes,
				CreatedAt:        fixedTime,
			},
		},
		{
			name: "not found",
			input: WorkoutSessionItem{
				ID:               "missing",
				WorkoutSessionID: "sess-1",
				ExerciseID:       "ex-1",
			},
			userID: "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE\s+workout_session_items\s+SET`).
					WithArgs("missing", "sess-1", "ex-1", 0, 0, 0.0, "", (*string)(nil), "user-1").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: apperrors.ErrNotFound,
		},
		{
			name: "database error",
			input: WorkoutSessionItem{
				ID:               "item-1",
				WorkoutSessionID: "sess-1",
				ExerciseID:       "ex-1",
			},
			userID: "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE\s+workout_session_items\s+SET`).
					WithArgs("item-1", "sess-1", "ex-1", 0, 0, 0.0, "", (*string)(nil), "user-1").
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

			got, err := repo.Update(context.Background(), tt.input, tt.userID)

			if tt.wantErr != nil {
				if !stderrors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.ID != tt.want.ID || got.Sets != tt.want.Sets {
					t.Fatalf("expected %+v, got %+v", tt.want, got)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPostgresWorkoutSessionItemRepository_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		userID  string
		mock    func(mock sqlmock.Sqlmock)
		wantErr error
	}{
		{
			name:   "success",
			id:     "item-1",
			userID: "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE\s+FROM\s+workout_session_items\s+WHERE\s+id\s*=\s*\$1`).
					WithArgs("item-1", "user-1").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:   "not found",
			id:     "missing",
			userID: "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE\s+FROM\s+workout_session_items\s+WHERE\s+id\s*=\s*\$1`).
					WithArgs("missing", "user-1").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: apperrors.ErrNotFound,
		},
		{
			name:   "database error",
			id:     "item-1",
			userID: "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE\s+FROM\s+workout_session_items\s+WHERE\s+id\s*=\s*\$1`).
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

			err := repo.Delete(context.Background(), tt.id, tt.userID)

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

func TestNewWorkoutSessionItemRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewWorkoutSessionItemRepository(sqlx.NewDb(db, "postgres"))
	if repo == nil {
		t.Fatal("NewWorkoutSessionItemRepository returned nil")
	}
}
