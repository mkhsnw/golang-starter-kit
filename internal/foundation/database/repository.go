package database

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
	FindAllCursor(ctx context.Context, cursor string, size int, preloads ...string) ([]T, error)
}

type Repository[T any] struct {
	DB *gorm.DB
}

func NewRepository[T any](db *gorm.DB) *Repository[T] {
	return &Repository[T]{DB: db}
}

func (r *Repository[T]) GetDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(TxKey).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return r.DB.WithContext(ctx)
}

func (r *Repository[T]) Create(ctx context.Context, entity *T) error {
	return r.GetDB(ctx).Create(entity).Error
}

func (r *Repository[T]) FindByID(ctx context.Context, id any, preloads ...string) (*T, error) {
	var entity T
	db := r.GetDB(ctx)

	for _, p := range preloads {
		db = db.Preload(p)
	}

	if err := db.First(&entity, "id = ?", id).Error; err != nil {
		return nil, err // biarkan gorm.ErrRecordNotFound naik, sudah dihandle di error handler
	}
	return &entity, nil
}

func (r *Repository[T]) Update(ctx context.Context, entity *T) error {
	return r.GetDB(ctx).Updates(entity).Error
}

func (r *Repository[T]) Delete(ctx context.Context, entity *T) error {
	return r.GetDB(ctx).Delete(entity).Error
}

func (r *Repository[T]) FindAllPaginated(ctx context.Context, page, size int, preloads ...string) ([]T, int64, error) {
	var entities []T
	var total int64

	dbCount := r.GetDB(ctx).Model(new(T))
	if err := dbCount.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	dbFetch := r.GetDB(ctx).Model(new(T))
	for _, p := range preloads {
		dbFetch = dbFetch.Preload(p)
	}

	offset := (page - 1) * size
	if err := dbFetch.Offset(offset).Limit(size).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

// FindAllCursor performs cursor-based pagination using IDs (works for ordered UUID v7 or numeric IDs).
func (r *Repository[T]) FindAllCursor(ctx context.Context, cursor string, size int, preloads ...string) ([]T, error) {
	var entities []T
	db := r.GetDB(ctx).Model(new(T))

	for _, p := range preloads {
		db = db.Preload(p)
	}

	if cursor != "" {
		db = db.Where("id > ?", cursor)
	}

	err := db.Order("id ASC").Limit(size).Find(&entities).Error
	return entities, err
}
