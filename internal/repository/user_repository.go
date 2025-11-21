package repository

import (
	"database/sql"
	"fmt"

	"github.com/avito/pr-reviewer-service/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetUser(userID string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(`
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE user_id = $1
	`, userID).Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) SetIsActive(userID string, isActive bool) (*models.User, error) {
	_, err := r.db.Exec(`
		UPDATE users
		SET is_active = $1, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $2
	`, isActive, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return r.GetUser(userID)
}

func (r *UserRepository) GetActiveTeamMembers(teamName string, excludeUserID string) ([]models.User, error) {
	query := `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE team_name = $1 AND is_active = true AND user_id != $2
		ORDER BY user_id
	`
	rows, err := r.db.Query(query, teamName, excludeUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active team members: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

func (r *UserRepository) BulkDeactivateTeamMembers(teamName string) error {
	_, err := r.db.Exec(`
		UPDATE users
		SET is_active = false, updated_at = CURRENT_TIMESTAMP
		WHERE team_name = $1
	`, teamName)
	return err
}
