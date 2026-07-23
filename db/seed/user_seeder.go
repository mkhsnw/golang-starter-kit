package seed

import (
	"log"

	"github.com/mkhsnw/rel/db/factory"
	"github.com/mkhsnw/rel/internal/module/user"
	"gorm.io/gorm"
)

type UserSeeder struct {
	userFactory *factory.UserFactory
}

func NewUserSeeder() *UserSeeder {
	return &UserSeeder{
		userFactory: factory.NewUserFactory(),
	}
}

func (s *UserSeeder) Name() string {
	return "UserSeeder"
}

func (s *UserSeeder) Seed(db *gorm.DB) error {
	// Clean existing users & refresh tokens
	if err := TruncateTables(db, "refresh_tokens", "users"); err != nil {
		return err
	}

	admin, err := s.userFactory.BuildDefaultAdmin()
	if err != nil {
		return err
	}

	regularUser, err := s.userFactory.BuildDefaultUser()
	if err != nil {
		return err
	}

	users := []*user.User{admin, regularUser}
	if err := db.Create(&users).Error; err != nil {
		return err
	}

	log.Println("-------------------------------------------")
	log.Println("👤 Seeded Users:")
	log.Printf("   1. [Admin]   ID: %s | Email: admin@example.com | Pass: Password123!\n", admin.ID)
	log.Printf("   2. [User]    ID: %s | Email: john.doe@example.com | Pass: Password123!\n", regularUser.ID)
	log.Println("-------------------------------------------")

	return nil
}
