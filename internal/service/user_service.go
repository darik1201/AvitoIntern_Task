package service

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/avito/pr-reviewer-service/internal/models"
	"github.com/avito/pr-reviewer-service/internal/repository"
)

type UserService struct {
	userRepo UserRepositoryInterface
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) SetIsActive(userID string, isActive bool) (*models.User, error) {
	user, err := s.userRepo.SetIsActive(userID, isActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	return user, nil
}

func (s *UserService) GetActiveTeamMembers(teamName string, excludeUserID string) ([]models.User, error) {
	return s.userRepo.GetActiveTeamMembers(teamName, excludeUserID)
}
