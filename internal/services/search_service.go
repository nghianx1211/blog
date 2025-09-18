package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"blog/internal/database"
	"blog/internal/models"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/google/uuid"
)

type SearchService struct {
	es *database.ElasticsearchClient
}

func (s *SearchService) DeletePost(ctx context.Context, id uuid.UUID) error {
	req := esapi.DeleteRequest{
		Index:      database.PostsIndex,
		DocumentID: id.String(),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, s.es)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("failed to delete post: %s", res.Status())
	}

	return nil
}

func (s *SearchService) SearchPosts(ctx context.Context, req *models.PostSearchRequest) (*models.PostSearchResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	from := (req.Page - 1) * req.Limit

	// Build search query
	query := s.buildSearchQuery(req.Query, req.Tags)
	
	searchReq := esapi.SearchRequest{
		Index: []string{database.PostsIndex},
		Body:  strings.NewReader(query),
		From:  &from,
		Size:  &req.Limit,
	}

	res, err := searchReq.Do(ctx, s.es)
	if err != nil {
		return nil, fmt.Errorf("failed to search posts: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search error: %s", res.Status())
	}

	var searchResult struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source models.Post `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode search result: %w", err)
	}

	posts := make([]models.PostResponse, len(searchResult.Hits.Hits))
	for i, hit := range searchResult.Hits.Hits {
		posts[i] = hit.Source.ToResponse()
	}

	return &models.PostSearchResponse{
		Posts:      posts,
		TotalCount: searchResult.Hits.Total.Value,
		Page:       req.Page,
		Limit:      req.Limit,
	}, nil
}

func (s *SearchService) buildSearchQuery(queryString, tags string) string {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{},
			},
		},
		"sort": []map[string]interface{}{
			{"created_at": map[string]string{"order": "desc"}},
		},
	}

	mustQueries := []interface{}{}

	// Add text search if query is provided
	if queryString != "" {
		mustQueries = append(mustQueries, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  queryString,
				"fields": []string{"title", "content"},
			},
		})
	}

	// Add tag filter if tags are provided
	if tags != "" {
		tagList := strings.Split(tags, ",")
		for i, tag := range tagList {
			tagList[i] = strings.TrimSpace(tag)
		}
		mustQueries = append(mustQueries, map[string]interface{}{
			"terms": map[string]interface{}{
				"tags": tagList,
			},
		})
	}

	// If no specific queries, match all
	if len(mustQueries) == 0 {
		mustQueries = append(mustQueries, map[string]interface{}{
			"match_all": map[string]interface{}{},
		})
	}

	query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = mustQueries

	queryJSON, _ := json.Marshal(query)
	return string(queryJSON)
}

func NewSearchService(es *database.ElasticsearchClient) *SearchService {
	return &SearchService{es: es}
}

func (s *SearchService) InitializeIndex(ctx context.Context) error {
	// Check if index exists
	req := esapi.IndicesExistsRequest{
		Index: []string{database.PostsIndex},
	}

	res, err := req.Do(ctx, s.es)
	if err != nil {
		return fmt.Errorf("failed to check if index exists: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		// Index already exists
		return nil
	}

	// Create index with mapping
	createReq := esapi.IndicesCreateRequest{
		Index: database.PostsIndex,
		Body:  strings.NewReader(database.GetPostsMapping()),
	}

	createRes, err := createReq.Do(ctx, s.es)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer createRes.Body.Close()

	if createRes.IsError() {
		return fmt.Errorf("failed to create index: %s", createRes.Status())
	}

	return nil
}

func (s *SearchService) IndexPost(ctx context.Context, post *models.Post) error {
	doc := post.ToElasticsearchDoc()
	data, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal post: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      database.PostsIndex,
		DocumentID: post.ID.String(),
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, s.es)
	if err != nil {
		return fmt.Errorf("failed to index post: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to index post: %s", res.Status())
	}

	return nil
}