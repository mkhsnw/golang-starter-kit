package user

import (
	"context"
	"errors"

	"github.com/mkhsnw/rel/internal/foundation/database"
	"gorm.io/gorm"
)

type UserRepository struct {
	*database.Repository[User]
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{Repository: database.NewRepository[User](db)}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	if err := r.DB.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
