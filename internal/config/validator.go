package config

import (
	"github.com/go-playground/validator"
)

func NewValidator(config *Config) *validator.Validate {
	return validator.New()
}
