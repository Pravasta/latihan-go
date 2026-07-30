package project

import (
	"errors"
	"testing"
)

func TestService_Create(t *testing.T) {
	t.Run("validation", func(t *testing.T) {
		cases := []struct {
			name        string
			ownerID     string
			projectName string
			description string
			wantErr     error
		}{
			{"missing owner id", "", "Name", "Desc", ErrInvalidOwnerID},
			{"missing name", "owner-1", "", "Desc", ErrInvalidProjectName},
			{"missing description", "owner-1", "Name", "", ErrInvalidProjectDescription},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				svc := NewService(&fakeStorage{})

				_, err := svc.Create(tc.ownerID, tc.projectName, tc.description)
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("got error %v, want %v", err, tc.wantErr)
				}
			})
		}
	})

	t.Run("success persists a new project", func(t *testing.T) {
		storage := &fakeStorage{}
		svc := NewService(storage)

		got, err := svc.Create("owner-1", "My Project", "desc")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID == "" {
			t.Fatal("expected a generated ID, got empty string")
		}
		if got.OwnerID != "owner-1" {
			t.Fatalf("got owner %q, want %q", got.OwnerID, "owner-1")
		}
		if len(storage.projects) != 1 {
			t.Fatalf("expected 1 project persisted, got %d", len(storage.projects))
		}
	})

	t.Run("load failure is returned to the caller", func(t *testing.T) {
		svc := NewService(&fakeStorage{loadErr: errors.New("disk error")})

		if _, err := svc.Create("owner-1", "Name", "Desc"); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("save failure is returned to the caller", func(t *testing.T) {
		svc := NewService(&fakeStorage{saveErr: errors.New("disk full")})

		if _, err := svc.Create("owner-1", "Name", "Desc"); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestService_GetByID(t *testing.T) {
	existing := Project{ID: "p1", OwnerID: "owner-1", Name: "Existing"}

	t.Run("found for the correct owner", func(t *testing.T) {
		svc := NewService(&fakeStorage{projects: []Project{existing}})

		got, err := svc.GetByID("owner-1", "p1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != "p1" {
			t.Fatalf("got id %q, want %q", got.ID, "p1")
		}
	})

	t.Run("not found for a different owner", func(t *testing.T) {
		svc := NewService(&fakeStorage{projects: []Project{existing}})

		_, err := svc.GetByID("owner-2", "p1")
		if !errors.Is(err, ErrProjectNotFound) {
			t.Fatalf("got error %v, want %v", err, ErrProjectNotFound)
		}
	})

	t.Run("not found for unknown id", func(t *testing.T) {
		svc := NewService(&fakeStorage{projects: []Project{existing}})

		_, err := svc.GetByID("owner-1", "does-not-exist")
		if !errors.Is(err, ErrProjectNotFound) {
			t.Fatalf("got error %v, want %v", err, ErrProjectNotFound)
		}
	})

	t.Run("missing owner id", func(t *testing.T) {
		svc := NewService(&fakeStorage{})

		_, err := svc.GetByID("", "p1")
		if !errors.Is(err, ErrInvalidOwnerID) {
			t.Fatalf("got error %v, want %v", err, ErrInvalidOwnerID)
		}
	})

	t.Run("missing project id", func(t *testing.T) {
		svc := NewService(&fakeStorage{})

		_, err := svc.GetByID("owner-1", "")
		if !errors.Is(err, ErrInvalidProjectID) {
			t.Fatalf("got error %v, want %v", err, ErrInvalidProjectID)
		}
	})

	t.Run("load failure is returned to the caller", func(t *testing.T) {
		svc := NewService(&fakeStorage{loadErr: errors.New("disk error")})

		if _, err := svc.GetByID("owner-1", "p1"); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestService_ListByOwner(t *testing.T) {
	projects := []Project{
		{ID: "p1", OwnerID: "owner-1"},
		{ID: "p2", OwnerID: "owner-2"},
		{ID: "p3", OwnerID: "owner-1"},
	}

	svc := NewService(&fakeStorage{projects: projects})

	got, err := svc.ListByOwner("owner-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d projects, want 2", len(got))
	}
	for _, p := range got {
		if p.OwnerID != "owner-1" {
			t.Fatalf("got project owned by %q, want owner-1", p.OwnerID)
		}
	}

	t.Run("missing owner id", func(t *testing.T) {
		svc := NewService(&fakeStorage{})

		_, err := svc.ListByOwner("")
		if !errors.Is(err, ErrInvalidOwnerID) {
			t.Fatalf("got error %v, want %v", err, ErrInvalidOwnerID)
		}
	})

	t.Run("load failure is returned to the caller", func(t *testing.T) {
		svc := NewService(&fakeStorage{loadErr: errors.New("disk error")})

		if _, err := svc.ListByOwner("owner-1"); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestService_Update(t *testing.T) {
	existing := Project{ID: "p1", OwnerID: "owner-1", Name: "Old", Description: "Old desc"}

	t.Run("validation", func(t *testing.T) {
		cases := []struct {
			name        string
			ownerID     string
			projectID   string
			projectName string
			description string
			wantErr     error
		}{
			{"missing owner id", "", "p1", "New", "New desc", ErrInvalidOwnerID},
			{"missing project id", "owner-1", "", "New", "New desc", ErrInvalidProjectID},
			{"missing name", "owner-1", "p1", "", "New desc", ErrInvalidProjectName},
			{"missing description", "owner-1", "p1", "New", "", ErrInvalidProjectDescription},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				svc := NewService(&fakeStorage{projects: []Project{existing}})

				_, err := svc.Update(tc.ownerID, tc.projectID, tc.projectName, tc.description)
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("got error %v, want %v", err, tc.wantErr)
				}
			})
		}
	})

	t.Run("load failure is returned to the caller", func(t *testing.T) {
		svc := NewService(&fakeStorage{loadErr: errors.New("disk error")})

		if _, err := svc.Update("owner-1", "p1", "New", "New desc"); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("save failure is returned to the caller", func(t *testing.T) {
		svc := NewService(&fakeStorage{projects: []Project{existing}, saveErr: errors.New("disk full")})

		if _, err := svc.Update("owner-1", "p1", "New", "New desc"); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("updates name and description", func(t *testing.T) {
		storage := &fakeStorage{projects: []Project{existing}}
		svc := NewService(storage)

		got, err := svc.Update("owner-1", "p1", "New", "New desc")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name != "New" || got.Description != "New desc" {
			t.Fatalf("got %+v, want updated name/description", got)
		}
		if storage.projects[0].Name != "New" {
			t.Fatalf("update was not persisted: %+v", storage.projects[0])
		}
	})

	t.Run("not found for a different owner", func(t *testing.T) {
		svc := NewService(&fakeStorage{projects: []Project{existing}})

		_, err := svc.Update("owner-2", "p1", "New", "New desc")
		if !errors.Is(err, ErrProjectNotFound) {
			t.Fatalf("got error %v, want %v", err, ErrProjectNotFound)
		}
	})
}

func TestService_Delete(t *testing.T) {
	existing := Project{ID: "p1", OwnerID: "owner-1"}

	t.Run("validation", func(t *testing.T) {
		cases := []struct {
			name      string
			ownerID   string
			projectID string
			wantErr   error
		}{
			{"missing owner id", "", "p1", ErrInvalidOwnerID},
			{"missing project id", "owner-1", "", ErrInvalidProjectID},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				svc := NewService(&fakeStorage{projects: []Project{existing}})

				err := svc.Delete(tc.ownerID, tc.projectID)
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("got error %v, want %v", err, tc.wantErr)
				}
			})
		}
	})

	t.Run("load failure is returned to the caller", func(t *testing.T) {
		svc := NewService(&fakeStorage{loadErr: errors.New("disk error")})

		if err := svc.Delete("owner-1", "p1"); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("save failure is returned to the caller", func(t *testing.T) {
		svc := NewService(&fakeStorage{projects: []Project{existing}, saveErr: errors.New("disk full")})

		if err := svc.Delete("owner-1", "p1"); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("deletes an existing project", func(t *testing.T) {
		storage := &fakeStorage{projects: []Project{existing}}
		svc := NewService(storage)

		if err := svc.Delete("owner-1", "p1"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(storage.projects) != 0 {
			t.Fatalf("expected project to be removed, got %d remaining", len(storage.projects))
		}
	})

	t.Run("not found for a different owner", func(t *testing.T) {
		svc := NewService(&fakeStorage{projects: []Project{existing}})

		err := svc.Delete("owner-2", "p1")
		if !errors.Is(err, ErrProjectNotFound) {
			t.Fatalf("got error %v, want %v", err, ErrProjectNotFound)
		}
	})
}
