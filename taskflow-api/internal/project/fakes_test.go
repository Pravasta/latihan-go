package project

// fakeStorage is a test double for Storage. It keeps projects in memory and
// lets a test force Load/Save to fail, so service_test.go can exercise error
// paths without touching the filesystem.
type fakeStorage struct {
	projects []Project
	loadErr  error
	saveErr  error
}

func (f *fakeStorage) Load() ([]Project, error) {
	if f.loadErr != nil {
		return nil, f.loadErr
	}
	return f.projects, nil
}

func (f *fakeStorage) Save(projects []Project) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.projects = projects
	return nil
}

// fakeService is a test double for Service, used by handler_test.go. Each
// method is a func field so a test can plug in only the behavior it needs
// without maintaining in-memory state.
type fakeService struct {
	createFn func(ownerID, name, description string) (*Project, error)
	listFn   func(ownerID string) ([]Project, error)
	getFn    func(ownerID, projectID string) (*Project, error)
	updateFn func(ownerID, projectID, name, description string) (*Project, error)
	deleteFn func(ownerID, projectID string) error
}

func (f *fakeService) Create(ownerID, name, description string) (*Project, error) {
	return f.createFn(ownerID, name, description)
}

func (f *fakeService) ListByOwner(ownerID string) ([]Project, error) {
	return f.listFn(ownerID)
}

func (f *fakeService) GetByID(ownerID, projectID string) (*Project, error) {
	return f.getFn(ownerID, projectID)
}

func (f *fakeService) Update(ownerID, projectID, name, description string) (*Project, error) {
	return f.updateFn(ownerID, projectID, name, description)
}

func (f *fakeService) Delete(ownerID, projectID string) error {
	return f.deleteFn(ownerID, projectID)
}
