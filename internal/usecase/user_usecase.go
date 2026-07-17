package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"github.com/mkhsnw/golang-starter-kit/internal/entity"
	"github.com/mkhsnw/golang-starter-kit/internal/exception"
	"github.com/mkhsnw/golang-starter-kit/internal/model"
	"github.com/mkhsnw/golang-starter-kit/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecase struct {
	Log                    *logrus.Logger
	JwtSecret              string
	JwtExpirationHours     int
	JwtRefreshSecret       string
	JwtRefreshExpDays      int
	UserRepository         repository.UserRepositoryInterface
	RefreshTokenRepository repository.RefreshTokenRepositoryInterface
}

func NewUserUsecase(log *logrus.Logger, jwtSecret string, jwtExpHours int, jwtRefreshSecret string, jwtRefreshExpDays int, userRepo repository.UserRepositoryInterface, refreshRepo repository.RefreshTokenRepositoryInterface) *UserUsecase {
	return &UserUsecase{
		Log:                    log,
		JwtSecret:              jwtSecret,
		JwtExpirationHours:     jwtExpHours,
		JwtRefreshSecret:       jwtRefreshSecret,
		JwtRefreshExpDays:      jwtRefreshExpDays,
		UserRepository:         userRepo,
		RefreshTokenRepository: refreshRepo,
	}
}

func (u *UserUsecase) Register(ctx context.Context, req *model.RegisterRequest) (*model.UserResponse, error) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Check if email already exists
	existingUser, err := u.UserRepository.FindByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, exception.Conflict("Email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		u.Log.Errorf("failed to hash password: %v", err)
		return nil, exception.NewResponseError(500, "INTERNAL_SERVER_ERROR", "Registration failed")
	}

	id, _ := uuid.NewV7()
	user := &entity.User{
		ID:       id.String(),
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := u.UserRepository.Create(ctx, user); err != nil {
		u.Log.Errorf("failed to create user: %v", err)
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
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Find user by email
	user, err := u.UserRepository.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, exception.Unauthorized("Invalid email or password")
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, exception.Unauthorized("Invalid email or password")
	}

	tokenString, err := u.generateAccessToken(user.ID, user.Email)
	if err != nil {
		u.Log.Errorf("failed to sign jwt token: %v", err)
		return nil, exception.NewResponseError(500, "INTERNAL_SERVER_ERROR", "Failed to generate token")
	}

	refreshToken, err := u.generateAndStoreRefreshToken(ctx, user.ID)
	if err != nil {
		u.Log.Errorf("failed to generate refresh token: %v", err)
		return nil, exception.NewResponseError(500, "INTERNAL_SERVER_ERROR", "Failed to generate refresh token")
	}

	return &model.TokenResponse{
		Token:        tokenString,
		RefreshToken: refreshToken,
	}, nil
}

func (u *UserUsecase) generateAccessToken(userID, email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    userID,
		"email": email,
		"exp":   time.Now().Add(time.Hour * time.Duration(u.JwtExpirationHours)).Unix(),
	})
	return token.SignedString([]byte(u.JwtSecret))
}

func (u *UserUsecase) generateAndStoreRefreshToken(ctx context.Context, userID string) (string, error) {
	rawToken := uuid.NewString() + uuid.NewString()
	hash := sha256.Sum256([]byte(rawToken))
	hashHex := hex.EncodeToString(hash[:])

	id, _ := uuid.NewV7()
	rt := &entity.RefreshToken{
		ID:        id.String(),
		UserId:    userID,
		TokenHash: hashHex,
		ExpiresAt: time.Now().AddDate(0, 0, u.JwtRefreshExpDays),
	}
	if err := u.RefreshTokenRepository.Create(ctx, rt); err != nil {
		return "", err
	}
	return rawToken, nil
}

func (u *UserUsecase) RefreshToken(ctx context.Context, rawToken string) (*model.TokenResponse, error) {
	hash := sha256.Sum256([]byte(rawToken))
	hashHex := hex.EncodeToString(hash[:])

	rt, err := u.RefreshTokenRepository.FindByTokenHash(ctx, hashHex)
	if err != nil {
		return nil, exception.Unauthorized("Invalid or expired refresh token")
	}

	user, err := u.UserRepository.FindByID(ctx, rt.UserId)
	if err != nil {
		return nil, exception.Unauthorized("User not found")
	}

	accessToken, err := u.generateAccessToken(user.ID, user.Email)
	if err != nil {
		u.Log.Errorf("failed to generate access token: %v", err)
		return nil, exception.NewResponseError(500, "INTERNAL_SERVER_ERROR", "failed to generate token")
	}

	return &model.TokenResponse{
		Token:        accessToken,
		RefreshToken: rawToken,
	}, nil
}

func (u *UserUsecase) Logout(ctx context.Context, userID string) error {
	return u.RefreshTokenRepository.RevokeAllForUser(ctx, userID)
}
func (u *UserUsecase) GetCurrentUser(ctx context.Context, userID string) (*model.UserResponse, error) {
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
