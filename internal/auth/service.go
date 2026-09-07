package auth

import (
	"context"
	"errors"
	"time"

	apperrors "github.com/mmk31585/workout-tracker/internal/app_errors"
	"github.com/mmk31585/workout-tracker/internal/user"
)

type UserRepository interface {
	Create(context.Context, user.User) (user.User, error)
	GetByEmail(context.Context, string) (user.User, error)
}

type AuthService struct {
	userRepo UserRepository
	jwt      JWTInterface
	tokenExp time.Duration
}

func NewAuthService(userRepo UserRepository, jwt JWTInterface, tokenExp time.Duration) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		jwt:      jwt,
		tokenExp: tokenExp,
	}
}

func (a *AuthService) Signup(ctx context.Context, email, password, displayName string) (string, error) {
	_, err := a.userRepo.GetByEmail(ctx, email)
	if err == nil {
		return "", apperrors.ErrConflict
	}
	if !errors.Is(err, apperrors.ErrNotFound) {
		return "", err
	}

	hashPass, err := HashPassword(password)
	if err != nil {
		return "", err
	}

	newUser := user.User{
		Email:        email,
		PasswordHash: hashPass,
		DisplayName:  displayName,
	}
	createdUser, err := a.userRepo.Create(ctx, newUser)
	if err != nil {
		return "", err
	}
	
	return a.jwt.GenerateToken(createdUser.ID)
}

func (a *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	usr, err := a.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", apperrors.ErrInvalidCredentials
	}
	if err := CheckPassword(usr.PasswordHash, password); err != nil {
		return "", apperrors.ErrInvalidCredentials
	}
	return a.jwt.GenerateToken(usr.ID)
}

func (a *AuthService) ValidateToken(token string) (string, error) {
	return a.jwt.ValidateToken(token)
}

func (a *AuthService) TokenExpiration() time.Duration {
	return a.tokenExp
}
