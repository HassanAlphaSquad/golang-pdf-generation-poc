package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/HassanAlphaSquad/golang-pdf-generation-poc/internal/api/models"
)

type Job struct {
	ID           string
	BatchID      string
	Status       models.JobStatus
	HTML         string
	Filename     string
	FilePath     string
	FileSize     int64
	ErrorMessage string
	Progress     int
	CreatedAt    time.Time
	CompletedAt  *time.Time
	Options      *models.PrintOptions
}

type JobStore struct {
	jobs      map[string]*Job
	batches   map[string][]string
	mu        sync.RWMutex
	outputDir string
}

func NewJobStore(outputDir string) (*JobStore, error) {

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	return &JobStore{
		jobs:      make(map[string]*Job),
		batches:   make(map[string][]string),
		outputDir: outputDir,
	}, nil
}

func (s *JobStore) CreateJob(id, html, filename string, opts *models.PrintOptions) *Job {
	s.mu.Lock()
	defer s.mu.Unlock()

	timestamp := time.Now().Format("20060102_150405")

	if filename == "" {
		filename = fmt.Sprintf("%s_%s.pdf", id, timestamp)
	} else {
		ext := filepath.Ext(filename)
		nameWithoutExt := filename[:len(filename)-len(ext)]
		filename = fmt.Sprintf("%s_%s%s", nameWithoutExt, timestamp, ext)
	}

	job := &Job{
		ID:        id,
		Status:    models.JobStatusPending,
		HTML:      html,
		Filename:  filename,
		FilePath:  filepath.Join(s.outputDir, filename),
		Progress:  0,
		CreatedAt: time.Now(),
		Options:   opts,
	}

	s.jobs[id] = job
	return job
}

func (s *JobStore) CreateBatch(batchID string, jobIDs []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.batches[batchID] = jobIDs

	for _, jobID := range jobIDs {
		if job, exists := s.jobs[jobID]; exists {
			job.BatchID = batchID
		}
	}
}

func (s *JobStore) GetJob(id string) (*Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, exists := s.jobs[id]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", id)
	}

	return job, nil
}

func (s *JobStore) GetBatch(batchID string) ([]*Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobIDs, exists := s.batches[batchID]
	if !exists {
		return nil, fmt.Errorf("batch not found: %s", batchID)
	}

	jobs := make([]*Job, 0, len(jobIDs))
	for _, jobID := range jobIDs {
		if job, exists := s.jobs[jobID]; exists {
			jobs = append(jobs, job)
		}
	}

	return jobs, nil
}

func (s *JobStore) UpdateJobStatus(id string, status models.JobStatus, errorMsg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[id]
	if !exists {
		return fmt.Errorf("job not found: %s", id)
	}

	job.Status = status
	job.ErrorMessage = errorMsg

	switch status {
	case models.JobStatusProcessing:
		job.Progress = 50
	case models.JobStatusCompleted:
		job.Progress = 100
		now := time.Now()
		job.CompletedAt = &now

		if info, err := os.Stat(job.FilePath); err == nil {
			job.FileSize = info.Size()
		}
	case models.JobStatusFailed:
		job.Progress = 0
		now := time.Now()
		job.CompletedAt = &now
	}

	return nil
}

func (s *JobStore) ListJobs(page, pageSize int) ([]*Job, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	allJobs := make([]*Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		allJobs = append(allJobs, job)
	}

	for i := 0; i < len(allJobs)-1; i++ {
		for j := 0; j < len(allJobs)-i-1; j++ {
			if allJobs[j].CreatedAt.Before(allJobs[j+1].CreatedAt) {
				allJobs[j], allJobs[j+1] = allJobs[j+1], allJobs[j]
			}
		}
	}

	total := len(allJobs)

	start := (page - 1) * pageSize
	if start >= total {
		return []*Job{}, total
	}

	end := start + pageSize
	if end > total {
		end = total
	}

	return allJobs[start:end], total
}

func (s *JobStore) CleanupOldJobs(olderThan time.Duration) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-olderThan)
	removed := 0

	for id, job := range s.jobs {
		if job.CreatedAt.Before(cutoff) {

			if job.FilePath != "" {
				os.Remove(job.FilePath)
			}
			delete(s.jobs, id)
			removed++
		}
	}

	for batchID, jobIDs := range s.batches {
		allRemoved := true
		for _, jobID := range jobIDs {
			if _, exists := s.jobs[jobID]; exists {
				allRemoved = false
				break
			}
		}
		if allRemoved {
			delete(s.batches, batchID)
		}
	}

	return removed
}

func (s *JobStore) GetStats() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := map[string]int{
		"total":      len(s.jobs),
		"pending":    0,
		"processing": 0,
		"completed":  0,
		"failed":     0,
		"batches":    len(s.batches),
	}

	for _, job := range s.jobs {
		switch job.Status {
		case models.JobStatusPending:
			stats["pending"]++
		case models.JobStatusProcessing:
			stats["processing"]++
		case models.JobStatusCompleted:
			stats["completed"]++
		case models.JobStatusFailed:
			stats["failed"]++
		}
	}

	return stats
}

func (s *JobStore) GetFilePath(id string) (string, error) {
	job, err := s.GetJob(id)
	if err != nil {
		return "", err
	}

	if job.Status != models.JobStatusCompleted {
		return "", fmt.Errorf("job not completed")
	}

	if _, err := os.Stat(job.FilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found")
	}

	return job.FilePath, nil
}
