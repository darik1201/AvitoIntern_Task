package service

import (
	"github.com/avito/pr-reviewer-service/internal/models"
)

type PRRepositoryInterface interface {
	PRExists(prID string) (bool, error)
	CreatePR(pr *models.PullRequest) error
	GetPR(prID string) (*models.PullRequest, error)
	MergePR(prID string) (*models.PullRequest, error)
	ReassignReviewer(prID, oldReviewerID, newReviewerID string) error
	GetPRsByReviewer(reviewerID string) ([]models.PullRequestShort, error)
	GetUserStats() ([]models.UserStat, error)
	GetPRStats() (*models.PRStat, error)
}

type UserRepositoryInterface interface {
	GetUser(userID string) (*models.User, error)
	SetIsActive(userID string, isActive bool) (*models.User, error)
	GetActiveTeamMembers(teamName string, excludeUserID string) ([]models.User, error)
	BulkDeactivateTeamMembers(teamName string) error
}

type TeamRepositoryInterface interface {
	TeamExists(teamName string) (bool, error)
	CreateTeam(team *models.Team) error
	GetTeam(teamName string) (*models.Team, error)
}

