package auth

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestPasswordService_HashPassword(t *testing.T) {

	tests := []struct {
		name        string
		password    string
		expectError bool
	}{
		{
			name:        "valid password",
			password:    "password123",
			expectError: false,
		},
		{
			name:        "empty password",
			password:    "",
			expectError: false,
		},
		{
			name:        "long password exceeds bcrypt limit (72 bytes)",
			password:    "this_is_a_very_long_password_that_exceeds_seventy_two_bytes_limit_of_bcrypt_algorithm",
			expectError: true,
		},
		{
			name:        "special characters",
			password:    "p@ssw0rd!#$%^&*()",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if hash == "" {
				t.Fatal("hash should not be empty")
			}

			// Verify it's a valid bcrypt hash (starts with $2a$)
			if len(hash) < 4 || hash[:4] != "$2a$" {
				t.Fatalf("hash should be bcrypt format, got: %s", hash)
			}
		})
	}
}

func TestPasswordService_CheckPassword(t *testing.T) {

	password := "testpassword123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	tests := []struct {
		name        string
		hash        string
		password    string
		expectError bool
	}{
		{
			name:        "correct password",
			hash:        hash,
			password:    password,
			expectError: false,
		},
		{
			name:        "wrong password",
			hash:        hash,
			password:    "wrongpassword",
			expectError: true,
		},
		{
			name:        "empty password",
			hash:        hash,
			password:    "",
			expectError: true,
		},
		{
			name:        "malformed hash",
			hash:        "not-a-valid-hash",
			password:    password,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPassword(tt.hash, tt.password)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				if err != bcrypt.ErrMismatchedHashAndPassword && tt.name != "malformed hash" {
					// bcrypt returns ErrMismatchedHashAndPassword for wrong password
					t.Logf("error: %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestPasswordService_HashDeterministic(t *testing.T) {

	password := "testpassword123"
	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("first hash failed: %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("second hash failed: %v", err)
	}

	// bcrypt uses salt, so hashes should be different each time
	if hash1 == hash2 {
		t.Fatal("bcrypt should generate different hashes for same password (different salts)")
	}

	// But both should verify correctly
	if err := CheckPassword(hash1, password); err != nil {
		t.Fatalf("first hash verification failed: %v", err)
	}
	if err := CheckPassword(hash2, password); err != nil {
		t.Fatalf("second hash verification failed: %v", err)
	}
}