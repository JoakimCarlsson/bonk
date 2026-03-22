package bonk

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/joakimcarlsson/bonk/proto"
	"golang.org/x/image/draw"
)

// ScreenshotBytes captures a screenshot and returns it as
// image bytes. Defaults to PNG unless AsJPEG is specified.
func (p *Page) ScreenshotBytes(
	opts ...ScreenshotOption,
) ([]byte, error) {
	cfg := &screenshotConfig{}
	for _, o := range opts {
		o(cfg)
	}

	params := proto.PageCaptureScreenshot().
		WithFormat(proto.PageCaptureScreenshotFormatPng)

	if cfg.fullPage {
		params = params.WithCaptureBeyondViewport(true)
	}

	res, err := params.Do(p.execCtx)
	if err != nil {
		return nil, err
	}

	data := res.Data
	if isBase64(data) {
		decoded, err := base64.StdEncoding.DecodeString(
			string(data),
		)
		if err == nil {
			data = decoded
		}
	}

	needsResize := cfg.maxWidth > 0
	needsJPEG := cfg.jpeg

	if needsResize || needsJPEG {
		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			return data, nil
		}

		if needsResize {
			img = resizeImage(img, cfg.maxWidth)
		}

		var buf bytes.Buffer
		if needsJPEG {
			q := int(cfg.quality)
			if q <= 0 {
				q = 80
			}
			err = jpeg.Encode(
				&buf, img, &jpeg.Options{Quality: q},
			)
		} else {
			err = png.Encode(&buf, img)
		}
		if err != nil {
			return data, nil
		}
		data = buf.Bytes()
	}

	return data, nil
}

func resizeImage(img image.Image, maxWidth int) image.Image {
	bounds := img.Bounds()
	w := bounds.Dx()
	if w <= maxWidth {
		return img
	}
	ratio := float64(maxWidth) / float64(w)
	newW := maxWidth
	newH := int(float64(bounds.Dy()) * ratio)
	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.ApproxBiLinear.Scale(
		dst, dst.Bounds(), img, bounds, draw.Over, nil,
	)
	return dst
}

// PDFBytes renders the page as a PDF and returns the raw bytes.
func (p *Page) PDFBytes() ([]byte, error) {
	res, err := proto.PagePrintToPDF().Do(p.execCtx)
	if err != nil {
		return nil, err
	}

	data := res.Data
	if isBase64(data) {
		decoded, err := base64.StdEncoding.DecodeString(
			string(data),
		)
		if err == nil {
			data = decoded
		}
	}

	return data, nil
}

// ScreenshotOption configures screenshot behavior.
type ScreenshotOption func(*screenshotConfig)

type screenshotConfig struct {
	fullPage bool
	quality  int64
	maxWidth int
	jpeg     bool
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

// MaxWidth sets the maximum width for the screenshot. If the
// captured image is wider, it is downscaled proportionally.
func MaxWidth(px int) ScreenshotOption {
	return func(c *screenshotConfig) {
		c.maxWidth = px
	}
}

// AsJPEG captures the screenshot as JPEG instead of PNG.
// Quality defaults to 80 if not set via ScreenshotQuality.
func AsJPEG() ScreenshotOption {
	return func(c *screenshotConfig) {
		c.jpeg = true
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
