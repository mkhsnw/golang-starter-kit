package repository

import (
	"github.com/mkhsnw/golang-starter-kit/internal/entity"
	"gorm.io/gorm"
)

type ProductRepository struct {
	*Repository[entity.Product]
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{
		Repository: NewRepository[entity.Product](db),
	}
}
