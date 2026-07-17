package usecase

import (
	"context"

	"github.com/mkhsnw/golang-starter-kit/internal/model"
)

type UserUsecaseInterface interface {
	Register(ctx context.Context, req *model.RegisterRequest) (*model.UserResponse, error)
	Login(ctx context.Context, req *model.LoginRequest) (*model.TokenResponse, error)
	GetCurrentUser(ctx context.Context, userID string) (*model.UserResponse, error)
	RefreshToken(ctx context.Context, rawToken string) (*model.TokenResponse, error)
	Logout(ctx context.Context, userID string) error
}

// @InjectUsecaseInterface
