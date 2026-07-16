package usecase

import (
	"context"

	"github.com/mkhsnw/golang-starter-kit/internal/model"
)

type UserUsecaseInterface interface {
	Register(ctx context.Context, req *model.RegisterRequest) (*model.UserResponse, error)
	Login(ctx context.Context, req *model.LoginRequest) (*model.TokenResponse, error)
	GetCurrentUser(ctx context.Context, userID string) (*model.UserResponse, error)
}

type ProductUsecaseInterface interface {
	Create(ctx context.Context, req *model.CreateProductRequest) (*model.ProductResponse, error)
	GetByID(ctx context.Context, id string) (*model.ProductResponse, error)
	GetAll(ctx context.Context, page, size int) ([]model.ProductResponse, int64, error)
	Update(ctx context.Context, id string, req *model.UpdateProductRequest) (*model.ProductResponse, error)
	Delete(ctx context.Context, id string) error
}

type OrderUsecaseInterface interface {
	Create(ctx context.Context, req *model.CreateOrderRequest) (*model.OrderResponse, error)
	GetByID(ctx context.Context, id string) (*model.OrderResponse, error)
	GetAll(ctx context.Context, cursor string, size int) ([]model.OrderResponse, *string, error)
	Update(ctx context.Context, id string, req *model.UpdateOrderRequest) (*model.OrderResponse, error)
	Delete(ctx context.Context, id string) error
}

// @InjectUsecaseInterface
