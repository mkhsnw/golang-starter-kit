package entity

import "time"

type Product struct {
	ID          uint64    `gorm:"primaryKey;column:id"`
	Name        string    `gorm:"column:name"`
	Description string    `gorm:"column:description"`
	Price       int       `gorm:"column:price"`
	Stock       int32     `gorm:"column:stock"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (e *Product) TableName() string {
	return "products"
}
