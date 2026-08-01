package task

import "taskflow-api/internal/project"

// Compile-time checks: if a fake drifts from the interface it stands in
// for (wrong method name, wrong signature), this fails the build instead
// of silently compiling as an unrelated type with no test ever catching it.
var (
	_ Storage        = (*fakeStorage)(nil)
	_ Service        = (*fakeService)(nil)
	_ ProjectService = (*fakeProjectService)(nil)
)

type fakeStorage struct {
	tasks   []Task
	loadErr error
	saveErr error
}

func (f *fakeStorage) Load() ([]Task, error) {
	if f.loadErr != nil {
		return nil, f.loadErr
	}
	return f.tasks, nil
}

func (f *fakeStorage) Save(tasks []Task) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.tasks = tasks
	return nil
}

type fakeService struct {
	createFn       func(ownerID, projectID, title, description string) (*Task, error)
	listFn         func(ownerID, projectID string, query TaskQuery) (*TaskListResult, error)
	getByIDFn      func(ownerID, projectID, taskID string) (*Task, error)
	updateFn       func(ownerID, projectID, taskID, title, description string) (*Task, error)
	updateStatusFn func(ownerID, projectID, taskID string, status TaskStatus) (*Task, error)
	deleteFn       func(ownerID, projectID, taskID string) error
}

func (f *fakeService) Create(ownerID, projectID, title, description string) (*Task, error) {
	return f.createFn(ownerID, projectID, title, description)
}

func (f *fakeService) List(ownerID, projectID string, query TaskQuery) (*TaskListResult, error) {
	return f.listFn(ownerID, projectID, query)
}

func (f *fakeService) GetByID(ownerID, projectID, taskID string) (*Task, error) {
	return f.getByIDFn(ownerID, projectID, taskID)
}

func (f *fakeService) Update(ownerID, projectID, taskID, title, description string) (*Task, error) {
	return f.updateFn(ownerID, projectID, taskID, title, description)
}

func (f *fakeService) UpdateStatus(ownerID, projectID, taskID string, status TaskStatus) (*Task, error) {
	return f.updateStatusFn(ownerID, projectID, taskID, status)
}

func (f *fakeService) Delete(ownerID, projectID, taskID string) error {
	return f.deleteFn(ownerID, projectID, taskID)
}

type fakeProjectService struct {
	getByIDFn func(ownerID, projectID string) (*project.Project, error)
}

func (f *fakeProjectService) GetByID(ownerID, projectID string) (*project.Project, error) {
	return f.getByIDFn(ownerID, projectID)
}
