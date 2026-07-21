package user

import (
	"context"
	"time"

	"github.com/mkhsnw/golang-starter-kit/internal/foundation/database"
	"gorm.io/gorm"
)

type RefreshTokenRepository struct {
	*database.Repository[RefreshToken]
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{Repository: database.NewRepository[RefreshToken](db)}
}

func (r *RefreshTokenRepository) FindByTokenHash(ctx context.Context, hash string) (*RefreshToken, error) {
	var rt RefreshToken
	err := r.GetDB(ctx).Where("token_hash = ? AND revoked = false AND expires_at > ?", hash, time.Now()).First(&rt).Error
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	return r.GetDB(ctx).Model(&RefreshToken{}).Where("user_id = ?", userID).Update("revoked", true).Error
}
