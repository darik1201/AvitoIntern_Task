package handler

import (
	"net/http"

	"github.com/avito/pr-reviewer-service/internal/models"
	"github.com/avito/pr-reviewer-service/internal/service"
	"github.com/gin-gonic/gin"
)

type TeamHandler struct {
	teamService TeamServiceInterface
}

func NewTeamHandler(teamService *service.TeamService) *TeamHandler {
	return &TeamHandler{teamService: teamService}
}

func (h *TeamHandler) AddTeam(c *gin.Context) {
	var team models.Team
	if err := c.ShouldBindJSON(&team); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if err := h.teamService.CreateTeam(&team); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"team": team})
}

func (h *TeamHandler) GetTeam(c *gin.Context) {
	teamName := c.Query("team_name")
	if teamName == "" {
		errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "team_name is required")
		return
	}

	team, err := h.teamService.GetTeam(teamName)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, team)
}

func (h *TeamHandler) BulkDeactivateTeam(c *gin.Context) {
	var req struct {
		TeamName string `json:"team_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if err := h.teamService.BulkDeactivateTeam(req.TeamName); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "team members deactivated successfully"})
}
