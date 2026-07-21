package user

import "time"

type RefreshToken struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)"`
	UserId    string    `gorm:"column:user_id"`
	TokenHash string    `gorm:"column:token_hash"`
	ExpiresAt time.Time `gorm:"column:expires_at"`
	Revoked   bool      `gorm:"column:revoked;default:false"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (r *RefreshToken) TableName() string { return "refresh_tokens" }
