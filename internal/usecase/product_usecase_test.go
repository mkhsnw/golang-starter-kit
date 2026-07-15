package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mkhsnw/golang-starter-kit/internal/entity"
	"github.com/mkhsnw/golang-starter-kit/internal/model"
	"github.com/mkhsnw/golang-starter-kit/internal/repository/mocks"
	"github.com/mkhsnw/golang-starter-kit/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─── Helper ──────────────────────────────────────────────────────────────────

func newProductUsecase(t *testing.T) (*usecase.ProductUsecase, *mocks.ProductRepositoryInterface) {
	t.Helper()
	repoMock := new(mocks.ProductRepositoryInterface)
	uc := usecase.NewProductUsecase(repoMock)
	return uc, repoMock
}

// ─── Create ───────────────────────────────────────────────────────────────────

func TestProductUsecase_Create_Success(t *testing.T) {
	uc, repoMock := newProductUsecase(t)

	req := &model.CreateProductRequest{}

	repoMock.On("Create", mock.Anything, mock.AnythingOfType("*entity.Product")).
		Return(nil).Once()

	res, err := uc.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	repoMock.AssertExpectations(t)
}

func TestProductUsecase_Create_DBError(t *testing.T) {
	uc, repoMock := newProductUsecase(t)

	req := &model.CreateProductRequest{}

	repoMock.On("Create", mock.Anything, mock.AnythingOfType("*entity.Product")).
		Return(errors.New("db error")).Once()

	res, err := uc.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	repoMock.AssertExpectations(t)
}

// ─── GetByID ──────────────────────────────────────────────────────────────────

func TestProductUsecase_GetByID_Success(t *testing.T) {
	uc, repoMock := newProductUsecase(t)

	item := &entity.Product{}
	repoMock.On("FindByID", mock.Anything, uint64(1)).
		Return(item, nil).Once()

	res, err := uc.GetByID(context.Background(), 1)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	repoMock.AssertExpectations(t)
}

func TestProductUsecase_GetByID_NotFound(t *testing.T) {
	uc, repoMock := newProductUsecase(t)

	repoMock.On("FindByID", mock.Anything, uint64(404)).
		Return(nil, errors.New("record not found")).Once()

	res, err := uc.GetByID(context.Background(), 404)

	assert.Error(t, err)
	assert.Nil(t, res)
	repoMock.AssertExpectations(t)
}

// ─── GetAll ───────────────────────────────────────────────────────────────────

func TestProductUsecase_GetAll_Success(t *testing.T) {
	uc, repoMock := newProductUsecase(t)

	items := []entity.Product{}
	repoMock.On("FindAllPaginated", mock.Anything, 1, 10).
		Return(items, int64(0), nil).Once()

	res, total, err := uc.GetAll(context.Background(), 1, 10)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, int64(0), total)
	repoMock.AssertExpectations(t)
}

// ─── Update ───────────────────────────────────────────────────────────────────

func TestProductUsecase_Update_NotFound(t *testing.T) {
	uc, repoMock := newProductUsecase(t)

	repoMock.On("FindByID", mock.Anything, uint64(99)).
		Return(nil, errors.New("record not found")).Once()

	res, err := uc.Update(context.Background(), 99, &model.UpdateProductRequest{})

	assert.Error(t, err)
	assert.Nil(t, res)
	repoMock.AssertExpectations(t)
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func TestProductUsecase_Delete_Success(t *testing.T) {
	uc, repoMock := newProductUsecase(t)

	item := &entity.Product{}
	repoMock.On("FindByID", mock.Anything, uint64(1)).Return(item, nil).Once()
	repoMock.On("Delete", mock.Anything, item).Return(nil).Once()

	err := uc.Delete(context.Background(), 1)

	assert.NoError(t, err)
	repoMock.AssertExpectations(t)
}

func TestProductUsecase_Delete_NotFound(t *testing.T) {
	uc, repoMock := newProductUsecase(t)

	repoMock.On("FindByID", mock.Anything, uint64(99)).
		Return(nil, errors.New("record not found")).Once()

	err := uc.Delete(context.Background(), 99)

	assert.Error(t, err)
	repoMock.AssertExpectations(t)
}
