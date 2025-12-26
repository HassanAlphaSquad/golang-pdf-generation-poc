package pdfgen

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

var (
	ErrInvalidHTML       = errors.New("invalid or empty HTML content")
	ErrInvalidOutputPath = errors.New("invalid output path")
)

type Generator struct {
	timeout time.Duration
}

func NewGenerator(timeout time.Duration) *Generator {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &Generator{timeout: timeout}
}

func (g *Generator) Close() {}

func (g *Generator) validateHTML(html string) error {
	if strings.TrimSpace(html) == "" {
		return ErrInvalidHTML
	}
	return nil
}

func (g *Generator) validateOutputPath(outputPath string) error {
	if strings.TrimSpace(outputPath) == "" {
		return ErrInvalidOutputPath
	}

	dir := filepath.Dir(outputPath)
	if dir != "." && dir != "/" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	if !strings.HasSuffix(strings.ToLower(outputPath), ".pdf") {
		return fmt.Errorf("%w: file must have .pdf extension", ErrInvalidOutputPath)
	}

	return nil
}

func (g *Generator) FromHTML(html string, outputPath string) error {
	return g.FromHTMLWithCustomOptions(html, outputPath, DefaultPrintOptions())
}

func (g *Generator) FromHTMLWithCustomOptions(html string, outputPath string, opts *PrintOptions) error {
	if err := g.validateHTML(html); err != nil {
		return err
	}
	if err := g.validateOutputPath(outputPath); err != nil {
		return err
	}

	if opts == nil {
		opts = DefaultPrintOptions()
	}

	return g.generatePDF(html, outputPath, opts.ToCDPParams(), opts.WaitBeforePrint)
}

func (g *Generator) generatePDF(html string, outputPath string, opts *page.PrintToPDFParams, waitTime time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()

	// Allocator options
	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.NoSandbox,
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, options...)
	defer allocCancel()

	taskCtx, taskCancel := chromedp.NewContext(allocCtx)
	defer taskCancel()

	if opts == nil {
		opts = page.PrintToPDF().WithPrintBackground(true)
	}

	if waitTime == 0 {
		waitTime = 1 * time.Second
	}

	var pdfData []byte

	if err := chromedp.Run(taskCtx,
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}
			return page.SetDocumentContent(frameTree.Frame.ID, html).Do(ctx)
		}),
		chromedp.Sleep(waitTime),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfData, _, err = opts.Do(ctx)
			return err
		}),
	); err != nil {
		return fmt.Errorf("failed to generate PDF: %w", err)
	}

	if len(pdfData) == 0 {
		return errors.New("PDF generation resulted in empty file")
	}

	return os.WriteFile(outputPath, pdfData, 0644)
}

func (g *Generator) FromURL(url string, outputPath string) error {
	return g.FromURLWithCustomOptions(url, outputPath, DefaultPrintOptions())
}

func (g *Generator) FromURLWithCustomOptions(url string, outputPath string, opts *PrintOptions) error {
	if strings.TrimSpace(url) == "" {
		return errors.New("URL cannot be empty")
	}
	if err := g.validateOutputPath(outputPath); err != nil {
		return err
	}

	if opts == nil {
		opts = DefaultPrintOptions()
	}

	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()

	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.NoSandbox,
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, options...)
	defer allocCancel()

	taskCtx, taskCancel := chromedp.NewContext(allocCtx)
	defer taskCancel()

	waitTime := opts.WaitBeforePrint
	if waitTime == 0 {
		waitTime = 2 * time.Second
	}

	var pdfData []byte

	if err := chromedp.Run(taskCtx,
		chromedp.Navigate(url),
		chromedp.Sleep(waitTime),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfData, _, err = opts.ToCDPParams().Do(ctx)
			return err
		}),
	); err != nil {
		return fmt.Errorf("failed to generate PDF from URL: %w", err)
	}

	if len(pdfData) == 0 {
		return errors.New("PDF generation resulted in empty file")
	}

	return os.WriteFile(outputPath, pdfData, 0644)
}

func (g *Generator) FromFile(htmlPath string, outputPath string) error {
	return g.FromFileWithCustomOptions(htmlPath, outputPath, DefaultPrintOptions())
}

func (g *Generator) FromFileWithCustomOptions(htmlPath string, outputPath string, opts *PrintOptions) error {
	htmlContent, err := os.ReadFile(htmlPath)
	if err != nil {
		return fmt.Errorf("failed to read HTML file %s: %w", htmlPath, err)
	}

	if opts == nil {
		opts = DefaultPrintOptions()
	}

	return g.FromHTMLWithCustomOptions(string(htmlContent), outputPath, opts)
}
