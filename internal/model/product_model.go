package model

import "time"

type CreateProductRequest struct {
	Name  string `json:"name" validate:"required"`
	Notes string `json:"notes"`
}

type UpdateProductRequest struct {
	Name  *string `json:"name"`
	Notes *string `json:"notes"`
}

type ProductResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
