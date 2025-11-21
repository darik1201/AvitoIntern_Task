package handler

import (
	"github.com/avito/pr-reviewer-service/internal/models"
)

type PRServiceInterface interface {
	CreatePR(prID, prName, authorID string) (*models.PullRequest, error)
	MergePR(prID string) (*models.PullRequest, error)
	ReassignReviewer(prID, oldReviewerID string) (string, *models.PullRequest, error)
	GetPRsByReviewer(reviewerID string) ([]models.PullRequestShort, error)
}

type TeamServiceInterface interface {
	CreateTeam(team *models.Team) error
	GetTeam(teamName string) (*models.Team, error)
	BulkDeactivateTeam(teamName string) error
}

type UserServiceInterface interface {
	SetIsActive(userID string, isActive bool) (*models.User, error)
}

type StatsServiceInterface interface {
	GetStats() (*models.StatsResponse, error)
}

