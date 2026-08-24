package scheduledworkout

import (
	"context"
	"database/sql"
	stderrors "errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	apperrors "github.com/mmk31585/workout-tracker/internal/app_errors"
	"github.com/mmk31585/workout-tracker/internal/config"
)

var errDB = stderrors.New("db error")

func newMockRepo(t *testing.T) (*PostgresScheduledWorkoutRepository, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	sqlxDB := sqlx.NewDb(db, "postgres")

	return &PostgresScheduledWorkoutRepository{db: sqlxDB}, mock
}

func setupTestConfig(t *testing.T) {
	t.Helper()
	config.SetForTest(&config.Config{
		DB: config.DBConfig{QueryTimeout: 5},
	})
	t.Cleanup(config.ResetForTest)
}

func uniqueViolation() error {
	return &pgconn.PgError{Code: "23505", Message: "duplicate key value violates unique constraint"}
}

func strPtr(s string) *string { return &s }

func TestPostgresScheduledWorkoutRepository_Create(t *testing.T) {
	fixedTime := time.Date(2025, 5, 30, 12, 0, 0, 0, time.UTC)
	scheduledAt := time.Date(2025, 6, 1, 9, 0, 0, 0, time.UTC)
	notes := "morning run"

	tests := []struct {
		name    string
		input   ScheduledWorkout
		mock    func(mock sqlmock.Sqlmock)
		want    ScheduledWorkout
		wantErr error
	}{
		{
			name: "success",
			input: ScheduledWorkout{
				UserID:        "user-1",
				WorkoutPlanID: "plan-1",
				ScheduledAt:   scheduledAt,
				Status:        ScheduleScheduled,
				Notes:         &notes,
			},
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "user_id", "workout_plan_id", "scheduled_at", "status", "notes", "created_at", "updated_at",
				}).AddRow("sw-1", "user-1", "plan-1", scheduledAt, ScheduleScheduled, notes, fixedTime, fixedTime)

				mock.ExpectQuery(`INSERT\s+INTO\s+scheduled_workouts`).
					WithArgs("user-1", "plan-1", scheduledAt, ScheduleScheduled, &notes).
					WillReturnRows(rows)
			},
			want: ScheduledWorkout{
				ID:            "sw-1",
				UserID:        "user-1",
				WorkoutPlanID: "plan-1",
				ScheduledAt:   scheduledAt,
				Status:        ScheduleScheduled,
				Notes:         &notes,
				CreatedAt:     fixedTime,
				UpdatedAt:     fixedTime,
			},
		},
		{
			name: "success with nil notes",
			input: ScheduledWorkout{
				UserID:        "user-1",
				WorkoutPlanID: "missing-plan",
				ScheduledAt:   scheduledAt,
				Status:        ScheduleScheduled,
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`INSERT\s+INTO\s+scheduled_workouts`).
					WithArgs("user-1", "missing-plan", scheduledAt, ScheduleScheduled, (*string)(nil)).
					WillReturnError(errDB)
			},
			wantErr: errDB,
		},
		{
			name: "database error",
			input: ScheduledWorkout{
				UserID:        "user-1",
				WorkoutPlanID: "plan-1",
				ScheduledAt:   scheduledAt,
				Status:        ScheduleScheduled,
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`INSERT\s+INTO\s+scheduled_workouts`).
					WithArgs("user-1", "plan-1", scheduledAt, ScheduleScheduled, (*string)(nil)).
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

			got, err := repo.Create(context.Background(), tt.input)

			if tt.wantErr != nil {
				if !stderrors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.ID != tt.want.ID || got.UserID != tt.want.UserID || got.WorkoutPlanID != tt.want.WorkoutPlanID {
					t.Fatalf("expected %+v, got %+v", tt.want, got)
				}
				if got.Status != tt.want.Status || !got.ScheduledAt.Equal(tt.want.ScheduledAt) {
					t.Fatalf("expected %+v, got %+v", tt.want, got)
				}
				if (got.Notes == nil) != (tt.want.Notes == nil) ||
					(got.Notes != nil && *got.Notes != *tt.want.Notes) {
					t.Fatalf("notes mismatch: expected %v, got %v", tt.want.Notes, got.Notes)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPostgresScheduledWorkoutRepository_GetByID(t *testing.T) {
	fixedTime := time.Date(2025, 5, 30, 12, 0, 0, 0, time.UTC)
	scheduledAt := time.Date(2025, 6, 1, 9, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		id      string
		userID  string
		mock    func(mock sqlmock.Sqlmock)
		want    ScheduledWorkout
		wantErr error
	}{
		{
			name:   "success scoped by user",
			id:     "sw-1",
			userID: "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "user_id", "workout_plan_id", "scheduled_at", "status", "notes", "created_at", "updated_at",
				}).AddRow("sw-1", "user-1", "plan-1", scheduledAt, ScheduleScheduled, nil, fixedTime, fixedTime)

				mock.ExpectQuery(`SELECT\s+.*FROM\s+scheduled_workouts\s+WHERE\s+id\s*=\s*\$1`).
					WithArgs("sw-1", "user-1").
					WillReturnRows(rows)
			},
			want: ScheduledWorkout{
				ID:            "sw-1",
				UserID:        "user-1",
				WorkoutPlanID: "plan-1",
				ScheduledAt:   scheduledAt,
				Status:        ScheduleScheduled,
				CreatedAt:     fixedTime,
				UpdatedAt:     fixedTime,
			},
		},
		{
			name:   "not found",
			id:     "sw-other-users",
			userID: "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT\s+.*FROM\s+scheduled_workouts\s+WHERE\s+id\s*=\s*\$1`).
					WithArgs("sw-other-users", "user-1").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: apperrors.ErrNotFound,
		},
		{
			name:   "database error",
			id:     "sw-1",
			userID: "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT\s+.*FROM\s+scheduled_workouts\s+WHERE\s+id\s*=\s*\$1`).
					WithArgs("sw-1", "user-1").
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

			got, err := repo.GetByID(context.Background(), tt.id, tt.userID)

			if tt.wantErr != nil {
				if !stderrors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.ID != tt.want.ID || got.UserID != tt.want.UserID || got.Status != tt.want.Status {
					t.Fatalf("expected %+v, got %+v", tt.want, got)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPostgresScheduledWorkoutRepository_GetByUserID(t *testing.T) {
	fixedTime := time.Date(2025, 5, 30, 12, 0, 0, 0, time.UTC)
	scheduledAt := time.Date(2025, 6, 1, 9, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		userID  string
		mock    func(mock sqlmock.Sqlmock)
		want    []ScheduledWorkout
		wantErr error
	}{
		{
			name:   "success",
			userID: "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "user_id", "workout_plan_id", "scheduled_at", "status", "notes", "created_at", "updated_at",
				}).
					AddRow("sw-1", "user-1", "plan-1", scheduledAt, ScheduleScheduled, nil, fixedTime, fixedTime).
					AddRow("sw-2", "user-1", "plan-2", scheduledAt, ScheduleCompleted, "done", fixedTime, fixedTime)

				mock.ExpectQuery(`SELECT\s+.*FROM\s+scheduled_workouts\s+WHERE\s+user_id\s*=\s*\$1`).
					WithArgs("user-1").
					WillReturnRows(rows)
			},
			want: []ScheduledWorkout{
				{ID: "sw-1", UserID: "user-1", WorkoutPlanID: "plan-1", ScheduledAt: scheduledAt, Status: ScheduleScheduled, CreatedAt: fixedTime, UpdatedAt: fixedTime},
				{ID: "sw-2", UserID: "user-1", WorkoutPlanID: "plan-2", ScheduledAt: scheduledAt, Status: ScheduleCompleted, Notes: strPtr("done"), CreatedAt: fixedTime, UpdatedAt: fixedTime},
			},
		},
		{
			name:   "empty result returns empty slice not error",
			userID: "user-empty",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "user_id", "workout_plan_id", "scheduled_at", "status", "notes", "created_at", "updated_at",
				})

				mock.ExpectQuery(`SELECT\s+.*FROM\s+scheduled_workouts\s+WHERE\s+user_id\s*=\s*\$1`).
					WithArgs("user-empty").
					WillReturnRows(rows)
			},
			want: []ScheduledWorkout{},
		},
		{
			name:   "database error",
			userID: "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT\s+.*FROM\s+scheduled_workouts\s+WHERE\s+user_id\s*=\s*\$1`).
					WithArgs("user-1").
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

			got, err := repo.GetByUserID(context.Background(), tt.userID)

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
					if got[i].ID != tt.want[i].ID || got[i].Status != tt.want[i].Status {
						t.Fatalf("row %d: expected %+v, got %+v", i, tt.want[i], got[i])
					}
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPostgresScheduledWorkoutRepository_Update(t *testing.T) {
	fixedTime := time.Date(2025, 5, 30, 12, 0, 0, 0, time.UTC)
	newScheduledAt := time.Date(2025, 6, 3, 18, 0, 0, 0, time.UTC)
	newNotes := "moved to evening"

	tests := []struct {
		name    string
		input   ScheduledWorkout
		userID  string
		mock    func(mock sqlmock.Sqlmock)
		want    ScheduledWorkout
		wantErr error
	}{
		{
			name: "success scoped by user",
			input: ScheduledWorkout{
				ID:          "sw-1",
				ScheduledAt: newScheduledAt,
				Notes:       &newNotes,
			},
			userID: "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "user_id", "workout_plan_id", "scheduled_at", "status", "notes", "created_at", "updated_at",
				}).AddRow("sw-1", "user-1", "plan-1", newScheduledAt, ScheduleScheduled, newNotes, fixedTime, fixedTime)

				mock.ExpectQuery(`UPDATE\s+scheduled_workouts\s+SET`).
					WithArgs("sw-1", newScheduledAt, &newNotes, "user-1").
					WillReturnRows(rows)
			},
			want: ScheduledWorkout{
				ID:            "sw-1",
				UserID:        "user-1",
				WorkoutPlanID: "plan-1",
				ScheduledAt:   newScheduledAt,
				Status:        ScheduleScheduled,
				Notes:         &newNotes,
				CreatedAt:     fixedTime,
				UpdatedAt:     fixedTime,
			},
		},
		{
			name: "not found or owned by another user",
			input: ScheduledWorkout{
				ID:          "sw-missing",
				ScheduledAt: newScheduledAt,
			},
			userID: "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE\s+scheduled_workouts\s+SET`).
					WithArgs("sw-missing", newScheduledAt, (*string)(nil), "user-1").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: apperrors.ErrNotFound,
		},
		{
			name: "database error",
			input: ScheduledWorkout{
				ID:          "sw-1",
				ScheduledAt: newScheduledAt,
			},
			userID: "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE\s+scheduled_workouts\s+SET`).
					WithArgs("sw-1", newScheduledAt, (*string)(nil), "user-1").
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

			got, err := repo.Update(context.Background(), tt.input, tt.userID)

			if tt.wantErr != nil {
				if !stderrors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.ID != tt.want.ID || got.ScheduledAt != tt.want.ScheduledAt || got.Status != tt.want.Status {
					t.Fatalf("expected %+v, got %+v", tt.want, got)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPostgresScheduledWorkoutRepository_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		userID  string
		mock    func(mock sqlmock.Sqlmock)
		wantErr error
	}{
		{
			name:   "success scoped by user",
			id:     "sw-1",
			userID: "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE\s+FROM\s+scheduled_workouts\s+WHERE\s+id\s*=\s*\$1`).
					WithArgs("sw-1", "user-1").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:   "not found or owned by another user",
			id:     "sw-missing",
			userID: "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE\s+FROM\s+scheduled_workouts\s+WHERE\s+id\s*=\s*\$1`).
					WithArgs("sw-missing", "user-1").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: apperrors.ErrNotFound,
		},
		{
			name:   "database error",
			id:     "sw-1",
			userID: "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE\s+FROM\s+scheduled_workouts\s+WHERE\s+id\s*=\s*\$1`).
					WithArgs("sw-1", "user-1").
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

// TestPostgresScheduledWorkoutRepository_Transition is the critical test:
// the atomic guard UPDATE ... WHERE id=$1 AND user_id=$3 AND status=$4.
// The service validates legality; the repo guarantees only a matching row moves.
func TestPostgresScheduledWorkoutRepository_Transition(t *testing.T) {
	fixedTime := time.Date(2025, 5, 30, 12, 0, 0, 0, time.UTC)
	scheduledAt := time.Date(2025, 6, 1, 9, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		id      string
		userID  string
		from    ScheduleStatus
		to      ScheduleStatus
		mock    func(mock sqlmock.Sqlmock)
		want    ScheduledWorkout
		wantErr error
	}{
		{
			name:   "success scheduled to in_progress",
			id:     "sw-1",
			userID: "user-1",
			from:   ScheduleScheduled,
			to:     ScheduleInProgress,
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "user_id", "workout_plan_id", "scheduled_at", "status", "notes", "created_at", "updated_at",
				}).AddRow("sw-1", "user-1", "plan-1", scheduledAt, ScheduleInProgress, nil, fixedTime, fixedTime)

				mock.ExpectQuery(`UPDATE\s+scheduled_workouts\s+SET`).
					WithArgs("sw-1", string(ScheduleInProgress), "user-1", string(ScheduleScheduled)).
					WillReturnRows(rows)
			},
			want: ScheduledWorkout{
				ID:            "sw-1",
				UserID:        "user-1",
				WorkoutPlanID: "plan-1",
				ScheduledAt:   scheduledAt,
				Status:        ScheduleInProgress,
				CreatedAt:     fixedTime,
				UpdatedAt:     fixedTime,
			},
		},
		{
			name:   "status mismatch matches zero rows returns not found",
			id:     "sw-1",
			userID: "user-1",
			from:   ScheduleCompleted,
			to:     ScheduleCancelled,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE\s+scheduled_workouts\s+SET`).
					WithArgs("sw-1", string(ScheduleCancelled), "user-1", string(ScheduleCompleted)).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: apperrors.ErrNotFound,
		},
		{
			name:   "another users workout is invisible",
			id:     "sw-someone-elses",
			userID: "user-1",
			from:   ScheduleScheduled,
			to:     ScheduleCancelled,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE\s+scheduled_workouts\s+SET`).
					WithArgs("sw-someone-elses", string(ScheduleCancelled), "user-1", string(ScheduleScheduled)).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: apperrors.ErrNotFound,
		},
		{
			name:   "database error",
			id:     "sw-1",
			userID: "user-1",
			from:   ScheduleScheduled,
			to:     ScheduleInProgress,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE\s+scheduled_workouts\s+SET`).
					WithArgs("sw-1", string(ScheduleInProgress), "user-1", string(ScheduleScheduled)).
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

			got, err := repo.Transition(context.Background(), tt.id, tt.userID, tt.from, tt.to)

			if tt.wantErr != nil {
				if !stderrors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.ID != tt.want.ID || got.Status != tt.want.Status {
					t.Fatalf("expected %+v, got %+v", tt.want, got)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestNewScheduledWorkoutRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewScheduledWorkoutRepository(sqlx.NewDb(db, "postgres"))
	if repo == nil {
		t.Fatal("NewScheduledWorkoutRepository returned nil")
	}
}
