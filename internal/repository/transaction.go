package repository

import (
	"context"
	"gorm.io/gorm"
)

type txKeyType struct{}

var TxKey = txKeyType{}

type TransactionManager interface {
	Run(ctx context.Context, fn func(ctxTx context.Context) error) error
}

type GormTransactionManager struct {
	db *gorm.DB
}

func NewGormTransactionManager(db *gorm.DB) *GormTransactionManager {
	return &GormTransactionManager{db: db}
}

func (m *GormTransactionManager) Run(ctx context.Context, fn func(ctxTx context.Context) error) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctxTx := context.WithValue(ctx, TxKey, tx)
		return fn(ctxTx)
	})
}
