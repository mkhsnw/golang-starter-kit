package model

import "time"

type CreateOrderRequest struct {
	UserId    string `json:"user_id" validate:"required"`
	ProductId string `json:"product_id" validate:"required"`
	Detail    string `json:"detail"`
	Amount    uint64 `json:"amount" validate:"required"`
	Total     int    `json:"total" validate:"required"`
}

type UpdateOrderRequest struct {
	UserId    *string `json:"user_id"`
	ProductId *string `json:"product_id"`
	Detail    *string `json:"detail"`
	Amount    *uint64 `json:"amount"`
	Total     *int    `json:"total"`
}

type OrderResponse struct {
	ID        string    `json:"id"`
	UserId    string    `json:"user_id"`
	ProductId string    `json:"product_id"`
	Detail    string    `json:"detail"`
	Amount    uint64    `json:"amount"`
	Total     int       `json:"total"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
