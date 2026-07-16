package entity

import (
	"time"
)

type User struct {
	ID        string    `gorm:"primaryKey;column:id;type:varchar(36)"`
	Name      string    `gorm:"column:name;type:varchar(100);not null"`
	Email     string    `gorm:"column:email;type:varchar(100);not null;unique"`
	Password  string    `gorm:"column:password;type:varchar(255);not null"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (u *User) TableName() string {
	return "users"
}
