package config

import (
	"encoding/json"
	"time"

	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/mkhsnw/golang-starter-kit/internal/exception"
	"github.com/mkhsnw/golang-starter-kit/internal/model"
	"gorm.io/gorm"
)

func NewFiber(config *Config) *fiber.App {

	app := fiber.New(fiber.Config{
		AppName:      config.App.Name,
		ErrorHandler: NewErrorHandler(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
		BodyLimit:    5 * 1024 * 1024,
	})

	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(helmet.New())
	
	// Performance Middlewares
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))
	
	// Security Middlewares
	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
	}))

	return app
}

func NewErrorHandler() fiber.ErrorHandler {
	return func(ctx fiber.Ctx, err error) error {
		// Default internal server error
		code := fiber.StatusInternalServerError
		var errs string = err.Error()

		// Handle Fiber native errors
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
			errs = e.Message
		}

		// Handle custom ResponseError
		if e, ok := err.(*exception.ResponseError); ok {
			code = e.Code
			errs = e.Message
		}

		// Handle Validator errors
		if e, ok := err.(validator.ValidationErrors); ok {
			code = fiber.StatusBadRequest
			errs = e.Error()
		}

		// Handle JSON Decode errors
		if _, ok := err.(*json.UnmarshalTypeError); ok {
			code = fiber.StatusBadRequest
			errs = "invalid json format"
		}
		if _, ok := err.(*json.SyntaxError); ok {
			code = fiber.StatusBadRequest
			errs = "invalid json syntax"
		}

		// Handle GORM Record Not Found
		if err == gorm.ErrRecordNotFound {
			code = fiber.StatusNotFound
			errs = "record not found"
		}

		return ctx.Status(code).JSON(model.WebResponse[any]{
			Errors: errs,
		})
	}
}
