package config

import (
	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v3"
	"github.com/mkhsnw/golang-starter-kit/internal/delivery/http/route"
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
	routes := route.RouteConfig{
		App: config.App,
	}
	routes.SetupRoutes()
}
