package database

import (
	"fmt"
	"log"
	"time"

	"blog/internal/config"
	"blog/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type GormDB struct {
	*gorm.DB
}

func NewGormDB(cfg *config.DatabaseConfig) (*GormDB, error) {
	gormConfig := &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Get underlying sql.DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return &GormDB{db}, nil
}

// Migrate runs auto migration for all models
func (db *GormDB) Migrate() error {
	// Enable UUID extension first
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		return fmt.Errorf("failed to create uuid-ossp extension: %w", err)
	}

	// Auto migrate tables in correct order (dependencies first)
	if err := db.AutoMigrate(
		&models.Post{},        // Must be first (referenced by ActivityLog)
		&models.ActivityLog{}, // Must be after Post
	); err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	log.Println("Database migration completed successfully")
	return nil
}

// Close closes the database connection
func (db *GormDB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}