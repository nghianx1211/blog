package database

import (
	"blog/internal/models"
	"fmt"
	"log"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		return fmt.Errorf("failed to create uuid-ossp extension: %w", err)
	}

	if err := db.AutoMigrate(
		&models.Post{},        
		&models.ActivityLog{},
	); err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	log.Println("✅ Database migration completed successfully")
	return nil
}
