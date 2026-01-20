package text

import (
	"testing"

	"github.com/boergens/gotypst/layout/inline"
)

func TestNewTextElem(t *testing.T) {
	te := New("Hello, World!")

	if te.Body != "Hello, World!" {
		t.Errorf("Body = %q, want %q", te.Body, "Hello, World!")
	}

	if te.Size != SizeFromPt(11) {
		t.Errorf("Size = %v, want 11pt", te.Size)
	}

	if te.Weight != FontWeightNormal {
		t.Errorf("Weight = %v, want normal", te.Weight)
	}

	if te.Style != FontStyleNormal {
		t.Errorf("Style = %v, want normal", te.Style)
	}

	if te.Stretch != FontStretchNormal {
		t.Errorf("Stretch = %v, want normal", te.Stretch)
	}

	if te.Spacing != 1.0 {
		t.Errorf("Spacing = %v, want 1.0", te.Spacing)
	}

	if !te.Fallback {
		t.Error("Fallback should be true by default")
	}
}

func TestTextElemBuilders(t *testing.T) {
	te := New("Test").
		WithFont("Helvetica", "Arial").
		WithSize(SizeFromPt(14)).
		WithWeight(FontWeightBold).
		WithStyle(FontStyleItalic).
		WithStretch(FontStretchCondensed).
		WithFill(Red)

	if len(te.Font) != 2 || te.Font[0] != "Helvetica" || te.Font[1] != "Arial" {
		t.Errorf("Font = %v, want [Helvetica Arial]", te.Font)
	}

	if te.Size != SizeFromPt(14) {
		t.Errorf("Size = %v, want 14pt", te.Size)
	}

	if te.Weight != FontWeightBold {
		t.Errorf("Weight = %v, want bold", te.Weight)
	}

	if te.Style != FontStyleItalic {
		t.Errorf("Style = %v, want italic", te.Style)
	}

	if te.Stretch != FontStretchCondensed {
		t.Errorf("Stretch = %v, want condensed", te.Stretch)
	}

	if te.Fill != Red {
		t.Errorf("Fill = %v, want red", te.Fill)
	}
}

func TestTextElemDecorations(t *testing.T) {
	te := New("Decorated").
		WithUnderline(NewUnderline()).
		WithStrikethrough(NewStrikethrough()).
		WithOverline(NewOverline())

	if !te.HasDecoration() {
		t.Error("HasDecoration should return true")
	}

	decos := te.Decorations()
	if len(decos) != 3 {
		t.Errorf("Decorations count = %d, want 3", len(decos))
	}
}

func TestTextElemNoDecorations(t *testing.T) {
	te := New("Plain")

	if te.HasDecoration() {
		t.Error("HasDecoration should return false for plain text")
	}

	decos := te.Decorations()
	if len(decos) != 0 {
		t.Errorf("Decorations count = %d, want 0", len(decos))
	}
}

func TestToFontVariant(t *testing.T) {
	te := New("Test").
		WithWeight(FontWeightBold).
		WithStyle(FontStyleItalic).
		WithStretch(FontStretchExpanded)

	variant := te.ToFontVariant()

	if variant.Weight != inline.FontWeightBold {
		t.Errorf("variant.Weight = %v, want bold", variant.Weight)
	}

	if variant.Style != inline.FontStyleItalic {
		t.Errorf("variant.Style = %v, want italic", variant.Style)
	}

	if variant.Stretch != inline.FontStretchExpanded {
		t.Errorf("variant.Stretch = %v, want expanded", variant.Stretch)
	}
}

func TestFontWeightStrings(t *testing.T) {
	tests := []struct {
		weight FontWeight
		want   string
	}{
		{FontWeightThin, "thin"},
		{FontWeightExtraLight, "extralight"},
		{FontWeightLight, "light"},
		{FontWeightNormal, "normal"},
		{FontWeightMedium, "medium"},
		{FontWeightSemiBold, "semibold"},
		{FontWeightBold, "bold"},
		{FontWeightExtraBold, "extrabold"},
		{FontWeightBlack, "black"},
	}

	for _, tt := range tests {
		if got := tt.weight.String(); got != tt.want {
			t.Errorf("FontWeight(%d).String() = %q, want %q", tt.weight, got, tt.want)
		}
	}
}

func TestFontStyleStrings(t *testing.T) {
	tests := []struct {
		style FontStyle
		want  string
	}{
		{FontStyleNormal, "normal"},
		{FontStyleItalic, "italic"},
		{FontStyleOblique, "oblique"},
	}

	for _, tt := range tests {
		if got := tt.style.String(); got != tt.want {
			t.Errorf("FontStyle(%d).String() = %q, want %q", tt.style, got, tt.want)
		}
	}
}

func TestFontStretchStrings(t *testing.T) {
	tests := []struct {
		stretch FontStretch
		want    string
	}{
		{FontStretchNormal, "normal"},
		{FontStretchCondensed, "condensed"},
		{FontStretchExpanded, "expanded"},
	}

	for _, tt := range tests {
		if got := tt.stretch.String(); got != tt.want {
			t.Errorf("FontStretch(%d).String() = %q, want %q", tt.stretch, got, tt.want)
		}
	}
}

func TestDirStrings(t *testing.T) {
	tests := []struct {
		dir  Dir
		want string
	}{
		{DirLTR, "ltr"},
		{DirRTL, "rtl"},
	}

	for _, tt := range tests {
		if got := tt.dir.String(); got != tt.want {
			t.Errorf("Dir(%d).String() = %q, want %q", tt.dir, got, tt.want)
		}
	}
}

func TestSizeConversion(t *testing.T) {
	s := SizeFromPt(12)

	if s.Points() != 12.0 {
		t.Errorf("Points() = %v, want 12.0", s.Points())
	}

	if s.ToAbs() != inline.Abs(12) {
		t.Errorf("ToAbs() = %v, want 12", s.ToAbs())
	}
}

func TestSizeFromEm(t *testing.T) {
	base := SizeFromPt(10)
	s := SizeFromEm(1.5, base)

	if s.Points() != 15.0 {
		t.Errorf("SizeFromEm(1.5, 10pt) = %vpt, want 15pt", s.Points())
	}
}
