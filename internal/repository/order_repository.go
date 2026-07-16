package repository

import (
	"github.com/mkhsnw/golang-starter-kit/internal/entity"
	"gorm.io/gorm"
)

type OrderRepository struct {
	*Repository[entity.Order]
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{
		Repository: NewRepository[entity.Order](db),
	}
}
