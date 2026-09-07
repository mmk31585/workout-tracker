package auth

import (
	"strings"
	"testing"
	"time"
)

func TestJWTService_GenerateToken(t *testing.T) {
	svc := NewJWTService([]byte("test-secret"), "test-issuer", 15*time.Minute)

	token, err := svc.GenerateToken("user-123")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}

	userID, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if userID != "user-123" {
		t.Fatalf("expected subject 'user-123', got %q", userID)
	}
}

func TestJWTService_ValidateToken_Expired(t *testing.T) {
	svc := NewJWTService([]byte("secret"), "test-issuer", -time.Hour)

	token, err := svc.GenerateToken("user-123")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	userID, err := svc.ValidateToken(token)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
	if userID != "" {
		t.Fatal("expected empty userID for expired token")
	}
}

func TestJWTService_ValidateToken_InvalidToken(t *testing.T) {
	svc := NewJWTService([]byte("secret"), "test-issuer", 15*time.Minute)

	userID, err := svc.ValidateToken("invalid-token")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
	if userID != "" {
		t.Fatal("expected empty userID for invalid token")
	}
}

func TestJWTService_ValidateToken_Tampered(t *testing.T) {
	svc := NewJWTService([]byte("secret"), "test-issuer", 15*time.Minute)

	// Create a valid token
	token, err := svc.GenerateToken("user-123")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Tamper with the token (modify the signature part - change multiple characters)
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatal("unexpected token format")
	}
	// Modify the signature by appending extra data
	tampered := parts[0] + "." + parts[1] + "." + parts[2] + "tampered"

	userID, err := svc.ValidateToken(tampered)
	if err == nil {
		t.Fatal("expected error for tampered token, got nil")
	}
	if userID != "" {
		t.Fatal("expected empty userID for tampered token")
	}
}

func TestJWTService_ValidateToken_WrongIssuer(t *testing.T) {
	svc1 := NewJWTService([]byte("secret"), "issuer-1", 15*time.Minute)
	svc2 := NewJWTService([]byte("secret"), "issuer-2", 15*time.Minute)

	token, err := svc1.GenerateToken("user-123")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Validate with different issuer
	userID, err := svc2.ValidateToken(token)
	if err == nil {
		t.Fatal("expected error for wrong issuer, got nil")
	}
	if userID != "" {
		t.Fatal("expected empty userID for wrong issuer")
	}
}
