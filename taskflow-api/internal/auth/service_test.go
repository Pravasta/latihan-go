package auth

import (
	"errors"
	"strings"
	"testing"
)

const validPassword = "Passw0rd!"

func testJWT() *JWTService {
	return NewJWTService("test-secret")
}

func TestService_CreateUser(t *testing.T) {
	t.Run("validation", func(t *testing.T) {
		cases := []struct {
			name     string
			userName string
			email    string
			password string
			wantErr  error
		}{
			{"missing name", "", "user@example.com", validPassword, ErrInvalidName},
			{"invalid email", "Alice", "not-an-email", validPassword, ErrInvalidEmail},
			{"password too short", "Alice", "user@example.com", "Aa1!", ErrInvalidPassword},
			{"password missing complexity", "Alice", "user@example.com", "alllowercase1", ErrInvalidPassword},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				svc := NewService(&fakeStorage{}, testJWT())

				_, err := svc.CreateUser(tc.userName, tc.email, tc.password)
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("got error %v, want %v", err, tc.wantErr)
				}
			})
		}
	})

	t.Run("success persists a new user with a hashed password", func(t *testing.T) {
		storage := &fakeStorage{}
		svc := NewService(storage, testJWT())

		got, err := svc.CreateUser("Alice", "alice@example.com", validPassword)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID == "" {
			t.Fatal("expected a generated ID, got empty string")
		}
		if got.PasswordHash == validPassword {
			t.Fatal("password was stored in plain text")
		}
		if !CheckPasswordHash(validPassword, got.PasswordHash) {
			t.Fatal("stored hash does not match the original password")
		}
		if len(storage.users) != 1 {
			t.Fatalf("expected 1 user persisted, got %d", len(storage.users))
		}
	})

	t.Run("email already exists", func(t *testing.T) {
		existing := User{ID: "u1", Email: "alice@example.com"}
		svc := NewService(&fakeStorage{users: []User{existing}}, testJWT())

		_, err := svc.CreateUser("Alice", "alice@example.com", validPassword)
		if !errors.Is(err, ErrEmailAlreadyExists) {
			t.Fatalf("got error %v, want %v", err, ErrEmailAlreadyExists)
		}
	})

	t.Run("password too long for bcrypt is returned to the caller", func(t *testing.T) {
		// bcrypt rejects inputs over 72 bytes; this still satisfies
		// IsValidPassword's complexity regex, so it reaches HashPassword.
		tooLong := strings.Repeat("Aa1!", 20)
		svc := NewService(&fakeStorage{}, testJWT())

		if _, err := svc.CreateUser("Alice", "alice@example.com", tooLong); err == nil {
			t.Fatal("expected an error for an over-length password, got nil")
		}
	})

	t.Run("load failure is returned to the caller", func(t *testing.T) {
		svc := NewService(&fakeStorage{loadErr: errors.New("disk error")}, testJWT())

		if _, err := svc.CreateUser("Alice", "alice@example.com", validPassword); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("save failure is returned to the caller", func(t *testing.T) {
		svc := NewService(&fakeStorage{saveErr: errors.New("disk full")}, testJWT())

		if _, err := svc.CreateUser("Alice", "alice@example.com", validPassword); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestService_Authenticate(t *testing.T) {
	hash, err := HashPassword(validPassword)
	if err != nil {
		t.Fatalf("failed to prepare test fixture: %v", err)
	}
	existing := User{ID: "u1", Email: "alice@example.com", PasswordHash: hash}

	t.Run("validation", func(t *testing.T) {
		cases := []struct {
			name     string
			email    string
			password string
			wantErr  error
		}{
			{"invalid email", "not-an-email", validPassword, ErrInvalidEmail},
			{"invalid password format", "alice@example.com", "short", ErrInvalidPassword},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				svc := NewService(&fakeStorage{users: []User{existing}}, testJWT())

				_, err := svc.Authenticate(tc.email, tc.password)
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("got error %v, want %v", err, tc.wantErr)
				}
			})
		}
	})

	t.Run("user not found", func(t *testing.T) {
		svc := NewService(&fakeStorage{users: []User{existing}}, testJWT())

		_, err := svc.Authenticate("nobody@example.com", validPassword)
		if !errors.Is(err, ErrUserNotFound) {
			t.Fatalf("got error %v, want %v", err, ErrUserNotFound)
		}
	})

	t.Run("wrong password returns invalid credentials, not invalid password format", func(t *testing.T) {
		svc := NewService(&fakeStorage{users: []User{existing}}, testJWT())

		_, err := svc.Authenticate("alice@example.com", "WrongPass1!")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("got error %v, want %v", err, ErrInvalidCredentials)
		}
	})

	t.Run("success returns a token containing the user id", func(t *testing.T) {
		jwtSvc := testJWT()
		svc := NewService(&fakeStorage{users: []User{existing}}, jwtSvc)

		token, err := svc.Authenticate("alice@example.com", validPassword)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		userID, err := jwtSvc.Parse(token)
		if err != nil {
			t.Fatalf("returned token does not parse: %v", err)
		}
		if userID != existing.ID {
			t.Fatalf("got user id %q, want %q", userID, existing.ID)
		}
	})

	t.Run("load failure is returned to the caller", func(t *testing.T) {
		svc := NewService(&fakeStorage{loadErr: errors.New("disk error")}, testJWT())

		if _, err := svc.Authenticate("alice@example.com", validPassword); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestService_Me(t *testing.T) {
	existing := User{ID: "u1", Email: "alice@example.com"}

	t.Run("missing user id", func(t *testing.T) {
		svc := NewService(&fakeStorage{}, testJWT())

		_, err := svc.Me("")
		if !errors.Is(err, ErrUserNotFound) {
			t.Fatalf("got error %v, want %v", err, ErrUserNotFound)
		}
	})

	t.Run("found", func(t *testing.T) {
		svc := NewService(&fakeStorage{users: []User{existing}}, testJWT())

		got, err := svc.Me("u1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != "u1" {
			t.Fatalf("got id %q, want %q", got.ID, "u1")
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc := NewService(&fakeStorage{users: []User{existing}}, testJWT())

		_, err := svc.Me("does-not-exist")
		if !errors.Is(err, ErrUserNotFound) {
			t.Fatalf("got error %v, want %v", err, ErrUserNotFound)
		}
	})

	t.Run("load failure is returned to the caller", func(t *testing.T) {
		svc := NewService(&fakeStorage{loadErr: errors.New("disk error")}, testJWT())

		if _, err := svc.Me("u1"); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
