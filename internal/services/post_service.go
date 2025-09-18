package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"blog/internal/database"
	"blog/internal/models"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type PostService struct {
	db          *database.PostgresDB
	cache       *CacheService
	searchSvc   *SearchService
	activitySvc *ActivityService
}

func NewPostService(db *database.PostgresDB, cache *CacheService, searchSvc *SearchService, activitySvc *ActivityService) *PostService {
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
		Tags:      pq.StringArray(req.Tags),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert post
	query := `
		INSERT INTO posts (id, title, content, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = tx.ExecContext(ctx, query, post.ID, post.Title, post.Content, post.Tags, post.CreatedAt, post.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	// Log activity
	if err := s.activitySvc.LogActivity(ctx, tx, models.ActionCreatePost, post.ID); err != nil {
		return nil, fmt.Errorf("failed to log activity: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Index in Elasticsearch
	if err := s.searchSvc.IndexPost(ctx, post); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to index post in Elasticsearch: %v\n", err)
	}

	return post, nil
}

func (s *PostService) GetPost(ctx context.Context, id uuid.UUID) (*models.Post, error) {
	// Try cache first
	if post, err := s.cache.GetPost(ctx, id); err == nil && post != nil {
		return post, nil
	}

	// Query database
	query := `
		SELECT id, title, content, tags, created_at, updated_at
		FROM posts
		WHERE id = $1
	`
	var post models.Post
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID, &post.Title, &post.Content, &post.Tags, &post.CreatedAt, &post.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("post not found")
		}
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	// Cache the result
	if err := s.cache.SetPost(ctx, &post); err != nil {
		fmt.Printf("Failed to cache post: %v\n", err)
	}

	return &post, nil
}

func (s *PostService) UpdatePost(ctx context.Context, id uuid.UUID, req *models.PostUpdateRequest) (*models.Post, error) {
	// Get existing post
	post, err := s.GetPost(ctx, id)
	if err != nil {
		return nil, err
	}

	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update fields if provided
	if req.Title != nil {
		post.Title = *req.Title
	}
	if req.Content != nil {
		post.Content = *req.Content
	}
	if req.Tags != nil {
		post.Tags = pq.StringArray(req.Tags)
	}
	post.UpdatedAt = time.Now()

	// Update in database
	query := `
		UPDATE posts
		SET title = $2, content = $3, tags = $4, updated_at = $5
		WHERE id = $1
	`
	_, err = tx.ExecContext(ctx, query, post.ID, post.Title, post.Content, post.Tags, post.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	// Log activity
	if err := s.activitySvc.LogActivity(ctx, tx, models.ActionUpdatePost, post.ID); err != nil {
		return nil, fmt.Errorf("failed to log activity: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Invalidate cache
	if err := s.cache.DeletePost(ctx, id); err != nil {
		fmt.Printf("Failed to invalidate cache: %v\n", err)
	}

	// Update in Elasticsearch
	if err := s.searchSvc.IndexPost(ctx, post); err != nil {
		fmt.Printf("Failed to update post in Elasticsearch: %v\n", err)
	}

	return post, nil
}

func (s *PostService) DeletePost(ctx context.Context, id uuid.UUID) error {
	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete post
	query := `DELETE FROM posts WHERE id = $1`
	result, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("post not found")
	}

	// Log activity
	if err := s.activitySvc.LogActivity(ctx, tx, models.ActionDeletePost, id); err != nil {
		return fmt.Errorf("failed to log activity: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Remove from cache
	if err := s.cache.DeletePost(ctx, id); err != nil {
		fmt.Printf("Failed to remove from cache: %v\n", err)
	}

	// Remove from Elasticsearch
	if err := s.searchSvc.DeletePost(ctx, id); err != nil {
		fmt.Printf("Failed to remove from Elasticsearch: %v\n", err)
	}

	return nil
}

func (s *PostService) SearchByTag(ctx context.Context, tag string) ([]models.Post, error) {
	query := `
		SELECT id, title, content, tags, created_at, updated_at
		FROM posts
		WHERE $1 = ANY(tags)
		ORDER BY created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to search posts by tag: %w", err)
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Tags, &post.CreatedAt, &post.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, post)
	}

	return posts, nil
}
