package cardrender

import (
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/go-rod/rod/lib/proto"
)

const (
	previewScale  = 2
	printScale    = 4
	bleedMM       = 3
	mmPerInch     = 25.4
	cssPixPerInch = 96.0
)

type RenderResult struct {
	PNGBase64 string
	PDFBase64 string
}

type Renderer struct {
	pool *ChromePool
}

func NewRenderer(pool *ChromePool) *Renderer {
	return &Renderer{pool: pool}
}

func BleedPx() int {
	return int(math.Round(float64(bleedMM) / mmPerInch * cssPixPerInch))
}

func (r *Renderer) RenderPNG(html string, width, height int) (string, error) {
	return r.screenshot(html, width, height, previewScale)
}

func (r *Renderer) Render(html string, width, height int) (*RenderResult, error) {
	pngBase64, err := r.screenshot(html, width, height, previewScale)
	if err != nil {
		return nil, err
	}

	pdfBase64, err := r.printPDF(html, width, height)
	if err != nil {
		return nil, err
	}

	return &RenderResult{PNGBase64: pngBase64, PDFBase64: pdfBase64}, nil
}

func (r *Renderer) RenderPrintPDF(html string, width, height int) (string, error) {
	return r.printPDF(html, width, height)
}

func (r *Renderer) screenshot(html string, width, height int, scale float64) (string, error) {
	page, err := r.pool.GetBrowser().Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return "", fmt.Errorf("create page: %w", err)
	}
	defer page.Close()

	if err := page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width: width, Height: height, DeviceScaleFactor: scale,
	}); err != nil {
		return "", fmt.Errorf("set viewport: %w", err)
	}

	if err := page.Navigate("about:blank"); err != nil {
		return "", fmt.Errorf("navigate: %w", err)
	}
	page.MustWaitStable()

	if err := page.SetDocumentContent(html); err != nil {
		return "", fmt.Errorf("set content: %w", err)
	}
	page.MustWaitStable()

	pngBytes, err := page.Screenshot(true, &proto.PageCaptureScreenshot{
		Format: proto.PageCaptureScreenshotFormatPng,
		Clip: &proto.PageViewport{
			X: 0, Y: 0, Width: float64(width), Height: float64(height), Scale: 1,
		},
		FromSurface: true, CaptureBeyondViewport: false,
	})
	if err != nil {
		return "", fmt.Errorf("screenshot: %w", err)
	}

	return base64.StdEncoding.EncodeToString(pngBytes), nil
}

func (r *Renderer) printPDF(html string, width, height int) (string, error) {
	bleed := BleedPx()
	fullW := width + 2*bleed
	fullH := height + 2*bleed

	printHTML := injectBleed(html, bleed, fullW, fullH)

	page, err := r.pool.GetBrowser().Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return "", fmt.Errorf("create page: %w", err)
	}
	defer page.Close()

	if err := page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width: fullW, Height: fullH, DeviceScaleFactor: printScale,
	}); err != nil {
		return "", fmt.Errorf("set viewport: %w", err)
	}

	if err := page.Navigate("about:blank"); err != nil {
		return "", fmt.Errorf("navigate: %w", err)
	}
	page.MustWaitStable()

	if err := page.SetDocumentContent(printHTML); err != nil {
		return "", fmt.Errorf("set content: %w", err)
	}
	page.MustWaitStable()

	paperW := float64(fullW) / cssPixPerInch
	paperH := float64(fullH) / cssPixPerInch

	pdfData, err := page.PDF(&proto.PagePrintToPDF{
		PrintBackground:         true,
		PreferCSSPageSize:       true,
		MarginTop:               floatPtr(0),
		MarginBottom:            floatPtr(0),
		MarginLeft:              floatPtr(0),
		MarginRight:             floatPtr(0),
		PaperWidth:              floatPtr(paperW),
		PaperHeight:             floatPtr(paperH),
		Scale:                   floatPtr(1),
		GenerateDocumentOutline: false,
	})
	if err != nil {
		return "", fmt.Errorf("pdf: %w", err)
	}

	pdfBytes, err := io.ReadAll(pdfData)
	if err != nil {
		return "", fmt.Errorf("read pdf: %w", err)
	}

	return base64.StdEncoding.EncodeToString(pdfBytes), nil
}

