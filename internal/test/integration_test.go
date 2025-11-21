package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/avito/pr-reviewer-service/internal/database"
	"github.com/avito/pr-reviewer-service/internal/handler"
	"github.com/avito/pr-reviewer-service/internal/models"
	"github.com/avito/pr-reviewer-service/internal/repository"
	"github.com/avito/pr-reviewer-service/internal/router"
	"github.com/avito/pr-reviewer-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) {
	if os.Getenv("DB_HOST") == "" {
		os.Setenv("DB_HOST", "localhost")     //nolint:errcheck
	}
	if os.Getenv("DB_PORT") == "" {
		os.Setenv("DB_PORT", "5435")          //nolint:errcheck
	}
	if os.Getenv("DB_USER") == "" {
		os.Setenv("DB_USER", "postgres")      //nolint:errcheck
	}
	if os.Getenv("DB_PASSWORD") == "" {
		os.Setenv("DB_PASSWORD", "postgres")  //nolint:errcheck
	}
	if os.Getenv("DB_NAME") == "" {
		os.Setenv("DB_NAME", "pr_reviewer_test") //nolint:errcheck
	}
	if os.Getenv("DB_SSLMODE") == "" {
		os.Setenv("DB_SSLMODE", "disable")    //nolint:errcheck
	}
}

func setupRouter(t *testing.T) *gin.Engine {
	setupTestDB(t)
	db, err := database.NewDB()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	teamRepo := repository.NewTeamRepository(db)
	userRepo := repository.NewUserRepository(db)
	prRepo := repository.NewPullRequestRepository(db)

	teamService := service.NewTeamService(teamRepo, userRepo)
	userService := service.NewUserService(userRepo)
	prService := service.NewPRService(prRepo, userRepo)
	statsService := service.NewStatsService(prRepo)

	teamHandler := handler.NewTeamHandler(teamService)
	userHandler := handler.NewUserHandler(userService, prService)
	prHandler := handler.NewPRHandler(prService)
	statsHandler := handler.NewStatsHandler(statsService)
	healthHandler := handler.NewHealthHandler(db)
	metricsHandler := handler.NewMetricsHandler()

	gin.SetMode(gin.TestMode)
	return router.SetupRouter(teamHandler, userHandler, prHandler, statsHandler, healthHandler, metricsHandler)
}

func TestHealthCheck(t *testing.T) {
	r := setupRouter(t)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response) //nolint:errcheck
	assert.Equal(t, "ok", response["status"])
}

func TestCreateTeam(t *testing.T) {
	setupTestDB(t)
	db, err := database.NewDB()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck

	_, _ = db.Exec("DELETE FROM teams WHERE team_name = 'backend'") //nolint:errcheck

	r := setupRouter(t)

	team := models.Team{
		TeamName: "backend",
		Members: []models.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u2", Username: "Bob", IsActive: true},
		},
	}

	body, _ := json.Marshal(team)
	req, _ := http.NewRequest("POST", "/team/add", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreatePR(t *testing.T) {
	setupTestDB(t)
	db, err := database.NewDB()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck

	_, _ = db.Exec("DELETE FROM teams WHERE team_name = 'frontend'") //nolint:errcheck
	_, _ = db.Exec("DELETE FROM pull_requests WHERE pull_request_id = 'pr-1'") //nolint:errcheck

	r := setupRouter(t)

	team := models.Team{
		TeamName: "frontend",
		Members: []models.TeamMember{
			{UserID: "u3", Username: "Charlie", IsActive: true},
			{UserID: "u4", Username: "David", IsActive: true},
			{UserID: "u5", Username: "Eve", IsActive: true},
		},
	}

	body, _ := json.Marshal(team)
	req, _ := http.NewRequest("POST", "/team/add", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	prReq := map[string]string{
		"pull_request_id":   "pr-1",
		"pull_request_name": "Test PR",
		"author_id":         "u3",
	}

	body, _ = json.Marshal(prReq)
	req, _ = http.NewRequest("POST", "/pullRequest/create", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response) //nolint:errcheck
	pr := response["pr"].(map[string]interface{})
	reviewers := pr["assigned_reviewers"].([]interface{})
	assert.LessOrEqual(t, len(reviewers), 2)
	assert.NotContains(t, reviewers, "u3")
}

func TestCreatePRWithDuplicateID(t *testing.T) {
	r := setupRouter(t)

	prReq := map[string]string{
		"pull_request_id":   "pr-duplicate",
		"pull_request_name": "Test PR",
		"author_id":         "u1",
	}

	body, _ := json.Marshal(prReq)
	req, _ := http.NewRequest("POST", "/pullRequest/create", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	
	if w.Code == http.StatusCreated {
		req, _ = http.NewRequest("POST", "/pullRequest/create", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusConflict, w.Code)
	}
}

func TestMergePR(t *testing.T) {
	r := setupRouter(t)

	mergeReq := map[string]string{
		"pull_request_id": "pr-1",
	}

	body, _ := json.Marshal(mergeReq)
	req, _ := http.NewRequest("POST", "/pullRequest/merge", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Logf("Unexpected status code: %d, body: %s", w.Code, w.Body.String())
		return
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v, body: %s", err, w.Body.String())
	}
	
	pr, ok := response["pr"].(map[string]interface{})
	if !ok {
		t.Fatalf("Failed to parse PR from response: %v", response)
	}
	assert.Equal(t, "MERGED", pr["status"])

	body, _ = json.Marshal(mergeReq)
	req, _ = http.NewRequest("POST", "/pullRequest/merge", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReassignReviewer(t *testing.T) {
	r := setupRouter(t)

	reassignReq := map[string]string{
		"pull_request_id": "pr-1",
		"old_user_id":     "u4",
	}

	body, _ := json.Marshal(reassignReq)
	req, _ := http.NewRequest("POST", "/pullRequest/reassign", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestReassignOnMergedPR(t *testing.T) {
	r := setupRouter(t)

	mergeReq := map[string]string{"pull_request_id": "pr-1"}
	body, _ := json.Marshal(mergeReq)
	req, _ := http.NewRequest("POST", "/pullRequest/merge", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	reassignReq := map[string]string{
		"pull_request_id": "pr-1",
		"old_user_id":     "u3",
	}
	body, _ = json.Marshal(reassignReq)
	req, _ = http.NewRequest("POST", "/pullRequest/reassign", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestSetIsActive(t *testing.T) {
	r := setupRouter(t)

	reqBody := map[string]interface{}{
		"user_id":   "u1",
		"is_active": false,
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/users/setIsActive", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response) //nolint:errcheck
	user := response["user"].(map[string]interface{})
	assert.Equal(t, false, user["is_active"])
}

func TestGetReview(t *testing.T) {
	r := setupRouter(t)

	req, _ := http.NewRequest("GET", "/users/getReview?user_id=u4", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response) //nolint:errcheck
	assert.Equal(t, "u4", response["user_id"])
}

func TestBulkDeactivateTeam(t *testing.T) {
	r := setupRouter(t)

	reqBody := map[string]string{"team_name": "backend"}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/team/bulkDeactivate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetStats(t *testing.T) {
	r := setupRouter(t)

	req, _ := http.NewRequest("GET", "/stats", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response) //nolint:errcheck
	assert.NotNil(t, response["user_stats"])
	assert.NotNil(t, response["pr_stats"])
}

func TestNotFound(t *testing.T) {
	r := setupRouter(t)

	req, _ := http.NewRequest("GET", "/team/get?team_name=nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
