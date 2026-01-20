package pdf

import (
	"testing"
)

func TestFontMap(t *testing.T) {
	fm := NewFontMap()

	// Register some glyphs (using nil face for testing)
	ref1 := fm.RegisterGlyph(nil, 65, 'A')
	ref2 := fm.RegisterGlyph(nil, 66, 'B')
	ref3 := fm.RegisterGlyph(nil, 65, 'A') // Duplicate

	// Same face should get same ref
	if ref1.ID != ref2.ID {
		t.Errorf("same face should have same ID, got %d and %d", ref1.ID, ref2.ID)
	}
	if ref1.ID != ref3.ID {
		t.Errorf("same face should have same ID on duplicate, got %d and %d", ref1.ID, ref3.ID)
	}

	// Check usage
	usage, ok := fm.GetUsage(nil)
	if !ok {
		t.Fatal("expected to find usage for nil face")
	}

	if len(usage.GlyphIDs) != 2 {
		t.Errorf("expected 2 unique glyphs, got %d", len(usage.GlyphIDs))
	}

	// Check glyph characters
	if chars, ok := usage.GlyphIDs[65]; !ok || len(chars) != 1 || chars[0] != 'A' {
		t.Errorf("expected glyph 65 to map to 'A', got %v", chars)
	}
	if chars, ok := usage.GlyphIDs[66]; !ok || len(chars) != 1 || chars[0] != 'B' {
		t.Errorf("expected glyph 66 to map to 'B', got %v", chars)
	}
}

func TestFontUsageAddGlyph(t *testing.T) {
	usage := &FontUsage{}

	// Add multiple characters mapping to same glyph
	usage.AddGlyph(100, 'X')
	usage.AddGlyph(100, 'x') // lowercase x might use same glyph in some fonts
	usage.AddGlyph(100, 'X') // Duplicate should not be added

	chars := usage.GlyphIDs[100]
	if len(chars) != 2 {
		t.Errorf("expected 2 characters for glyph 100, got %d", len(chars))
	}

	// Check order
	ordered := usage.OrderedGlyphIDs()
	if len(ordered) != 1 || ordered[0] != 100 {
		t.Errorf("expected ordered glyphs [100], got %v", ordered)
	}
}

func TestFontUsageSortedGlyphIDs(t *testing.T) {
	usage := &FontUsage{}

	// Add glyphs in non-sorted order
	usage.AddGlyph(300, 'C')
	usage.AddGlyph(100, 'A')
	usage.AddGlyph(200, 'B')

	sorted := usage.SortedGlyphIDs()
	if len(sorted) != 3 {
		t.Fatalf("expected 3 glyphs, got %d", len(sorted))
	}
	if sorted[0] != 100 || sorted[1] != 200 || sorted[2] != 300 {
		t.Errorf("expected sorted [100, 200, 300], got %v", sorted)
	}

	// Original order should be preserved in OrderedGlyphIDs
	ordered := usage.OrderedGlyphIDs()
	if ordered[0] != 300 || ordered[1] != 100 || ordered[2] != 200 {
		t.Errorf("expected insertion order [300, 100, 200], got %v", ordered)
	}
}

