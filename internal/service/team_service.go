package service

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/avito/pr-reviewer-service/internal/models"
	"github.com/avito/pr-reviewer-service/internal/repository"
)

type TeamService struct {
	teamRepo TeamRepositoryInterface
	userRepo UserRepositoryInterface
}

func NewTeamService(teamRepo *repository.TeamRepository, userRepo *repository.UserRepository) *TeamService {
	return &TeamService{teamRepo: teamRepo, userRepo: userRepo}
}

func (s *TeamService) CreateTeam(team *models.Team) error {
	exists, err := s.teamRepo.TeamExists(team.TeamName)
	if err != nil {
		return fmt.Errorf("failed to check team existence: %w", err)
	}
	if exists {
		return fmt.Errorf("team already exists")
	}

	return s.teamRepo.CreateTeam(team)
}

func (s *TeamService) GetTeam(teamName string) (*models.Team, error) {
	team, err := s.teamRepo.GetTeam(teamName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("team not found")
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}
	return team, nil
}

func (s *TeamService) BulkDeactivateTeam(teamName string) error {
	_, err := s.teamRepo.GetTeam(teamName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("team not found")
		}
		return fmt.Errorf("failed to get team: %w", err)
	}

	return s.userRepo.BulkDeactivateTeamMembers(teamName)
}
