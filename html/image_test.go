package html

import (
	"bytes"
	"strings"
	"testing"

	"github.com/boergens/gotypst/layout"
	"github.com/boergens/gotypst/layout/flow"
	"github.com/boergens/gotypst/layout/pages"
)

func TestEncodeFlowImage(t *testing.T) {
	// Create a minimal JPEG header (just enough to test encoding)
	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F'}

	img := &flow.FrameItemImage{
		Data:   jpegData,
		Format: "jpeg",
		Width:  100,
		Height: 50,
		RenderSize: layout.Size{
			Width:  200,
			Height: 100,
		},
	}

	ref, err := EncodeFlowImage(img, "test-img", ImageEncodingDataURL)
	if err != nil {
		t.Fatalf("EncodeFlowImage failed: %v", err)
	}

	if ref.ID != "test-img" {
		t.Errorf("expected ID 'test-img', got %q", ref.ID)
	}

	if ref.Width != 200 {
		t.Errorf("expected width 200, got %v", ref.Width)
	}

	if ref.Height != 100 {
		t.Errorf("expected height 100, got %v", ref.Height)
	}

	if !strings.HasPrefix(ref.DataURL, "data:image/jpeg;base64,") {
		t.Errorf("expected data URL to start with 'data:image/jpeg;base64,', got %q", ref.DataURL[:50])
	}
}

func TestEncodePagesImageJPEG(t *testing.T) {
	// Create a minimal JPEG for testing
	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0}

	item := &pages.ImageItem{
		Image: pages.Image{
			Data:             jpegData,
			Format:           pages.ImageFormatJPEG,
			Width:            100,
			Height:           100,
			BitsPerComponent: 8,
			ColorSpace:       pages.ColorSpaceDeviceRGB,
		},
		Size: layout.Size{
			Width:  150,
			Height: 150,
		},
	}

	ref, err := EncodePagesImage(item, "pages-img", ImageEncodingDataURL)
	if err != nil {
		t.Fatalf("EncodePagesImage failed: %v", err)
	}

	if !strings.HasPrefix(ref.DataURL, "data:image/jpeg;base64,") {
		t.Errorf("expected JPEG data URL, got %q", ref.DataURL)
	}
}

func TestEncodePagesImagePNG(t *testing.T) {
	// PNG magic bytes
	pngData := []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}

	item := &pages.ImageItem{
		Image: pages.Image{
			Data:             pngData,
			Format:           pages.ImageFormatPNG,
			Width:            100,
			Height:           100,
			BitsPerComponent: 8,
			ColorSpace:       pages.ColorSpaceDeviceRGB,
		},
		Size: layout.Size{
			Width:  150,
			Height: 150,
		},
	}

	ref, err := EncodePagesImage(item, "png-img", ImageEncodingDataURL)
	if err != nil {
		t.Fatalf("EncodePagesImage failed: %v", err)
	}

	if !strings.HasPrefix(ref.DataURL, "data:image/png;base64,") {
		t.Errorf("expected PNG data URL, got %q", ref.DataURL)
	}
}

func TestEncodePagesImageRawFails(t *testing.T) {
	// Raw format should fail for data URL encoding
	item := &pages.ImageItem{
		Image: pages.Image{
			Data:             []byte{0xFF, 0x00, 0x00}, // RGB pixel
			Format:           pages.ImageFormatRaw,
			Width:            1,
			Height:           1,
			BitsPerComponent: 8,
			ColorSpace:       pages.ColorSpaceDeviceRGB,
		},
		Size: layout.Size{
			Width:  100,
			Height: 100,
		},
	}

	_, err := EncodePagesImage(item, "raw-img", ImageEncodingDataURL)
	if err == nil {
		t.Error("expected error for raw image format, got nil")
	}
}

func TestRenderImageTag(t *testing.T) {
	ref := &ImageRef{
		ID:      "test",
		Width:   100,
		Height:  50,
		DataURL: "data:image/png;base64,ABC123",
	}

	tag := RenderImageTag(ref)

	if !strings.Contains(tag, `src="data:image/png;base64,ABC123"`) {
		t.Errorf("expected data URL in src, got %q", tag)
	}

	if !strings.Contains(tag, `width="100.00"`) {
		t.Errorf("expected width attribute, got %q", tag)
	}

	if !strings.Contains(tag, `height="50.00"`) {
		t.Errorf("expected height attribute, got %q", tag)
	}
}

func TestRenderPositionedImage(t *testing.T) {
	ref := &ImageRef{
		ID:      "test",
		Width:   100,
		Height:  50,
		DataURL: "data:image/png;base64,ABC123",
	}

	tag := RenderPositionedImage(ref, 10, 20)

	if !strings.Contains(tag, "position:absolute") {
		t.Errorf("expected absolute positioning, got %q", tag)
	}

	if !strings.Contains(tag, "left:10.00px") {
		t.Errorf("expected left position, got %q", tag)
	}

	if !strings.Contains(tag, "top:20.00px") {
		t.Errorf("expected top position, got %q", tag)
	}
}

func TestFormatToMIME(t *testing.T) {
	tests := []struct {
		format   string
		expected string
	}{
		{"jpeg", "image/jpeg"},
		{"jpg", "image/jpeg"},
		{"png", "image/png"},
		{"gif", "image/gif"},
		{"webp", "image/webp"},
		{"svg", "image/svg+xml"},
		{"unknown", ""},
	}

	for _, tt := range tests {
		result := formatToMIME(tt.format)
		if result != tt.expected {
			t.Errorf("formatToMIME(%q) = %q, want %q", tt.format, result, tt.expected)
		}
	}
}

func TestExportWithImage(t *testing.T) {
	// Create a simple document with an image
	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F'}

	doc := &pages.PagedDocument{
		Pages: []pages.Page{
			{
				Frame: pages.Frame{
					Size: layout.Size{Width: 612, Height: 792}, // Letter size
					Items: []pages.PositionedItem{
						{
							Pos: layout.Point{X: 100, Y: 100},
							Item: pages.ImageItem{
								Image: pages.Image{
									Data:             jpegData,
									Format:           pages.ImageFormatJPEG,
									Width:            200,
									Height:           150,
									BitsPerComponent: 8,
									ColorSpace:       pages.ColorSpaceDeviceRGB,
								},
								Size: layout.Size{Width: 200, Height: 150},
							},
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := Export(doc, &buf)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	html := buf.String()

	// Verify HTML structure
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("expected DOCTYPE declaration")
	}

	if !strings.Contains(html, `class="page"`) {
		t.Error("expected page class")
	}

	// Verify image is rendered with base64 data URL
	if !strings.Contains(html, "data:image/jpeg;base64,") {
		t.Error("expected base64 encoded image")
	}
}
