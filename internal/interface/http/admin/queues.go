// Package admin provides HTTP handlers for administrative endpoints.
package admin

import (
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/middleware"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/response"
	"github.com/iruldev/golang-api-hexagonal/internal/observability"
	"github.com/iruldev/golang-api-hexagonal/internal/runtimeutil"
	"github.com/iruldev/golang-api-hexagonal/internal/worker"
)

// QueueStatsResponse wraps queue stats for JSON response.
type QueueStatsResponse struct {
	Aggregate runtimeutil.AggregateStats `json:"aggregate"`
	Queues    []runtimeutil.QueueInfo    `json:"queues"`
}

// JobListResponse wraps job list for JSON response.
type JobListResponse struct {
	Jobs       []runtimeutil.JobInfo  `json:"jobs"`
	Pagination runtimeutil.Pagination `json:"pagination"`
}

// FailedJobListResponse wraps failed job list for JSON response.
type FailedJobListResponse struct {
	FailedJobs []runtimeutil.FailedJobInfo `json:"failed_jobs"`
	Pagination runtimeutil.Pagination      `json:"pagination"`
}

// RetryJobResponse is the response for retry operations.
type RetryJobResponse struct {
	Message string `json:"message"`
	TaskID  string `json:"task_id"`
	Queue   string `json:"queue"`
}

// DeleteJobResponse is the response for delete operations.
type DeleteJobResponse struct {
	Message string `json:"message"`
	TaskID  string `json:"task_id"`
	Queue   string `json:"queue"`
}

// QueuesHandler provides HTTP handlers for job queue inspection.
// Requires admin role (enforced at route level via RBAC middleware).
type QueuesHandler struct {
	inspector runtimeutil.QueueInspector
	logger    *zap.Logger
}

// NewQueuesHandler creates a new QueuesHandler.
// The inspector is required; logger is optional (can be nil).
func NewQueuesHandler(inspector runtimeutil.QueueInspector, logger *zap.Logger) *QueuesHandler {
	return &QueuesHandler{
		inspector: inspector,
		logger:    logger,
	}
}

// GetQueueStats handles GET /admin/queues/stats
// Returns aggregate and per-queue statistics.
func (h *QueuesHandler) GetQueueStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.inspector.GetQueueStats(r.Context())
	if err != nil {
		log.Printf("Failed to get queue stats: %v", err)
		response.Error(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to get queue stats")
		return
	}

	response.Success(w, QueueStatsResponse{
		Aggregate: stats.Aggregate,
		Queues:    stats.Queues,
	})
}

