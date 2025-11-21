package handler

import (
	"net/http"

	"github.com/avito/pr-reviewer-service/internal/service"
	"github.com/gin-gonic/gin"
)

type PRHandler struct {
	prService PRServiceInterface
}

func NewPRHandler(prService *service.PRService) *PRHandler {
	return &PRHandler{prService: prService}
}

func (h *PRHandler) CreatePR(c *gin.Context) {
	var req struct {
		PullRequestID   string `json:"pull_request_id" binding:"required"`
		PullRequestName string `json:"pull_request_name" binding:"required"`
		AuthorID        string `json:"author_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	pr, err := h.prService.CreatePR(req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"pr": pr})
}

func (h *PRHandler) MergePR(c *gin.Context) {
	var req struct {
		PullRequestID string `json:"pull_request_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	pr, err := h.prService.MergePR(req.PullRequestID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"pr": pr})
}

func (h *PRHandler) ReassignReviewer(c *gin.Context) {
	var req struct {
		PullRequestID string `json:"pull_request_id" binding:"required"`
		OldUserID     string `json:"old_user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	newReviewerID, pr, err := h.prService.ReassignReviewer(req.PullRequestID, req.OldUserID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pr":         pr,
		"replaced_by": newReviewerID,
	})
}
