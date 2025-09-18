package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ActivityLog struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()" db:"id"`
	Action   string    `json:"action" gorm:"type:varchar(50);not null" db:"action"`
	PostID   uuid.UUID `json:"post_id" gorm:"type:uuid;not null;index" db:"post_id"`
	LoggedAt time.Time `json:"logged_at" gorm:"autoCreateTime;index:,sort:desc" db:"logged_at"`

	Post Post `json:"post,omitempty" gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE"`
}

func (ActivityLog) TableName() string {
	return "activity_logs"
}

func (a *ActivityLog) AfterAutoMigrate(tx *gorm.DB) error {
	if err := tx.Exec("CREATE INDEX IF NOT EXISTS idx_activity_logs_post_id ON activity_logs(post_id)").Error; err != nil {
		return err
	}
	
	if err := tx.Exec("CREATE INDEX IF NOT EXISTS idx_activity_logs_logged_at ON activity_logs(logged_at DESC)").Error; err != nil {
		return err
	}
	
	return nil
}

const (
	ActionCreatePost = "new_post"
	ActionUpdatePost = "update_post"
	ActionDeletePost = "delete_post"
	ActionViewPost   = "view_post"
)

func NewActivityLog(action string, postID uuid.UUID) *ActivityLog {
	return &ActivityLog{
		ID:       uuid.New(),
		Action:   action,
		PostID:   postID,
		LoggedAt: time.Now(),
	}
}
