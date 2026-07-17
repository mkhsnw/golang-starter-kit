package repository

import (
	"context"

	"github.com/mkhsnw/golang-starter-kit/internal/entity"
)

type UserRepositoryInterface interface {
	RepositoryInterface[entity.User]
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
}

// @InjectRepositoryInterface
