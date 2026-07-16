package usecase

import (
	"context"
	"github.com/google/uuid"
	"time"

	"github.com/mkhsnw/golang-starter-kit/internal/entity"
	"github.com/mkhsnw/golang-starter-kit/internal/exception"
	"github.com/mkhsnw/golang-starter-kit/internal/model"
	"github.com/mkhsnw/golang-starter-kit/internal/repository"
	"github.com/mkhsnw/golang-starter-kit/internal/util"
	"github.com/sirupsen/logrus"
)

type ProductUsecase struct {
	Log               *logrus.Logger
	ProductRepository repository.ProductRepositoryInterface
}

func NewProductUsecase(log *logrus.Logger, repo repository.ProductRepositoryInterface) *ProductUsecase {
	return &ProductUsecase{Log: log, ProductRepository: repo}
}

func (u *ProductUsecase) log(ctx context.Context) *logrus.Entry {
	if reqID, ok := ctx.Value(util.ContextKeyRequestID).(string); ok {
		return u.Log.WithField("requestid", reqID)
	}
	return u.Log.WithField("requestid", "unknown")
}

func (u *ProductUsecase) Create(ctx context.Context, req *model.CreateProductRequest) (*model.ProductResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	product := &entity.Product{
		ID:    uuid.New().String(),
		Name:  req.Name,
		Notes: req.Notes,
	}

	if err := u.ProductRepository.Create(ctx, product); err != nil {
		u.log(ctx).Errorf("failed to create product: %v", err)
		return nil, exception.NewResponseError(500, "INTERNAL_SERVER_ERROR", "failed to create product")
	}

	res := toProductResponse(product)
	return &res, nil
}

func (u *ProductUsecase) GetByID(ctx context.Context, id string) (*model.ProductResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	product, err := u.ProductRepository.FindByID(ctx, id)
	if err != nil {
		return nil, exception.NotFound("Product not found")
	}

	res := toProductResponse(product)
	return &res, nil
}
func (u *ProductUsecase) GetAll(ctx context.Context, page, size int) ([]model.ProductResponse, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	items, total, err := u.ProductRepository.FindAllPaginated(ctx, page, size)
	if err != nil {
		u.log(ctx).Errorf("failed to fetch products: %v", err)
		return nil, 0, exception.NewResponseError(500, "INTERNAL_SERVER_ERROR", "failed to fetch products")
	}

	responses := make([]model.ProductResponse, len(items))
	for i := range items {
		responses[i] = toProductResponse(&items[i])
	}
	return responses, total, nil
}

func (u *ProductUsecase) Update(ctx context.Context, id string, req *model.UpdateProductRequest) (*model.ProductResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	product, err := u.ProductRepository.FindByID(ctx, id)
	if err != nil {
		return nil, exception.NotFound("Product not found")
	}

	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Notes != nil {
		product.Notes = *req.Notes
	}

	if err := u.ProductRepository.Update(ctx, product); err != nil {
		u.log(ctx).Errorf("failed to update product: %v", err)
		return nil, exception.NewResponseError(500, "INTERNAL_SERVER_ERROR", "failed to update product")
	}

	res := toProductResponse(product)
	return &res, nil
}

func (u *ProductUsecase) Delete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	product, err := u.ProductRepository.FindByID(ctx, id)
	if err != nil {
		return exception.NotFound("Product not found")
	}
	if err := u.ProductRepository.Delete(ctx, product); err != nil {
		u.log(ctx).Errorf("failed to delete product: %v", err)
		return exception.NewResponseError(500, "INTERNAL_SERVER_ERROR", "failed to delete product")
	}
	return nil
}

func toProductResponse(e *entity.Product) model.ProductResponse {
	return model.ProductResponse{
		ID:        e.ID,
		Name:      e.Name,
		Notes:     e.Notes,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}
