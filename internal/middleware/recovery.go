package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log.Error().
			Str("error", getString(recovered)).
			Str("path", c.Request.URL.Path).
			Str("method", c.Request.Method).
			Str("request_id", c.GetString("request_id")).
			Msg("Panic recovered")

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Internal server error",
			},
		})
		c.Abort()
	})
}

func getString(val interface{}) string {
	if str, ok := val.(string); ok {
		return str
	}
	if err, ok := val.(error); ok {
		return err.Error()
	}
	return "unknown error"
}