// ListJobs handles GET /admin/queues/{queue}/jobs
// Returns jobs in a specific queue with pagination.
func (h *QueuesHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	queueName := chi.URLParam(r, "queue")

	if !worker.IsValidQueue(queueName) {
		response.Error(w, http.StatusBadRequest, "ERR_BAD_REQUEST", "Invalid queue name")
		return
	}

	page, pageSize := parsePagination(r)

	jobs, err := h.inspector.GetJobsInQueue(r.Context(), queueName, page, pageSize)
	if err != nil {
		log.Printf("Failed to list jobs for queue %s: %v", queueName, err)
		response.Error(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to list jobs")
		return
	}

	response.Success(w, JobListResponse{
		Jobs:       jobs.Jobs,
		Pagination: jobs.Pagination,
	})
}

// ListFailedJobs handles GET /admin/queues/{queue}/failed
// Returns failed jobs in a specific queue with pagination.
func (h *QueuesHandler) ListFailedJobs(w http.ResponseWriter, r *http.Request) {
	queueName := chi.URLParam(r, "queue")

	if !worker.IsValidQueue(queueName) {
		response.Error(w, http.StatusBadRequest, "ERR_BAD_REQUEST", "Invalid queue name")
		return
	}

	page, pageSize := parsePagination(r)

	failedJobs, err := h.inspector.GetFailedJobs(r.Context(), queueName, page, pageSize)
	if err != nil {
		log.Printf("Failed to list failed jobs for queue %s: %v", queueName, err)
		response.Error(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to list failed jobs")
		return
	}

	response.Success(w, FailedJobListResponse{
		FailedJobs: failedJobs.FailedJobs,
		Pagination: failedJobs.Pagination,
	})
}

// DeleteFailedJob handles DELETE /admin/queues/{queue}/failed/{task_id}
// Deletes a failed task from the queue.
func (h *QueuesHandler) DeleteFailedJob(w http.ResponseWriter, r *http.Request) {
	queueName := chi.URLParam(r, "queue")
	taskID := chi.URLParam(r, "task_id")

	if !worker.IsValidQueue(queueName) {
		response.Error(w, http.StatusBadRequest, "ERR_BAD_REQUEST", "Invalid queue name")
		return
	}

	if taskID == "" {
		response.Error(w, http.StatusBadRequest, "ERR_BAD_REQUEST", "Task ID is required")
		return
	}

	err := h.inspector.DeleteFailedJob(r.Context(), queueName, taskID)
	if err != nil {
		if err == runtimeutil.ErrTaskNotFound {
			response.Error(w, http.StatusNotFound, "ERR_NOT_FOUND", "Task not found")
			return
		}
		log.Printf("Failed to delete task %s from queue %s: %v", taskID, queueName, err)
		response.Error(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to delete task")
		return
	}

	// Audit logging
	h.logQueueAction(r, queueName, taskID, "delete_failed_job")

	response.Success(w, DeleteJobResponse{
		Message: "Task deleted successfully",
		TaskID:  taskID,
		Queue:   queueName,
	})
}

// RetryFailedJob handles POST /admin/queues/{queue}/failed/{task_id}/retry
// Requeues a failed task for retry.
func (h *QueuesHandler) RetryFailedJob(w http.ResponseWriter, r *http.Request) {
	queueName := chi.URLParam(r, "queue")
	taskID := chi.URLParam(r, "task_id")

	if !worker.IsValidQueue(queueName) {
		response.Error(w, http.StatusBadRequest, "ERR_BAD_REQUEST", "Invalid queue name")
		return
	}

	if taskID == "" {
		response.Error(w, http.StatusBadRequest, "ERR_BAD_REQUEST", "Task ID is required")
		return
	}

	_, err := h.inspector.RetryFailedJob(r.Context(), queueName, taskID)
	if err != nil {
		if err == runtimeutil.ErrTaskNotFound {
			response.Error(w, http.StatusNotFound, "ERR_NOT_FOUND", "Task not found")
			return
		}
		log.Printf("Failed to retry task %s in queue %s: %v", taskID, queueName, err)
		response.Error(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to retry task")
		return
	}

	// Audit logging
	h.logQueueAction(r, queueName, taskID, "retry_failed_job")

	response.Success(w, RetryJobResponse{
		Message: "Task requeued for retry",
		TaskID:  taskID,
		Queue:   queueName,
	})
}

// parsePagination extracts page and page_size from query parameters.
func parsePagination(r *http.Request) (page, pageSize int) {
	page = 1
	pageSize = 20

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 {
			pageSize = parsed
			if pageSize > 100 {
				pageSize = 100
			}
		}
	}
	return page, pageSize
}

// logQueueAction logs a queue action event with audit logging.
func (h *QueuesHandler) logQueueAction(r *http.Request, queue, taskID, actionType string) {
	// Get actor from claims
	claims, err := middleware.FromContext(r.Context())
	actorID := claims.UserID
	if actorID == "" {
		actorID = "unknown"
	}
	if err != nil {
		log.Printf("Warning: could not get claims from context for audit log: %v", err)
	}

	// Determine action type
	action := observability.ActionUpdate
	if actionType == "delete_failed_job" {
		action = observability.ActionDelete
	}

	auditEvent := observability.NewAuditEvent(
		r.Context(),
		action,
		"job_queue:"+queue+":"+taskID,
		actorID,
		map[string]any{
			"action_type": actionType,
			"queue":       queue,
			"task_id":     taskID,
		},
	)
	observability.LogAudit(r.Context(), h.logger, auditEvent)
}
