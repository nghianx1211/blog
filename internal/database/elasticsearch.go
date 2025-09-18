package database

import (
	"blog/internal/config"

	"github.com/elastic/go-elasticsearch/v8"
)

type ElasticsearchClient struct {
	*elasticsearch.Client
}

func NewElasticsearchClient(cfg *config.ElasticsearchConfig) (*ElasticsearchClient, error) {
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{cfg.URL},
	})
	if err != nil {
		return nil, err
	}

	return &ElasticsearchClient{es}, nil
}

const PostsIndex = "posts"

func GetPostsMapping() string {
	return `{
		"mappings": {
			"properties": {
				"id": {"type": "keyword"},
				"title": {
					"type": "text",
					"analyzer": "standard"
				},
				"content": {
					"type": "text",
					"analyzer": "standard"
				},
				"tags": {"type": "keyword"},
				"created_at": {"type": "date"},
				"updated_at": {"type": "date"}
			}
		}
	}`
}
