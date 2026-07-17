package repository

import (
	"context"
	"time"

	"github.com/mkhsnw/golang-starter-kit/internal/entity"
	"gorm.io/gorm"
)

type RefreshTokenRepositoryInterface interface {
	RepositoryInterface[entity.RefreshToken]
	FindByTokenHash(ctx context.Context, hash string) (*entity.RefreshToken, error)
	RevokeAllForUser(ctx context.Context, userID string) error
}

type RefreshTokenRepository struct {
	*Repository[entity.RefreshToken]
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{Repository: NewRepository[entity.RefreshToken](db)}
}

func (r *RefreshTokenRepository) FindByTokenHash(ctx context.Context, hash string) (*entity.RefreshToken, error) {
	var rt entity.RefreshToken
	err := r.getDB(ctx).Where("token_hash = ? AND revoked = false AND expires_at > ?", hash, time.Now()).First(&rt).Error
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	return r.getDB(ctx).Model(&entity.RefreshToken{}).Where("user_id = ?", userID).Update("revoked", true).Error
}
