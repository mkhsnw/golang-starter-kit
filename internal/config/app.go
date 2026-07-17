package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/mkhsnw/golang-starter-kit/internal/delivery/http/controller"
	"github.com/mkhsnw/golang-starter-kit/internal/delivery/http/middleware"
	"github.com/mkhsnw/golang-starter-kit/internal/delivery/http/route"
	"github.com/mkhsnw/golang-starter-kit/internal/repository"
	"github.com/mkhsnw/golang-starter-kit/internal/usecase"
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
	txManager := repository.NewGormTransactionManager(config.Database)
	_ = txManager

	// Repositories
	userRepo := repository.NewUserRepository(config.Database)
	refreshTokenRepo := repository.NewRefreshTokenRepository(config.Database)
	// @InjectRepo

	// Usecases
	userUsecase := usecase.NewUserUsecase(config.Logger, config.Config.JWT.Secret, config.Config.JWT.ExpirationHours, config.Config.JWT.RefreshSecret, config.Config.JWT.RefreshExpirationDays, userRepo, refreshTokenRepo)
	// @InjectUsecase

	// Controllers
	userController := controller.NewUserController(userUsecase, config.Validator)
	// @InjectController

	// Middlewares
	authMiddleware := middleware.NewAuthMiddleware(config.Config.JWT.Secret)

	// Setup Routes
	routes := route.RouteConfig{
		App:            config.App,
		UserController: userController,
		AuthMiddleware: authMiddleware,
		// @InjectRouteConfig
	}
	routes.SetupRoutes()
}
