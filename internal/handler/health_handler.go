package handler

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db *sql.DB
}

func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	status := gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	if h.db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := h.db.PingContext(ctx); err != nil {
			status["database"] = "unhealthy"
			c.JSON(http.StatusServiceUnavailable, status)
			return
		}
		status["database"] = "healthy"
	}

	c.JSON(http.StatusOK, status)
}
