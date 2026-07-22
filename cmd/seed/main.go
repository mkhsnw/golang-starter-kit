package main

import (
	"log"
	"os"
	"strings"

	"github.com/mkhsnw/golang-starter-kit/db/seed"
	"github.com/mkhsnw/golang-starter-kit/internal/config"
)

func main() {
	// Parse CLI count flag if provided (e.g. task seed count=50 or --count 50)
	for i, arg := range os.Args {
		if strings.HasPrefix(arg, "count=") {
			os.Setenv("SEED_COUNT", strings.TrimPrefix(arg, "count="))
		} else if (arg == "--count" || arg == "-c") && i+1 < len(os.Args) {
			os.Setenv("SEED_COUNT", os.Args[i+1])
		}
	}

	appConfig := config.NewConfig()
	logger := config.NewLogrus(appConfig)
	db := config.NewDatabase(appConfig, logger)

	sqlDB, err := db.DB()
	if err == nil && sqlDB != nil {
		defer sqlDB.Close()
	}

	if err := seed.Execute(db); err != nil {
		log.Fatalf("❌ Seeding failed: %v", err)
	}

	log.Println("✨ Database seeding finished successfully.")
}
