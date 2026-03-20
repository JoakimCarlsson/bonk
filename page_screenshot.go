package bonk

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"

	"github.com/joakimcarlsson/bonk/proto"
)

// ScreenshotOption configures screenshot behavior.
type ScreenshotOption func(*screenshotConfig)

type screenshotConfig struct {
	fullPage bool
	quality  int64
}

// FullPage captures the full scrollable page.
func FullPage() ScreenshotOption {
	return func(c *screenshotConfig) {
		c.fullPage = true
	}
}

// ScreenshotQuality sets the JPEG/WebP quality (0-100).
func ScreenshotQuality(q int) ScreenshotOption {
	return func(c *screenshotConfig) {
		c.quality = int64(q)
	}
}

// Screenshot captures a screenshot and saves it to the given path.
func (p *Page) Screenshot(path string, opts ...ScreenshotOption) error {
	cfg := &screenshotConfig{}
	for _, o := range opts {
		o(cfg)
	}

	params := proto.PageCaptureScreenshot()

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		params = params.WithFormat(proto.PageCaptureScreenshotFormatJpeg)
		if cfg.quality > 0 {
			params = params.WithQuality(cfg.quality)
		}
	case ".webp":
		params = params.WithFormat(proto.PageCaptureScreenshotFormatWebp)
		if cfg.quality > 0 {
			params = params.WithQuality(cfg.quality)
		}
	default:
		params = params.WithFormat(proto.PageCaptureScreenshotFormatPng)
	}

	if cfg.fullPage {
		params = params.WithCaptureBeyondViewport(true)
	}

	res, err := params.Do(p.execCtx)
	if err != nil {
		return err
	}

	data := res.Data
	if isBase64(data) {
		decoded, err := base64.StdEncoding.DecodeString(string(data))
		if err == nil {
			data = decoded
		}
	}

	return os.WriteFile(path, data, 0o644)
}

// PDF saves the page as a PDF file.
func (p *Page) PDF(path string) error {
	res, err := proto.PagePrintToPDF().Do(p.execCtx)
	if err != nil {
		return err
	}

	data := res.Data
	if isBase64(data) {
		decoded, err := base64.StdEncoding.DecodeString(string(data))
		if err == nil {
			data = decoded
		}
	}

	return os.WriteFile(path, data, 0o644)
}

func isBase64(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	for _, b := range data[:min(100, len(data))] {
		if b > 127 {
			return false
		}
	}
	return true
}
