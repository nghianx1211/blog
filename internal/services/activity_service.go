package services

import (
	"context"
	"database/sql"
	"fmt"

	"blog/internal/models"

	"github.com/google/uuid"
)

type ActivityService struct{}

func NewActivityService() *ActivityService {
	return &ActivityService{}
}

func (s *ActivityService) LogActivity(ctx context.Context, tx *sql.Tx, action string, postID uuid.UUID) error {
	log := models.NewActivityLog(action, postID)
	
	query := `
		INSERT INTO activity_logs (id, action, post_id, logged_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := tx.ExecContext(ctx, query, log.ID, log.Action, log.PostID, log.LoggedAt)
	if err != nil {
		return fmt.Errorf("failed to log activity: %w", err)
	}

	return nil
}