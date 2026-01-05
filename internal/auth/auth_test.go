package auth

import (
	"testing"
	"time"
)

func TestJWTManager_GenerateAndValidateToken(t *testing.T) {
	manager := NewJWTManager("test-secret-key", time.Hour)

	t.Run("generate and validate valid token", func(t *testing.T) {
		userID := int(123)
		token, err := manager.GenerateToken(userID)
		if err != nil {
			t.Fatalf("GenerateToken() error = %v", err)
		}

		if token == "" {
			t.Fatal("GenerateToken() returned empty token")
		}

		gotUserID, err := manager.ValidateToken(token)
		if err != nil {
			t.Fatalf("ValidateToken() error = %v", err)
		}

		if gotUserID != userID {
			t.Errorf("ValidateToken() = %v, want %v", gotUserID, userID)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		_, err := manager.ValidateToken("invalid-token")
		if err == nil {
			t.Error("ValidateToken() expected error for invalid token")
		}
	})

	t.Run("expired token", func(t *testing.T) {
		shortManager := NewJWTManager("test-secret-key", time.Millisecond)
		token, err := shortManager.GenerateToken(123)
		if err != nil {
			t.Fatalf("GenerateToken() error = %v", err)
		}

		// Ждём истечения токена
		time.Sleep(10 * time.Millisecond)

		_, err = shortManager.ValidateToken(token)
		if err == nil {
			t.Error("ValidateToken() expected error for expired token")
		}
	})

	t.Run("wrong secret key", func(t *testing.T) {
		token, _ := manager.GenerateToken(123)
		otherManager := NewJWTManager("different-secret", time.Hour)

		_, err := otherManager.ValidateToken(token)
		if err == nil {
			t.Error("ValidateToken() expected error for wrong secret key")
		}
	})
}

func TestHashPassword(t *testing.T) {
	password := "test-password-123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if hash == "" {
		t.Fatal("HashPassword() returned empty hash")
	}

	if hash == password {
		t.Error("HashPassword() hash should not equal password")
	}

	// Проверяем, что разные вызовы дают разные хеши (из-за соли)
	hash2, _ := HashPassword(password)
	if hash == hash2 {
		t.Error("HashPassword() should generate different hashes for same password")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "test-password-123"
	hash, _ := HashPassword(password)

	t.Run("correct password", func(t *testing.T) {
		if !CheckPassword(password, hash) {
			t.Error("CheckPassword() = false, want true for correct password")
		}
	})

	t.Run("incorrect password", func(t *testing.T) {
		if CheckPassword("wrong-password", hash) {
			t.Error("CheckPassword() = true, want false for incorrect password")
		}
	})

	t.Run("empty password", func(t *testing.T) {
		if CheckPassword("", hash) {
			t.Error("CheckPassword() = true, want false for empty password")
		}
	})

	t.Run("invalid hash", func(t *testing.T) {
		if CheckPassword(password, "invalid-hash") {
			t.Error("CheckPassword() = true, want false for invalid hash")
		}
	})
}
