package entity

import "time"

type Product struct {
	ID        string    `gorm:"primaryKey;type:varchar(36);column:id"`
	Name      string    `gorm:"column:name"`
	Notes     string    `gorm:"column:notes"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (e *Product) TableName() string {
	return "products"
}
