package project

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service interface {
	Create(ownerID, name, description string) (*Project, error)
	ListByOwner(ownerID string) ([]Project, error)
	GetByID(ownerID, projectID string) (*Project, error)
	Update(
		ownerID,
		projectID,
		name,
		description string,
	) (*Project, error)
	Delete(ownerID, projectID string) error
}

type service struct {
	storage Storage
}

// Create implements Service.
func (s *service) Create(ownerID string, name string, description string) (*Project, error) {
	if ownerID == "" {
		return nil, ErrInvalidOwnerID
	}

	if name == "" {
		return nil, ErrInvalidProjectName
	}

	if description == "" {
		return nil, ErrInvalidProjectDescription
	}

	projects, err := s.storage.Load()
	if err != nil {
		fmt.Printf("[Service Create] Failed to load projects: %v\n", err)
		return nil, err
	}

	timeNow := time.Now()

	project := &Project{
		ID:          uuid.NewString(),
		OwnerID:     ownerID,
		Name:        name,
		Description: description,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	}

	projects = append(projects, *project)

	err = s.storage.Save(projects)
	if err != nil {
		fmt.Printf("[Service Create] Failed to save projects: %v\n", err)
		return nil, err
	}

	return project, nil
}

// Delete implements Service.
func (s *service) Delete(ownerID string, projectID string) error {
	if ownerID == "" {
		return ErrInvalidOwnerID
	}

	if projectID == "" {
		return ErrInvalidProjectID
	}

	projects, err := s.storage.Load()
	if err != nil {
		fmt.Printf("[Service Delete] Failed to load projects: %v\n", err)
		return err
	}

	for i, p := range projects {
		if p.ID == projectID && p.OwnerID == ownerID {
			projects = append(projects[:i], projects[i+1:]...)
			err = s.storage.Save(projects)
			if err != nil {
				fmt.Printf("[Service Delete] Failed to save projects: %v\n", err)
				return err
			}
			return nil
		}
	}

	fmt.Printf("[Service Delete] Project with ID %s not found for owner %s\n", projectID, ownerID)
	return ErrProjectNotFound
}

// GetByID implements Service.
func (s *service) GetByID(ownerID string, projectID string) (*Project, error) {
	if ownerID == "" {
		return nil, ErrInvalidOwnerID
	}

	if projectID == "" {
		return nil, ErrInvalidProjectID
	}

	projects, err := s.storage.Load()
	if err != nil {
		fmt.Printf("[Service GetByID] Failed to load projects: %v\n", err)
		return nil, err
	}

	for _, p := range projects {
		if p.ID == projectID && p.OwnerID == ownerID {
			return &p, nil
		}
	}

	fmt.Printf("[Service GetByID] Project with ID %s not found for owner %s\n", projectID, ownerID)
	return nil, ErrProjectNotFound
}

// ListByOwner implements Service.
func (s *service) ListByOwner(ownerID string) ([]Project, error) {
	if ownerID == "" {
		return nil, ErrInvalidOwnerID
	}

	projects, err := s.storage.Load()
	if err != nil {
		fmt.Printf("[Service ListByOwner] Failed to load projects: %v\n", err)
		return nil, err
	}

	var ownerProjects []Project
	for _, p := range projects {
		if p.OwnerID == ownerID {
			ownerProjects = append(ownerProjects, p)
		}
	}

	return ownerProjects, nil
}

// Update implements Service.
func (s *service) Update(ownerID string, projectID string, name string, description string) (*Project, error) {
	if ownerID == "" {
		return nil, ErrInvalidOwnerID
	}

	if projectID == "" {
		return nil, ErrInvalidProjectID
	}

	if name == "" {
		return nil, ErrInvalidProjectName
	}

	if description == "" {
		return nil, ErrInvalidProjectDescription
	}

	projects, err := s.storage.Load()
	if err != nil {
		fmt.Printf("[Service Update] Failed to load projects: %v\n", err)
		return nil, err
	}

	for i, p := range projects {
		if p.ID == projectID && p.OwnerID == ownerID {
			p.Name = name
			p.Description = description
			p.UpdatedAt = time.Now()
			projects[i] = p

			err = s.storage.Save(projects)
			if err != nil {
				fmt.Printf("[Service Update] Failed to save projects: %v\n", err)
				return nil, err
			}
			return &p, nil
		}
	}

	fmt.Printf("[Service Update] Project with ID %s not found for owner %s\n", projectID, ownerID)
	return nil, ErrProjectNotFound
}

func NewService(storage Storage) Service {
	return &service{storage: storage}
}
