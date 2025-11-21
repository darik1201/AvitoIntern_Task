package repository

import (
	"database/sql"
	"fmt"

	"github.com/avito/pr-reviewer-service/internal/models"
)

type TeamRepository struct {
	db *sql.DB
}

func NewTeamRepository(db *sql.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) CreateTeam(team *models.Team) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	_, err = tx.Exec("INSERT INTO teams (team_name) VALUES ($1)", team.TeamName)
	if err != nil {
		return fmt.Errorf("failed to create team: %w", err)
	}

	for _, member := range team.Members {
		_, err = tx.Exec(`
			INSERT INTO users (user_id, username, team_name, is_active)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_id) 
			DO UPDATE SET username = $2, team_name = $3, is_active = $4, updated_at = CURRENT_TIMESTAMP
		`, member.UserID, member.Username, team.TeamName, member.IsActive)
		if err != nil {
			return fmt.Errorf("failed to create/update user: %w", err)
		}
	}

	return tx.Commit()
}

func (r *TeamRepository) GetTeam(teamName string) (*models.Team, error) {
	rows, err := r.db.Query(`
		SELECT user_id, username, is_active
		FROM users
		WHERE team_name = $1
		ORDER BY username
	`, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	var team models.Team
	team.TeamName = teamName
	team.Members = []models.TeamMember{}

	for rows.Next() {
		var member models.TeamMember
		if err := rows.Scan(&member.UserID, &member.Username, &member.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan member: %w", err)
		}
		team.Members = append(team.Members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating members: %w", err)
	}

	var exists bool
	err = r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)", teamName).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check team existence: %w", err)
	}

	if !exists && len(team.Members) == 0 {
		return nil, sql.ErrNoRows
	}

	return &team, nil
}

func (r *TeamRepository) TeamExists(teamName string) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)", teamName).Scan(&exists)
	return exists, err
}

func (r *TeamRepository) GetTeamMembers(teamName string) ([]models.User, error) {
	rows, err := r.db.Query(`
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE team_name = $1
	`, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get team members: %w", err)
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
