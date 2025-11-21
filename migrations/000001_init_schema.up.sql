CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE teams (
    team_name VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE users (
    user_id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    team_name VARCHAR(255) NOT NULL REFERENCES teams(team_name) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE pull_requests (
    pull_request_id VARCHAR(255) PRIMARY KEY,
    pull_request_name VARCHAR(255) NOT NULL,
    author_id VARCHAR(255) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'MERGED')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    merged_at TIMESTAMP
);

CREATE TABLE pull_request_reviewers (
    pull_request_id VARCHAR(255) NOT NULL REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    reviewer_id VARCHAR(255) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    assigned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (pull_request_id, reviewer_id)
);

CREATE INDEX idx_users_team_name ON users(team_name);
CREATE INDEX idx_users_is_active ON users(is_active);
CREATE INDEX idx_pull_requests_author_id ON pull_requests(author_id);
CREATE INDEX idx_pull_requests_status ON pull_requests(status);
CREATE INDEX idx_pull_request_reviewers_reviewer_id ON pull_request_reviewers(reviewer_id);
CREATE INDEX idx_pull_request_reviewers_pr_id ON pull_request_reviewers(pull_request_id);
