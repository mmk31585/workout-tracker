package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	apperrors "github.com/mmk31585/workout-tracker/internal/app_errors"
	"github.com/mmk31585/workout-tracker/internal/db"
)

type UserRepository interface {
	Create(context.Context, User) (User, error)
	GetByID(context.Context, string) (User, error)
	GetByEmail(context.Context, string) (User, error)
	Update(context.Context, User) (User, error)
}
type PostgresUserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *PostgresUserRepository {
	return &PostgresUserRepository{
		db: db,
	}
}
func (r *PostgresUserRepository) Create(ctx context.Context, user User) (User, error) {
	var created User
	query := `
		INSERT INTO users
		(
			id,
			email,
			display_name,
			password_hash,
			created_at,
			updated_at
		)
		VALUES (gen_random_uuid(),$1,$2,$3,NOW(),NOW())
		RETURNING
			id,
			email,
			display_name,
			password_hash,
			created_at,
			updated_at
	`
	ctx, cancel := db.QueryTimeoutContext(ctx)
	defer cancel()
	err := r.db.QueryRowContext(
		ctx,
		query,
		user.Email,
		user.DisplayName,
		user.PasswordHash,
	).Scan(
		&created.ID,
		&created.Email,
		&created.DisplayName,
		&created.PasswordHash,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		if apperrors.IsUniqueRow(err) {
			return User{}, apperrors.ErrConflict
		}
		return User{}, fmt.Errorf("repository: failed to create user: %w", err)
	}
	return created, nil
}
func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (User, error) {
	var tempUser User
	query := `
		SELECT 
			id,
			email,
			display_name,
			password_hash,
			created_at,
			updated_at
		FROM users
		WHERE id = $1
	`
	ctx, cancel := db.QueryTimeoutContext(ctx)
	defer cancel()
	err := r.db.GetContext(ctx, &tempUser, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, apperrors.ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("repository: failed to get user by id: %w", err)
	}
	return tempUser, nil
}
func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (User, error) {
	var u User
	query := `
		SELECT
			id,
			email,
			display_name,
			password_hash,
			created_at,
			updated_at
		FROM users
		WHERE email = $1
	`
	ctx, cancel := db.QueryTimeoutContext(ctx)
	defer cancel()
	err := r.db.GetContext(ctx, &u, query, email)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, apperrors.ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("repository: failed to get user by email: %w", err)
	}
	return u, nil
}
func (r *PostgresUserRepository) Update(ctx context.Context, user User) (User, error) {
	var tempUser User
	query := `
		UPDATE users
		SET 
			email = $2,
			display_name = $3,
			password_hash = $4,
			updated_at = NOW()
		WHERE id = $1
		RETURNING
			id,
			email,
			display_name,
			password_hash,
			created_at,
			updated_at
	`
	ctx, cancel := db.QueryTimeoutContext(ctx)
	defer cancel()
	err := r.db.QueryRowContext(
		ctx,
		query,
		user.ID,
		user.Email,
		user.DisplayName,
		user.PasswordHash,
	).Scan(
		&tempUser.ID,
		&tempUser.Email,
		&tempUser.DisplayName,
		&tempUser.PasswordHash,
		&tempUser.CreatedAt,
		&tempUser.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, apperrors.ErrNotFound
	}
	if err != nil {
		if apperrors.IsUniqueRow(err) {
			return User{}, apperrors.ErrConflict
		}
		return User{}, fmt.Errorf("repository: failed to update user: %w", err)
	}
	return tempUser, nil
}
