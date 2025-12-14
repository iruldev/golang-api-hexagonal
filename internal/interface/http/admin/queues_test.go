package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/middleware"
	"github.com/iruldev/golang-api-hexagonal/internal/runtimeutil"
)

// mockQueueInspector implements QueueInspector for testing.
type mockQueueInspector struct {
	getQueueStatsFunc   func(ctx context.Context) (*runtimeutil.QueueStats, error)
	getJobsInQueueFunc  func(ctx context.Context, queueName string, page, pageSize int) (*runtimeutil.JobList, error)
	getFailedJobsFunc   func(ctx context.Context, queueName string, page, pageSize int) (*runtimeutil.FailedJobList, error)
	deleteFailedJobFunc func(ctx context.Context, queueName, taskID string) error
	retryFailedJobFunc  func(ctx context.Context, queueName, taskID string) (*runtimeutil.JobInfo, error)
}

func (m *mockQueueInspector) GetQueueStats(ctx context.Context) (*runtimeutil.QueueStats, error) {
	if m.getQueueStatsFunc != nil {
		return m.getQueueStatsFunc(ctx)
	}
	return &runtimeutil.QueueStats{
		Aggregate: runtimeutil.AggregateStats{TotalEnqueued: 100},
		Queues:    []runtimeutil.QueueInfo{{Name: "default", Size: 100}},
	}, nil
}

func (m *mockQueueInspector) GetJobsInQueue(ctx context.Context, queueName string, page, pageSize int) (*runtimeutil.JobList, error) {
	if m.getJobsInQueueFunc != nil {
		return m.getJobsInQueueFunc(ctx, queueName, page, pageSize)
	}
	return &runtimeutil.JobList{
		Jobs:       []runtimeutil.JobInfo{},
		Pagination: runtimeutil.Pagination{Page: page, PageSize: pageSize, Total: 0, TotalPages: 1},
	}, nil
}

func (m *mockQueueInspector) GetFailedJobs(ctx context.Context, queueName string, page, pageSize int) (*runtimeutil.FailedJobList, error) {
	if m.getFailedJobsFunc != nil {
		return m.getFailedJobsFunc(ctx, queueName, page, pageSize)
	}
	return &runtimeutil.FailedJobList{
		FailedJobs: []runtimeutil.FailedJobInfo{},
		Pagination: runtimeutil.Pagination{Page: page, PageSize: pageSize, Total: 0, TotalPages: 1},
	}, nil
}

func (m *mockQueueInspector) DeleteFailedJob(ctx context.Context, queueName, taskID string) error {
	if m.deleteFailedJobFunc != nil {
		return m.deleteFailedJobFunc(ctx, queueName, taskID)
	}
	return nil
}

func (m *mockQueueInspector) RetryFailedJob(ctx context.Context, queueName, taskID string) (*runtimeutil.JobInfo, error) {
	if m.retryFailedJobFunc != nil {
		return m.retryFailedJobFunc(ctx, queueName, taskID)
	}
	return &runtimeutil.JobInfo{TaskID: taskID, Queue: queueName, State: "pending"}, nil
}

// setupQueuesRouter creates a chi router with the queues handler for testing.
func setupQueuesRouter(inspector runtimeutil.QueueInspector) *chi.Mux {
	r := chi.NewRouter()
	handler := NewQueuesHandler(inspector, zap.NewNop())

	r.Route("/admin/queues", func(r chi.Router) {
		r.Get("/stats", handler.GetQueueStats)
		r.Get("/{queue}/jobs", handler.ListJobs)
		r.Get("/{queue}/failed", handler.ListFailedJobs)
		r.Delete("/{queue}/failed/{task_id}", handler.DeleteFailedJob)
		r.Post("/{queue}/failed/{task_id}/retry", handler.RetryFailedJob)
	})

	return r
}

// withQueueAdminClaims adds admin claims to the request context.
func withQueueAdminClaims(r *http.Request) *http.Request {
	claims := middleware.Claims{
		UserID: "admin-user-id",
		Roles:  []string{"admin"},
	}
	ctx := middleware.NewContext(r.Context(), claims)
	return r.WithContext(ctx)
}

