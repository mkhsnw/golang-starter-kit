package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mkhsnw/rel/docs"
	"github.com/mkhsnw/rel/internal/config"
)

// @title Golang Starter Kit API
// @version 1.0
// @description This is a sample API for Golang Starter Kit.
// @host localhost:3000
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	appConfig := config.NewConfig()
	log := config.NewLogrus(appConfig)
	db := config.NewDatabase(appConfig, log)
	validator := config.NewValidator(appConfig)
	redisStorage := config.NewRedisStorage(appConfig)
	app := config.NewHTTP(appConfig, db, log, redisStorage)

	config.Bootstrap(&config.BootstrapConfig{
		Config:    appConfig,
		Logger:    log,
		Database:  db,
		App:       app,
		Validator: validator,
	})
	go func() {
		port := appConfig.App.Port
		err := app.Listen(fmt.Sprintf(":%d", port))
		if err != nil {
			log.Fatalf("Failed to start server %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
	}

	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}

	if redisStorage != nil {
		_ = redisStorage.Close()
	}

	log.Info("Server exited gracefully")
}
