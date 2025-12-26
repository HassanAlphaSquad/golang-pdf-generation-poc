package pdfgen

import (
	"time"

	"github.com/chromedp/cdproto/page"
)

type PageSize struct {
	Width  float64
	Height float64
}

var (
	PageSizeA4     = PageSize{Width: 8.27, Height: 11.69}
	PageSizeLetter = PageSize{Width: 8.5, Height: 11.0}
	PageSizeLegal  = PageSize{Width: 8.5, Height: 14.0}
)

type PrintOptions struct {
	PageSize          PageSize
	Landscape         bool
	PrintBackground   bool
	MarginTop         float64
	MarginBottom      float64
	MarginLeft        float64
	MarginRight       float64
	Scale             float64
	PageRanges        string
	GenerateTaggedPDF bool
	WaitBeforePrint   time.Duration
}

func DefaultPrintOptions() *PrintOptions {
	return &PrintOptions{
		PageSize:          PageSizeA4,
		Landscape:         false,
		PrintBackground:   true,
		MarginTop:         0.4,
		MarginBottom:      0.4,
		MarginLeft:        0.4,
		MarginRight:       0.4,
		Scale:             1.0,
		GenerateTaggedPDF: false,
		WaitBeforePrint:   0,
	}
}

func (o *PrintOptions) ToCDPParams() *page.PrintToPDFParams {
	params := page.PrintToPDF().
		WithPrintBackground(o.PrintBackground).
		WithLandscape(o.Landscape).
		WithMarginTop(o.MarginTop).
		WithMarginBottom(o.MarginBottom).
		WithMarginLeft(o.MarginLeft).
		WithMarginRight(o.MarginRight).
		WithScale(o.Scale).
		WithGenerateTaggedPDF(o.GenerateTaggedPDF)

	if o.PageSize.Width > 0 && o.PageSize.Height > 0 {
		params = params.WithPaperWidth(o.PageSize.Width).WithPaperHeight(o.PageSize.Height)
	}

	if o.PageRanges != "" {
		params = params.WithPageRanges(o.PageRanges)
	}

	return params
}
