package handler

import (
	"net/http"

	"github.com/avito/pr-reviewer-service/internal/service"
	"github.com/gin-gonic/gin"
)

type StatsHandler struct {
	statsService StatsServiceInterface
}

func NewStatsHandler(statsService *service.StatsService) *StatsHandler {
	return &StatsHandler{statsService: statsService}
}

func (h *StatsHandler) GetStats(c *gin.Context) {
	stats, err := h.statsService.GetStats()
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, stats)
}
