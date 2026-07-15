package config

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDatabase(config *Config, log *logrus.Logger) *gorm.DB {
	dbUsername := config.Database.Username
	dbPassword := config.Database.Password
	dbHost := config.Database.Host
	dbPort := config.Database.Port
	dbName := config.Database.Name
	dbMaxIdle := config.Database.Pool.MaxIdle
	dbMaxOpen := config.Database.Pool.MaxOpen
	dbMaxLifetime := config.Database.Pool.MaxLifetime

	databaseUrl := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUsername, dbPassword, dbHost, dbPort, dbName)

	gormLogLevel := logger.Info
	if config.App.Environment == "production" {
		gormLogLevel = logger.Warn
	}

	db, err := gorm.Open(mysql.Open(databaseUrl), &gorm.Config{
		Logger: logger.New(&logrusWriter{Logger: log}, logger.Config{
			SlowThreshold:             time.Millisecond * 500,
			Colorful:                  true,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			LogLevel:                  gormLogLevel,
		}),
	})

	if err != nil {
		log.Fatalf("Failed to connect database %v", err)
	}

	connection, err := db.DB()

	if err != nil {
		log.Fatalf("Failed to get connection from database %v", err)
	}

	connection.SetMaxIdleConns(dbMaxIdle)
	connection.SetMaxOpenConns(dbMaxOpen)
	lifetime, err := time.ParseDuration(dbMaxLifetime)
	if err != nil {
		log.Fatalf("Failed to parse database maxLifetime %v", err)
	}
	connection.SetConnMaxLifetime(lifetime)

	return db
}

type logrusWriter struct {
	Logger *logrus.Logger
}

func (l *logrusWriter) Printf(message string, args ...interface{}) {
	l.Logger.Tracef(message, args...)
}
