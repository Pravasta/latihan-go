package auth

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	t.Run("hash differs from the plain text password", func(t *testing.T) {
		hash, err := HashPassword("Passw0rd!")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if hash == "Passw0rd!" {
			t.Fatal("password was not hashed")
		}
	})

	t.Run("rejects passwords over bcrypt's 72-byte limit", func(t *testing.T) {
		tooLong := strings.Repeat("a", 73)

		if _, err := HashPassword(tooLong); err == nil {
			t.Fatal("expected an error for an over-length password, got nil")
		}
	})
}

func TestCheckPasswordHash(t *testing.T) {
	hash, err := HashPassword("Passw0rd!")
	if err != nil {
		t.Fatalf("failed to prepare test fixture: %v", err)
	}

	t.Run("correct password matches", func(t *testing.T) {
		if !CheckPasswordHash("Passw0rd!", hash) {
			t.Fatal("expected the correct password to match its hash")
		}
	})

	t.Run("wrong password does not match", func(t *testing.T) {
		if CheckPasswordHash("WrongPassword1!", hash) {
			t.Fatal("expected a wrong password not to match")
		}
	})
}
