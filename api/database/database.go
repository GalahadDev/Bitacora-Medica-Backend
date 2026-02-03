package database

import (
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(dbUrl string) {
	var err error

	// Configuración de GORM
	gormConfig := &gorm.Config{
		Logger:      logger.Default.LogMode(logger.Info),
		PrepareStmt: true,
	}

	DB, err = gorm.Open(postgres.Open(dbUrl), gormConfig)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		panic("Failed to connect to database")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		slog.Error("Failed to get generic database object", "error", err)
		panic("Failed to configure connection pool")
	}

	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	slog.Info("Database connection established successfully with pool limits")
}

func GetDB() *gorm.DB {
	return DB
}
