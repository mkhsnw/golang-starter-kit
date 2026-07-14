package repository

import (
	"context"
	"gorm.io/gorm"
)

type Repository[T any] struct {
	DB *gorm.DB
}

func NewRepository[T any](db *gorm.DB) *Repository[T] {
	return &Repository[T]{DB: db}
}

func (r *Repository[T]) Create(ctx context.Context, entity *T) error {
	return r.DB.WithContext(ctx).Create(entity).Error
}

func (r *Repository[T]) FindByID(ctx context.Context, id any) (*T, error) {
	var entity T
	if err := r.DB.WithContext(ctx).First(&entity, id).Error; err != nil {
		return nil, err // biarkan gorm.ErrRecordNotFound naik, sudah dihandle di error handler
	}
	return &entity, nil
}

func (r *Repository[T]) Update(ctx context.Context, entity *T) error {
	return r.DB.WithContext(ctx).Save(entity).Error
}

func (r *Repository[T]) Delete(ctx context.Context, entity *T) error {
	return r.DB.WithContext(ctx).Delete(entity).Error
}

func (r *Repository[T]) FindAllPaginated(ctx context.Context, page, size int) ([]T, int64, error) {
	var entities []T
	var total int64

	db := r.DB.WithContext(ctx).Model(new(T))

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	if err := db.Offset(offset).Limit(size).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}