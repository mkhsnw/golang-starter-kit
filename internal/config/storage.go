package config

import (
	"github.com/gofiber/storage/redis/v3"
)

func NewRedisStorage(config *Config) *redis.Storage {
	return redis.New(redis.Config{
		Host:     config.Redis.Host,
		Port:     config.Redis.Port,
		Password: config.Redis.Password,
		Database: config.Redis.Database,
	})
}
