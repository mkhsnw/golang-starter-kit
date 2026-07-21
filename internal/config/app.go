package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"

	"github.com/mkhsnw/golang-starter-kit/internal/middleware"

	"github.com/mkhsnw/golang-starter-kit/internal/foundation/database"
	"github.com/mkhsnw/golang-starter-kit/internal/module/user"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type BootstrapConfig struct {
	Config    *Config
	Logger    *logrus.Logger
	Database  *gorm.DB
	App       *fiber.App
	Validator *validator.Validate
}

func Bootstrap(config *BootstrapConfig) {
	// Transaction Manager
	txManager := database.NewGormTransactionManager(config.Database)
	_ = txManager

	// Repositories
	userRepo := user.NewUserRepository(config.Database)
	refreshTokenRepo := user.NewRefreshTokenRepository(config.Database)
	// @InjectRepo

	// Usecases
	userUsecase := user.NewUserService(config.Logger, config.Config.JWT.Secret, config.Config.JWT.ExpirationHours, config.Config.JWT.RefreshSecret, config.Config.JWT.RefreshExpirationDays, userRepo, refreshTokenRepo)
	// @InjectUsecase

	// Controllers
	userController := user.NewUserController(userUsecase, config.Validator)
	// @InjectController

	// Middlewares
	authMiddleware := middleware.NewAuthMiddleware(config.Config.JWT.Secret)

	// Setup Routes
	api := config.App.Group("/api/v1")
	user.SetupRoutes(api, userController, authMiddleware)
	// @InjectRouteConfig
}
