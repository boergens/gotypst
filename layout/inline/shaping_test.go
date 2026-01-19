package inline

import (
	"testing"

	"github.com/go-text/typesetting/language"
)

func TestEmConversions(t *testing.T) {
	tests := []struct {
		name     string
		em       Em
		size     Abs
		expected Abs
	}{
		{"zero em at any size", 0, 12, 0},
		{"1em at 12pt", 1, 12, 12},
		{"0.5em at 12pt", 0.5, 12, 6},
		{"1em at 16pt", 1, 16, 16},
		{"2em at 10pt", 2, 10, 20},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.em.At(tc.size)
			if got != tc.expected {
				t.Errorf("got %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestEmFromAbs(t *testing.T) {
	tests := []struct {
		name     string
		abs      Abs
		size     Abs
		expected Em
	}{
		{"12pt at 12pt", 12, 12, 1},
		{"6pt at 12pt", 6, 12, 0.5},
		{"24pt at 12pt", 24, 12, 2},
		{"zero size returns zero", 12, 0, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := EmFromAbs(tc.abs, tc.size)
			if got != tc.expected {
				t.Errorf("got %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestRangeContains(t *testing.T) {
	r := Range{Start: 5, End: 10}

	tests := []struct {
		index    int
		expected bool
	}{
		{4, false},
		{5, true},
		{7, true},
		{9, true},
		{10, false},
		{11, false},
	}

	for _, tc := range tests {
		got := r.Contains(tc.index)
		if got != tc.expected {
			t.Errorf("Range{5,10}.Contains(%d) = %v, want %v", tc.index, got, tc.expected)
		}
	}
}

func TestIsSpace(t *testing.T) {
	tests := []struct {
		char     rune
		expected bool
	}{
		{' ', true},
		{'\u00A0', true},  // NBSP
		{'\u3000', true},  // Ideographic space
		{'a', false},
		{'\t', false},
		{'\n', false},
	}

	for _, tc := range tests {
		got := isSpace(tc.char)
		if got != tc.expected {
			t.Errorf("isSpace(%q) = %v, want %v", tc.char, got, tc.expected)
		}
	}
}

func TestIsOfCJScript(t *testing.T) {
	tests := []struct {
		char     rune
		expected bool
	}{
		{'中', true},      // Han
		{'あ', true},      // Hiragana
		{'ア', true},      // Katakana
		{'\u30FC', true}, // Katakana prolonged sound mark
		{'a', false},
		{'1', false},
	}

	for _, tc := range tests {
		got := IsOfCJScript(tc.char)
		if got != tc.expected {
			t.Errorf("IsOfCJScript(%q) = %v, want %v", tc.char, got, tc.expected)
		}
	}
}

func TestCJKPunctStyle(t *testing.T) {
	tests := []struct {
		lang     Lang
		region   *Region
		expected CJKPunctStyle
	}{
		{LangChinese, nil, CJKPunctStyleGB},
		{LangChinese, regionPtr("TW"), CJKPunctStyleCNS},
		{LangChinese, regionPtr("HK"), CJKPunctStyleCNS},
		{LangChinese, regionPtr("CN"), CJKPunctStyleGB},
		{LangJapanese, nil, CJKPunctStyleJIS},
		{LangEnglish, nil, CJKPunctStyleGB},
	}

	for _, tc := range tests {
		got := GetCJKPunctStyle(tc.lang, tc.region)
		if got != tc.expected {
			t.Errorf("GetCJKPunctStyle(%q, %v) = %v, want %v", tc.lang, tc.region, got, tc.expected)
		}
	}
}

func regionPtr(r string) *Region {
	region := Region(r)
	return &region
}

func TestGlyphs(t *testing.T) {
	glyphs := []ShapedGlyph{
		{Char: 'H', Range: Range{0, 1}},
		{Char: 'e', Range: Range{1, 2}},
		{Char: 'l', Range: Range{2, 3}},
		{Char: 'l', Range: Range{3, 4}},
		{Char: 'o', Range: Range{4, 5}},
	}

	g := NewGlyphsFromVec(glyphs)

	t.Run("Len", func(t *testing.T) {
		if g.Len() != 5 {
			t.Errorf("Len() = %d, want 5", g.Len())
		}
	})

	t.Run("At", func(t *testing.T) {
		if g.At(0).Char != 'H' {
			t.Errorf("At(0).Char = %q, want 'H'", g.At(0).Char)
		}
		if g.At(4).Char != 'o' {
			t.Errorf("At(4).Char = %q, want 'o'", g.At(4).Char)
		}
	})

	t.Run("Last", func(t *testing.T) {
		if g.Last().Char != 'o' {
			t.Errorf("Last().Char = %q, want 'o'", g.Last().Char)
		}
	})

	t.Run("All", func(t *testing.T) {
		all := g.All()
		if len(all) != 5 {
			t.Errorf("len(All()) = %d, want 5", len(all))
		}
	})
}

func TestGlyphsTrim(t *testing.T) {
	glyphs := []ShapedGlyph{
		{Char: ' ', Range: Range{0, 1}},
		{Char: 'H', Range: Range{1, 2}},
		{Char: 'i', Range: Range{2, 3}},
		{Char: ' ', Range: Range{3, 4}},
		{Char: ' ', Range: Range{4, 5}},
	}

	g := NewGlyphsFromVec(glyphs)
	g.Trim(func(sg *ShapedGlyph) bool {
		return sg.Char == ' '
	})

	if g.Len() != 2 {
		t.Errorf("After trim, Len() = %d, want 2", g.Len())
	}
	if g.At(0).Char != 'H' {
		t.Errorf("After trim, At(0).Char = %q, want 'H'", g.At(0).Char)
	}
	if g.At(1).Char != 'i' {
		t.Errorf("After trim, At(1).Char = %q, want 'i'", g.At(1).Char)
	}
}

func TestShapedGlyphIsSpace(t *testing.T) {
	tests := []struct {
		char     rune
		expected bool
	}{
		{' ', true},
		{'\u00A0', true},
		{'a', false},
	}

	for _, tc := range tests {
		g := ShapedGlyph{Char: tc.char}
		got := g.IsSpace()
		if got != tc.expected {
			t.Errorf("ShapedGlyph{Char: %q}.IsSpace() = %v, want %v", tc.char, got, tc.expected)
		}
	}
}

func TestShapedTextEmpty(t *testing.T) {
	st := &ShapedText{
		Base:   0,
		Text:   "Hello",
		Dir:    DirLTR,
		Lang:   LangEnglish,
		Glyphs: NewGlyphsFromVec([]ShapedGlyph{{Char: 'H'}}),
	}

	empty := st.Empty()
	if empty.Text != "" {
		t.Errorf("Empty().Text = %q, want empty", empty.Text)
	}
	if empty.Dir != DirLTR {
		t.Errorf("Empty().Dir = %v, want DirLTR", empty.Dir)
	}
	if empty.Lang != LangEnglish {
		t.Errorf("Empty().Lang = %v, want LangEnglish", empty.Lang)
	}
	if empty.Glyphs.Len() != 0 {
		t.Errorf("Empty().Glyphs.Len() = %d, want 0", empty.Glyphs.Len())
	}
}

func TestShapedTextWidth(t *testing.T) {
	glyphs := []ShapedGlyph{
		{XAdvance: 0.5, Size: 12},
		{XAdvance: 0.5, Size: 12},
		{XAdvance: 0.5, Size: 12},
	}

	st := &ShapedText{
		Glyphs: NewGlyphsFromVec(glyphs),
	}

	width := st.Width()
	expected := Abs(18) // 0.5 * 12 * 3 = 18
	if width != expected {
		t.Errorf("Width() = %v, want %v", width, expected)
	}
}

func TestGetScript(t *testing.T) {
	tests := []struct {
		char     rune
		expected language.Script
	}{
		{'A', language.Latin},
		{'中', language.Han},
		{'あ', language.Hiragana},
		{'ア', language.Katakana},
		{'α', language.Greek},
		{'а', language.Cyrillic},
	}

	for _, tc := range tests {
		got := getScript(tc.char)
		if got != tc.expected {
			t.Errorf("getScript(%q) = %v, want %v", tc.char, got, tc.expected)
		}
	}
}

func TestIsDefaultIgnorable(t *testing.T) {
	tests := []struct {
		char     rune
		expected bool
	}{
		{'\u00AD', true},  // Soft hyphen
		{'\u200B', true},  // Zero-width space
		{'\u200C', true},  // ZWNJ
		{'\u200D', true},  // ZWJ
		{'\uFEFF', true},  // BOM
		{'a', false},
		{' ', false},
	}

	for _, tc := range tests {
		got := isDefaultIgnorable(tc.char)
		if got != tc.expected {
			t.Errorf("isDefaultIgnorable(%U) = %v, want %v", tc.char, got, tc.expected)
		}
	}
}

func TestDirIsPositive(t *testing.T) {
	if !DirLTR.IsPositive() {
		t.Error("DirLTR.IsPositive() = false, want true")
	}
	if DirRTL.IsPositive() {
		t.Error("DirRTL.IsPositive() = true, want false")
	}
}

func TestAdjustability(t *testing.T) {
	g := &ShapedGlyph{
		Adjustability: Adjustability{
			Stretchability: [2]Em{0.1, 0.2},
			Shrinkability:  [2]Em{0.05, 0.1},
		},
	}

	stretch := g.Stretchability()
	if stretch[0] != 0.1 || stretch[1] != 0.2 {
		t.Errorf("Stretchability() = %v, want [0.1, 0.2]", stretch)
	}

	shrink := g.Shrinkability()
	if shrink[0] != 0.05 || shrink[1] != 0.1 {
		t.Errorf("Shrinkability() = %v, want [0.05, 0.1]", shrink)
	}
}

func TestShrinkLeftRight(t *testing.T) {
	g := &ShapedGlyph{
		XAdvance: 1.0,
		XOffset:  0,
		Adjustability: Adjustability{
			Shrinkability: [2]Em{0.5, 0.5},
		},
	}

	g.ShrinkLeft(0.1)
	if g.XOffset != -0.1 {
		t.Errorf("After ShrinkLeft, XOffset = %v, want -0.1", g.XOffset)
	}
	if g.XAdvance != 0.9 {
		t.Errorf("After ShrinkLeft, XAdvance = %v, want 0.9", g.XAdvance)
	}
	if g.Adjustability.Shrinkability[0] != 0.4 {
		t.Errorf("After ShrinkLeft, Shrinkability[0] = %v, want 0.4", g.Adjustability.Shrinkability[0])
	}

	g.ShrinkRight(0.2)
	if g.XAdvance != 0.7 {
		t.Errorf("After ShrinkRight, XAdvance = %v, want 0.7", g.XAdvance)
	}
	if g.Adjustability.Shrinkability[1] != 0.3 {
		t.Errorf("After ShrinkRight, Shrinkability[1] = %v, want 0.3", g.Adjustability.Shrinkability[1])
	}
}

func TestShapedTextJustifiables(t *testing.T) {
	glyphs := []ShapedGlyph{
		{IsJustifiable: true},
		{IsJustifiable: false},
		{IsJustifiable: true},
		{IsJustifiable: true},
	}

	st := &ShapedText{
		Glyphs: NewGlyphsFromVec(glyphs),
	}

	count := st.Justifiables()
	if count != 3 {
		t.Errorf("Justifiables() = %d, want 3", count)
	}
}

func TestShapedTextStretchability(t *testing.T) {
	glyphs := []ShapedGlyph{
		{Size: 12, Adjustability: Adjustability{Stretchability: [2]Em{0.1, 0.1}}},
		{Size: 12, Adjustability: Adjustability{Stretchability: [2]Em{0.1, 0.1}}},
	}

	st := &ShapedText{
		Glyphs: NewGlyphsFromVec(glyphs),
	}

	// Total stretchability = (0.1 + 0.1) * 12 + (0.1 + 0.1) * 12 = 4.8
	stretch := st.Stretchability()
	expected := Abs(4.8)
	// Use approximate comparison due to floating point precision
	if abs(float64(stretch)-float64(expected)) > 0.0001 {
		t.Errorf("Stretchability() = %v, want approximately %v", stretch, expected)
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func TestToFixed(t *testing.T) {
	// 26.6 fixed point: multiply by 64
	tests := []struct {
		input    float64
		expected int
	}{
		{1.0, 64},
		{0.5, 32},
		{12.0, 768},
	}

	for _, tc := range tests {
		got := toFixed(tc.input)
		if int(got) != tc.expected {
			t.Errorf("toFixed(%v) = %v, want %v", tc.input, got, tc.expected)
		}
	}
}

func TestMin(t *testing.T) {
	if min(1.0, 2.0) != 1.0 {
		t.Error("min(1.0, 2.0) != 1.0")
	}
	if min(3.0, 2.0) != 2.0 {
		t.Error("min(3.0, 2.0) != 2.0")
	}
	if min(1.5, 1.5) != 1.5 {
		t.Error("min(1.5, 1.5) != 1.5")
	}
}

func TestShapedTextString(t *testing.T) {
	st := &ShapedText{
		Text:   "Hello",
		Dir:    DirLTR,
		Glyphs: NewGlyphsFromVec([]ShapedGlyph{{}, {}, {}, {}, {}}),
	}

	s := st.String()
	if s == "" {
		t.Error("String() returned empty")
	}
}
