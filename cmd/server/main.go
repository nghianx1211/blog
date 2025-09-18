package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"blog/internal/config"
	"blog/internal/database"
	"blog/internal/handlers"
	"blog/internal/middleware"
	"blog/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize databases
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	redis, err := database.NewRedisClient(&cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis.Close()

	es, err := database.NewElasticsearchClient(&cfg.Elasticsearch)
	if err != nil {
		log.Fatalf("Failed to connect to Elasticsearch: %v", err)
	}

	// Initialize services
	cacheService := services.NewCacheService(redis)
	searchService := services.NewSearchService(es)
	activityService := services.NewActivityService()
	postService := services.NewPostService(db, cacheService, searchService, activityService)

	// Initialize Elasticsearch index
	ctx := context.Background()
	if err := searchService.InitializeIndex(ctx); err != nil {
		log.Fatalf("Failed to initialize Elasticsearch index: %v", err)
	}

	// Initialize handlers
	postHandler := handlers.NewPostHandler(postService)
	searchHandler := handlers.NewSearchHandler(searchService)

	// Setup Gin router
	router := setupRouter(postHandler, searchHandler)

	// Start server
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Server started on port %s", cfg.Server.Port)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}

func setupRouter(postHandler *handlers.PostHandler, searchHandler *handlers.SearchHandler) *gin.Engine {
	router := gin.New()

	// Middleware
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.CORSMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Posts endpoints
		api.POST("/posts", postHandler.CreatePost)
		api.GET("/posts/:id", postHandler.GetPost)
		api.PUT("/posts/:id", postHandler.UpdatePost)
		api.DELETE("/posts/:id", postHandler.DeletePost)

		// Search endpoints
		api.GET("/posts/search", searchHandler.SearchPosts)
		api.GET("/posts/search-by-tag", postHandler.SearchPostsByTag)
	}

	return router
}