package html

import (
	"encoding/base64"
	"fmt"

	"github.com/boergens/gotypst/layout"
	"github.com/boergens/gotypst/layout/flow"
	"github.com/boergens/gotypst/layout/pages"
)

// ImageEncoding specifies how images should be embedded in HTML output.
type ImageEncoding int

const (
	// ImageEncodingDataURL embeds images as base64 data URLs.
	// This produces self-contained HTML but larger file sizes.
	ImageEncodingDataURL ImageEncoding = iota
	// ImageEncodingExternal uses external file references.
	// Requires images to be saved separately.
	ImageEncodingExternal
)

// ImageRef represents a reference to an image in the HTML output.
type ImageRef struct {
	// ID is a unique identifier for this image.
	ID string
	// Width is the rendered width.
	Width layout.Abs
	// Height is the rendered height.
	Height layout.Abs
	// DataURL is the base64 encoded data URL (if using DataURL encoding).
	DataURL string
	// Path is the external file path (if using External encoding).
	Path string
}

// EncodeFlowImage encodes a flow.FrameItemImage to an HTML image reference.
func EncodeFlowImage(img *flow.FrameItemImage, id string, encoding ImageEncoding) (*ImageRef, error) {
	ref := &ImageRef{
		ID:     id,
		Width:  img.RenderSize.Width,
		Height: img.RenderSize.Height,
	}

	if encoding == ImageEncodingDataURL {
		dataURL, err := encodeDataURL(img.Data, img.Format)
		if err != nil {
			return nil, err
		}
		ref.DataURL = dataURL
	}

	return ref, nil
}

// EncodePagesImage encodes a pages.ImageItem to an HTML image reference.
func EncodePagesImage(item *pages.ImageItem, id string, encoding ImageEncoding) (*ImageRef, error) {
	ref := &ImageRef{
		ID:     id,
		Width:  item.Size.Width,
		Height: item.Size.Height,
	}

	if encoding == ImageEncodingDataURL {
		format := imageFormatToMIME(item.Image.Format)
		dataURL, err := encodeDataURLFromPages(&item.Image, format)
		if err != nil {
			return nil, err
		}
		ref.DataURL = dataURL
	}

	return ref, nil
}

// encodeDataURL encodes image data as a base64 data URL.
func encodeDataURL(data []byte, format string) (string, error) {
	mime := formatToMIME(format)
	if mime == "" {
		return "", fmt.Errorf("unsupported image format: %s", format)
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", mime, encoded), nil
}

// encodeDataURLFromPages encodes a pages.Image as a base64 data URL.
func encodeDataURLFromPages(img *pages.Image, mime string) (string, error) {
	if mime == "" {
		return "", fmt.Errorf("unsupported image format")
	}

	// For raw format, we need to convert to a web-compatible format.
	// Raw RGB data would need to be encoded as PNG or similar.
	// For JPEG and PNG, we can use the data directly.
	if img.Format == pages.ImageFormatRaw {
		// Raw images can't be directly embedded - they need encoding.
		// For now, return an error. A full implementation would
		// re-encode as PNG.
		return "", fmt.Errorf("raw image format requires encoding before HTML embedding")
	}

	encoded := base64.StdEncoding.EncodeToString(img.Data)
	return fmt.Sprintf("data:%s;base64,%s", mime, encoded), nil
}

// formatToMIME converts a flow image format string to MIME type.
func formatToMIME(format string) string {
	switch format {
	case "jpeg", "jpg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	case "svg":
		return "image/svg+xml"
	default:
		return ""
	}
}

// imageFormatToMIME converts a pages.ImageFormat to MIME type.
func imageFormatToMIME(format pages.ImageFormat) string {
	switch format {
	case pages.ImageFormatJPEG:
		return "image/jpeg"
	case pages.ImageFormatPNG:
		return "image/png"
	case pages.ImageFormatRaw:
		// Raw format has no direct MIME equivalent
		return ""
	default:
		return ""
	}
}

// RenderImageTag generates an HTML img element for the image reference.
func RenderImageTag(ref *ImageRef) string {
	var src string
	if ref.DataURL != "" {
		src = ref.DataURL
	} else if ref.Path != "" {
		src = ref.Path
	} else {
		return ""
	}

	// Convert layout.Abs to CSS pixels (assuming 72 DPI for Typst points).
	// layout.Abs is in points, CSS pixels are also typically 1:1 with points
	// for web display at standard resolution.
	widthPx := float64(ref.Width)
	heightPx := float64(ref.Height)

	return fmt.Sprintf(
		`<img src="%s" width="%.2f" height="%.2f" style="display:block;" />`,
		src, widthPx, heightPx,
	)
}

// RenderPositionedImage generates an absolutely positioned image element.
func RenderPositionedImage(ref *ImageRef, x, y layout.Abs) string {
	var src string
	if ref.DataURL != "" {
		src = ref.DataURL
	} else if ref.Path != "" {
		src = ref.Path
	} else {
		return ""
	}

	widthPx := float64(ref.Width)
	heightPx := float64(ref.Height)
	xPx := float64(x)
	yPx := float64(y)

	return fmt.Sprintf(
		`<img src="%s" style="position:absolute;left:%.2fpx;top:%.2fpx;width:%.2fpx;height:%.2fpx;" />`,
		src, xPx, yPx, widthPx, heightPx,
	)
}
