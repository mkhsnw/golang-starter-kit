package usecase

import (
	"context"
	"time"

	"github.com/mkhsnw/golang-starter-kit/internal/entity"
	"github.com/mkhsnw/golang-starter-kit/internal/exception"
	"github.com/mkhsnw/golang-starter-kit/internal/model"
	"github.com/mkhsnw/golang-starter-kit/internal/repository"
)

type ProductUsecase struct {
	ProductRepository repository.ProductRepositoryInterface
}

func NewProductUsecase(repo repository.ProductRepositoryInterface) *ProductUsecase {
	return &ProductUsecase{ProductRepository: repo}
}

func (u *ProductUsecase) Create(ctx context.Context, req *model.CreateProductRequest) (*model.ProductResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	product := &entity.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
	}

	if err := u.ProductRepository.Create(ctx, product); err != nil {
		return nil, exception.NewResponseError(500, "failed to create product")
	}

	res := toProductResponse(product)
	return &res, nil
}

func (u *ProductUsecase) GetByID(ctx context.Context, id uint64) (*model.ProductResponse, error) {
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
		return nil, 0, exception.NewResponseError(500, "failed to fetch products")
	}

	responses := make([]model.ProductResponse, len(items))
	for i := range items {
		responses[i] = toProductResponse(&items[i])
	}
	return responses, total, nil
}

func (u *ProductUsecase) Update(ctx context.Context, id uint64, req *model.UpdateProductRequest) (*model.ProductResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	product, err := u.ProductRepository.FindByID(ctx, id)
	if err != nil {
		return nil, exception.NotFound("Product not found")
	}

	product.Name = req.Name
	product.Description = req.Description
	product.Price = req.Price
	product.Stock = req.Stock

	if err := u.ProductRepository.Update(ctx, product); err != nil {
		return nil, exception.NewResponseError(500, "failed to update product")
	}

	res := toProductResponse(product)
	return &res, nil
}

func (u *ProductUsecase) Delete(ctx context.Context, id uint64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	product, err := u.ProductRepository.FindByID(ctx, id)
	if err != nil {
		return exception.NotFound("Product not found")
	}
	if err := u.ProductRepository.Delete(ctx, product); err != nil {
		return exception.NewResponseError(500, "failed to delete product")
	}
	return nil
}

func toProductResponse(e *entity.Product) model.ProductResponse {
	return model.ProductResponse{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
		Price:       e.Price,
		Stock:       e.Stock,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}
