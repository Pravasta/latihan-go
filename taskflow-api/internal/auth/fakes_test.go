package auth

// Compile-time checks: if a fake drifts from the interface it stands in
// for (wrong method name, wrong signature), this fails the build instead
// of silently compiling as an unrelated type with no test ever catching it.
var (
	_ Storage = (*fakeStorage)(nil)
	_ Service = (*fakeService)(nil)
)

type fakeStorage struct {
	users   []User
	loadErr error
	saveErr error
}

func (f *fakeStorage) Load() ([]User, error) {
	if f.loadErr != nil {
		return nil, f.loadErr
	}
	return f.users, nil
}

func (f *fakeStorage) Save(users []User) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.users = users
	return nil
}

type fakeService struct {
	createUserFn   func(name, email, password string) (*User, error)
	authenticateFn func(email, password string) (string, error)
	meFn           func(userID string) (*User, error)
}

func (f *fakeService) CreateUser(name, email, password string) (*User, error) {
	return f.createUserFn(name, email, password)
}

func (f *fakeService) Authenticate(email, password string) (string, error) {
	return f.authenticateFn(email, password)
}

func (f *fakeService) Me(userID string) (*User, error) {
	return f.meFn(userID)
}
