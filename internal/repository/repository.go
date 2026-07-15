package repository

import (
	"context"
	"gorm.io/gorm"
)

type RepositoryInterface[T any] interface {
	Create(ctx context.Context, entity *T) error
	FindByID(ctx context.Context, id any, preloads ...string) (*T, error)
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, entity *T) error
	FindAllPaginated(ctx context.Context, page, size int, preloads ...string) ([]T, int64, error)
	FindAllCursor(ctx context.Context, cursor uint64, size int, preloads ...string) ([]T, error)
}

type Repository[T any] struct {
	DB *gorm.DB
}

func NewRepository[T any](db *gorm.DB) *Repository[T] {
	return &Repository[T]{DB: db}
}

func (r *Repository[T]) Create(ctx context.Context, entity *T) error {
	return r.DB.WithContext(ctx).Create(entity).Error
}

func (r *Repository[T]) FindByID(ctx context.Context, id any, preloads ...string) (*T, error) {
	var entity T
	db := r.DB.WithContext(ctx)

	for _, p := range preloads {
		db = db.Preload(p)
	}

	if err := db.First(&entity, id).Error; err != nil {
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

func (r *Repository[T]) FindAllPaginated(ctx context.Context, page, size int, preloads ...string) ([]T, int64, error) {
	var entities []T
	var total int64

	db := r.DB.WithContext(ctx).Model(new(T))

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	for _, p := range preloads {
		db = db.Preload(p)
	}

	offset := (page - 1) * size
	if err := db.Offset(offset).Limit(size).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

func (r *Repository[T]) FindAllCursor(ctx context.Context, cursor uint64, size int, preloads ...string) ([]T, error) {
	var entities []T
	db := r.DB.WithContext(ctx).Model(new(T))

	for _, p := range preloads {
		db = db.Preload(p)
	}

	if cursor > 0 {
		db = db.Where("id > ?", cursor)
	}

	err := db.Order("id ASC").Limit(size).Find(&entities).Error
	return entities, err
}
