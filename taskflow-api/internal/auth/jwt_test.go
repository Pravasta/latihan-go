package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTService_GenerateAndParse(t *testing.T) {
	t.Run("round trip returns the original user id", func(t *testing.T) {
		svc := NewJWTService("test-secret")

		token, err := svc.Generate("user-1")
		if err != nil {
			t.Fatalf("unexpected error generating token: %v", err)
		}

		userID, err := svc.Parse(token)
		if err != nil {
			t.Fatalf("unexpected error parsing token: %v", err)
		}
		if userID != "user-1" {
			t.Fatalf("got user id %q, want %q", userID, "user-1")
		}
	})

	t.Run("malformed token is rejected", func(t *testing.T) {
		svc := NewJWTService("test-secret")

		if _, err := svc.Parse("not-a-jwt"); err == nil {
			t.Fatal("expected an error, got nil")
		}
	})

	t.Run("token signed with a different secret is rejected", func(t *testing.T) {
		issuer := NewJWTService("secret-a")
		verifier := NewJWTService("secret-b")

		token, err := issuer.Generate("user-1")
		if err != nil {
			t.Fatalf("unexpected error generating token: %v", err)
		}

		if _, err := verifier.Parse(token); err == nil {
			t.Fatal("expected a signature error, got nil")
		}
	})

	t.Run("expired token is rejected", func(t *testing.T) {
		svc := NewJWTService("test-secret")

		claims := jwt.MapClaims{
			"user_id": "user-1",
			"exp":     time.Now().Add(-time.Hour).Unix(),
		}
		expired := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := expired.SignedString(svc.secret)
		if err != nil {
			t.Fatalf("failed to prepare test fixture: %v", err)
		}

		if _, err := svc.Parse(signed); err == nil {
			t.Fatal("expected an expiration error, got nil")
		}
	})

	t.Run("token with a non-string user_id claim is rejected", func(t *testing.T) {
		svc := NewJWTService("test-secret")

		claims := jwt.MapClaims{
			"user_id": 12345,
			"exp":     time.Now().Add(time.Hour).Unix(),
		}
		malformed := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := malformed.SignedString(svc.secret)
		if err != nil {
			t.Fatalf("failed to prepare test fixture: %v", err)
		}

		if _, err := svc.Parse(signed); err == nil {
			t.Fatal("expected an error, got nil")
		}
	})

	t.Run("token signed with a non-HMAC method is rejected", func(t *testing.T) {
		svc := NewJWTService("test-secret")

		claims := jwt.MapClaims{
			"user_id": "user-1",
			"exp":     time.Now().Add(time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
		signed, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
		if err != nil {
			t.Fatalf("failed to prepare test fixture: %v", err)
		}

		if _, err := svc.Parse(signed); err == nil {
			t.Fatal("expected an error for a non-HMAC signing method, got nil")
		}
	})
}
