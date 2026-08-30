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
	"github.com/mmk31585/workout-tracker/internal/config"
)

var errDB = stderrors.New("db error")

func newMockRepo(t *testing.T) (*PostgresWorkoutSessionRepository, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	sqlxDB := sqlx.NewDb(db, "postgres")

	return &PostgresWorkoutSessionRepository{db: sqlxDB}, mock
}

func newMockItemRepo(t *testing.T) (*PostgresWorkoutSessionItemRepository, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	sqlxDB := sqlx.NewDb(db, "postgres")

	return &PostgresWorkoutSessionItemRepository{db: sqlxDB}, mock
}

func setupTestConfig(t *testing.T) {
	t.Helper()
	config.SetForTest(&config.Config{
		DB: config.DBConfig{QueryTimeout: 5},
	})
	t.Cleanup(config.ResetForTest)
}


func TestPostgresWorkoutSessionRepository_Create(t *testing.T) {
	fixedTime := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	performedAt := time.Date(2025, 1, 5, 10, 0, 0, 0, time.UTC)
	notes := "felt strong"

	tests := []struct {
		name    string
		input   WorkoutSession
		mock    func(mock sqlmock.Sqlmock)
		want    WorkoutSession
		wantErr bool
	}{
		{
			name: "success",
			input: WorkoutSession{
				UserID:             "user-1",
				ScheduledWorkoutID: "sw-1",
				PerformedAt:        &performedAt,
				OverallNotes:       &notes,
			},
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "user_id", "scheduled_workout_id", "performed_at", "overall_notes", "created_at", "updated_at",
				}).AddRow("sess-1", "user-1", "sw-1", performedAt, notes, fixedTime, fixedTime)

				mock.ExpectQuery(`INSERT\s+INTO\s+workout_sessions`).
					WithArgs("user-1", "sw-1", &performedAt, &notes).
					WillReturnRows(rows)
			},
			want: WorkoutSession{
				ID:                 "sess-1",
				UserID:             "user-1",
				ScheduledWorkoutID: "sw-1",
				PerformedAt:        &performedAt,
				OverallNotes:       &notes,
				CreatedAt:          fixedTime,
				UpdatedAt:          fixedTime,
			},
		},
		{
			name: "success with nil performed_at uses NOW",
			input: WorkoutSession{
				UserID:             "user-1",
				ScheduledWorkoutID: "sw-1",
				PerformedAt:        nil,
				OverallNotes:       nil,
			},
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "user_id", "scheduled_workout_id", "performed_at", "overall_notes", "created_at", "updated_at",
				}).AddRow("sess-1", "user-1", "sw-1", fixedTime, nil, fixedTime, fixedTime)

				mock.ExpectQuery(`INSERT\s+INTO\s+workout_sessions`).
					WithArgs("user-1", "sw-1", (*time.Time)(nil), (*string)(nil)).
					WillReturnRows(rows)
			},
			want: WorkoutSession{
				ID:                 "sess-1",
				UserID:             "user-1",
				ScheduledWorkoutID: "sw-1",
				PerformedAt:        &fixedTime,
				OverallNotes:       nil,
				CreatedAt:          fixedTime,
				UpdatedAt:          fixedTime,
			},
		},
		{
			name: "database error",
			input: WorkoutSession{
				UserID:             "user-1",
				ScheduledWorkoutID: "sw-1",
				PerformedAt:        &performedAt,
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`INSERT\s+INTO\s+workout_sessions`).
					WithArgs("user-1", "sw-1", &performedAt, (*string)(nil)).
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
				if got.ID != tt.want.ID || got.UserID != tt.want.UserID || got.ScheduledWorkoutID != tt.want.ScheduledWorkoutID {
					t.Fatalf("expected %+v, got %+v", tt.want, got)
				}
				if tt.want.PerformedAt != nil && got.PerformedAt != nil && !tt.want.PerformedAt.Equal(*got.PerformedAt) {
					t.Fatalf("performed_at mismatch: want %v got %v", *tt.want.PerformedAt, *got.PerformedAt)
				}
				if (tt.want.PerformedAt == nil) != (got.PerformedAt == nil) {
					t.Fatalf("performed_at nil mismatch: want %v got %v", tt.want.PerformedAt, got.PerformedAt)
				}
				if got.CreatedAt != tt.want.CreatedAt || got.UpdatedAt != tt.want.UpdatedAt {
					t.Fatalf("timestamps mismatch: want %v/%v got %v/%v", tt.want.CreatedAt, tt.want.UpdatedAt, got.CreatedAt, got.UpdatedAt)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPostgresWorkoutSessionRepository_GetByID(t *testing.T) {
	fixedTime := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	performedAt := time.Date(2025, 1, 5, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		id      string
		userID  string
		mock    func(mock sqlmock.Sqlmock)
		want    WorkoutSession
		wantErr error
	}{
		{
			name:   "success",
			id:     "sess-1",
			userID: "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "user_id", "scheduled_workout_id", "performed_at", "overall_notes", "created_at", "updated_at",
				}).AddRow("sess-1", "user-1", "sw-1", performedAt, "notes", fixedTime, fixedTime)

				mock.ExpectQuery(`SELECT\s+.*FROM\s+workout_sessions\s+WHERE\s+id\s*=\s*\$1`).
					WithArgs("sess-1", "user-1").
					WillReturnRows(rows)
			},
			want: WorkoutSession{
				ID:                 "sess-1",
				UserID:             "user-1",
				ScheduledWorkoutID: "sw-1",
				PerformedAt:        &performedAt,
				OverallNotes:       new("notes"),
				CreatedAt:          fixedTime,
				UpdatedAt:          fixedTime,
			},
		},
		{
			name:   "not found",
			id:     "missing",
			userID: "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT\s+.*FROM\s+workout_sessions\s+WHERE\s+id\s*=\s*\$1`).
					WithArgs("missing", "user-1").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: apperrors.ErrNotFound,
		},
		{
			name:   "database error",
			id:     "sess-1",
			userID: "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT\s+.*FROM\s+workout_sessions\s+WHERE\s+id\s*=\s*\$1`).
					WithArgs("sess-1", "user-1").
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
				if got.ID != tt.want.ID || got.UserID != tt.want.UserID {
					t.Fatalf("expected %+v, got %+v", tt.want, got)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPostgresWorkoutSessionRepository_GetByUserID(t *testing.T) {
	fixedTime := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	performedAt := time.Date(2025, 1, 5, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		userID  string
		mock    func(mock sqlmock.Sqlmock)
		want    []WorkoutSession
		wantErr error
	}{
		{
			name:   "success",
			userID: "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "user_id", "scheduled_workout_id", "performed_at", "overall_notes", "created_at", "updated_at",
				}).
					AddRow("sess-1", "user-1", "sw-1", performedAt, "notes", fixedTime, fixedTime).
					AddRow("sess-2", "user-1", "sw-2", performedAt, nil, fixedTime, fixedTime)

				mock.ExpectQuery(`SELECT\s+.*FROM\s+workout_sessions\s+WHERE\s+user_id\s*=\s*\$1`).
					WithArgs("user-1").
					WillReturnRows(rows)
			},
			want: []WorkoutSession{
				{ID: "sess-1", UserID: "user-1", ScheduledWorkoutID: "sw-1", PerformedAt: &performedAt, OverallNotes: new("notes"), CreatedAt: fixedTime, UpdatedAt: fixedTime},
				{ID: "sess-2", UserID: "user-1", ScheduledWorkoutID: "sw-2", PerformedAt: &performedAt, CreatedAt: fixedTime, UpdatedAt: fixedTime},
			},
		},
		{
			name:   "empty result",
			userID: "user-empty",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "user_id", "scheduled_workout_id", "performed_at", "overall_notes", "created_at", "updated_at",
				})

				mock.ExpectQuery(`SELECT\s+.*FROM\s+workout_sessions\s+WHERE\s+user_id\s*=\s*\$1`).
					WithArgs("user-empty").
					WillReturnRows(rows)
			},
			want: []WorkoutSession{},
		},
		{
			name:   "database error",
			userID: "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT\s+.*FROM\s+workout_sessions\s+WHERE\s+user_id\s*=\s*\$1`).
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
					if got[i].ID != tt.want[i].ID || got[i].UserID != tt.want[i].UserID {
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

func TestPostgresWorkoutSessionRepository_GetByScheduledWorkoutID(t *testing.T) {
	fixedTime := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	performedAt := time.Date(2025, 1, 5, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name               string
		scheduledWorkoutID string
		userID             string
		mock               func(mock sqlmock.Sqlmock)
		want               WorkoutSession
		wantErr            error
	}{
		{
			name:               "success",
			scheduledWorkoutID: "sw-1",
			userID:             "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "user_id", "scheduled_workout_id", "performed_at", "overall_notes", "created_at", "updated_at",
				}).AddRow("sess-1", "user-1", "sw-1", performedAt, "notes", fixedTime, fixedTime)

				mock.ExpectQuery(`SELECT\s+.*FROM\s+workout_sessions`).
					WithArgs("sw-1", "user-1").
					WillReturnRows(rows)
			},
			want: WorkoutSession{
				ID:                 "sess-1",
				UserID:             "user-1",
				ScheduledWorkoutID: "sw-1",
				PerformedAt:        &performedAt,
				OverallNotes:       new("notes"),
				CreatedAt:          fixedTime,
				UpdatedAt:          fixedTime,
			},
		},
		{
			name:               "not found",
			scheduledWorkoutID: "missing",
			userID:             "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT\s+.*FROM\s+workout_sessions`).
					WithArgs("missing", "user-1").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: apperrors.ErrNotFound,
		},
		{
			name:               "database error",
			scheduledWorkoutID: "sw-1",
			userID:             "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT\s+.*FROM\s+workout_sessions`).
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

			got, err := repo.GetByScheduledWorkoutID(context.Background(), tt.scheduledWorkoutID, tt.userID)

			if tt.wantErr != nil {
				if !stderrors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.ID != tt.want.ID || got.ScheduledWorkoutID != tt.want.ScheduledWorkoutID {
					t.Fatalf("expected %+v, got %+v", tt.want, got)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPostgresWorkoutSessionRepository_Update(t *testing.T) {
	fixedTime := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	performedAt := time.Date(2025, 1, 6, 10, 0, 0, 0, time.UTC)
	notes := "updated notes"

	tests := []struct {
		name    string
		input   WorkoutSession
		userID  string
		mock    func(mock sqlmock.Sqlmock)
		want    WorkoutSession
		wantErr error
	}{
		{
			name: "success",
			input: WorkoutSession{
				ID:           "sess-1",
				PerformedAt:  &performedAt,
				OverallNotes: &notes,
			},
			userID: "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "user_id", "scheduled_workout_id", "performed_at", "overall_notes", "created_at", "updated_at",
				}).AddRow("sess-1", "user-1", "sw-1", performedAt, notes, fixedTime, fixedTime)

				mock.ExpectQuery(`UPDATE\s+workout_sessions\s+SET`).
					WithArgs("sess-1", &performedAt, &notes, "user-1").
					WillReturnRows(rows)
			},
			want: WorkoutSession{
				ID:                 "sess-1",
				UserID:             "user-1",
				ScheduledWorkoutID: "sw-1",
				PerformedAt:        &performedAt,
				OverallNotes:       &notes,
				CreatedAt:          fixedTime,
				UpdatedAt:          fixedTime,
			},
		},
		{
			name: "not found",
			input: WorkoutSession{
				ID:          "missing",
				PerformedAt: &performedAt,
			},
			userID: "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE\s+workout_sessions\s+SET`).
					WithArgs("missing", &performedAt, (*string)(nil), "user-1").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: apperrors.ErrNotFound,
		},
		{
			name: "database error",
			input: WorkoutSession{
				ID:          "sess-1",
				PerformedAt: &performedAt,
			},
			userID: "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE\s+workout_sessions\s+SET`).
					WithArgs("sess-1", &performedAt, (*string)(nil), "user-1").
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
				if got.ID != tt.want.ID {
					t.Fatalf("expected %+v, got %+v", tt.want, got)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPostgresWorkoutSessionRepository_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		userID  string
		mock    func(mock sqlmock.Sqlmock)
		wantErr error
	}{
		{
			name:   "success",
			id:     "sess-1",
			userID: "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE\s+FROM\s+workout_sessions\s+WHERE\s+id\s*=\s*\$1`).
					WithArgs("sess-1", "user-1").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:   "not found",
			id:     "missing",
			userID: "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE\s+FROM\s+workout_sessions\s+WHERE\s+id\s*=\s*\$1`).
					WithArgs("missing", "user-1").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: apperrors.ErrNotFound,
		},
		{
			name:   "database error",
			id:     "sess-1",
			userID: "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE\s+FROM\s+workout_sessions\s+WHERE\s+id\s*=\s*\$1`).
					WithArgs("sess-1", "user-1").
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

func TestNewWorkoutSessionRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewWorkoutSessionRepository(sqlx.NewDb(db, "postgres"))
	if repo == nil {
		t.Fatal("NewWorkoutSessionRepository returned nil")
	}
}