func (r *Renderer) RenderMultiPagePDF(frontHTML, insideHTML string, width, height int) (string, error) {
	combinedHTML := buildTwoPageHTML(frontHTML, insideHTML, width, height)

	bleed := BleedPx()
	fullW := width + 2*bleed
	fullH := height + 2*bleed

	page, err := r.pool.GetBrowser().Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return "", fmt.Errorf("create page: %w", err)
	}
	defer page.Close()

	if err := page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width: fullW, Height: fullH, DeviceScaleFactor: printScale,
	}); err != nil {
		return "", fmt.Errorf("set viewport: %w", err)
	}

	if err := page.Navigate("about:blank"); err != nil {
		return "", fmt.Errorf("navigate: %w", err)
	}
	page.MustWaitStable()

	if err := page.SetDocumentContent(combinedHTML); err != nil {
		return "", fmt.Errorf("set content: %w", err)
	}
	page.MustWaitStable()

	paperW := float64(fullW) / cssPixPerInch
	paperH := float64(fullH) / cssPixPerInch

	pdfData, err := page.PDF(&proto.PagePrintToPDF{
		PrintBackground:         true,
		PreferCSSPageSize:       false,
		MarginTop:               floatPtr(0),
		MarginBottom:            floatPtr(0),
		MarginLeft:              floatPtr(0),
		MarginRight:             floatPtr(0),
		PaperWidth:              floatPtr(paperW),
		PaperHeight:             floatPtr(paperH),
		Scale:                   floatPtr(1),
		GenerateDocumentOutline: false,
	})
	if err != nil {
		return "", fmt.Errorf("pdf: %w", err)
	}

	pdfBytes, err := io.ReadAll(pdfData)
	if err != nil {
		return "", fmt.Errorf("read pdf: %w", err)
	}

	return base64.StdEncoding.EncodeToString(pdfBytes), nil
}

func buildTwoPageHTML(frontHTML, insideHTML string, width, height int) string {
	bleed := BleedPx()
	fullW := width + 2*bleed
	fullH := height + 2*bleed

	extractBody := func(html string) string {
		if idx := strings.Index(html, "<body>"); idx >= 0 {
			body := html[idx+6:]
			if end := strings.Index(body, "</body>"); end >= 0 {
				return body[:end]
			}
			return body
		}
		return html
	}

	extractHead := func(html string) string {
		start := strings.Index(html, "<style>")
		end := strings.Index(html, "</style>")
		if start >= 0 && end >= 0 {
			return html[start : end+8]
		}
		return ""
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta charset="utf-8"><style>
* { margin: 0; padding: 0; box-sizing: border-box; }
@page { size: %dpx %dpx; margin: 0; }
.page {
  width: %dpx; height: %dpx;
  padding: %dpx;
  page-break-after: always;
  overflow: hidden;
  position: relative;
}
.page:last-child { page-break-after: auto; }
.page-inner { width: %dpx; height: %dpx; overflow: hidden; position: relative; }
</style>
%s
%s
</head><body>
<div class="page"><div class="page-inner">%s</div></div>
<div class="page"><div class="page-inner">%s</div></div>
</body></html>`,
		fullW, fullH,
		fullW, fullH,
		bleed,
		width, height,
		extractHead(frontHTML),
		extractHead(insideHTML),
		extractBody(frontHTML),
		extractBody(insideHTML),
	)
}

func injectBleed(html string, bleed, fullW, fullH int) string {
	bleedCSS := fmt.Sprintf(`<style>
html, body { margin: 0; padding: 0; width: %dpx; height: %dpx; overflow: hidden; }
body { padding: %dpx; background: inherit; }
.trim-mark { position: fixed; background: #000; z-index: 99999; print-color-adjust: exact; }
.tm-tl-h { top: %dpx; left: 0; width: %dpx; height: 0.25pt; }
.tm-tl-v { top: 0; left: %dpx; width: 0.25pt; height: %dpx; }
.tm-tr-h { top: %dpx; right: 0; width: %dpx; height: 0.25pt; }
.tm-tr-v { top: 0; right: %dpx; width: 0.25pt; height: %dpx; }
.tm-bl-h { bottom: %dpx; left: 0; width: %dpx; height: 0.25pt; }
.tm-bl-v { bottom: 0; left: %dpx; width: 0.25pt; height: %dpx; }
.tm-br-h { bottom: %dpx; right: 0; width: %dpx; height: 0.25pt; }
.tm-br-v { bottom: 0; right: %dpx; width: 0.25pt; height: %dpx; }
</style>`,
		fullW, fullH, bleed,
		bleed, bleed-2,
		bleed, bleed-2,
		bleed, bleed-2,
		bleed, bleed-2,
		bleed, bleed-2,
		bleed, bleed-2,
		bleed, bleed-2,
		bleed, bleed-2,
	)

	trimMarks := `<div class="trim-mark tm-tl-h"></div><div class="trim-mark tm-tl-v"></div>` +
		`<div class="trim-mark tm-tr-h"></div><div class="trim-mark tm-tr-v"></div>` +
		`<div class="trim-mark tm-bl-h"></div><div class="trim-mark tm-bl-v"></div>` +
		`<div class="trim-mark tm-br-h"></div><div class="trim-mark tm-br-v"></div>`

	result := strings.Replace(html, "</head>", bleedCSS+"</head>", 1)
	result = strings.Replace(result, "<body>", "<body>"+trimMarks, 1)
	if !strings.Contains(result, "<body>") {
		result = strings.Replace(result, "</head>", "</head><body>"+trimMarks, 1)
	}

	return result
}

func floatPtr(f float64) *float64 {
	return &f
}
