package handler

import (
	"net/http"

	"github.com/avito/pr-reviewer-service/internal/models"
	"github.com/gin-gonic/gin"
)

func errorResponse(c *gin.Context, code int, errorCode, message string) {
	c.JSON(code, models.ErrorResponse{
		Error: struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}{
			Code:    errorCode,
			Message: message,
		},
	})
}

func handleError(c *gin.Context, err error) {
	switch err.Error() {
	case "team already exists":
		errorResponse(c, http.StatusBadRequest, "TEAM_EXISTS", "team_name already exists")
	case "team not found":
		errorResponse(c, http.StatusNotFound, "NOT_FOUND", "team not found")
	case "user not found":
		errorResponse(c, http.StatusNotFound, "NOT_FOUND", "user not found")
	case "PR already exists":
		errorResponse(c, http.StatusConflict, "PR_EXISTS", "PR id already exists")
	case "PR not found":
		errorResponse(c, http.StatusNotFound, "NOT_FOUND", "PR not found")
	case "author not found":
		errorResponse(c, http.StatusNotFound, "NOT_FOUND", "author not found")
	case "cannot reassign on merged PR":
		errorResponse(c, http.StatusConflict, "PR_MERGED", "cannot reassign on merged PR")
	case "reviewer is not assigned to this PR":
		errorResponse(c, http.StatusConflict, "NOT_ASSIGNED", "reviewer is not assigned to this PR")
	case "no active replacement candidate in team":
		errorResponse(c, http.StatusConflict, "NO_CANDIDATE", "no active replacement candidate in team")
	default:
		errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}
}
