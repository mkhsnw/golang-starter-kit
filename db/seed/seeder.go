package seed

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"gorm.io/gorm"
)

// Seeder defines the interface for database seeders
type Seeder interface {
	Name() string
	Seed(db *gorm.DB) error
}

// Registry manages and executes registered seeders
type Registry struct {
	seeders []Seeder
}

func NewRegistry() *Registry {
	return &Registry{
		seeders: make([]Seeder, 0),
	}
}

func (r *Registry) Register(s Seeder) {
	r.seeders = append(r.seeders, s)
}

func (r *Registry) RunAll(db *gorm.DB) error {
	for _, s := range r.seeders {
		log.Printf("🌱 Running Seeder: %s...\n", s.Name())
		if err := s.Seed(db); err != nil {
			return fmt.Errorf("failed running seeder %s: %w", s.Name(), err)
		}
		log.Printf("✅ Completed Seeder: %s\n", s.Name())
	}
	return nil
}

// GetSeedCount checks OS Environment "SEED_COUNT" or returns fallback default count
func GetSeedCount(defaultCount int) int {
	if countStr := os.Getenv("SEED_COUNT"); countStr != "" {
		if c, err := strconv.Atoi(countStr); err == nil && c > 0 {
			return c
		}
	}
	return defaultCount
}

// TruncateTables truncates specified tables with foreign key checks temporarily disabled
func TruncateTables(db *gorm.DB, tables ...string) error {
	if err := db.Exec("SET FOREIGN_KEY_CHECKS = 0").Error; err != nil {
		return fmt.Errorf("failed to disable foreign key checks: %w", err)
	}

	defer func() {
		_ = db.Exec("SET FOREIGN_KEY_CHECKS = 1")
	}()

	for _, table := range tables {
		log.Printf("🧹 Truncating table: %s...\n", table)
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE `%s`", table)).Error; err != nil {
			// Fallback to DELETE FROM if TRUNCATE fails (e.g. SQLite or specific constraints)
			if errDel := db.Exec(fmt.Sprintf("DELETE FROM `%s`", table)).Error; errDel != nil {
				return fmt.Errorf("failed to truncate/delete table %s: %w", table, errDel)
			}
		}
	}

	return nil
}

// Execute is the main entry point to run all application seeders
func Execute(db *gorm.DB) error {
	registry := NewRegistry()

	// Register all seeders in order
	registry.Register(NewUserSeeder())
	// @InjectSeeder

	log.Println("===========================================")
	log.Println("🚀 STARTING DATABASE SEEDING PROCESS")
	log.Println("===========================================")

	if err := registry.RunAll(db); err != nil {
		return err
	}

	log.Println("===========================================")
	log.Println("🎉 DATABASE SEEDING COMPLETED SUCCESSFULLY")
	log.Println("===========================================")
	return nil
}
