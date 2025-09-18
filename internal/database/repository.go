package database

import (
	"blog/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PostRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Create(post *models.Post) error {
	return r.db.Create(post).Error
}

func (r *PostRepository) GetByID(id uuid.UUID) (*models.Post, error) {
	var post models.Post
	err := r.db.Preload("ActivityLogs").First(&post, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *PostRepository) Update(post *models.Post) error {
	return r.db.Save(post).Error
}

func (r *PostRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Post{}, "id = ?", id).Error
}

func (r *PostRepository) Search(req models.PostSearchRequest) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	query := r.db.Model(&models.Post{})

	// Add search conditions
	if req.Query != "" {
		query = query.Where("title ILIKE ? OR content ILIKE ?", "%"+req.Query+"%", "%"+req.Query+"%")
	}

	if req.Tags != "" {
		query = query.Where("? = ANY(tags)", req.Tags)
	}

	// Count total
	query.Count(&total)

	// Apply pagination
	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}
	if req.Page > 0 {
		query = query.Offset((req.Page - 1) * req.Limit)
	}

	// Order by created_at DESC
	query = query.Order("created_at DESC")

	err := query.Find(&posts).Error
	return posts, total, err
}

type ActivityLogRepository struct {
	db *gorm.DB
}

func NewActivityLogRepository(db *gorm.DB) *ActivityLogRepository {
	return &ActivityLogRepository{db: db}
}

func (r *ActivityLogRepository) Create(log *models.ActivityLog) error {
	return r.db.Create(log).Error
}

func (r *ActivityLogRepository) GetByPostID(postID uuid.UUID) ([]models.ActivityLog, error) {
	var logs []models.ActivityLog
	err := r.db.Where("post_id = ?", postID).Order("logged_at DESC").Find(&logs).Error
	return logs, err
}