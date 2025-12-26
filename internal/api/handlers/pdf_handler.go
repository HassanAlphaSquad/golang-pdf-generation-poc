package handlers

import (
	"fmt"
	"log"

	"github.com/HassanAlphaSquad/golang-pdf-generation-poc/internal/api/models"
	"github.com/HassanAlphaSquad/golang-pdf-generation-poc/internal/storage"
	"github.com/HassanAlphaSquad/golang-pdf-generation-poc/pkg/pdfgen"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type PDFHandler struct {
	generator *pdfgen.Generator
	store     *storage.JobStore
}

func NewPDFHandler(generator *pdfgen.Generator, store *storage.JobStore) *PDFHandler {
	return &PDFHandler{
		generator: generator,
		store:     store,
	}
}

// @Summary Generate PDF from HTML
// @Description Generate a PDF document from HTML content
// @Tags PDF
// @Accept json
// @Produce json
// @Param request body models.GeneratePDFRequest true "PDF generation request"
// @Success 202 {object} models.GeneratePDFResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/pdf/generate [post]
func (h *PDFHandler) GeneratePDF(c *fiber.Ctx) error {
	var req models.GeneratePDFRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
			Code:    fiber.StatusBadRequest,
		})
	}

	if req.HTML == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Validation failed",
			Message: "HTML content is required",
			Code:    fiber.StatusBadRequest,
		})
	}

	jobID := uuid.New().String()
	job := h.store.CreateJob(jobID, req.HTML, req.Filename, req.Options)

	go h.processJob(job)

	return c.Status(fiber.StatusAccepted).JSON(models.GeneratePDFResponse{
		JobID:     jobID,
		Status:    models.JobStatusPending,
		Message:   "PDF generation job queued successfully",
		CreatedAt: job.CreatedAt,
	})
}

// @Summary Generate PDF from URL
// @Description Generate a PDF document from a web URL
// @Tags PDF
// @Accept json
// @Produce json
// @Param request body models.GenerateFromURLRequest true "PDF generation from URL request"
// @Success 202 {object} models.GeneratePDFResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/pdf/generate/url [post]
func (h *PDFHandler) GenerateFromURL(c *fiber.Ctx) error {
	var req models.GenerateFromURLRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
			Code:    fiber.StatusBadRequest,
		})
	}

	if req.URL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Validation failed",
			Message: "URL is required",
			Code:    fiber.StatusBadRequest,
		})
	}

	jobID := uuid.New().String()
	html := fmt.Sprintf("URL:%s", req.URL)
	job := h.store.CreateJob(jobID, html, req.Filename, req.Options)

	go h.processURLJob(job, req.URL)

	return c.Status(fiber.StatusAccepted).JSON(models.GeneratePDFResponse{
		JobID:     jobID,
		Status:    models.JobStatusPending,
		Message:   "PDF generation from URL queued successfully",
		CreatedAt: job.CreatedAt,
	})
}

// @Summary Get job status
// @Description Get the status of a PDF generation job
// @Tags PDF
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} models.JobStatusResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/pdf/status/{id} [get]
func (h *PDFHandler) GetJobStatus(c *fiber.Ctx) error {
	jobID := c.Params("id")

	job, err := h.store.GetJob(jobID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error:   "Not found",
			Message: err.Error(),
			Code:    fiber.StatusNotFound,
		})
	}

	response := models.JobStatusResponse{
		JobID:        job.ID,
		Status:       job.Status,
		Filename:     job.Filename,
		FileSize:     job.FileSize,
		ErrorMessage: job.ErrorMessage,
		Progress:     job.Progress,
		CreatedAt:    job.CreatedAt,
		CompletedAt:  job.CompletedAt,
	}

	if job.Status == models.JobStatusCompleted {
		response.DownloadURL = fmt.Sprintf("/api/pdf/download/%s", job.ID)
	}

	return c.JSON(response)
}

