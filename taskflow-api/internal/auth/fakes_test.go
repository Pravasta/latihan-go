package auth

import "context"

// Compile-time checks: if a fake drifts from the interface it stands in
// for (wrong method name, wrong signature), this fails the build instead
// of silently compiling as an unrelated type with no test ever catching it.
var (
	_ Repository = (*fakeRepository)(nil)
	_ Service    = (*fakeService)(nil)
)

type fakeRepository struct {
	users     []User
	createErr error
	findErr   error
}

func (f *fakeRepository) Create(ctx context.Context, user *User) (*User, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	for _, u := range f.users {
		if u.Email == user.Email {
			return nil, ErrEmailAlreadyExists
		}
	}
	f.users = append(f.users, *user)
	return user, nil
}

func (f *fakeRepository) FindByID(ctx context.Context, id string) (*User, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	for _, u := range f.users {
		if u.ID == id {
			return &u, nil
		}
	}
	return nil, ErrUserNotFound
}

func (f *fakeRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	for _, u := range f.users {
		if u.Email == email {
			return &u, nil
		}
	}
	return nil, ErrUserNotFound
}

type fakeService struct {
	createUserFn   func(ctx context.Context, name, email, password string) (*User, error)
	authenticateFn func(ctx context.Context, email, password string) (string, error)
	meFn           func(ctx context.Context, userID string) (*User, error)
}

func (f *fakeService) CreateUser(ctx context.Context, name, email, password string) (*User, error) {
	return f.createUserFn(ctx, name, email, password)
}

func (f *fakeService) Authenticate(ctx context.Context, email, password string) (string, error) {
	return f.authenticateFn(ctx, email, password)
}

func (f *fakeService) Me(ctx context.Context, userID string) (*User, error) {
	return f.meFn(ctx, userID)
}
