package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Post struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()" db:"id"`
	Title     string         `json:"title" gorm:"type:varchar(255);not null" db:"title"`
	Content   string         `json:"content" gorm:"type:text;not null" db:"content"`
	Tags      pq.StringArray `json:"tags" gorm:"type:text[];default:'{}'" db:"tags"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime" db:"updated_at"`

	// Relationships
	ActivityLogs []ActivityLog `json:"activity_logs,omitempty" gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for GORM
func (Post) TableName() string {
	return "posts"
}

// GORM hooks for custom indexes
func (p *Post) AfterAutoMigrate(tx *gorm.DB) error {
	// Create GIN index for tags array
	if err := tx.Exec("CREATE INDEX IF NOT EXISTS idx_posts_tags ON posts USING GIN(tags)").Error; err != nil {
		return err
	}
	
	// Create index on created_at for ordering
	if err := tx.Exec("CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at DESC)").Error; err != nil {
		return err
	}
	
	return nil
}

type PostSearchRequest struct {
	Query string `json:"query" form:"q"`
	Tags  string `json:"tags" form:"tags"`
	Limit int    `json:"limit" form:"limit"`
	Page  int    `json:"page" form:"page"`
}

type PostCreateRequest struct {
	Title   string   `json:"title" binding:"required"`
	Content string   `json:"content" binding:"required"`
	Tags    []string `json:"tags"`
}

type PostUpdateRequest struct {
	Title   *string  `json:"title"`
	Content *string  `json:"content"`
	Tags    []string `json:"tags"`
}

type PostResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PostSearchResponse struct {
	Posts      []PostResponse `json:"posts"`
	TotalCount int64          `json:"total_count"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
}

// ToResponse converts Post model to PostResponse
func (p *Post) ToResponse() PostResponse {
	return PostResponse{
		ID:        p.ID.String(),
		Title:     p.Title,
		Content:   p.Content,
		Tags:      []string(p.Tags),
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

// ToElasticsearchDoc converts Post to Elasticsearch document
func (p *Post) ToElasticsearchDoc() map[string]interface{} {
	return map[string]interface{}{
		"id":         p.ID.String(),
		"title":      p.Title,
		"content":    p.Content,
		"tags":       []string(p.Tags),
		"created_at": p.CreatedAt,
		"updated_at": p.UpdatedAt,
	}
}

// ToJSON converts Post to JSON string
func (p *Post) ToJSON() (string, error) {
	data, err := json.Marshal(p)
	return string(data), err
}
