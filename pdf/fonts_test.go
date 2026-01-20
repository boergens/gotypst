package pdf

import (
	"strings"
	"testing"

	"github.com/boergens/gotypst/font"
)

func TestFontManager(t *testing.T) {
	fm := NewFontManager()

	if fm.HasFonts() {
		t.Error("new font manager should have no fonts")
	}

	// Create a mock font
	f := &font.Font{Info: font.FontInfo{Family: "TestFont"}}

	// Get or create should create a new entry
	pdfFont := fm.GetOrCreateFont(f)
	if pdfFont == nil {
		t.Fatal("expected non-nil PDF font")
	}
	if pdfFont.Name != "F1" {
		t.Errorf("expected name F1, got %s", pdfFont.Name)
	}
	if pdfFont.Font != f {
		t.Error("PDF font should reference original font")
	}

	// Second call should return same entry
	pdfFont2 := fm.GetOrCreateFont(f)
	if pdfFont2 != pdfFont {
		t.Error("expected same PDF font entry")
	}

	if !fm.HasFonts() {
		t.Error("font manager should have fonts")
	}

	// Add a second font
	f2 := &font.Font{Info: font.FontInfo{Family: "TestFont2"}}
	pdfFont3 := fm.GetOrCreateFont(f2)
	if pdfFont3.Name != "F2" {
		t.Errorf("expected name F2, got %s", pdfFont3.Name)
	}

	// Check fonts list
	fonts := fm.Fonts()
	if len(fonts) != 2 {
		t.Errorf("expected 2 fonts, got %d", len(fonts))
	}
}

func TestFontManagerAddGlyph(t *testing.T) {
	fm := NewFontManager()
	f := &font.Font{Info: font.FontInfo{Family: "TestFont"}}

	fm.AddGlyph(f, 1)
	fm.AddGlyph(f, 5)
	fm.AddGlyph(f, 10)
	fm.AddGlyph(f, 5) // duplicate

	pdfFont := fm.GetOrCreateFont(f)
	if pdfFont.Glyphs.Len() != 3 {
		t.Errorf("expected 3 glyphs, got %d", pdfFont.Glyphs.Len())
	}
}

func TestGenerateSubsetPrefix(t *testing.T) {
	tests := []struct {
		index    int
		expected string
	}{
		{0, "AAAAAA"},
		{1, "AAAAAB"},
		{25, "AAAAAZ"},
		{26, "AAAABA"},
		{27, "AAAABB"},
	}

	for _, tt := range tests {
		result := generateSubsetPrefix(tt.index)
		if result != tt.expected {
			t.Errorf("generateSubsetPrefix(%d) = %s, expected %s", tt.index, result, tt.expected)
		}
	}
}

