package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/avito/pr-reviewer-service/internal/database"
	"github.com/avito/pr-reviewer-service/internal/handler"
	"github.com/avito/pr-reviewer-service/internal/repository"
	"github.com/avito/pr-reviewer-service/internal/router"
	"github.com/avito/pr-reviewer-service/internal/service"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	db, err := database.NewDB()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close() //nolint:errcheck

	teamRepo := repository.NewTeamRepository(db)
	userRepo := repository.NewUserRepository(db)
	prRepo := repository.NewPullRequestRepository(db)

	teamService := service.NewTeamService(teamRepo, userRepo)
	userService := service.NewUserService(userRepo)
	prService := service.NewPRService(prRepo, userRepo)
	statsService := service.NewStatsService(prRepo)

	teamHandler := handler.NewTeamHandler(teamService)
	userHandler := handler.NewUserHandler(userService, prService)
	prHandler := handler.NewPRHandler(prService)
	statsHandler := handler.NewStatsHandler(statsService)
	healthHandler := handler.NewHealthHandler(db)
	metricsHandler := handler.NewMetricsHandler()

	r := router.SetupRouter(teamHandler, userHandler, prHandler, statsHandler, healthHandler, metricsHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Str("port", port).Msg("Server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}