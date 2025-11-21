package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/avito/pr-reviewer-service/internal/models"
)

type PullRequestRepository struct {
	db *sql.DB
}

func NewPullRequestRepository(db *sql.DB) *PullRequestRepository {
	return &PullRequestRepository{db: db}
}

func (r *PullRequestRepository) CreatePR(pr *models.PullRequest) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	now := time.Now()
	_, err = tx.Exec(`
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, pr.PullRequestID, pr.PullRequestName, pr.AuthorID, pr.Status, now)
	if err != nil {
		return fmt.Errorf("failed to create PR: %w", err)
	}

	for _, reviewerID := range pr.AssignedReviewers {
		_, err = tx.Exec(`
			INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id, assigned_at)
			VALUES ($1, $2, $3)
		`, pr.PullRequestID, reviewerID, now)
		if err != nil {
			return fmt.Errorf("failed to assign reviewer: %w", err)
		}
	}

	return tx.Commit()
}

func (r *PullRequestRepository) GetPR(prID string) (*models.PullRequest, error) {
	var pr models.PullRequest
	var createdAt, mergedAt sql.NullTime

	err := r.db.QueryRow(`
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE pull_request_id = $1
	`, prID).Scan(
		&pr.PullRequestID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&pr.Status,
		&createdAt,
		&mergedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR: %w", err)
	}

	if createdAt.Valid {
		pr.CreatedAt = &createdAt.Time
	}
	if mergedAt.Valid {
		pr.MergedAt = &mergedAt.Time
	}

	rows, err := r.db.Query(`
		SELECT reviewer_id
		FROM pull_request_reviewers
		WHERE pull_request_id = $1
		ORDER BY assigned_at
	`, prID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviewers: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	pr.AssignedReviewers = []string{}
	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, fmt.Errorf("failed to scan reviewer: %w", err)
		}
		pr.AssignedReviewers = append(pr.AssignedReviewers, reviewerID)
	}

	return &pr, rows.Err()
}

func (r *PullRequestRepository) PRExists(prID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)", prID).Scan(&exists)
	return exists, err
}

func (r *PullRequestRepository) MergePR(prID string) (*models.PullRequest, error) {
	now := time.Now()
	_, err := r.db.Exec(`
		UPDATE pull_requests
		SET status = 'MERGED', merged_at = $1
		WHERE pull_request_id = $2 AND status != 'MERGED'
	`, now, prID)
	if err != nil {
		return nil, fmt.Errorf("failed to merge PR: %w", err)
	}

	return r.GetPR(prID)
}

func (r *PullRequestRepository) ReassignReviewer(prID, oldReviewerID, newReviewerID string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	var exists bool
	err = tx.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM pull_request_reviewers
			WHERE pull_request_id = $1 AND reviewer_id = $2
		)
	`, prID, oldReviewerID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check reviewer assignment: %w", err)
	}
	if !exists {
		return fmt.Errorf("reviewer not assigned")
	}

	_, err = tx.Exec(`
		DELETE FROM pull_request_reviewers
		WHERE pull_request_id = $1 AND reviewer_id = $2
	`, prID, oldReviewerID)
	if err != nil {
		return fmt.Errorf("failed to remove old reviewer: %w", err)
	}

	_, err = tx.Exec(`
		INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id, assigned_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
	`, prID, newReviewerID)
	if err != nil {
		return fmt.Errorf("failed to add new reviewer: %w", err)
	}

	return tx.Commit()
}

func (r *PullRequestRepository) GetPRsByReviewer(reviewerID string) ([]models.PullRequestShort, error) {
	rows, err := r.db.Query(`
		SELECT p.pull_request_id, p.pull_request_name, p.author_id, p.status
		FROM pull_requests p
		INNER JOIN pull_request_reviewers prr ON p.pull_request_id = prr.pull_request_id
		WHERE prr.reviewer_id = $1
		ORDER BY p.created_at DESC
	`, reviewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get PRs by reviewer: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	var prs []models.PullRequestShort
	for rows.Next() {
		var pr models.PullRequestShort
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status); err != nil {
			return nil, fmt.Errorf("failed to scan PR: %w", err)
		}
		prs = append(prs, pr)
	}

	return prs, rows.Err()
}

func (r *PullRequestRepository) GetUserStats() ([]models.UserStat, error) {
	rows, err := r.db.Query(`
		SELECT u.user_id, u.username, COUNT(prr.reviewer_id) as assigned_count
		FROM users u
		LEFT JOIN pull_request_reviewers prr ON u.user_id = prr.reviewer_id
		GROUP BY u.user_id, u.username
		ORDER BY assigned_count DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	var stats []models.UserStat
	for rows.Next() {
		var stat models.UserStat
		if err := rows.Scan(&stat.UserID, &stat.Username, &stat.AssignedCount); err != nil {
			return nil, fmt.Errorf("failed to scan user stat: %w", err)
		}
		stats = append(stats, stat)
	}

	return stats, rows.Err()
}

func (r *PullRequestRepository) GetPRStats() (*models.PRStat, error) {
	var stats models.PRStat
	err := r.db.QueryRow(`
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'OPEN') as open,
			COUNT(*) FILTER (WHERE status = 'MERGED') as merged
		FROM pull_requests
	`).Scan(&stats.TotalPRs, &stats.OpenPRs, &stats.MergedPRs)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR stats: %w", err)
	}
	return &stats, nil
}
