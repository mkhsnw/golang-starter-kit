package factory

import (
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/mkhsnw/rel/internal/module/user"
	"golang.org/x/crypto/bcrypt"
)

// UserFactory provides helper methods for creating User instances
type UserFactory struct{}

func NewUserFactory() *UserFactory {
	return &UserFactory{}
}

// BuildUser creates a new User entity with default or overridden values
func (f *UserFactory) BuildUser(name, email, plainPassword string) (*user.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &user.User{
		ID:        uuid.New().String(),
		Name:      name,
		Email:     email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// BuildRandomUser creates a user with randomized realistic fake data
func (f *UserFactory) BuildRandomUser(plainPassword string) (*user.User, error) {
	return f.BuildUser(gofakeit.Name(), gofakeit.Email(), plainPassword)
}

// BuildDefaultAdmin returns a pre-configured Admin user
func (f *UserFactory) BuildDefaultAdmin() (*user.User, error) {
	return f.BuildUser("Admin User", "admin@example.com", "Password123!")
}

// BuildDefaultUser returns a pre-configured regular test user
func (f *UserFactory) BuildDefaultUser() (*user.User, error) {
	return f.BuildUser("John Doe", "john.doe@example.com", "Password123!")
}

// BuildBatch builds multiple randomized test users
func (f *UserFactory) BuildBatch(count int, overrides ...func(int, *user.User)) []*user.User {
	items := make([]*user.User, 0, count)
	for i := 1; i <= count; i++ {
		u, err := f.BuildRandomUser("Password123!")
		if err != nil {
			continue
		}
		for _, override := range overrides {
			override(i, u)
		}
		items = append(items, u)
	}
	return items
}
