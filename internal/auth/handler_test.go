package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/mmk31585/workout-tracker/internal/app_errors"
	"github.com/mmk31585/workout-tracker/internal/user"
)

type authResponseEnvelope struct {
	Data AuthResponse `json:"data"`
}

type errorEnvelope struct {
	Error string `json:"error"`
}

func TestAuthHandler_Signup_Success(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSvc := NewJWTService([]byte("test-secret"), "test-issuer", 15*time.Minute)

	authSvc := NewAuthService(mockRepo, jwtSvc, 15*time.Minute)
	handler := NewAuthHandler(authSvc)

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

	reqBody := SignupRequest{
		Email:       "test@example.com",
		Password:    "testpassword123",
		DisplayName: "Test User",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.Signup(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp authResponseEnvelope
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Data.AccessToken)
	assert.Equal(t, "Bearer", resp.Data.TokenType)
	assert.Equal(t, int(15*time.Minute.Seconds()), resp.Data.ExpiresIn)
}

func TestAuthHandler_Signup_DuplicateEmail(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSvc := NewJWTService([]byte("test-secret"), "test-issuer", 15*time.Minute)

	authSvc := NewAuthService(mockRepo, jwtSvc, 15*time.Minute)
	handler := NewAuthHandler(authSvc)

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

	reqBody := SignupRequest{
		Email:       "test@example.com",
		Password:    "testpassword123",
		DisplayName: "Test User",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.Signup(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)

	var resp errorEnvelope
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "resource conflict", resp.Error)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSvc := NewJWTService([]byte("test-secret"), "test-issuer", 15*time.Minute)

	authSvc := NewAuthService(mockRepo, jwtSvc, 15*time.Minute)
	handler := NewAuthHandler(authSvc)

	// Hash the password we'll use for testing
	testPassword := "testpassword123"
	hash, err := HashPassword(testPassword)
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

	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: testPassword,
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.Login(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp authResponseEnvelope
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Data.AccessToken)
	assert.Equal(t, "Bearer", resp.Data.TokenType)
	assert.Equal(t, int(15*time.Minute.Seconds()), resp.Data.ExpiresIn)
}

func TestAuthHandler_Login_WrongPassword(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSvc := NewJWTService([]byte("test-secret"), "test-issuer", 15*time.Minute)

	authSvc := NewAuthService(mockRepo, jwtSvc, 15*time.Minute)
	handler := NewAuthHandler(authSvc)

	// Hash the correct password
	testPassword := "testpassword123"
	hash, err := HashPassword(testPassword)
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

	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.Login(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var resp errorEnvelope
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp.Error, "invalid credentials")
}

func TestAuthHandler_Login_UserNotFound(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSvc := NewJWTService([]byte("test-secret"), "test-issuer", 15*time.Minute)

	authSvc := NewAuthService(mockRepo, jwtSvc, 15*time.Minute)
	handler := NewAuthHandler(authSvc)

	// Mock the GetByEmail to return ErrNotFound (user doesn't exist)
	mockRepo.GetByEmailFn = func(ctx context.Context, email string) (user.User, error) {
		return user.User{}, apperrors.ErrNotFound
	}

	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "testpassword123",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.Login(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var resp errorEnvelope
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp.Error, "invalid credentials")
}

func TestAuthHandler_Logout(t *testing.T) {
	mockRepo := &MockUserRepository{}
	jwtSvc := NewJWTService([]byte("test-secret"), "test-issuer", 15*time.Minute)

	authSvc := NewAuthService(mockRepo, jwtSvc, 15*time.Minute)
	handler := NewAuthHandler(authSvc)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	rr := httptest.NewRecorder()
	handler.Logout(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}
