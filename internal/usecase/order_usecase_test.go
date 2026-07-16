package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mkhsnw/golang-starter-kit/internal/entity"
	"github.com/mkhsnw/golang-starter-kit/internal/model"
	"github.com/mkhsnw/golang-starter-kit/internal/repository/mocks"
	"github.com/mkhsnw/golang-starter-kit/internal/usecase"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─── Helper ──────────────────────────────────────────────────────────────────

type dummyTxManager struct{}

func (d *dummyTxManager) Run(ctx context.Context, fn func(ctxTx context.Context) error) error {
	return fn(ctx)
}

func newOrderUsecase(t *testing.T) (*usecase.OrderUsecase, *mocks.OrderRepositoryInterface) {
	t.Helper()
	repoMock := new(mocks.OrderRepositoryInterface)
	logger := logrus.New()
	uc := usecase.NewOrderUsecase(logger, &dummyTxManager{}, repoMock)
	return uc, repoMock
}

// ─── Create ───────────────────────────────────────────────────────────────────

func TestOrderUsecase_Create_Success(t *testing.T) {
	uc, repoMock := newOrderUsecase(t)

	req := &model.CreateOrderRequest{}

	repoMock.On("Create", mock.Anything, mock.AnythingOfType("*entity.Order")).
		Return(nil).Once()

	res, err := uc.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	repoMock.AssertExpectations(t)
}

func TestOrderUsecase_Create_DBError(t *testing.T) {
	uc, repoMock := newOrderUsecase(t)

	req := &model.CreateOrderRequest{}

	repoMock.On("Create", mock.Anything, mock.AnythingOfType("*entity.Order")).
		Return(errors.New("db error")).Once()

	res, err := uc.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	repoMock.AssertExpectations(t)
}

// ─── GetByID ──────────────────────────────────────────────────────────────────

func TestOrderUsecase_GetByID_Success(t *testing.T) {
	uc, repoMock := newOrderUsecase(t)

	item := &entity.Order{}
	repoMock.On("FindByID", mock.Anything, "uuid-1").
		Return(item, nil).Once()

	res, err := uc.GetByID(context.Background(), "uuid-1")

	assert.NoError(t, err)
	assert.NotNil(t, res)
	repoMock.AssertExpectations(t)
}

func TestOrderUsecase_GetByID_NotFound(t *testing.T) {
	uc, repoMock := newOrderUsecase(t)

	repoMock.On("FindByID", mock.Anything, "uuid-404").
		Return(nil, errors.New("record not found")).Once()

	res, err := uc.GetByID(context.Background(), "uuid-404")

	assert.Error(t, err)
	assert.Nil(t, res)
	repoMock.AssertExpectations(t)
}

// ─── GetAll ───────────────────────────────────────────────────────────────────

func TestOrderUsecase_GetAll_Success(t *testing.T) {
	uc, repoMock := newOrderUsecase(t)

	items := []entity.Order{}
	repoMock.On("FindAllCursor", mock.Anything, "", 10).
		Return(items, nil).Once()

	res, nextCursor, err := uc.GetAll(context.Background(), "", 10)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Nil(t, nextCursor)
	repoMock.AssertExpectations(t)
}

// ─── Update ───────────────────────────────────────────────────────────────────

func TestOrderUsecase_Update_NotFound(t *testing.T) {
	uc, repoMock := newOrderUsecase(t)

	repoMock.On("FindByID", mock.Anything, "uuid-99").
		Return(nil, errors.New("record not found")).Once()

	res, err := uc.Update(context.Background(), "uuid-99", &model.UpdateOrderRequest{})

	assert.Error(t, err)
	assert.Nil(t, res)
	repoMock.AssertExpectations(t)
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func TestOrderUsecase_Delete_Success(t *testing.T) {
	uc, repoMock := newOrderUsecase(t)

	item := &entity.Order{}
	repoMock.On("FindByID", mock.Anything, "uuid-1").Return(item, nil).Once()
	repoMock.On("Delete", mock.Anything, item).Return(nil).Once()

	err := uc.Delete(context.Background(), "uuid-1")

	assert.NoError(t, err)
	repoMock.AssertExpectations(t)
}

func TestOrderUsecase_Delete_NotFound(t *testing.T) {
	uc, repoMock := newOrderUsecase(t)

	repoMock.On("FindByID", mock.Anything, "uuid-99").
		Return(nil, errors.New("record not found")).Once()

	err := uc.Delete(context.Background(), "uuid-99")

	assert.Error(t, err)
	repoMock.AssertExpectations(t)
}
