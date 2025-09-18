package services

import (
	"context"
	"fmt"
	"time"

	"blog/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PostService struct {
	db          *gorm.DB
	cache       *CacheService
	searchSvc   *SearchService
	activitySvc *ActivityService
}

func NewPostService(db *gorm.DB, cache *CacheService, searchSvc *SearchService, activitySvc *ActivityService) *PostService {
	return &PostService{
		db:          db,
		cache:       cache,
		searchSvc:   searchSvc,
		activitySvc: activitySvc,
	}
}

func (s *PostService) CreatePost(ctx context.Context, req *models.PostCreateRequest) (*models.Post, error) {
	post := &models.Post{
		ID:        uuid.New(),
		Title:     req.Title,
		Content:   req.Content,
		Tags:      req.Tags,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(post).Error; err != nil {
			return fmt.Errorf("failed to create post: %w", err)
		}

		if err := s.activitySvc.LogActivity(ctx, tx, models.ActionCreatePost, post.ID); err != nil {
			return fmt.Errorf("failed to log activity: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if err := s.searchSvc.IndexPost(ctx, post); err != nil {
		fmt.Printf("[WARN] Failed to index post in Elasticsearch: %v\n", err)
	}

	return post, nil
}

func (s *PostService) GetPost(ctx context.Context, id uuid.UUID) (*models.Post, error) {
	if post, err := s.cache.GetPost(ctx, id); err == nil && post != nil {
		return post, nil
	}

	var post models.Post
	if err := s.db.WithContext(ctx).First(&post, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("post not found")
		}
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	if err := s.cache.SetPost(ctx, &post); err != nil {
		fmt.Printf("Failed to cache post: %v\n", err)
	}

	return &post, nil
}

func (s *PostService) UpdatePost(ctx context.Context, id uuid.UUID, req *models.PostUpdateRequest) (*models.Post, error) {
	var post models.Post
	if err := s.db.WithContext(ctx).First(&post, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("post not found")
		}
		return nil, err
	}

	if req.Title != nil {
		post.Title = *req.Title
	}
	if req.Content != nil {
		post.Content = *req.Content
	}
	if req.Tags != nil {
		post.Tags = req.Tags
	}
	post.UpdatedAt = time.Now()

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&post).Error; err != nil {
			return fmt.Errorf("failed to update post: %w", err)
		}
		if err := s.activitySvc.LogActivity(ctx, tx, models.ActionUpdatePost, post.ID); err != nil {
			return fmt.Errorf("failed to log activity: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if err := s.cache.DeletePost(ctx, id); err != nil {
		fmt.Printf("Failed to invalidate cache: %v\n", err)
	}

	if err := s.searchSvc.IndexPost(ctx, &post); err != nil {
		fmt.Printf("Failed to update Elasticsearch: %v\n", err)
	}

	return &post, nil
}

func (s *PostService) DeletePost(ctx context.Context, id uuid.UUID) error {
    err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        if err := s.activitySvc.LogActivity(ctx, tx, models.ActionDeletePost, id); err != nil {
            return fmt.Errorf("failed to log activity: %w", err)
        }

        if err := tx.Delete(&models.Post{}, "id = ?", id).Error; err != nil {
            return fmt.Errorf("failed to delete post: %w", err)
        }

        return nil
    })
    if err != nil {
        return err
    }

    if err := s.cache.DeletePost(ctx, id); err != nil {
        fmt.Printf("Failed to remove from cache: %v\n", err)
    }
    if err := s.searchSvc.DeletePost(ctx, id); err != nil {
        fmt.Printf("Failed to remove from Elasticsearch: %v\n", err)
    }

    return nil
}

func (s *PostService) SearchByTag(ctx context.Context, tag string) ([]models.Post, error) {
	var posts []models.Post
	if err := s.db.WithContext(ctx).
		Where("? = ANY(tags)", tag).
		Order("created_at DESC").
		Find(&posts).Error; err != nil {
		return nil, fmt.Errorf("failed to search posts by tag: %w", err)
	}
	return posts, nil
}