func TestSanitizePostScriptName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Arial", "Arial"},
		{"Times New Roman", "TimesNewRoman"},
		{"Arial-Bold", "Arial-Bold"},
		{"Font_Name", "Font_Name"},
		{"日本語フォント", "Font"}, // Non-ASCII becomes "Font"
		{"", "Font"},
		{"123", "123"},
	}

	for _, tt := range tests {
		result := sanitizePostScriptName(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizePostScriptName(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestEncodeGlyphString(t *testing.T) {
	tests := []struct {
		glyphs   []uint16
		expected string
	}{
		{[]uint16{}, "<>"},
		{[]uint16{0}, "<0000>"},
		{[]uint16{1}, "<0001>"},
		{[]uint16{255}, "<00FF>"},
		{[]uint16{256}, "<0100>"},
		{[]uint16{1, 2, 3}, "<000100020003>"},
		{[]uint16{0xFFFF}, "<FFFF>"},
	}

	for _, tt := range tests {
		result := EncodeGlyphString(tt.glyphs)
		if result != tt.expected {
			t.Errorf("EncodeGlyphString(%v) = %s, expected %s", tt.glyphs, result, tt.expected)
		}
	}
}

func TestEncodeGlyphID(t *testing.T) {
	tests := []struct {
		glyph    uint16
		expected []byte
	}{
		{0, []byte{0x00, 0x00}},
		{1, []byte{0x00, 0x01}},
		{256, []byte{0x01, 0x00}},
		{0xFFFF, []byte{0xFF, 0xFF}},
	}

	for _, tt := range tests {
		result := EncodeGlyphID(tt.glyph)
		if len(result) != 2 {
			t.Errorf("EncodeGlyphID(%d) length = %d, expected 2", tt.glyph, len(result))
		}
		if result[0] != tt.expected[0] || result[1] != tt.expected[1] {
			t.Errorf("EncodeGlyphID(%d) = %v, expected %v", tt.glyph, result, tt.expected)
		}
	}
}

func TestBuildFontResources(t *testing.T) {
	fm := NewFontManager()

	// Empty manager
	resources := fm.BuildFontResources()
	if len(resources) != 0 {
		t.Errorf("expected empty resources, got %d entries", len(resources))
	}

	// Add fonts
	f1 := &font.Font{Info: font.FontInfo{Family: "Font1"}}
	f2 := &font.Font{Info: font.FontInfo{Family: "Font2"}}
	pdfFont1 := fm.GetOrCreateFont(f1)
	pdfFont2 := fm.GetOrCreateFont(f2)

	// Simulate ref assignment
	pdfFont1.Ref = Ref{ID: 10}
	pdfFont2.Ref = Ref{ID: 20}

	resources = fm.BuildFontResources()
	if len(resources) != 2 {
		t.Errorf("expected 2 font resources, got %d", len(resources))
	}

	if ref, ok := resources[Name("F1")].(Ref); !ok || ref.ID != 10 {
		t.Error("expected F1 to reference object 10")
	}
	if ref, ok := resources[Name("F2")].(Ref); !ok || ref.ID != 20 {
		t.Error("expected F2 to reference object 20")
	}
}

func TestBuildToUnicodeCMap(t *testing.T) {
	fm := NewFontManager()
	f := &font.Font{Info: font.FontInfo{Family: "TestFont"}}
	pdfFont := fm.GetOrCreateFont(f)

	// Add some glyphs
	pdfFont.Glyphs.Add(0)
	pdfFont.Glyphs.Add(65) // 'A'
	pdfFont.Glyphs.Add(66) // 'B'

	// Create a mock subset
	pdfFont.Subset = &font.SubsettedFont{
		GlyphMapping: map[uint16]uint16{
			0:  0,
			65: 1,
			66: 2,
		},
	}

	cmap := fm.buildToUnicodeCMap(pdfFont)

	// Check that it contains expected elements
	cmapStr := string(cmap)
	if !strings.Contains(cmapStr, "beginbfchar") {
		t.Error("CMap should contain beginbfchar")
	}
	if !strings.Contains(cmapStr, "endbfchar") {
		t.Error("CMap should contain endbfchar")
	}
	if !strings.Contains(cmapStr, "/CIDInit") {
		t.Error("CMap should contain /CIDInit")
	}
	if !strings.Contains(cmapStr, "endcmap") {
		t.Error("CMap should contain endcmap")
	}
}

func TestWriteType1Font(t *testing.T) {
	w := NewWriter()
	fm := NewFontManager()

	f := &font.Font{Info: font.FontInfo{Family: "TestFont"}}
	pdfFont := fm.GetOrCreateFont(f)

	err := fm.writeType1Font(w, pdfFont)
	if err != nil {
		t.Fatalf("writeType1Font failed: %v", err)
	}

	// Check that ref was assigned
	if pdfFont.Ref.ID == 0 {
		t.Error("expected ref to be assigned")
	}

	// Check that object was added
	if len(w.objects) != 1 {
		t.Errorf("expected 1 object, got %d", len(w.objects))
	}
}
