package services

import (
	"context"
	"fmt"

	"blog/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ActivityService struct{}

func NewActivityService() *ActivityService {
	return &ActivityService{}
}

func (s *ActivityService) LogActivity(ctx context.Context, tx *gorm.DB, action string, postID uuid.UUID) error {
	log := models.NewActivityLog(action, postID)

	if err := tx.WithContext(ctx).Create(&log).Error; err != nil {
		return fmt.Errorf("failed to log activity: %w", err)
	}

	return nil
}
