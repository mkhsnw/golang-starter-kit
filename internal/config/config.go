package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	App      AppConfig      `mapstructure:"app" validate:"required"`
	Database DatabaseConfig `mapstructure:"database" validate:"required"`
	Log      LogConfig      `mapstructure:"log"`
	JWT      JwtConfig      `mapstructure:"jwt" validate:"required"`
	Redis    RedisConfig    `mapstructure:"redis"`
}

type AppConfig struct {
	Name        string `mapstructure:"name" validate:"required"`
	Environment string `mapstructure:"environment" validate:"required"`
	Port        int    `mapstructure:"port" validate:"required,gt=0"`
	Url         string `mapstructure:"url"`
}

type DatabasePoolConfig struct {
	MaxIdle     int    `mapstructure:"maxIdle"`
	MaxOpen     int    `mapstructure:"maxOpen"`
	MaxLifetime string `mapstructure:"maxLifetime"`
}

type DatabaseConfig struct {
	Port     int                `mapstructure:"port" validate:"required,gt=0"`
	Username string             `mapstructure:"username" validate:"required"`
	Password string             `mapstructure:"password"`
	Host     string             `mapstructure:"host" validate:"required"`
	Name     string             `mapstructure:"name" validate:"required"`
	Pool     DatabasePoolConfig `mapstructure:"pool"`
}

type LogConfig struct {
	Level int `mapstructure:"level"`
}

type JwtConfig struct {
	Secret                string `mapstructure:"secret" validate:"required,min=16"`
	ExpirationHours       int    `mapstructure:"expiration_hours"`
	RefreshSecret         string `mapstructure:"refresh_secret"`
	RefreshExpirationDays int    `mapstructure:"refresh_expiration_days"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	Database int    `mapstructure:"database"`
}

// Validate performs fail-fast validation on configuration fields.
func (c *Config) Validate() error {
	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		if valErrs, ok := err.(validator.ValidationErrors); ok {
			var missing []string
			for _, ve := range valErrs {
				missing = append(missing, fmt.Sprintf("%s (rule: %s)", ve.Namespace(), ve.Tag()))
			}
			return fmt.Errorf("CONFIG VALIDATION ERROR - Invalid or missing fields:\n  - %s", strings.Join(missing, "\n  - "))
		}
		return err
	}
	return nil
}
