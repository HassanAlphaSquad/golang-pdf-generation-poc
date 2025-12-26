package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"

	_ "github.com/HassanAlphaSquad/golang-pdf-generation-poc/docs"
	"github.com/HassanAlphaSquad/golang-pdf-generation-poc/internal/api/handlers"
	"github.com/HassanAlphaSquad/golang-pdf-generation-poc/internal/api/middleware"
	"github.com/HassanAlphaSquad/golang-pdf-generation-poc/internal/storage"
	"github.com/HassanAlphaSquad/golang-pdf-generation-poc/pkg/pdfgen"
)

// @title PDF Generation API
// @version 1.0
// @description PDF generation API using headless Chrome
// @host localhost:3000
// @BasePath /
// @schemes http
// @tag.name PDF
// @tag.description PDF generation endpoints
// @tag.name Health
// @tag.description Health check endpoints

const (
	Version          = "1.0.0"
	DefaultPort      = "3000"
	DefaultOutputDir = "./output"
)

func main() {
	port := getEnv("PORT", DefaultPort)
	outputDir := getEnv("OUTPUT_DIR", DefaultOutputDir)

	store, err := storage.NewJobStore(outputDir)
	if err != nil {
		log.Fatalf("Failed to create job store: %v", err)
	}

	generator := pdfgen.NewGenerator(60 * time.Second)

	app := fiber.New(fiber.Config{
		AppName:               "PDF Generation API",
		ErrorHandler:          middleware.ErrorHandler(),
		DisableStartupMessage: false,
		ServerHeader:          "PDF-API",
		BodyLimit:             10 * 1024 * 1024,
		ReadTimeout:           30 * time.Second,
		WriteTimeout:          60 * time.Second,
	})

	app.Use(recover.New())
	app.Use(middleware.CORS())
	app.Use(middleware.RequestLogger())
	app.Use(compress.New())

	pdfHandler := handlers.NewPDFHandler(generator, store)
	healthHandler := handlers.NewHealthHandler(store, Version)

	app.Get("/health", healthHandler.HealthCheck)
	app.Get("/swagger/*", swagger.HandlerDefault)

	v1 := app.Group("/api")
	pdf := v1.Group("/pdf")
	pdf.Post("/generate", pdfHandler.GeneratePDF)
	pdf.Post("/generate/url", pdfHandler.GenerateFromURL)
	pdf.Get("/status/:id", pdfHandler.GetJobStatus)
	pdf.Get("/download/:id", pdfHandler.DownloadPDF)
	pdf.Get("/jobs", pdfHandler.ListJobs)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": "PDF Generation API",
			"version": Version,
			"status":  "running",
			"docs":    "/swagger/index.html",
		})
	})

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			removed := store.CleanupOldJobs(24 * time.Hour)
			if removed > 0 {
				log.Printf("Cleaned up %d old jobs", removed)
			}
		}
	}()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")

		// Close browser instance
		generator.Close()

		if err := app.Shutdown(); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
		log.Println("Server stopped")
	}()

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting PDF Generation API v%s on %s", Version, addr)
	log.Printf("Swagger docs: http://localhost:%s/swagger/index.html", port)
	log.Printf("Health check: http://localhost:%s/health", port)

	if err := app.Listen(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
