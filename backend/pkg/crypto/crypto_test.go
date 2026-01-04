package crypto

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "testPassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if hash == "" {
		t.Error("HashPassword() returned empty hash")
	}

	// Hash should be bcrypt format (starts with $2a$ or $2b$)
	if !strings.HasPrefix(hash, "$2") {
		t.Errorf("HashPassword() hash format incorrect: %s", hash)
	}
}

func TestCheckPassword(t *testing.T) {
	password := "testPassword123!"
	wrongPassword := "wrongPassword!"

	hash, _ := HashPassword(password)

	if !CheckPassword(password, hash) {
		t.Error("CheckPassword() should return true for correct password")
	}

	if CheckPassword(wrongPassword, hash) {
		t.Error("CheckPassword() should return false for wrong password")
	}
}

func TestSHA256Hash(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple", "hello"},
		{"empty", ""},
		{"long", "this is a much longer string with special chars: !@#$%^&*()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := SHA256Hash(tt.input)
			hash2 := SHA256Hash(tt.input)

			if hash1 != hash2 {
				t.Error("SHA256Hash() should be deterministic")
			}

			if len(hash1) != 64 { // SHA256 produces 32 bytes = 64 hex chars
				t.Errorf("SHA256Hash() length = %d, want 64", len(hash1))
			}
		})
	}

	// Different inputs should produce different hashes
	h1 := SHA256Hash("input1")
	h2 := SHA256Hash("input2")
	if h1 == h2 {
		t.Error("SHA256Hash() different inputs should produce different hashes")
	}
}

func TestEncryptor(t *testing.T) {
	key := "01234567890123456789012345678901" // 32 bytes for AES-256

	enc, err := NewEncryptor(key)
	if err != nil {
		t.Fatalf("NewEncryptor() error = %v", err)
	}

	plaintext := "sensitive patient data"

	encrypted, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	if encrypted == plaintext {
		t.Error("Encrypt() should produce different output than input")
	}

	decrypted, err := enc.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Decrypt() = %v, want %v", decrypted, plaintext)
	}
}

func TestEncryptor_InvalidKey(t *testing.T) {
	_, err := NewEncryptor("short")
	if err == nil {
		t.Error("NewEncryptor() should fail with short key")
	}
}
