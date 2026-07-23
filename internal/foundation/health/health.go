package health

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/storage/redis/v3"
	"gorm.io/gorm"
)

type HealthHandler struct {
	DB    *gorm.DB
	Redis *redis.Storage
}

func NewHealthHandler(db *gorm.DB, rds *redis.Storage) *HealthHandler {
	return &HealthHandler{DB: db, Redis: rds}
}

// HealthCheck (Liveness probe) - returns HTTP 200 if the process is alive.
func (h *HealthHandler) HealthCheck(c fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":    "UP",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// ReadinessCheck (Readiness probe) - verifies DB & Redis connectivity.
func (h *HealthHandler) ReadinessCheck(c fiber.Ctx) error {
	dbStatus := "UP"
	if h.DB != nil {
		sqlDB, err := h.DB.DB()
		if err != nil || sqlDB.PingContext(c.Context()) != nil {
			dbStatus = "DOWN"
		}
	} else {
		dbStatus = "DOWN"
	}

	redisStatus := "N/A"
	if h.Redis != nil {
		if err := h.Redis.Set("health_ping", []byte("1"), 5*time.Second); err != nil {
			redisStatus = "DOWN"
		} else {
			redisStatus = "UP"
		}
	}

	isHealthy := dbStatus == "UP" && (redisStatus == "UP" || redisStatus == "N/A")
	statusCode := fiber.StatusOK
	if !isHealthy {
		statusCode = fiber.StatusServiceUnavailable
	}

	statusStr := "UP"
	if !isHealthy {
		statusStr = "DOWN"
	}

	return c.Status(statusCode).JSON(fiber.Map{
		"status": statusStr,
		"services": fiber.Map{
			"database": dbStatus,
			"redis":    redisStatus,
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// SetupRoutes registers health check endpoints.
func SetupRoutes(router fiber.Router, handler *HealthHandler) {
	router.Get("/health", handler.HealthCheck)
	router.Get("/ready", handler.ReadinessCheck)
	router.Get("/live", handler.HealthCheck)
}
