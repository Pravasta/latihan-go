package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestStorage_Load(t *testing.T) {
	t.Run("missing file returns empty slice, not an error", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "does-not-exist.json")
		s := NewStorage(path)

		got, err := s.Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("got %d users, want 0", len(got))
		}
	})

	t.Run("empty file returns empty slice, not an error", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "users.json")
		if err := os.WriteFile(path, []byte(""), 0644); err != nil {
			t.Fatalf("failed to seed empty file: %v", err)
		}
		s := NewStorage(path)

		got, err := s.Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("got %d users, want 0", len(got))
		}
	})

	t.Run("malformed json returns an error", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "users.json")
		if err := os.WriteFile(path, []byte("not-json"), 0644); err != nil {
			t.Fatalf("failed to seed malformed file: %v", err)
		}
		s := NewStorage(path)

		if _, err := s.Load(); err == nil {
			t.Fatal("expected an error, got nil")
		}
	})

	t.Run("unreadable file returns an error distinct from not-exist", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("running as root: file permissions do not block reads")
		}

		path := filepath.Join(t.TempDir(), "users.json")
		if err := os.WriteFile(path, []byte("[]"), 0000); err != nil {
			t.Fatalf("failed to seed unreadable file: %v", err)
		}
		s := NewStorage(path)

		if _, err := s.Load(); err == nil {
			t.Fatal("expected an error, got nil")
		}
	})
}

func TestStorage_Save(t *testing.T) {
	t.Run("round trip preserves data", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "users.json")
		s := NewStorage(path)

		want := []User{
			{
				ID:           "u1",
				Name:         "Alice",
				Email:        "alice@example.com",
				PasswordHash: "hashed",
				CreatedAt:    time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC),
			},
		}

		if err := s.Save(want); err != nil {
			t.Fatalf("unexpected error saving: %v", err)
		}

		got, err := s.Load()
		if err != nil {
			t.Fatalf("unexpected error loading: %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	})

	t.Run("save writes valid, human-readable json", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "users.json")
		s := NewStorage(path)

		if err := s.Save([]User{{ID: "u1"}}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read written file: %v", err)
		}

		var got []User
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatalf("written file is not valid json: %v", err)
		}
	})

	t.Run("save fails for an unwritable path", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "no-such-dir", "users.json")
		s := NewStorage(path)

		if err := s.Save([]User{}); err == nil {
			t.Fatal("expected an error, got nil")
		}
	})
}
