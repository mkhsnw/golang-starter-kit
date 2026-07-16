package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mkhsnw/golang-starter-kit/internal/entity"
	"github.com/mkhsnw/golang-starter-kit/internal/exception"
	"github.com/mkhsnw/golang-starter-kit/internal/model"
	"github.com/mkhsnw/golang-starter-kit/internal/repository"
	"github.com/mkhsnw/golang-starter-kit/internal/util"
	"github.com/sirupsen/logrus"
)

type OrderUsecase struct {
	Log             *logrus.Logger
	TxManager       repository.TransactionManager
	OrderRepository repository.OrderRepositoryInterface
}

func NewOrderUsecase(log *logrus.Logger, txManager repository.TransactionManager, repo repository.OrderRepositoryInterface) *OrderUsecase {
	return &OrderUsecase{
		Log:             log,
		TxManager:       txManager,
		OrderRepository: repo,
	}
}

func (u *OrderUsecase) log(ctx context.Context) *logrus.Entry {
	if reqID, ok := ctx.Value(util.ContextKeyRequestID).(string); ok {
		return u.Log.WithField("requestid", reqID)
	}
	return u.Log.WithField("requestid", "unknown")
}

func (u *OrderUsecase) Create(ctx context.Context, req *model.CreateOrderRequest) (*model.OrderResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	order := &entity.Order{
		ID: func() string {
			id, _ := uuid.NewV7()
			return id.String()
		}(),
		UserId:    req.UserId,
		ProductId: req.ProductId,
		Detail:    req.Detail,
		Amount:    req.Amount,
		Total:     req.Total,
	}
	var res model.OrderResponse
	err := u.TxManager.Run(ctx, func(ctxTx context.Context) error {
		if err := u.OrderRepository.Create(ctxTx, order); err != nil {
			u.log(ctxTx).Errorf("failed to create order: %v", err)
			return exception.NewResponseError(500, "INTERNAL_SERVER_ERROR", "failed to create order")
		}
		res = toOrderResponse(order)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (u *OrderUsecase) GetByID(ctx context.Context, id string) (*model.OrderResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	order, err := u.OrderRepository.FindByID(ctx, id)
	if err != nil {
		return nil, exception.NotFound("Order not found")
	}

	res := toOrderResponse(order)
	return &res, nil
}

func (u *OrderUsecase) GetAll(ctx context.Context, cursor string, size int) ([]model.OrderResponse, *string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	items, err := u.OrderRepository.FindAllCursor(ctx, cursor, size)
	if err != nil {
		u.log(ctx).Errorf("failed to fetch orders: %v", err)
		return nil, nil, exception.NewResponseError(500, "INTERNAL_SERVER_ERROR", "failed to fetch orders")
	}

	responses := make([]model.OrderResponse, len(items))
	for i := range items {
		responses[i] = toOrderResponse(&items[i])
	}

	var nextCursor *string
	if len(items) == size {
		nextCursor = &items[len(items)-1].ID
	}

	return responses, nextCursor, nil
}

func (u *OrderUsecase) Update(ctx context.Context, id string, req *model.UpdateOrderRequest) (*model.OrderResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	order, err := u.OrderRepository.FindByID(ctx, id)
	if err != nil {
		return nil, exception.NotFound("Order not found")
	}

	if req.UserId != nil {
		order.UserId = *req.UserId
	}
	if req.ProductId != nil {
		order.ProductId = *req.ProductId
	}
	if req.Detail != nil {
		order.Detail = *req.Detail
	}
	if req.Amount != nil {
		order.Amount = *req.Amount
	}
	if req.Total != nil {
		order.Total = *req.Total
	}
	var res model.OrderResponse
	err = u.TxManager.Run(ctx, func(ctxTx context.Context) error {
		if err := u.OrderRepository.Update(ctxTx, order); err != nil {
			u.log(ctxTx).Errorf("failed to update order: %v", err)
			return exception.NewResponseError(500, "INTERNAL_SERVER_ERROR", "failed to update order")
		}
		res = toOrderResponse(order)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (u *OrderUsecase) Delete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	order, err := u.OrderRepository.FindByID(ctx, id)
	if err != nil {
		return exception.NotFound("Order not found")
	}
	return u.TxManager.Run(ctx, func(ctxTx context.Context) error {
		if err := u.OrderRepository.Delete(ctxTx, order); err != nil {
			u.log(ctxTx).Errorf("failed to delete order: %v", err)
			return exception.NewResponseError(500, "INTERNAL_SERVER_ERROR", "failed to delete order")
		}
		return nil
	})
}

func toOrderResponse(e *entity.Order) model.OrderResponse {
	return model.OrderResponse{
		ID:        e.ID,
		UserId:    e.UserId,
		ProductId: e.ProductId,
		Detail:    e.Detail,
		Amount:    e.Amount,
		Total:     e.Total,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}
