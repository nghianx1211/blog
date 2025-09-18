package handlers

import (
	"net/http"

	"blog/internal/models"
	"blog/internal/services"
	"blog/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PostHandler struct {
	postService *services.PostService
}

func NewPostHandler(postService *services.PostService) *PostHandler {
	return &PostHandler{postService: postService}
}

func (h *PostHandler) CreatePost(c *gin.Context) {
	var req models.PostCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	post, err := h.postService.CreatePost(c.Request.Context(), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create post", err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Post created successfully", post.ToResponse())
}

func (h *PostHandler) GetPost(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid post ID", err)
		return
	}

	post, err := h.postService.GetPost(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "post not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "Post not found", err)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get post", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Post retrieved successfully", post.ToResponse())
}

func (h *PostHandler) UpdatePost(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid post ID", err)
		return
	}

	var req models.PostUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	post, err := h.postService.UpdatePost(c.Request.Context(), id, &req)
	if err != nil {
		if err.Error() == "post not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "Post not found", err)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update post", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Post updated successfully", post.ToResponse())
}

func (h *PostHandler) DeletePost(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid post ID", err)
		return
	}

	err = h.postService.DeletePost(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "post not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "Post not found", err)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete post", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Post deleted successfully", nil)
}

func (h *PostHandler) SearchPostsByTag(c *gin.Context) {
	tag := c.Query("tag")
	if tag == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tag parameter is required", nil)
		return
	}

	posts, err := h.postService.SearchByTag(c.Request.Context(), tag)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to search posts", err)
		return
	}

	responses := make([]models.PostResponse, len(posts))
	for i, post := range posts {
		responses[i] = post.ToResponse()
	}

	utils.SuccessResponse(c, http.StatusOK, "Posts retrieved successfully", responses)
}
