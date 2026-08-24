package user

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

func newMockRepo(t *testing.T) (*PostgresUserRepository, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	sqlxDB := sqlx.NewDb(db, "postgres")

	return &PostgresUserRepository{db: sqlxDB}, mock
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

func TestPostgresUserRepository_Create(t *testing.T) {
	fixedTime := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		name    string
		input   User
		mock    func(mock sqlmock.Sqlmock)
		want    User
		wantErr error
	}{
		{
			name: "success",
			input: User{
				Email:        "user@test.com",
				DisplayName:  "Test User",
				PasswordHash: "$2a$12$hashedpassword",
			},
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "email", "display_name", "password_hash", "created_at", "updated_at",
				}).AddRow("user-1", "user@test.com", "Test User", "$2a$12$hashedpassword", fixedTime, fixedTime)

				mock.ExpectQuery(`INSERT\s+INTO\s+users`).
					WithArgs("user@test.com", "Test User", "$2a$12$hashedpassword").
					WillReturnRows(rows)
			},
			want: User{
				ID:           "user-1",
				Email:        "user@test.com",
				DisplayName:  "Test User",
				PasswordHash: "$2a$12$hashedpassword",
				CreatedAt:    fixedTime,
				UpdatedAt:    fixedTime,
			},
		},
		{
			name: "email already exists returns conflict",
			input: User{
				Email:        "taken@test.com",
				DisplayName:  "Test User",
				PasswordHash: "$2a$12$hashedpassword",
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`INSERT\s+INTO\s+users`).
					WithArgs("taken@test.com", "Test User", "$2a$12$hashedpassword").
					WillReturnError(uniqueViolation())
			},
			wantErr: apperrors.ErrConflict,
		},
		{
			name: "database error",
			input: User{
				Email:        "user@test.com",
				DisplayName:  "Test User",
				PasswordHash: "$2a$12$hashedpassword",
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`INSERT\s+INTO\s+users`).
					WithArgs("user@test.com", "Test User", "$2a$12$hashedpassword").
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

func TestPostgresUserRepository_GetByID(t *testing.T) {
	fixedTime := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		name    string
		id      string
		mock    func(mock sqlmock.Sqlmock)
		want    User
		wantErr error
	}{
		{
			name: "success",
			id:   "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "email", "display_name", "password_hash", "created_at", "updated_at",
				}).AddRow("user-1", "user@test.com", "Test User", "$2a$12$hash", fixedTime, fixedTime)

				mock.ExpectQuery(`SELECT\s+.*FROM\s+users\s+WHERE\s+id\s*=\s*\$1`).
					WithArgs("user-1").
					WillReturnRows(rows)
			},
			want: User{
				ID:           "user-1",
				Email:        "user@test.com",
				DisplayName:  "Test User",
				PasswordHash: "$2a$12$hash",
				CreatedAt:    fixedTime,
				UpdatedAt:    fixedTime,
			},
		},
		{
			name: "not found",
			id:   "missing",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT\s+.*FROM\s+users\s+WHERE\s+id\s*=\s*\$1`).
					WithArgs("missing").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: apperrors.ErrNotFound,
		},
		{
			name: "database error",
			id:   "user-1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT\s+.*FROM\s+users\s+WHERE\s+id\s*=\s*\$1`).
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

			got, err := repo.GetByID(context.Background(), tt.id)

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

func TestPostgresUserRepository_GetByEmail(t *testing.T) {
	fixedTime := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		name    string
		email   string
		mock    func(mock sqlmock.Sqlmock)
		want    User
		wantErr error
	}{
		{
			name:  "success",
			email: "user@test.com",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "email", "display_name", "password_hash", "created_at", "updated_at",
				}).AddRow("user-1", "user@test.com", "Test User", "$2a$12$hash", fixedTime, fixedTime)

				mock.ExpectQuery(`SELECT\s+.*FROM\s+users\s+WHERE\s+email\s*=\s*\$1`).
					WithArgs("user@test.com").
					WillReturnRows(rows)
			},
			want: User{
				ID:           "user-1",
				Email:        "user@test.com",
				DisplayName:  "Test User",
				PasswordHash: "$2a$12$hash",
				CreatedAt:    fixedTime,
				UpdatedAt:    fixedTime,
			},
		},
		{
			name:  "not found",
			email: "nobody@test.com",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT\s+.*FROM\s+users\s+WHERE\s+email\s*=\s*\$1`).
					WithArgs("nobody@test.com").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: apperrors.ErrNotFound,
		},
		{
			name:  "database error",
			email: "user@test.com",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT\s+.*FROM\s+users\s+WHERE\s+email\s*=\s*\$1`).
					WithArgs("user@test.com").
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

			got, err := repo.GetByEmail(context.Background(), tt.email)

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

func TestPostgresUserRepository_Update(t *testing.T) {
	fixedTime := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		name    string
		input   User
		mock    func(mock sqlmock.Sqlmock)
		want    User
		wantErr error
	}{
		{
			name: "success",
			input: User{
				ID:           "user-1",
				Email:        "new@test.com",
				DisplayName:  "New Name",
				PasswordHash: "$2a$12$newhash",
			},
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "email", "display_name", "password_hash", "created_at", "updated_at",
				}).AddRow("user-1", "new@test.com", "New Name", "$2a$12$newhash", fixedTime, fixedTime)

				mock.ExpectQuery(`UPDATE\s+users\s+SET`).
					WithArgs("user-1", "new@test.com", "New Name", "$2a$12$newhash").
					WillReturnRows(rows)
			},
			want: User{
				ID:           "user-1",
				Email:        "new@test.com",
				DisplayName:  "New Name",
				PasswordHash: "$2a$12$newhash",
				CreatedAt:    fixedTime,
				UpdatedAt:    fixedTime,
			},
		},
		{
			name: "email taken by another user returns conflict",
			input: User{
				ID:           "user-1",
				Email:        "taken@test.com",
				DisplayName:  "New Name",
				PasswordHash: "$2a$12$newhash",
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE\s+users\s+SET`).
					WithArgs("user-1", "taken@test.com", "New Name", "$2a$12$newhash").
					WillReturnError(uniqueViolation())
			},
			wantErr: apperrors.ErrConflict,
		},
		{
			name: "not found",
			input: User{
				ID:           "missing",
				Email:        "new@test.com",
				DisplayName:  "New Name",
				PasswordHash: "$2a$12$newhash",
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE\s+users\s+SET`).
					WithArgs("missing", "new@test.com", "New Name", "$2a$12$newhash").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: apperrors.ErrNotFound,
		},
		{
			name: "database error",
			input: User{
				ID:           "user-1",
				Email:        "new@test.com",
				DisplayName:  "New Name",
				PasswordHash: "$2a$12$newhash",
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE\s+users\s+SET`).
					WithArgs("user-1", "new@test.com", "New Name", "$2a$12$newhash").
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

			got, err := repo.Update(context.Background(), tt.input)

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

func TestNewUserRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewUserRepository(sqlx.NewDb(db, "postgres"))
	if repo == nil {
		t.Fatal("NewUserRepository returned nil")
	}
}
