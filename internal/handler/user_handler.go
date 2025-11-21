package handler

import (
	"net/http"

	"github.com/avito/pr-reviewer-service/internal/service"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService UserServiceInterface
	prService   PRServiceInterface
}

func NewUserHandler(userService *service.UserService, prService *service.PRService) *UserHandler {
	return &UserHandler{userService: userService, prService: prService}
}

func (h *UserHandler) SetIsActive(c *gin.Context) {
	var req struct {
		UserID   string `json:"user_id" binding:"required"`
		IsActive bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	user, err := h.userService.SetIsActive(req.UserID, req.IsActive)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *UserHandler) GetReview(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "user_id is required")
		return
	}

	prs, err := h.prService.GetPRsByReviewer(userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":       userID,
		"pull_requests": prs,
	})
}
