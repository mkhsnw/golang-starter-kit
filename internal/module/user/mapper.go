package user

import (
	"github.com/mkhsnw/golang-starter-kit/internal/module/user/dto"
		)

func ToUserResponse(user *User) *dto.UserResponse {
	if user == nil {
		return nil
	}
	return &dto.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
