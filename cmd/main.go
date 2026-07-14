package main

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/mkhsnw/golang-starter-kit/internal/config"
)

func main() {
	appConfig := config.NewConfig()
	log := config.NewLogrus(appConfig)
	db := config.NewDatabase(appConfig, log)
	validator := config.NewValidator(appConfig)
	app := config.NewFiber(appConfig)

	config.Bootstrap(&config.BootstrapConfig{
		Config:    appConfig,
		Logger:    log,
		Database:  db,
		App:       app,
		Validator: validator,
	})
	port := appConfig.App.Port
	err := app.Listen(fmt.Sprintf(":%d", port), fiber.ListenConfig{
		EnablePrefork: true,
	})
	if err != nil {
		log.Fatalf("Failed to start server %v", err)
	}

}