// @Summary Download generated PDF
// @Description Download a completed PDF file
// @Tags PDF
// @Produce application/pdf
// @Param id path string true "Job ID"
// @Success 200 {file} binary
// @Failure 404 {object} models.ErrorResponse
// @Router /api/pdf/download/{id} [get]
func (h *PDFHandler) DownloadPDF(c *fiber.Ctx) error {
	jobID := c.Params("id")

	filePath, err := h.store.GetFilePath(jobID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error:   "Not found",
			Message: err.Error(),
			Code:    fiber.StatusNotFound,
		})
	}
	defer h.store.ReleaseFile(jobID)

	job, _ := h.store.GetJob(jobID)
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", job.Filename))

	return c.SendFile(filePath)
}

// @Summary List all jobs
// @Description Get a paginated list of all PDF generation jobs
// @Tags PDF
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} models.ListJobsResponse
// @Router /api/pdf/jobs [get]
func (h *PDFHandler) ListJobs(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 20)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	jobs, total := h.store.ListJobs(page, pageSize)

	jobStatuses := make([]models.JobStatusResponse, 0, len(jobs))
	for _, job := range jobs {
		status := models.JobStatusResponse{
			JobID:        job.ID,
			Status:       job.Status,
			Filename:     job.Filename,
			FileSize:     job.FileSize,
			ErrorMessage: job.ErrorMessage,
			Progress:     job.Progress,
			CreatedAt:    job.CreatedAt,
			CompletedAt:  job.CompletedAt,
		}

		if job.Status == models.JobStatusCompleted {
			status.DownloadURL = fmt.Sprintf("/api/pdf/download/%s", job.ID)
		}

		jobStatuses = append(jobStatuses, status)
	}

	totalPages := (total + pageSize - 1) / pageSize

	return c.JSON(models.ListJobsResponse{
		Jobs:       jobStatuses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

func (h *PDFHandler) processJob(job *storage.Job) {
	h.store.UpdateJobStatus(job.ID, models.JobStatusProcessing, "")

	opts := convertPrintOptions(job.Options)
	err := h.generator.FromHTMLWithCustomOptions(job.HTML, job.FilePath, opts)

	if err != nil {
		log.Printf("Job %s failed: %v", job.ID, err)
		h.store.UpdateJobStatus(job.ID, models.JobStatusFailed, err.Error())
	} else {
		log.Printf("Job %s completed successfully", job.ID)
		h.store.UpdateJobStatus(job.ID, models.JobStatusCompleted, "")
	}
}

func (h *PDFHandler) processURLJob(job *storage.Job, url string) {
	h.store.UpdateJobStatus(job.ID, models.JobStatusProcessing, "")

	opts := convertPrintOptions(job.Options)
	err := h.generator.FromURLWithCustomOptions(url, job.FilePath, opts)

	if err != nil {
		log.Printf("URL Job %s failed: %v", job.ID, err)
		h.store.UpdateJobStatus(job.ID, models.JobStatusFailed, err.Error())
	} else {
		log.Printf("URL Job %s completed successfully", job.ID)
		h.store.UpdateJobStatus(job.ID, models.JobStatusCompleted, "")
	}
}

func convertPrintOptions(opts *models.PrintOptions) *pdfgen.PrintOptions {
	if opts == nil {
		return pdfgen.DefaultPrintOptions()
	}

	pdfOpts := pdfgen.DefaultPrintOptions()
	pdfOpts.Landscape = opts.Landscape
	pdfOpts.PrintBackground = opts.PrintBackground
	pdfOpts.MarginTop = opts.MarginTop
	pdfOpts.MarginBottom = opts.MarginBottom
	pdfOpts.MarginLeft = opts.MarginLeft
	pdfOpts.MarginRight = opts.MarginRight
	pdfOpts.Scale = opts.Scale

	switch opts.PageSize {
	case "A4":
		pdfOpts.PageSize = pdfgen.PageSizeA4
	case "Letter":
		pdfOpts.PageSize = pdfgen.PageSizeLetter
	case "Legal":
		pdfOpts.PageSize = pdfgen.PageSizeLegal
	default:
		pdfOpts.PageSize = pdfgen.PageSizeA4
	}

	return pdfOpts
}
