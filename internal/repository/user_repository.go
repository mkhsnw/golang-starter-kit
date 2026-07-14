package repository

import (
	"context"

	"github.com/mkhsnw/golang-starter-kit/internal/entity"
	"gorm.io/gorm"
)

type UserRepository struct {
	*Repository[entity.User]
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		Repository: NewRepository[entity.User](db),
	}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	if err := r.DB.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