func TestDetectFontFormat(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected FontFormat
	}{
		{"TrueType", []byte{0x00, 0x01, 0x00, 0x00, 0x00}, FontFormatTTF},
		{"TrueType Mac", []byte("true" + "\x00"), FontFormatTTF},
		{"OpenType CFF", []byte("OTTO\x00"), FontFormatOTF},
		{"WOFF", []byte("wOFF\x00"), FontFormatWOFF},
		{"WOFF2", []byte("wOF2\x00"), FontFormatWOFF2},
		{"Unknown", []byte{0xFF, 0xFF, 0xFF, 0xFF}, FontFormatUnknown},
		{"Too short", []byte{0x00}, FontFormatUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectFontFormat(tt.data)
			if got != tt.expected {
				t.Errorf("detectFontFormat() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsCJK(t *testing.T) {
	tests := []struct {
		r    rune
		want bool
	}{
		{'A', false},
		{'中', true},  // CJK Unified Ideograph
		{'あ', true},  // Hiragana
		{'ア', true},  // Katakana
		{'한', true},  // Hangul
		{'1', false},
		{'α', false}, // Greek
	}

	for _, tt := range tests {
		t.Run(string(tt.r), func(t *testing.T) {
			if got := isCJK(tt.r); got != tt.want {
				t.Errorf("isCJK(%q) = %v, want %v", tt.r, got, tt.want)
			}
		})
	}
}

func TestPdfName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Arial", "Arial"},
		{"Times New Roman", "Times#20New#20Roman"},
		{"Font/Bold", "Font#2FBold"},
		{"Test#Name", "Test#23Name"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := pdfName(tt.input)
			if got != tt.expected {
				t.Errorf("pdfName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestPdfString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello", "(Hello)"},
		{"Hello (World)", "(Hello \\(World\\))"},
		{"Back\\slash", "(Back\\\\slash)"},
		{"New\nLine", "(New\\nLine)"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := pdfString(tt.input)
			if got != tt.expected {
				t.Errorf("pdfString(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestRuneToUTF16BE(t *testing.T) {
	tests := []struct {
		r        rune
		expected string
	}{
		{'A', "0041"},
		{'中', "4E2D"},
		{0x1F600, "D83DDE00"}, // Emoji (surrogate pair)
	}

	for _, tt := range tests {
		t.Run(string(tt.r), func(t *testing.T) {
			got := runeToUTF16BE(tt.r)
			if got != tt.expected {
				t.Errorf("runeToUTF16BE(%q) = %q, want %q", tt.r, got, tt.expected)
			}
		})
	}
}

func TestBuildWidthsArray(t *testing.T) {
	widths := map[uint16]int{
		65: 600, // A
		66: 700, // B
	}

	result := buildWidthsArray(widths)

	// Should have 256 entries
	if len(result) < 10 {
		t.Errorf("widths array too short: %s", result)
	}

	// Should start with [ and end with ]
	if result[0] != '[' || result[len(result)-1] != ']' {
		t.Errorf("widths array should be bracketed: %s", result)
	}
}

func TestBuildCIDWidths(t *testing.T) {
	tests := []struct {
		name   string
		widths map[uint16]int
	}{
		{"empty", map[uint16]int{}},
		{"single", map[uint16]int{100: 500}},
		{"consecutive", map[uint16]int{100: 500, 101: 500, 102: 500}},
		{"mixed", map[uint16]int{100: 500, 101: 600, 200: 700}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildCIDWidths(tt.widths)
			if result[0] != '[' || result[len(result)-1] != ']' {
				t.Errorf("CID widths should be bracketed: %s", result)
			}
		})
	}
}

func TestBuildToUnicodeCMap(t *testing.T) {
	mapping := map[uint16][]rune{
		65: {'A'},
		66: {'B'},
	}

	cmap := buildToUnicodeCMap(mapping)

	// Check required CMap elements
	if !contains(cmap, "begincmap") {
		t.Error("CMap should contain begincmap")
	}
	if !contains(cmap, "endcmap") {
		t.Error("CMap should contain endcmap")
	}
	if !contains(cmap, "beginbfchar") {
		t.Error("CMap should contain beginbfchar")
	}
	if !contains(cmap, "<41> <0041>") {
		t.Error("CMap should map glyph 65 to U+0041")
	}
}

func TestBuildCIDToUnicodeCMap(t *testing.T) {
	mapping := map[uint16][]rune{
		1000: {'中'},
		1001: {'文'},
	}

	cmap := buildCIDToUnicodeCMap(mapping)

	// Check CID-specific elements
	if !contains(cmap, "<0000> <FFFF>") {
		t.Error("CID CMap should have 2-byte codespace")
	}
	if !contains(cmap, "<03E8>") { // 1000 in hex
		t.Error("CID CMap should contain glyph 1000")
	}
}

func TestCompressDecompress(t *testing.T) {
	original := []byte("Hello, World! This is a test of compression.")

	compressed, err := CompressStream(original)
	if err != nil {
		t.Fatalf("CompressStream failed: %v", err)
	}

	decompressed, err := DecompressStream(compressed)
	if err != nil {
		t.Fatalf("DecompressStream failed: %v", err)
	}

	if string(decompressed) != string(original) {
		t.Errorf("roundtrip failed: got %q, want %q", decompressed, original)
	}
}

func TestCalculateChecksum(t *testing.T) {
	// Test with known values
	data := []byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02}
	sum := calculateChecksum(data)

	expected := uint32(0x00000001 + 0x00000002)
	if sum != expected {
		t.Errorf("calculateChecksum() = %08X, want %08X", sum, expected)
	}

	// Test with padding
	data2 := []byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00}
	sum2 := calculateChecksum(data2)
	if sum2 == 0 {
		t.Error("checksum should not be zero for non-empty data")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