func TestQueuesHandler_GetQueueStats(t *testing.T) {
	t.Parallel()

	t.Run("returns aggregate stats", func(t *testing.T) {
		// Arrange
		inspector := &mockQueueInspector{
			getQueueStatsFunc: func(ctx context.Context) (*runtimeutil.QueueStats, error) {
				return &runtimeutil.QueueStats{
					Aggregate: runtimeutil.AggregateStats{
						TotalEnqueued: 150,
						TotalActive:   5,
						TotalPending:  100,
					},
					Queues: []runtimeutil.QueueInfo{
						{Name: "critical", Size: 50, Active: 2},
						{Name: "default", Size: 80, Active: 3},
						{Name: "low", Size: 20, Active: 0},
					},
				}, nil
			},
		}
		router := setupQueuesRouter(inspector)

		req := httptest.NewRequest(http.MethodGet, "/admin/queues/stats", nil)
		req = withQueueAdminClaims(req)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		require.Equal(t, http.StatusOK, rr.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.True(t, resp["success"].(bool))
		data := resp["data"].(map[string]interface{})
		aggregate := data["aggregate"].(map[string]interface{})
		assert.Equal(t, float64(150), aggregate["total_enqueued"])
		queues := data["queues"].([]interface{})
		assert.Len(t, queues, 3)
	})

	t.Run("returns 500 on internal error", func(t *testing.T) {
		// Arrange
		inspector := &mockQueueInspector{
			getQueueStatsFunc: func(ctx context.Context) (*runtimeutil.QueueStats, error) {
				return nil, errors.New("redis connection failed")
			},
		}
		router := setupQueuesRouter(inspector)

		req := httptest.NewRequest(http.MethodGet, "/admin/queues/stats", nil)
		req = withQueueAdminClaims(req)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		require.Equal(t, http.StatusInternalServerError, rr.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.False(t, resp["success"].(bool))
	})
}

func TestQueuesHandler_ListJobs(t *testing.T) {
	t.Parallel()

	t.Run("returns job list with pagination", func(t *testing.T) {
		// Arrange
		inspector := &mockQueueInspector{
			getJobsInQueueFunc: func(ctx context.Context, queueName string, page, pageSize int) (*runtimeutil.JobList, error) {
				return &runtimeutil.JobList{
					Jobs: []runtimeutil.JobInfo{
						{
							TaskID:         "task-123",
							Type:           "note:archive",
							PayloadPreview: `{"note_id":"abc"}`,
							State:          "pending",
							Queue:          queueName,
						},
					},
					Pagination: runtimeutil.Pagination{Page: page, PageSize: pageSize, Total: 1, TotalPages: 1},
				}, nil
			},
		}
		router := setupQueuesRouter(inspector)

		req := httptest.NewRequest(http.MethodGet, "/admin/queues/default/jobs?page=1&page_size=10", nil)
		req = withQueueAdminClaims(req)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		require.Equal(t, http.StatusOK, rr.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.True(t, resp["success"].(bool))
		data := resp["data"].(map[string]interface{})
		jobs := data["jobs"].([]interface{})
		assert.Len(t, jobs, 1)
		pagination := data["pagination"].(map[string]interface{})
		assert.Equal(t, float64(1), pagination["page"])
	})

	t.Run("returns 400 for invalid queue name", func(t *testing.T) {
		// Arrange
		inspector := &mockQueueInspector{}
		router := setupQueuesRouter(inspector)

		req := httptest.NewRequest(http.MethodGet, "/admin/queues/invalid_queue/jobs", nil)
		req = withQueueAdminClaims(req)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		require.Equal(t, http.StatusBadRequest, rr.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.False(t, resp["success"].(bool))
		errData := resp["error"].(map[string]interface{})
		assert.Equal(t, "Invalid queue name", errData["message"])
	})
}

func TestQueuesHandler_ListFailedJobs(t *testing.T) {
	t.Parallel()

	t.Run("returns failed jobs", func(t *testing.T) {
		// Arrange
		inspector := &mockQueueInspector{
			getFailedJobsFunc: func(ctx context.Context, queueName string, page, pageSize int) (*runtimeutil.FailedJobList, error) {
				return &runtimeutil.FailedJobList{
					FailedJobs: []runtimeutil.FailedJobInfo{
						{
							TaskID:         "task-456",
							Type:           "note:export",
							PayloadPreview: `{"note_id":"xyz"}`,
							ErrorMessage:   "connection timeout",
							FailedAt:       time.Now(),
							RetryCount:     3,
							MaxRetry:       3,
						},
					},
					Pagination: runtimeutil.Pagination{Page: 1, PageSize: 20, Total: 1, TotalPages: 1},
				}, nil
			},
		}
		router := setupQueuesRouter(inspector)

		req := httptest.NewRequest(http.MethodGet, "/admin/queues/critical/failed", nil)
		req = withQueueAdminClaims(req)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		require.Equal(t, http.StatusOK, rr.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.True(t, resp["success"].(bool))
		data := resp["data"].(map[string]interface{})
		failedJobs := data["failed_jobs"].([]interface{})
		assert.Len(t, failedJobs, 1)
		firstJob := failedJobs[0].(map[string]interface{})
		assert.Equal(t, "connection timeout", firstJob["error_message"])
	})

	t.Run("returns 400 for invalid queue name", func(t *testing.T) {
		// Arrange
		inspector := &mockQueueInspector{}
		router := setupQueuesRouter(inspector)

		req := httptest.NewRequest(http.MethodGet, "/admin/queues/bad_queue/failed", nil)
		req = withQueueAdminClaims(req)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		require.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestQueuesHandler_DeleteFailedJob(t *testing.T) {
	t.Parallel()

	t.Run("deletes task successfully", func(t *testing.T) {
		// Arrange
		deleteCalled := false
		inspector := &mockQueueInspector{
			deleteFailedJobFunc: func(ctx context.Context, queueName, taskID string) error {
				deleteCalled = true
				assert.Equal(t, "default", queueName)
				assert.Equal(t, "task-789", taskID)
				return nil
			},
		}
		router := setupQueuesRouter(inspector)

		req := httptest.NewRequest(http.MethodDelete, "/admin/queues/default/failed/task-789", nil)
		req = withQueueAdminClaims(req)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		require.Equal(t, http.StatusOK, rr.Code)
		assert.True(t, deleteCalled)

		var resp map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.True(t, resp["success"].(bool))
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, "Task deleted successfully", data["message"])
	})

	t.Run("returns 404 for task not found", func(t *testing.T) {
		// Arrange
		inspector := &mockQueueInspector{
			deleteFailedJobFunc: func(ctx context.Context, queueName, taskID string) error {
				return runtimeutil.ErrTaskNotFound
			},
		}
		router := setupQueuesRouter(inspector)

		req := httptest.NewRequest(http.MethodDelete, "/admin/queues/default/failed/unknown-task", nil)
		req = withQueueAdminClaims(req)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		require.Equal(t, http.StatusNotFound, rr.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)

		errData := resp["error"].(map[string]interface{})
		assert.Equal(t, "ERR_NOT_FOUND", errData["code"])
	})

	t.Run("returns 400 for invalid queue name", func(t *testing.T) {
		// Arrange
		inspector := &mockQueueInspector{}
		router := setupQueuesRouter(inspector)

		req := httptest.NewRequest(http.MethodDelete, "/admin/queues/wrong_queue/failed/task-123", nil)
		req = withQueueAdminClaims(req)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		require.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestQueuesHandler_RetryFailedJob(t *testing.T) {
	t.Parallel()

	t.Run("retries task successfully", func(t *testing.T) {
		// Arrange
		retryCalled := false
		inspector := &mockQueueInspector{
			retryFailedJobFunc: func(ctx context.Context, queueName, taskID string) (*runtimeutil.JobInfo, error) {
				retryCalled = true
				return &runtimeutil.JobInfo{
					TaskID: taskID,
					Queue:  queueName,
					State:  "pending",
				}, nil
			},
		}
		router := setupQueuesRouter(inspector)

		req := httptest.NewRequest(http.MethodPost, "/admin/queues/critical/failed/task-999/retry", nil)
		req = withQueueAdminClaims(req)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		require.Equal(t, http.StatusOK, rr.Code)
		assert.True(t, retryCalled)

		var resp map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.True(t, resp["success"].(bool))
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, "Task requeued for retry", data["message"])
		assert.Equal(t, "task-999", data["task_id"])
	})

	t.Run("returns 404 for task not found", func(t *testing.T) {
		// Arrange
		inspector := &mockQueueInspector{
			retryFailedJobFunc: func(ctx context.Context, queueName, taskID string) (*runtimeutil.JobInfo, error) {
				return nil, runtimeutil.ErrTaskNotFound
			},
		}
		router := setupQueuesRouter(inspector)

		req := httptest.NewRequest(http.MethodPost, "/admin/queues/default/failed/missing-task/retry", nil)
		req = withQueueAdminClaims(req)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		require.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("returns 400 for invalid queue name", func(t *testing.T) {
		// Arrange
		inspector := &mockQueueInspector{}
		router := setupQueuesRouter(inspector)

		req := httptest.NewRequest(http.MethodPost, "/admin/queues/unknown/failed/task-123/retry", nil)
		req = withQueueAdminClaims(req)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		require.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
