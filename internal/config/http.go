package config

import (
	"encoding/json"
	"strings"
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
	"github.com/gofiber/storage/redis/v3"
	"github.com/mkhsnw/golang-starter-kit/internal/foundation/exception"
	"github.com/mkhsnw/golang-starter-kit/internal/foundation/mapper"
	"github.com/mkhsnw/golang-starter-kit/internal/foundation/response"
	"github.com/mkhsnw/golang-starter-kit/internal/middleware"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func NewHTTP(config *Config, db *gorm.DB, log *logrus.Logger, redisStorage *redis.Storage) *fiber.App {
	sensitiveBodyPaths := []string{"/api/v1/auth/login", "/api/v1/auth/register"}

	app := fiber.New(fiber.Config{
		AppName:      config.App.Name,
		ErrorHandler: newErrorHandler(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
		BodyLimit:    5 * 1024 * 1024,
	})

	// Generate X-Request-ID (Correlation ID) untuk setiap request
	app.Use(requestid.New())

	// Propagate request ID into standard Go context.Context
	app.Use(middleware.RequestContextMiddleware())

	// Logger diformat untuk menyertakan Request ID (Format JSON)
	app.Use(logger.New(logger.Config{
		Format: `{"time":"${time}","ip":"${ip}","requestid":"${locals:requestid}","status":${status},"method":"${method}","path":"${path}","latency":"${latency}"}` + "\n",
	}))

	// Middleware log raw body jika terjadi internal server error (5xx)
	app.Use(func(ctx fiber.Ctx) error {
		err := ctx.Next()
		if ctx.Response().StatusCode() >= fiber.StatusInternalServerError {
			isSensitive := false
			for _, p := range sensitiveBodyPaths {
				if ctx.Path() == p {
					isSensitive = true
					break
				}
			}
			if isSensitive {
				log.Errorf("[5xx Error] Method: %s | Path: %s | Body: [REDACTED - sensitive endpoint]", ctx.Method(), ctx.Path())
			} else if body := ctx.Body(); len(body) > 0 {
				log.Errorf("[5xx Error Details] Method: %s | Path: %s | Body: %s", ctx.Method(), ctx.Path(), string(body))
			}
		}
		return err
	})

	app.Use(recover.New())
	app.Use(newCORSConfig(config))
	app.Use(helmet.New())

	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		Storage:    redisStorage,
		Next: func(ctx fiber.Ctx) bool {
			// Skip limiter for swagger docs
			return strings.HasPrefix(ctx.Path(), "/api/v1/docs")
		},
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

func newCORSConfig(cfg *Config) fiber.Handler {
	if cfg.App.Environment == "production" {
		return cors.New(cors.Config{
			AllowOrigins: []string{cfg.App.Url},
			AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
			AllowHeaders: []string{"Origin", "Content-type", "Accept", "Authorization"},
		})
	}
	return cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3000"},
	})
}

func newErrorHandler() fiber.ErrorHandler {
	return func(ctx fiber.Ctx, err error) error {
		// Default internal server error
		apiErr := exception.New(exception.INTERNAL_ERROR, err.Error())

		// Handle Fiber native errors
		if e, ok := err.(*fiber.Error); ok {
			apiErr = exception.New(exception.INTERNAL_ERROR, e.Message)
			apiErr.Status = e.Code
		}

		// Handle custom APIError
		if e, ok := err.(*exception.APIError); ok {
			apiErr = e
		}

		// Handle Validator errors
		if e, ok := err.(validator.ValidationErrors); ok {
			errFields := mapper.MapValidation(e)
			apiErr = exception.New(exception.VALIDATION_ERROR, "Invalid input", errFields)
		}

		// Handle JSON Decode errors
		if _, ok := err.(*json.UnmarshalTypeError); ok {
			apiErr = exception.New(exception.VALIDATION_ERROR, "invalid json format")
		}
		if _, ok := err.(*json.SyntaxError); ok {
			apiErr = exception.New(exception.VALIDATION_ERROR, "invalid json syntax")
		}

		// Handle GORM Record Not Found
		if err == gorm.ErrRecordNotFound {
			apiErr = exception.New(exception.DATABASE_ERROR, "record not found")
			apiErr.Status = fiber.StatusNotFound
		}

		return response.Error(apiErr).Send(ctx)
	}
}
