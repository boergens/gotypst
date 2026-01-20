package pdf

import (
	"bytes"
	"strings"
	"testing"

	"github.com/go-text/typesetting/font"
)

func TestFontCollector_RegisterFont(t *testing.T) {
	fc := NewFontCollector()

	// Test with nil face (should still work, creates entry)
	var face *font.Face

	// Register a font
	name := fc.RegisterFont(face, "/path/to/font.ttf")
	if name != "F1" {
		t.Errorf("expected F1, got %s", name)
	}

	// Same font should return same name
	name2 := fc.RegisterFont(face, "/path/to/font.ttf")
	if name2 != "F1" {
		t.Errorf("expected F1 for same font, got %s", name2)
	}

	// Different face should get different name
	var face2 *font.Face
	name3 := fc.RegisterFont(face2, "/path/to/other.ttf")
	// Note: nil faces are the same pointer, so we can't really test different faces
	// without actual font data. The important thing is the API works.
	if name3 == "" {
		t.Error("expected non-empty name for second font")
	}
}

func TestFontCollector_RecordGlyph(t *testing.T) {
	fc := NewFontCollector()

	// Register a nil face first
	var face *font.Face
	fc.RegisterFont(face, "")

	// Recording glyphs should not panic with nil face
	fc.RecordGlyph(face, 65, 'A')
	fc.RecordGlyph(face, 66, 'B')

	// Check that glyphs were recorded
	fonts := fc.Fonts()
	if len(fonts) != 1 {
		t.Fatalf("expected 1 font, got %d", len(fonts))
	}

	ef := fonts[0]
	if len(ef.UsedGlyphs) != 2 {
		t.Errorf("expected 2 used glyphs, got %d", len(ef.UsedGlyphs))
	}

	if r, ok := ef.UsedGlyphs[65]; !ok || r != 'A' {
		t.Errorf("expected glyph 65 -> 'A', got %c", r)
	}
}

func TestFontCollector_FontName(t *testing.T) {
	fc := NewFontCollector()

	// Test with nil face - should auto-register
	var face *font.Face
	name := fc.FontName(face)

	// Should return /F1 (with leading slash)
	if name != "/F1" {
		t.Errorf("expected /F1, got %s", name)
	}

	// Second call should return same name
	name2 := fc.FontName(face)
	if name2 != "/F1" {
		t.Errorf("expected /F1, got %s", name2)
	}
}

func TestFontCollector_SetFontPath(t *testing.T) {
	fc := NewFontCollector()

	var face *font.Face
	path := "/test/path/font.ttf"

	fc.SetFontPath(face, path)

	// Now register/get name - it should have the path
	fc.FontName(face)

	fonts := fc.Fonts()
	if len(fonts) != 1 {
		t.Fatalf("expected 1 font, got %d", len(fonts))
	}

	if fonts[0].Path != path {
		t.Errorf("expected path %s, got %s", path, fonts[0].Path)
	}
}

func TestWriteToUnicodeCMap(t *testing.T) {
	fc := NewFontCollector()
	var face *font.Face
	fc.RegisterFont(face, "")

	// Record some glyphs
	fc.RecordGlyph(face, 65, 'A')
	fc.RecordGlyph(face, 66, 'B')
	fc.RecordGlyph(face, 67, 'C')

	ef := fc.Fonts()[0]

	// Create a writer mock
	w := NewWriter()
	fe := NewFontEmitter(w)

	var buf bytes.Buffer
	fe.writeToUnicodeCMap(&buf, ef)

	cmap := buf.String()

	// Check for required CMap components
	if !strings.Contains(cmap, "begincmap") {
		t.Error("CMap missing begincmap")
	}
	if !strings.Contains(cmap, "endcmap") {
		t.Error("CMap missing endcmap")
	}
	if !strings.Contains(cmap, "/CIDSystemInfo") {
		t.Error("CMap missing CIDSystemInfo")
	}
	if !strings.Contains(cmap, "beginbfchar") {
		t.Error("CMap missing beginbfchar")
	}
	if !strings.Contains(cmap, "endbfchar") {
		t.Error("CMap missing endbfchar")
	}
	if !strings.Contains(cmap, "<41>") {
		t.Error("CMap missing glyph 65 (0x41) mapping")
	}
	if !strings.Contains(cmap, "<0041>") {
		t.Error("CMap missing Unicode A (0x0041)")
	}
}

func TestSanitizeFontName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Arial", "Arial"},
		{"Times New Roman", "TimesNewRoman"},
		{"Helvetica-Bold", "Helvetica-Bold"},
		{"My Font (Regular)", "MyFontRegular"},
		{"", "Font"},
		{"A_B-C123", "A_B-C123"},
	}

	for _, tt := range tests {
		result := sanitizeFontName(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeFontName(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestFontEmitter_EmitFontsEmpty(t *testing.T) {
	w := NewWriter()
	fc := w.FontCollector()

	fe := NewFontEmitter(w)
	resources, err := fe.EmitFonts(fc)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Empty collector should produce empty resources
	if len(resources) != 0 {
		t.Errorf("expected empty resources, got %d", len(resources))
	}
}

func TestFontEmitter_EmitFontsFallback(t *testing.T) {
	w := NewWriter()
	fc := w.FontCollector()

	// Register a nil face (will use fallback)
	var face *font.Face
	fc.RegisterFont(face, "")

	fe := NewFontEmitter(w)
	resources, err := fe.EmitFonts(fc)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have one font resource
	if len(resources) != 1 {
		t.Errorf("expected 1 font resource, got %d", len(resources))
	}

	// Check that F1 is in resources
	if _, ok := resources[Name("F1")]; !ok {
		t.Error("expected F1 in resources")
	}
}

func TestFontMetricsDefaults(t *testing.T) {
	w := NewWriter()
	fe := NewFontEmitter(w)

	// Test with nil face
	ef := &EmbeddedFont{Face: nil}
	metrics := fe.getFontMetrics(ef)

	// Should have reasonable defaults
	if metrics.Ascent == 0 {
		t.Error("expected non-zero ascent")
	}
	if metrics.Descent == 0 {
		t.Error("expected non-zero descent")
	}
	if metrics.Flags == 0 {
		t.Error("expected non-zero flags")
	}
}

func TestGetWidthsArrayDefaults(t *testing.T) {
	w := NewWriter()
	fe := NewFontEmitter(w)

	// Test with nil face
	ef := &EmbeddedFont{Face: nil}
	widths := fe.getWidthsArray(ef)

	// Should have 256 entries
	if len(widths) != 256 {
		t.Errorf("expected 256 widths, got %d", len(widths))
	}

	// All should be default width (600)
	for i, w := range widths {
		if w != Int(600) {
			t.Errorf("width[%d] = %v, expected 600", i, w)
		}
	}
}
