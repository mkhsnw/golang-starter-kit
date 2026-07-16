package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/mkhsnw/golang-starter-kit/internal/config"
)

func main() {
	appConfig := config.NewConfig()

	dsn := fmt.Sprintf("mysql://%s:%s@tcp(%s:%d)/%s",
		appConfig.Database.Username,
		appConfig.Database.Password,
		appConfig.Database.Host,
		appConfig.Database.Port,
		appConfig.Database.Name,
	)

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}
	migrationPath := filepath.Join(cwd, "db", "migration")
	sourceURL := "file://" + filepath.ToSlash(migrationPath)

	m, err := migrate.New(
		sourceURL,
		dsn,
	)
	if err != nil {
		log.Fatalf("Migration failed to initialize: %v", err)
	}

	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go [up|down|version]")
	}

	cmd := os.Args[1]
	switch cmd {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to run migrate up: %v", err)
		}
		log.Println("Migrate up success")
	case "down":
		if err := m.Steps(-1); err != nil {
			log.Fatalf("Failed to run migrate down: %v", err)
		}
		log.Println("Migrate down success")
	case "version":
		version, dirty, err := m.Version()
		if err != nil && err != migrate.ErrNilVersion {
			log.Fatalf("Failed to get migration version: %v", err)
		}
		log.Printf("Current migration version: %v, dirty: %v\n", version, dirty)
	default:
		log.Fatalf("Unknown command: %s", cmd)
	}
}
