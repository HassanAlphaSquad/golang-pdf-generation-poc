package models

import "time"

type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

type GeneratePDFRequest struct {
	HTML     string        `json:"html"`
	Filename string        `json:"filename"`
	Options  *PrintOptions `json:"options,omitempty"`
}

type GenerateFromURLRequest struct {
	URL      string        `json:"url"`
	Filename string        `json:"filename"`
	Options  *PrintOptions `json:"options,omitempty"`
}

type PrintOptions struct {
	Landscape       bool    `json:"landscape"`
	PageSize        string  `json:"page_size"`
	MarginTop       float64 `json:"margin_top"`
	MarginBottom    float64 `json:"margin_bottom"`
	MarginLeft      float64 `json:"margin_left"`
	MarginRight     float64 `json:"margin_right"`
	PrintBackground bool    `json:"print_background"`
	Scale           float64 `json:"scale"`
}

type GeneratePDFResponse struct {
	JobID     string    `json:"job_id"`
	Status    JobStatus `json:"status"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type JobStatusResponse struct {
	JobID        string     `json:"job_id"`
	Status       JobStatus  `json:"status"`
	Filename     string     `json:"filename,omitempty"`
	FileSize     int64      `json:"file_size,omitempty"`
	DownloadURL  string     `json:"download_url,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
	Progress     int        `json:"progress"`
	CreatedAt    time.Time  `json:"created_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type HealthResponse struct {
	Status    string            `json:"status"`
	Version   string            `json:"version"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

type ListJobsResponse struct {
	Jobs       []JobStatusResponse `json:"jobs"`
	Total      int                 `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int                 `json:"total_pages"`
}
