package usecase

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mkhsnw/golang-starter-kit/internal/entity"
	"github.com/mkhsnw/golang-starter-kit/internal/exception"
	"github.com/mkhsnw/golang-starter-kit/internal/model"
	"github.com/mkhsnw/golang-starter-kit/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecase struct {
	JwtSecret          string
	JwtExpirationHours int
	UserRepository     repository.UserRepositoryInterface
}

func NewUserUsecase(jwtSecret string, jwtExpHours int, userRepo repository.UserRepositoryInterface) *UserUsecase {
	return &UserUsecase{
		JwtSecret:          jwtSecret,
		JwtExpirationHours: jwtExpHours,
		UserRepository:     userRepo,
	}
}

func (u *UserUsecase) Register(ctx context.Context, req *model.RegisterRequest) (*model.UserResponse, error) {
	// Check if email already exists
	existingUser, err := u.UserRepository.FindByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, exception.Conflict("Email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, exception.NewResponseError(500, "INTERNAL_SERVER_ERROR", "Registration failed")
	}

	user := &entity.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := u.UserRepository.Create(ctx, user); err != nil {
		return nil, exception.NewResponseError(500, "INTERNAL_SERVER_ERROR", "Registration failed")
	}

	return &model.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (u *UserUsecase) Login(ctx context.Context, req *model.LoginRequest) (*model.TokenResponse, error) {
	// Find user by email
	user, err := u.UserRepository.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, exception.Unauthorized("Invalid email or password")
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, exception.Unauthorized("Invalid email or password")
	}

	// Generate JWT Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(time.Hour * time.Duration(u.JwtExpirationHours)).Unix(),
	})

	tokenString, err := token.SignedString([]byte(u.JwtSecret))
	if err != nil {
		return nil, exception.NewResponseError(500, "INTERNAL_SERVER_ERROR", "Failed to generate token")
	}

	return &model.TokenResponse{
		Token: tokenString,
	}, nil
}
func (u *UserUsecase) GetCurrentUser(ctx context.Context, userID uint64) (*model.UserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user, err := u.UserRepository.FindByID(ctx, userID)
	if err != nil {
		return nil, exception.NotFound("User not found")
	}

	return &model.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}
