package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	apperrors "github.com/mmk31585/workout-tracker/internal/app_errors"
	"github.com/mmk31585/workout-tracker/internal/user"
)

// MockUserRepository is a mock implementation of UserRepository for testing
type MockUserRepository struct {
	CreateFn     func(ctx context.Context, user user.User) (user.User, error)
	GetByIDFn    func(ctx context.Context, id string) (user.User, error)
	GetByEmailFn func(ctx context.Context, email string) (user.User, error)
	UpdateFn     func(ctx context.Context, user user.User) (user.User, error)
}

// Create implements the UserRepository interface
func (m MockUserRepository) Create(ctx context.Context, user user.User) (user.User, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, user)
	}
	return user, errors.New("not implemented")
}

// GetByID implements the UserRepository interface
func (m MockUserRepository) GetByID(ctx context.Context, id string) (user.User, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return user.User{}, errors.New("not implemented")
}

// GetByEmail implements the UserRepository interface
func (m MockUserRepository) GetByEmail(ctx context.Context, email string) (user.User, error) {
	if m.GetByEmailFn != nil {
		return m.GetByEmailFn(ctx, email)
	}
	return user.User{}, errors.New("not implemented")
}

// Update implements the UserRepository interface
func (m MockUserRepository) Update(ctx context.Context, user user.User) (user.User, error) {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, user)
	}
	return user, errors.New("not implemented")
}

func TestAuthService_Signup_Success(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSvc := NewJWTService([]byte("test-secret"), "test-issuer", 15*time.Minute)

	authSvc := NewAuthService(mockRepo, jwtSvc, 15*time.Minute)

	// Mock the GetByEmail to return ErrNotFound (user doesn't exist)
	mockRepo.GetByEmailFn = func(ctx context.Context, email string) (user.User, error) {
		return user.User{}, apperrors.ErrNotFound
	}

	// Mock the Create to return a user
	testUser := user.User{
		ID:           "test-user-id",
		Email:        "test@example.com",
		PasswordHash: "hashed-password",
		DisplayName:  "Test User",
	}
	mockRepo.CreateFn = func(ctx context.Context, user user.User) (user.User, error) {
		return testUser, nil
	}

	email := "test@example.com"
	password := "testpassword123"
	displayName := "Test User"

	token, err := authSvc.Signup(context.Background(), email, password, displayName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestAuthService_Signup_DuplicateEmail(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSvc := NewJWTService([]byte("test-secret"), "test-issuer", 15*time.Minute)

	authSvc := NewAuthService(mockRepo, jwtSvc, 15*time.Minute)

	// Mock the GetByEmail to return an existing user
	existingUser := user.User{
		ID:           "existing-user-id",
		Email:        "test@example.com",
		PasswordHash: "existing-hash",
		DisplayName:  "Existing User",
	}
	mockRepo.GetByEmailFn = func(ctx context.Context, email string) (user.User, error) {
		return existingUser, nil
	}

	email := "test@example.com"
	password := "testpassword123"
	displayName := "Test User"

	token, err := authSvc.Signup(context.Background(), email, password, displayName)
	if err == nil {
		t.Fatal("expected ErrConflict for duplicate email, got nil")
	}
	if !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	if token != "" {
		t.Fatal("expected empty token for duplicate email")
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSvc := NewJWTService([]byte("test-secret"), "test-issuer", 15*time.Minute)

	authSvc := NewAuthService(mockRepo, jwtSvc, 15*time.Minute)

	// Hash the password we'll use for testing
	testPassword := "testpassword123"
	hash, err :=HashPassword(testPassword)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	// Mock the GetByEmail to return a user with the hashed password
	testUser := user.User{
		ID:           "test-user-id",
		Email:        "test@example.com",
		PasswordHash: hash,
		DisplayName:  "Test User",
	}
	mockRepo.GetByEmailFn = func(ctx context.Context, email string) (user.User, error) {
		return testUser, nil
	}

	email := "test@example.com"
	password := testPassword

	token, err := authSvc.Login(context.Background(), email, password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}

	// Verify the token is valid
	userID, err := authSvc.ValidateToken(token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}
	if userID != testUser.ID {
		t.Fatalf("expected user ID %s, got %s", testUser.ID, userID)
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSvc := NewJWTService([]byte("test-secret"), "test-issuer", 15*time.Minute)

	authSvc := NewAuthService(mockRepo, jwtSvc, 15*time.Minute)

	// Hash the correct password
	correctPassword := "correctpassword123"
	hash, err := HashPassword(correctPassword)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	// Mock the GetByEmail to return a user with the hashed password
	testUser := user.User{
		ID:           "test-user-id",
		Email:        "test@example.com",
		PasswordHash: hash,
		DisplayName:  "Test User",
	}
	mockRepo.GetByEmailFn = func(ctx context.Context, email string) (user.User, error) {
		return testUser, nil
	}

	email := "test@example.com"
	password := "wrongpassword"

	token, err := authSvc.Login(context.Background(), email, password)
	if err == nil {
		t.Fatal("expected ErrInvalidCredentials for wrong password, got nil")
	}
	if !errors.Is(err, apperrors.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
	if token != "" {
		t.Fatal("expected empty token for wrong password")
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSvc := NewJWTService([]byte("test-secret"), "test-issuer", 15*time.Minute)

	authSvc := NewAuthService(mockRepo, jwtSvc, 15*time.Minute)

	// Mock the GetByEmail to return ErrNotFound (user doesn't exist)
	mockRepo.GetByEmailFn = func(ctx context.Context, email string) (user.User, error) {
		return user.User{}, apperrors.ErrNotFound
	}

	email := "test@example.com"
	password := "testpassword123"

	token, err := authSvc.Login(context.Background(), email, password)
	if err == nil {
		t.Fatal("expected ErrInvalidCredentials for non-existent user, got nil")
	}
	if !errors.Is(err, apperrors.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
	if token != "" {
		t.Fatal("expected empty token for non-existent user")
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSvc := NewJWTService([]byte("test-secret"), "test-issuer", 15*time.Minute)

	authSvc := NewAuthService(mockRepo, jwtSvc, 15*time.Minute)

	// Generate a token
	token, err := jwtSvc.GenerateToken("test-user-id")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Validate it
	userID, err := authSvc.ValidateToken(token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}
	if userID != "test-user-id" {
		t.Fatalf("expected user ID 'test-user-id', got %s", userID)
	}
}

func TestAuthService_ValidateToken_Expired(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSvc := NewJWTService([]byte("test-secret"), "test-issuer", -time.Hour) // Expired

	authSvc := NewAuthService(mockRepo, jwtSvc, -time.Hour)

	// Generate an expired token
	token, err := jwtSvc.GenerateToken("test-user-id")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Validate it - should fail
	userID, err := authSvc.ValidateToken(token)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
	if userID != "" {
		t.Fatal("expected empty userID for expired token")
	}
}