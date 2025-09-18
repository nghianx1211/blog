package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"blog/internal/database"
	"blog/internal/models"

	"github.com/google/uuid"
)

type CacheService struct {
	redis *database.RedisClient
}

func NewCacheService(redis *database.RedisClient) *CacheService {
	return &CacheService{redis: redis}
}

const (
	postCacheKeyPrefix = "post:"
	postCacheTTL       = 5 * time.Minute
)

func (s *CacheService) GetPost(ctx context.Context, id uuid.UUID) (*models.Post, error) {
	key := postCacheKeyPrefix + id.String()

	result, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var post models.Post
	if err := json.Unmarshal([]byte(result), &post); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached post: %w", err)
	}

	return &post, nil
}

func (s *CacheService) SetPost(ctx context.Context, post *models.Post) error {
	key := postCacheKeyPrefix + post.ID.String()

	data, err := json.Marshal(post)
	if err != nil {
		return fmt.Errorf("failed to marshal post: %w", err)
	}

	return s.redis.Set(ctx, key, data, postCacheTTL).Err()
}

func (s *CacheService) DeletePost(ctx context.Context, id uuid.UUID) error {
	key := postCacheKeyPrefix + id.String()
	return s.redis.Del(ctx, key).Err()
}

func (s *CacheService) InvalidatePostCache(ctx context.Context, id uuid.UUID) error {
	key := postCacheKeyPrefix + id.String()
	return s.redis.Del(ctx, key).Err()
}
