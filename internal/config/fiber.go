package config

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/mkhsnw/golang-starter-kit/internal/exception"
	"github.com/mkhsnw/golang-starter-kit/internal/model"
	"gorm.io/gorm"
)

func NewFiber(config *Config, db *gorm.DB) *fiber.App {

	app := fiber.New(fiber.Config{
		AppName:      config.App.Name,
		ErrorHandler: NewErrorHandler(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
		BodyLimit:    5 * 1024 * 1024,
	})

	// Generate X-Request-ID (Correlation ID) untuk setiap request
	app.Use(requestid.New())

	// Logger diformat untuk menyertakan Request ID
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${ip} | ${locals:requestid} | ${status} - ${method} ${path}\n",
	}))

	app.Use(recover.New())
	app.Use(NewCORSConfig(config))
	app.Use(helmet.New())

	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
	}))

	// Health check yang benar-benar cek koneksi DB
	app.Get("/health", func(ctx fiber.Ctx) error {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "error",
				"detail": "database unreachable",
			})
		}
		return ctx.JSON(fiber.Map{"status": "ok"})
	})

	return app
}

func NewCORSConfig(config *Config) fiber.Handler {
	if config.App.Environment == "production" {
		return cors.New(cors.Config{
			AllowOrigins: []string{config.App.Url},
			AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
			AllowHeaders: []string{"Origin", "Content-type", "Accept", "Authorization"},
		})
	}
	return cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3000"},
	})
}

func NewErrorHandler() fiber.ErrorHandler {
	return func(ctx fiber.Ctx, err error) error {
		// Default internal server error
		code := fiber.StatusInternalServerError
		errCode := "INTERNAL_SERVER_ERROR"
		errMessage := err.Error()
		var errFields []model.FieldError

		// Handle Fiber native errors
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
			errCode = "FRAMEWORK_ERROR"
			errMessage = e.Message
		}

		// Handle custom ResponseError
		if e, ok := err.(*exception.ResponseError); ok {
			code = e.Code
			errCode = e.AppCode
			errMessage = e.Message
		}

		// Handle Validator errors
		if e, ok := err.(validator.ValidationErrors); ok {
			code = fiber.StatusBadRequest
			errCode = "VALIDATION_ERROR"
			errMessage = "Invalid input"
			for _, errField := range e {
				errFields = append(errFields, model.FieldError{
					Field:   errField.Field(),
					Message: fmt.Sprintf("failed on '%s' validation", errField.Tag()),
				})
			}
		}

		// Handle JSON Decode errors
		if _, ok := err.(*json.UnmarshalTypeError); ok {
			code = fiber.StatusBadRequest
			errCode = "BAD_REQUEST"
			errMessage = "invalid json format"
		}
		if _, ok := err.(*json.SyntaxError); ok {
			code = fiber.StatusBadRequest
			errCode = "BAD_REQUEST"
			errMessage = "invalid json syntax"
		}

		// Handle GORM Record Not Found
		if err == gorm.ErrRecordNotFound {
			code = fiber.StatusNotFound
			errCode = "NOT_FOUND"
			errMessage = "record not found"
		}

		return ctx.Status(code).JSON(model.WebResponse[any]{
			Error: &model.ErrorDetail{
				Code:    errCode,
				Message: errMessage,
				Fields:  errFields,
			},
		})
	}
}
