package service

import (
	"github.com/avito/pr-reviewer-service/internal/models"
	"github.com/avito/pr-reviewer-service/internal/repository"
)

type StatsService struct {
	prRepo PRRepositoryInterface
}

func NewStatsService(prRepo *repository.PullRequestRepository) *StatsService {
	return &StatsService{prRepo: prRepo}
}

func (s *StatsService) GetStats() (*models.StatsResponse, error) {
	userStats, err := s.prRepo.GetUserStats()
	if err != nil {
		return nil, err
	}

	prStats, err := s.prRepo.GetPRStats()
	if err != nil {
		return nil, err
	}

	return &models.StatsResponse{
		UserStats: userStats,
		PRStats:   *prStats,
	}, nil
}
