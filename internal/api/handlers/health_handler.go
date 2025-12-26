package handlers

import (
	"fmt"
	"time"

	"github.com/HassanAlphaSquad/golang-pdf-generation-poc/internal/api/models"
	"github.com/HassanAlphaSquad/golang-pdf-generation-poc/internal/storage"
	"github.com/gofiber/fiber/v2"
)

type HealthHandler struct {
	store   *storage.JobStore
	version string
}

func NewHealthHandler(store *storage.JobStore, version string) *HealthHandler {
	return &HealthHandler{
		store:   store,
		version: version,
	}
}

// @Summary 		Health check
// @Description 	Get the health status of the service
// @Tags 			Health
// @Produce 		json
// @Success 		200 {object} models.HealthResponse
// @Router 			/health [get]
func (h *HealthHandler) HealthCheck(c *fiber.Ctx) error {
	stats := h.store.GetStats()
	services := map[string]string{
		"pdf_generator":   "healthy",
		"job_store":       "healthy",
		"total_jobs":      fmt.Sprintf("%d", stats["total"]),
		"pending_jobs":    fmt.Sprintf("%d", stats["pending"]),
		"processing_jobs": fmt.Sprintf("%d", stats["processing"]),
		"completed_jobs":  fmt.Sprintf("%d", stats["completed"]),
		"failed_jobs":     fmt.Sprintf("%d", stats["failed"]),
	}

	return c.JSON(models.HealthResponse{
		Status:    "healthy",
		Version:   h.version,
		Timestamp: time.Now(),
		Services:  services,
	})
}
