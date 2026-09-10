package auth_test

import (
	"testing"
	
	"laboratory-management/internal/auth"
)

// TestPasswordHashing tests password hashing functionality
func TestPasswordHashing(t *testing.T) {
	password := "testPassword123"
	
	// Test password hashing
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	
	if hash == "" {
		t.Error("HashPassword returned empty string")
	}
	
	if hash == password {
		t.Error("HashPassword returned the same password (not hashed)")
	}
	
	// Test password checking
	isValid := auth.CheckPassword(password, hash)
	if !isValid {
		t.Error("CheckPassword failed to validate correct password")
	}
	
	// Test wrong password
	isValid = auth.CheckPassword("wrongPassword", hash)
	if isValid {
		t.Error("CheckPassword validated wrong password")
	}
}

// TestSessionTokenGeneration tests session token generation
func TestSessionTokenGeneration(t *testing.T) {
	token, err := auth.GenerateSessionToken()
	if err != nil {
		t.Fatalf("GenerateSessionToken failed: %v", err)
	}
	
	if token == "" {
		t.Error("GenerateSessionToken returned empty string")
	}
	
	// Test that tokens are unique
	token2, err := auth.GenerateSessionToken()
	if err != nil {
		t.Fatalf("GenerateSessionToken failed on second call: %v", err)
	}
	
	if token == token2 {
		t.Error("GenerateSessionToken generated duplicate tokens")
	}
}