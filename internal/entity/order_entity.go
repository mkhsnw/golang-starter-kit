package entity

import "time"

type Order struct {
	ID        string    `gorm:"primaryKey;type:varchar(36);column:id"`
	UserId    string    `gorm:"column:user_id"`
	ProductId string    `gorm:"column:product_id"`
	Detail    string    `gorm:"column:detail"`
	Amount    uint64    `gorm:"column:amount"`
	Total     int       `gorm:"column:total"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (e *Order) TableName() string {
	return "orders"
}
