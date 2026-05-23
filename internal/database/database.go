package database

import (
	"fmt"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"tz-app/internal/config"
	"tz-app/internal/logger"
	"tz-app/internal/models"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode)

	logger.Info("Connecting to database",
		zap.String("host", cfg.DBHost),
		zap.String("port", cfg.DBPort),
		zap.String("database", cfg.DBName),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Error("Failed to connect to database", zap.Error(err))
		return nil, err
	}

	logger.Info("Database connected successfully")
	return db, nil
}

// AutoMigrate выполняет автоматическую миграцию (для разработки)
func AutoMigrate(db *gorm.DB) error {
	logger.Info("Running auto migration...")

	err := db.AutoMigrate(&models.Department{}, &models.Employee{})
	if err != nil {
		logger.Error("Auto migration failed", zap.Error(err))
		return err
	}

	logger.Info("Auto migration completed")
	return nil
}
