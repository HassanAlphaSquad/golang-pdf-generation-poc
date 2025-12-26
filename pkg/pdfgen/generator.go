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
	return &Generator{
		timeout: timeout,
	}
}

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

	return g.generateFromHTML(html, outputPath, opts.ToCDPParams(), opts.WaitBeforePrint)
}

func (g *Generator) generateFromHTML(html string, outputPath string, opts *page.PrintToPDFParams, waitTime time.Duration) error {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, g.timeout)
	defer cancel()

	if opts == nil {
		opts = page.PrintToPDF().WithPrintBackground(true)
	}

	var pdfBuf []byte
	actions := []chromedp.Action{
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return fmt.Errorf("failed to get frame tree: %w", err)
			}
			return page.SetDocumentContent(frameTree.Frame.ID, html).Do(ctx)
		}),
	}

	if waitTime > 0 {
		actions = append(actions, chromedp.Sleep(waitTime))
	}

	actions = append(actions, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		pdfBuf, _, err = opts.Do(ctx)
		if err != nil {
			return fmt.Errorf("failed to generate PDF: %w", err)
		}
		return nil
	}))

	if err := chromedp.Run(ctx, actions...); err != nil {
		return fmt.Errorf("chromedp execution failed: %w", err)
	}

	if len(pdfBuf) == 0 {
		return errors.New("generated PDF buffer is empty")
	}

	if err := os.WriteFile(outputPath, pdfBuf, 0644); err != nil {
		return fmt.Errorf("failed to write PDF file: %w", err)
	}

	return nil
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

	return g.generateFromURL(url, outputPath, opts.ToCDPParams(), opts.WaitBeforePrint)
}

func (g *Generator) generateFromURL(url string, outputPath string, opts *page.PrintToPDFParams, waitTime time.Duration) error {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, g.timeout)
	defer cancel()

	if opts == nil {
		opts = page.PrintToPDF().WithPrintBackground(true)
	}

	var pdfBuf []byte
	actions := []chromedp.Action{
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
	}

	if waitTime > 0 {
		actions = append(actions, chromedp.Sleep(waitTime))
	}

	actions = append(actions, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		pdfBuf, _, err = opts.Do(ctx)
		if err != nil {
			return fmt.Errorf("failed to generate PDF: %w", err)
		}
		return nil
	}))

	if err := chromedp.Run(ctx, actions...); err != nil {
		return fmt.Errorf("chromedp execution failed: %w", err)
	}

	if len(pdfBuf) == 0 {
		return errors.New("generated PDF buffer is empty")
	}

	if err := os.WriteFile(outputPath, pdfBuf, 0644); err != nil {
		return fmt.Errorf("failed to write PDF file: %w", err)
	}

	return nil
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
