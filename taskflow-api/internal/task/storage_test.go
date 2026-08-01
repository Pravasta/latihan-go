package task

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
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
			t.Fatalf("got %d tasks, want 0", len(got))
		}
	})

	t.Run("empty file returns empty slice, not an error", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "tasks.json")
		if err := os.WriteFile(path, []byte(""), 0644); err != nil {
			t.Fatalf("failed to seed empty file: %v", err)
		}
		s := NewStorage(path)

		got, err := s.Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("got %d tasks, want 0", len(got))
		}
	})

	t.Run("malformed json returns an error", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "tasks.json")
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

		path := filepath.Join(t.TempDir(), "tasks.json")
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
		path := filepath.Join(t.TempDir(), "tasks.json")
		s := NewStorage(path)

		originalTasks := []Task{
			{ID: "1", Title: "Task 1", Description: "Description 1"},
			{ID: "2", Title: "Task 2", Description: "Description 2"},
		}

		if err := s.Save(originalTasks); err != nil {
			t.Fatalf("unexpected error saving tasks: %v", err)
		}

		got, err := s.Load()
		if err != nil {
			t.Fatalf("unexpected error loading tasks: %v", err)
		}

		if len(got) != len(originalTasks) {
			t.Fatalf("got %d tasks, want %d", len(got), len(originalTasks))
		}

		for i, task := range got {
			if task != originalTasks[i] {
				t.Errorf("task at index %d does not match: got %+v, want %+v", i, task, originalTasks[i])
			}
		}
	})

	t.Run("save writes valid, human-readable JSON", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "tasks.json")
		s := NewStorage(path)

		tasks := []Task{
			{ID: "1", Title: "Task 1", Description: "Description 1"},
		}

		if err := s.Save(tasks); err != nil {
			t.Fatalf("unexpected error saving tasks: %v", err)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("unexpected error reading file: %v", err)
		}

		var loadedTasks []Task
		if err := json.Unmarshal(data, &loadedTasks); err != nil {
			t.Fatalf("unexpected error unmarshaling JSON: %v", err)
		}
	})

	t.Run("save fails for an unwritable path", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "no-such-dir", "tasks.json")
		s := NewStorage(path)

		if err := s.Save([]Task{{ID: "1"}}); err == nil {
			t.Fatal("expected an error, got nil")
		}
	})
}
