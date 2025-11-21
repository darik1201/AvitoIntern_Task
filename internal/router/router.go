package router

import (
	"os"

	"github.com/avito/pr-reviewer-service/internal/handler"
	"github.com/avito/pr-reviewer-service/internal/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(
	teamHandler *handler.TeamHandler,
	userHandler *handler.UserHandler,
	prHandler *handler.PRHandler,
	statsHandler *handler.StatsHandler,
	healthHandler *handler.HealthHandler,
	metricsHandler *handler.MetricsHandler,
) *gin.Engine {
	r := gin.New()

	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.PrometheusMetrics())

	r.GET("/health", healthHandler.HealthCheck)
	r.GET("/metrics", metricsHandler.Metrics)

	swaggerGroup := r.Group("/swagger")
	swaggerHandler := handler.NewSwaggerHandler("openapi.yml")
	
	if swaggerJSON, err := os.Stat("docs/swagger.json"); err == nil && swaggerJSON.Size() > 100 {
		swaggerGroup.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	} else {
		swaggerGroup.GET("/openapi.json", swaggerHandler.ServeOpenAPI)
		swaggerGroup.GET("", swaggerHandler.ServeSwaggerUI)
		swaggerGroup.GET("/", swaggerHandler.ServeSwaggerUI)
	}

	teams := r.Group("/team")
	{
		teams.POST("/add", teamHandler.AddTeam)
		teams.GET("/get", teamHandler.GetTeam)
		teams.POST("/bulkDeactivate", teamHandler.BulkDeactivateTeam)
	}

	users := r.Group("/users")
	{
		users.POST("/setIsActive", userHandler.SetIsActive)
		users.GET("/getReview", userHandler.GetReview)
	}

	prs := r.Group("/pullRequest")
	{
		prs.POST("/create", prHandler.CreatePR)
		prs.POST("/merge", prHandler.MergePR)
		prs.POST("/reassign", prHandler.ReassignReviewer)
	}

	r.GET("/stats", statsHandler.GetStats)

	return r
}
