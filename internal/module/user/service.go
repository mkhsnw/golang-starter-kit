package user

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mkhsnw/rel/internal/foundation/exception"
	"github.com/mkhsnw/rel/internal/foundation/logger"
	"github.com/mkhsnw/rel/internal/module/user/dto"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

var (
	ErrEmailAlreadyExists    = exception.New(exception.EMAIL_ALREADY_EXISTS, "Email already exists")
	ErrRegistrationFailed    = exception.New(exception.INTERNAL_ERROR, "Registration failed")
	ErrInvalidCredentials    = exception.New(exception.UNAUTHORIZED, "Invalid email or password")
	ErrTokenGenerationFailed = exception.New(exception.INTERNAL_ERROR, "Failed to generate token")
	ErrInvalidRefreshToken   = exception.New(exception.UNAUTHORIZED, "Invalid or expired refresh token")
	ErrUserNotFound          = exception.New(exception.USER_NOT_FOUND, "User not found")
)

type UserService struct {
	Log                    *logrus.Logger
	JwtSecret              string
	JwtExpirationHours     int
	JwtRefreshSecret       string
	JwtRefreshExpDays      int
	UserRepository         *UserRepository
	RefreshTokenRepository *RefreshTokenRepository
}

func NewUserService(log *logrus.Logger, jwtSecret string, jwtExpHours int, jwtRefreshSecret string, jwtRefreshExpDays int, userRepo *UserRepository, refreshRepo *RefreshTokenRepository) *UserService {
	return &UserService{
		Log:                    log,
		JwtSecret:              jwtSecret,
		JwtExpirationHours:     jwtExpHours,
		JwtRefreshSecret:       jwtRefreshSecret,
		JwtRefreshExpDays:      jwtRefreshExpDays,
		UserRepository:         userRepo,
		RefreshTokenRepository: refreshRepo,
	}
}

func (u *UserService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.UserResponse, error) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Check if email already exists
	existingUser, err := u.UserRepository.FindByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.WithContext(ctx, u.Log).Errorf("failed to hash password: %v", err)
		return nil, ErrRegistrationFailed
	}

	id, _ := uuid.NewV7()
	user := &User{
		ID:       id.String(),
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := u.UserRepository.Create(ctx, user); err != nil {
		logger.WithContext(ctx, u.Log).Errorf("failed to create user: %v", err)
		return nil, ErrRegistrationFailed
	}

	return ToUserResponse(user), nil
}

func (u *UserService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.TokenResponse, error) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Find user by email
	user, err := u.UserRepository.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	tokenString, err := u.generateAccessToken(user.ID, user.Email)
	if err != nil {
		logger.WithContext(ctx, u.Log).Errorf("failed to sign jwt token: %v", err)
		return nil, ErrTokenGenerationFailed
	}

	refreshToken, err := u.generateAndStoreRefreshToken(ctx, user.ID)
	if err != nil {
		logger.WithContext(ctx, u.Log).Errorf("failed to generate refresh token: %v", err)
		return nil, ErrTokenGenerationFailed
	}

	return &dto.TokenResponse{
		Token:        tokenString,
		RefreshToken: refreshToken,
	}, nil
}

func (u *UserService) generateAccessToken(userID, email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    userID,
		"email": email,
		"exp":   time.Now().Add(time.Hour * time.Duration(u.JwtExpirationHours)).Unix(),
	})
	return token.SignedString([]byte(u.JwtSecret))
}

func (u *UserService) generateAndStoreRefreshToken(ctx context.Context, userID string) (string, error) {
	rawToken := uuid.NewString() + uuid.NewString()
	hash := sha256.Sum256([]byte(rawToken))
	hashHex := hex.EncodeToString(hash[:])

	id, _ := uuid.NewV7()
	rt := &RefreshToken{
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

func (u *UserService) RefreshToken(ctx context.Context, rawToken string) (*dto.TokenResponse, error) {
	hash := sha256.Sum256([]byte(rawToken))
	hashHex := hex.EncodeToString(hash[:])

	rt, err := u.RefreshTokenRepository.FindByTokenHash(ctx, hashHex)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	user, err := u.UserRepository.FindByID(ctx, rt.UserId)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, err := u.generateAccessToken(user.ID, user.Email)
	if err != nil {
		logger.WithContext(ctx, u.Log).Errorf("failed to generate access token: %v", err)
		return nil, ErrTokenGenerationFailed
	}

	return &dto.TokenResponse{
		Token:        accessToken,
		RefreshToken: rawToken,
	}, nil
}

func (u *UserService) Logout(ctx context.Context, userID string) error {
	return u.RefreshTokenRepository.RevokeAllForUser(ctx, userID)
}
func (u *UserService) GetCurrentUser(ctx context.Context, userID string) (*dto.UserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user, err := u.UserRepository.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return ToUserResponse(user), nil
}
