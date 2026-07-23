package database

import (
	"time"

	"gorm.io/gorm"
)

// BaseEntity contains standard primary key and timestamp fields for domain models.
type BaseEntity struct {
	ID        string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}
