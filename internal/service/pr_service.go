package service

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/avito/pr-reviewer-service/internal/models"
	"github.com/avito/pr-reviewer-service/internal/repository"
)

type PRService struct {
	prRepo   PRRepositoryInterface
	userRepo UserRepositoryInterface
}

func NewPRService(prRepo *repository.PullRequestRepository, userRepo *repository.UserRepository) *PRService {
	return &PRService{prRepo: prRepo, userRepo: userRepo}
}

func (s *PRService) CreatePR(prID, prName, authorID string) (*models.PullRequest, error) {
	exists, err := s.prRepo.PRExists(prID)
	if err != nil {
		return nil, fmt.Errorf("failed to check PR existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("PR already exists")
	}

	author, err := s.userRepo.GetUser(authorID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("author not found")
		}
		return nil, fmt.Errorf("failed to get author: %w", err)
	}

	candidates, err := s.userRepo.GetActiveTeamMembers(author.TeamName, authorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}

	maxReviewers := 2
	if len(candidates) < maxReviewers {
		maxReviewers = len(candidates)
	}

	var reviewerIDs []string
	if maxReviewers > 0 {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		selected := r.Perm(len(candidates))[:maxReviewers]
		for _, idx := range selected {
			reviewerIDs = append(reviewerIDs, candidates[idx].UserID)
		}
	}

	pr := &models.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            models.StatusOpen,
		AssignedReviewers: reviewerIDs,
	}

	if err := s.prRepo.CreatePR(pr); err != nil {
		return nil, fmt.Errorf("failed to create PR: %w", err)
	}

	return s.prRepo.GetPR(prID)
}

func (s *PRService) MergePR(prID string) (*models.PullRequest, error) {
	pr, err := s.prRepo.GetPR(prID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("PR not found")
		}
		return nil, fmt.Errorf("failed to get PR: %w", err)
	}

	if pr.Status == models.StatusMerged {
		return pr, nil
	}

	return s.prRepo.MergePR(prID)
}

func (s *PRService) ReassignReviewer(prID, oldReviewerID string) (string, *models.PullRequest, error) {
	pr, err := s.prRepo.GetPR(prID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil, fmt.Errorf("PR not found")
		}
		return "", nil, fmt.Errorf("failed to get PR: %w", err)
	}

	if pr.Status == models.StatusMerged {
		return "", nil, fmt.Errorf("cannot reassign on merged PR")
	}

	found := false
	for _, reviewerID := range pr.AssignedReviewers {
		if reviewerID == oldReviewerID {
			found = true
			break
		}
	}

	if !found {
		return "", nil, fmt.Errorf("reviewer is not assigned to this PR")
	}

	oldReviewer, err := s.userRepo.GetUser(oldReviewerID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get old reviewer: %w", err)
	}

	candidates, err := s.userRepo.GetActiveTeamMembers(oldReviewer.TeamName, oldReviewerID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get team members: %w", err)
	}

	excludeIDs := make(map[string]bool)
	for _, reviewerID := range pr.AssignedReviewers {
		excludeIDs[reviewerID] = true
	}
	excludeIDs[pr.AuthorID] = true

	var available []models.User
	for _, candidate := range candidates {
		if !excludeIDs[candidate.UserID] {
			available = append(available, candidate)
		}
	}

	if len(available) == 0 {
		return "", nil, fmt.Errorf("no active replacement candidate in team")
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	newReviewer := available[r.Intn(len(available))]

	if err := s.prRepo.ReassignReviewer(prID, oldReviewerID, newReviewer.UserID); err != nil {
		return "", nil, fmt.Errorf("failed to reassign reviewer: %w", err)
	}

	updatedPR, err := s.prRepo.GetPR(prID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get updated PR: %w", err)
	}

	return newReviewer.UserID, updatedPR, nil
}

func (s *PRService) GetPRsByReviewer(reviewerID string) ([]models.PullRequestShort, error) {
	_, err := s.userRepo.GetUser(reviewerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return s.prRepo.GetPRsByReviewer(reviewerID)
}
