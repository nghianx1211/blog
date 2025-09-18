package handlers

import (
	"net/http"

	"blog/internal/models"
	"blog/internal/services"
	"blog/internal/utils"

	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	searchService *services.SearchService
}

func NewSearchHandler(searchService *services.SearchService) *SearchHandler {
	return &SearchHandler{searchService: searchService}
}

func (h *SearchHandler) SearchPosts(c *gin.Context) {
	var req models.PostSearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters", err)
		return
	}

	result, err := h.searchService.SearchPosts(c.Request.Context(), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to search posts", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Search completed successfully", result)
}
