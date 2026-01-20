package html

import (
	"bytes"
	"strings"
	"testing"

	"github.com/boergens/gotypst/layout"
	"github.com/boergens/gotypst/layout/pages"
)

func TestRenderEmptyDocument(t *testing.T) {
	doc := &pages.PagedDocument{
		Pages: []pages.Page{},
	}

	var buf bytes.Buffer
	if err := Export(doc, &buf); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("missing DOCTYPE")
	}
	if !strings.Contains(html, "<title>Document</title>") {
		t.Error("missing default title")
	}
}

func TestRenderSinglePage(t *testing.T) {
	doc := &pages.PagedDocument{
		Pages: []pages.Page{
			{
				Frame: pages.Frame{
					Size: layout.Size{Width: 595, Height: 842}, // A4
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := Export(doc, &buf); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `class="page"`) {
		t.Error("missing page div")
	}
	if !strings.Contains(html, "595.00pt") {
		t.Error("missing page width")
	}
}

func TestRenderDocumentWithTitle(t *testing.T) {
	title := "Test Document"
	doc := &pages.PagedDocument{
		Pages: []pages.Page{},
		Info: pages.DocumentInfo{
			Title: &title,
		},
	}

	var buf bytes.Buffer
	if err := Export(doc, &buf); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "<title>Test Document</title>") {
		t.Error("missing custom title")
	}
}

func TestRenderTextItem(t *testing.T) {
	doc := &pages.PagedDocument{
		Pages: []pages.Page{
			{
				Frame: pages.Frame{
					Size: layout.Size{Width: 595, Height: 842},
					Items: []pages.PositionedItem{
						{
							Pos: layout.Point{X: 72, Y: 72},
							Item: pages.TextItem{
								Text:     "Hello World",
								FontSize: 12,
							},
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := Export(doc, &buf); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "Hello World") {
		t.Error("missing text content")
	}
	if !strings.Contains(html, `class="text"`) {
		t.Error("missing text class")
	}
}

func TestRenderNestedFrame(t *testing.T) {
	doc := &pages.PagedDocument{
		Pages: []pages.Page{
			{
				Frame: pages.Frame{
					Size: layout.Size{Width: 595, Height: 842},
					Items: []pages.PositionedItem{
						{
							Pos: layout.Point{X: 50, Y: 50},
							Item: pages.GroupItem{
								Frame: pages.Frame{
									Size: layout.Size{Width: 200, Height: 100},
								},
							},
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := Export(doc, &buf); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, `class="frame"`) {
		t.Error("missing nested frame div")
	}
}

func TestEscapeHTML(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"<script>", "&lt;script&gt;"},
		{"a & b", "a &amp; b"},
		{`"quoted"`, "&quot;quoted&quot;"},
		{"it's", "it&#39;s"},
	}

	for _, tt := range tests {
		result := escapeHTML(tt.input)
		if result != tt.expected {
			t.Errorf("escapeHTML(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestRenderPageWithFill(t *testing.T) {
	doc := &pages.PagedDocument{
		Pages: []pages.Page{
			{
				Frame: pages.Frame{
					Size: layout.Size{Width: 595, Height: 842},
				},
				Fill: &pages.Paint{
					Color: &pages.Color{R: 255, G: 0, B: 0, A: 255},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := Export(doc, &buf); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "background-color: rgba(255, 0, 0,") {
		t.Error("missing background color")
	}
}
