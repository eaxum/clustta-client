package handlers

import (
	"errors"
	"sync"
	"time"

	"clustta/services"

	"github.com/google/uuid"
)

const (
	jobQueued     = "queued"
	jobRunning    = "running"
	jobCancelling = "cancelling"
	jobCancelled  = "cancelled"
	jobSucceeded  = "succeeded"
	jobFailed     = "failed"
)

type jobResponse struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Status      string `json:"status"`
	Message     string `json:"message,omitempty"`
	Progress    int    `json:"progress"`
	Error       string `json:"error,omitempty"`
	Result      any    `json:"result,omitempty"`
	Cancellable bool   `json:"cancellable"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type bridgeJob struct {
	jobResponse
	idempotencyKey  string
	cancelRequested bool
}

type jobOperation func(update func(message string, progress int)) (any, error)

var (
	jobsMu           sync.RWMutex
	jobs             = map[string]*bridgeJob{}
	idempotentJobs   = map[string]string{}
	jobOperationMu   sync.Mutex
	errJobNotFound   = errors.New("job not found")
	errNotCancelable = errors.New("job cannot be cancelled")
)

func startJob(kind, idempotencyKey string, cancellable bool, operation jobOperation) jobResponse {
	jobsMu.Lock()
	if idempotencyKey != "" {
		if existingID, ok := idempotentJobs[idempotencyKey]; ok {
			existing := jobs[existingID].jobResponse
			jobsMu.Unlock()
			return existing
		}
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	job := &bridgeJob{
		jobResponse: jobResponse{
			ID:          uuid.NewString(),
			Kind:        kind,
			Status:      jobQueued,
			Cancellable: cancellable,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		idempotencyKey: idempotencyKey,
	}
	jobs[job.ID] = job
	if idempotencyKey != "" {
		idempotentJobs[idempotencyKey] = job.ID
	}
	response := job.jobResponse
	jobsMu.Unlock()

	go runJob(job.ID, operation)
	return response
}

func runJob(jobID string, operation jobOperation) {
	jobOperationMu.Lock()
	defer jobOperationMu.Unlock()

	jobsMu.Lock()
	job, ok := jobs[jobID]
	if !ok || job.cancelRequested {
		if ok {
			setJobStatus(job, jobCancelled)
		}
		jobsMu.Unlock()
		return
	}
	setJobStatus(job, jobRunning)
	jobsMu.Unlock()

	result, err := operation(func(message string, progress int) {
		updateJob(jobID, message, progress)
	})

	jobsMu.Lock()
	defer jobsMu.Unlock()
	job, ok = jobs[jobID]
	if !ok {
		return
	}
	if job.cancelRequested {
		setJobStatus(job, jobCancelled)
		return
	}
	if err != nil {
		job.Error = err.Error()
		setJobStatus(job, jobFailed)
		return
	}
	job.Result = result
	job.Progress = 100
	setJobStatus(job, jobSucceeded)
}

func updateJob(jobID, message string, progress int) {
	jobsMu.Lock()
	defer jobsMu.Unlock()
	job, ok := jobs[jobID]
	if !ok || job.Status != jobRunning {
		return
	}
	if progress < 0 {
		progress = 0
	}
	if progress > 99 {
		progress = 99
	}
	job.Message = message
	job.Progress = progress
	job.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
}

func getJob(jobID string) (jobResponse, error) {
	jobsMu.RLock()
	defer jobsMu.RUnlock()
	job, ok := jobs[jobID]
	if !ok {
		return jobResponse{}, errJobNotFound
	}
	return job.jobResponse, nil
}

func cancelJob(jobID string) (jobResponse, error) {
	jobsMu.Lock()
	job, ok := jobs[jobID]
	if !ok {
		jobsMu.Unlock()
		return jobResponse{}, errJobNotFound
	}
	if !job.Cancellable {
		jobsMu.Unlock()
		return jobResponse{}, errNotCancelable
	}
	if job.Status == jobSucceeded || job.Status == jobFailed || job.Status == jobCancelled {
		response := job.jobResponse
		jobsMu.Unlock()
		return response, nil
	}

	job.cancelRequested = true
	if job.Status == jobQueued {
		setJobStatus(job, jobCancelled)
	} else {
		setJobStatus(job, jobCancelling)
	}
	response := job.jobResponse
	jobsMu.Unlock()

	if response.Status == jobCancelling {
		syncService := &services.SyncService{}
		syncService.CancelSync()
	}
	return response, nil
}

func setJobStatus(job *bridgeJob, status string) {
	job.Status = status
	job.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
}
