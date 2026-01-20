package html

import (
	"strings"
	"testing"

	"github.com/boergens/gotypst/layout"
	"github.com/boergens/gotypst/layout/inline"
)

func TestColorCSS(t *testing.T) {
	tests := []struct {
		name  string
		color *Color
		want  string
	}{
		{
			name:  "nil color",
			color: nil,
			want:  "",
		},
		{
			name:  "opaque black",
			color: &Color{R: 0, G: 0, B: 0, A: 255},
			want:  "rgb(0, 0, 0)",
		},
		{
			name:  "opaque white",
			color: &Color{R: 255, G: 255, B: 255, A: 255},
			want:  "rgb(255, 255, 255)",
		},
		{
			name:  "opaque red",
			color: &Color{R: 255, G: 0, B: 0, A: 255},
			want:  "rgb(255, 0, 0)",
		},
		{
			name:  "semi-transparent blue",
			color: &Color{R: 0, G: 0, B: 255, A: 128},
			want:  "rgba(0, 0, 255, 0.502)",
		},
		{
			name:  "transparent",
			color: &Color{R: 0, G: 0, B: 0, A: 0},
			want:  "rgba(0, 0, 0, 0.000)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.color.CSS()
			if got != tt.want {
				t.Errorf("Color.CSS() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestColorHex(t *testing.T) {
	tests := []struct {
		name  string
		color *Color
		want  string
	}{
		{
			name:  "nil color",
			color: nil,
			want:  "",
		},
		{
			name:  "black",
			color: &Color{R: 0, G: 0, B: 0, A: 255},
			want:  "#000000",
		},
		{
			name:  "white",
			color: &Color{R: 255, G: 255, B: 255, A: 255},
			want:  "#ffffff",
		},
		{
			name:  "red",
			color: &Color{R: 255, G: 0, B: 0, A: 255},
			want:  "#ff0000",
		},
		{
			name:  "custom color",
			color: &Color{R: 18, G: 52, B: 86, A: 255},
			want:  "#123456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.color.Hex()
			if got != tt.want {
				t.Errorf("Color.Hex() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStyleBuilderPosition(t *testing.T) {
	s := NewStyleBuilder()
	s.Absolute().Left(10).Top(20)

	got := s.String()
	if !strings.Contains(got, "position: absolute") {
		t.Errorf("expected 'position: absolute' in %q", got)
	}
	if !strings.Contains(got, "left: 10pt") {
		t.Errorf("expected 'left: 10pt' in %q", got)
	}
	if !strings.Contains(got, "top: 20pt") {
		t.Errorf("expected 'top: 20pt' in %q", got)
	}
}

func TestStyleBuilderSize(t *testing.T) {
	s := NewStyleBuilder()
	s.Size(layout.Size{Width: 100, Height: 50})

	got := s.String()
	if !strings.Contains(got, "width: 100pt") {
		t.Errorf("expected 'width: 100pt' in %q", got)
	}
	if !strings.Contains(got, "height: 50pt") {
		t.Errorf("expected 'height: 50pt' in %q", got)
	}
}

func TestStyleBuilderColor(t *testing.T) {
	red := &Color{R: 255, G: 0, B: 0, A: 255}
	blue := &Color{R: 0, G: 0, B: 255, A: 128}

	s := NewStyleBuilder()
	s.Color(red).Background(blue)

	got := s.String()
	if !strings.Contains(got, "color: rgb(255, 0, 0)") {
		t.Errorf("expected text color in %q", got)
	}
	if !strings.Contains(got, "background: rgba(0, 0, 255,") {
		t.Errorf("expected background color in %q", got)
	}
}

func TestStyleBuilderFont(t *testing.T) {
	s := NewStyleBuilder()
	s.FontSize(12).FontFamily("serif").FontWeight(inline.FontWeightBold)

	got := s.String()
	if !strings.Contains(got, "font-size: 12pt") {
		t.Errorf("expected font-size in %q", got)
	}
	if !strings.Contains(got, "font-family: serif") {
		t.Errorf("expected font-family in %q", got)
	}
	if !strings.Contains(got, "font-weight: 700") {
		t.Errorf("expected font-weight in %q", got)
	}
}

func TestStyleBuilderFontStyle(t *testing.T) {
	tests := []struct {
		name  string
		style inline.FontStyle
		want  string
	}{
		{"normal", inline.FontStyleNormal, "font-style: normal"},
		{"italic", inline.FontStyleItalic, "font-style: italic"},
		{"oblique", inline.FontStyleOblique, "font-style: oblique"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStyleBuilder()
			s.FontStyle(tt.style)
			got := s.String()
			if !strings.Contains(got, tt.want) {
				t.Errorf("expected %q in %q", tt.want, got)
			}
		})
	}
}

func TestStyleBuilderFontStretch(t *testing.T) {
	tests := []struct {
		name    string
		stretch inline.FontStretch
		want    string
	}{
		{"normal", inline.FontStretchNormal, "font-stretch: normal"},
		{"condensed", inline.FontStretchCondensed, "font-stretch: condensed"},
		{"expanded", inline.FontStretchExpanded, "font-stretch: expanded"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStyleBuilder()
			s.FontStretch(tt.stretch)
			got := s.String()
			if !strings.Contains(got, tt.want) {
				t.Errorf("expected %q in %q", tt.want, got)
			}
		})
	}
}

func TestStyleBuilderTextAlign(t *testing.T) {
	tests := []struct {
		name  string
		align layout.Alignment
		want  string
	}{
		{"start", layout.AlignStart, "text-align: start"},
		{"center", layout.AlignCenter, "text-align: center"},
		{"end", layout.AlignEnd, "text-align: end"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStyleBuilder()
			s.TextAlign(tt.align)
			got := s.String()
			if !strings.Contains(got, tt.want) {
				t.Errorf("expected %q in %q", tt.want, got)
			}
		})
	}
}

func TestStyleBuilderDirection(t *testing.T) {
	tests := []struct {
		name string
		dir  layout.Dir
		want string
	}{
		{"ltr", layout.DirLTR, "direction: ltr"},
		{"rtl", layout.DirRTL, "direction: rtl"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStyleBuilder()
			s.Direction(tt.dir)
			got := s.String()
			if !strings.Contains(got, tt.want) {
				t.Errorf("expected %q in %q", tt.want, got)
			}
		})
	}
}

func TestStyleBuilderBorder(t *testing.T) {
	red := &Color{R: 255, G: 0, B: 0, A: 255}

	s := NewStyleBuilder()
	s.Border(2, "solid", red).BorderRadius(5)

	got := s.String()
	if !strings.Contains(got, "border: 2pt solid rgb(255, 0, 0)") {
		t.Errorf("expected border in %q", got)
	}
	if !strings.Contains(got, "border-radius: 5pt") {
		t.Errorf("expected border-radius in %q", got)
	}
}

func TestStyleBuilderMarginPadding(t *testing.T) {
	s := NewStyleBuilder()
	s.MarginSides(1, 2, 3, 4).PaddingSides(5, 6, 7, 8)

	got := s.String()
	if !strings.Contains(got, "margin: 1pt 2pt 3pt 4pt") {
		t.Errorf("expected margin in %q", got)
	}
	if !strings.Contains(got, "padding: 5pt 6pt 7pt 8pt") {
		t.Errorf("expected padding in %q", got)
	}
}

func TestStyleBuilderTextDecoration(t *testing.T) {
	s := NewStyleBuilder()
	red := &Color{R: 255, G: 0, B: 0, A: 255}
	s.TextDecorationLine("underline").
		TextDecorationColor(red).
		TextDecorationStyle("solid").
		TextDecorationThickness(2)

	got := s.String()
	if !strings.Contains(got, "text-decoration-line: underline") {
		t.Errorf("expected text-decoration-line in %q", got)
	}
	if !strings.Contains(got, "text-decoration-color: rgb(255, 0, 0)") {
		t.Errorf("expected text-decoration-color in %q", got)
	}
	if !strings.Contains(got, "text-decoration-style: solid") {
		t.Errorf("expected text-decoration-style in %q", got)
	}
	if !strings.Contains(got, "text-decoration-thickness: 2pt") {
		t.Errorf("expected text-decoration-thickness in %q", got)
	}
}

func TestStyleBuilderReset(t *testing.T) {
	s := NewStyleBuilder()
	s.Color(&Color{R: 255, G: 0, B: 0, A: 255})

	if s.String() == "" {
		t.Error("expected non-empty string before reset")
	}

	s.Reset()

	if s.String() != "" {
		t.Errorf("expected empty string after reset, got %q", s.String())
	}
}

func TestStyleBuilderTransform(t *testing.T) {
	s := NewStyleBuilder()
	s.Transform("rotate(45deg)").TransformOrigin("center")

	got := s.String()
	if !strings.Contains(got, "transform: rotate(45deg)") {
		t.Errorf("expected transform in %q", got)
	}
	if !strings.Contains(got, "transform-origin: center") {
		t.Errorf("expected transform-origin in %q", got)
	}
}

func TestStyleBuilderRaw(t *testing.T) {
	s := NewStyleBuilder()
	s.Raw("custom-property", "custom-value")

	got := s.String()
	if !strings.Contains(got, "custom-property: custom-value") {
		t.Errorf("expected custom property in %q", got)
	}
}

func TestStyleBuilderChaining(t *testing.T) {
	s := NewStyleBuilder()
	result := s.Absolute().Left(10).Top(20).Width(100).Height(50)

	// Verify chaining returns the same builder
	if result != s {
		t.Error("chaining should return the same builder")
	}
}

func TestFormatPt(t *testing.T) {
	tests := []struct {
		v    layout.Abs
		want string
	}{
		{0, "0pt"},
		{10, "10pt"},
		{12.5, "12.50pt"},
		{100, "100pt"},
	}

	for _, tt := range tests {
		got := formatPt(tt.v)
		if got != tt.want {
			t.Errorf("formatPt(%v) = %q, want %q", tt.v, got, tt.want)
		}
	}
}

func TestLineCap(t *testing.T) {
	tests := []struct {
		cap  inline.LineCap
		want string
	}{
		{inline.LineCapButt, "butt"},
		{inline.LineCapRound, "round"},
		{inline.LineCapSquare, "square"},
	}

	for _, tt := range tests {
		got := LineCap(tt.cap)
		if got != tt.want {
			t.Errorf("LineCap(%v) = %q, want %q", tt.cap, got, tt.want)
		}
	}
}

func TestLineJoin(t *testing.T) {
	tests := []struct {
		join inline.LineJoin
		want string
	}{
		{inline.LineJoinMiter, "miter"},
		{inline.LineJoinRound, "round"},
		{inline.LineJoinBevel, "bevel"},
	}

	for _, tt := range tests {
		got := LineJoin(tt.join)
		if got != tt.want {
			t.Errorf("LineJoin(%v) = %q, want %q", tt.join, got, tt.want)
		}
	}
}

func TestCSSClass(t *testing.T) {
	s := NewStyleBuilder()
	s.Color(&Color{R: 255, G: 0, B: 0, A: 255})

	class := CSSClass{Name: "highlight", Styles: s}
	got := class.String()

	if !strings.HasPrefix(got, ".highlight {") {
		t.Errorf("expected class to start with '.highlight {', got %q", got)
	}
	if !strings.Contains(got, "color: rgb(255, 0, 0)") {
		t.Errorf("expected color in class, got %q", got)
	}
	if !strings.HasSuffix(got, "}") {
		t.Errorf("expected class to end with '}', got %q", got)
	}
}

func TestCSSClassWithPseudo(t *testing.T) {
	s := NewStyleBuilder()
	s.Raw("content", `""`)

	class := CSSClass{Name: "test", Styles: s, PseudoElem: "::before"}
	got := class.String()

	if !strings.Contains(got, ".test::before") {
		t.Errorf("expected '.test::before' in %q", got)
	}
}

func TestStylesheet(t *testing.T) {
	ss := NewStylesheet()

	style1 := NewStyleBuilder()
	style1.Color(&Color{R: 255, G: 0, B: 0, A: 255})
	ss.AddClass("red", style1)

	style2 := NewStyleBuilder()
	style2.FontSize(14)
	ss.AddClass("text", style2)

	ss.AddRaw("body { margin: 0; }")

	got := ss.String()

	if !strings.Contains(got, ".red {") {
		t.Errorf("expected .red class in %q", got)
	}
	if !strings.Contains(got, ".text {") {
		t.Errorf("expected .text class in %q", got)
	}
	if !strings.Contains(got, "body { margin: 0; }") {
		t.Errorf("expected raw CSS in %q", got)
	}
}

func TestBaseStyles(t *testing.T) {
	ss := BaseStyles()
	got := ss.String()

	// Check for reset styles
	if !strings.Contains(got, "box-sizing: border-box") {
		t.Errorf("expected box-sizing reset in %q", got)
	}

	// Check for page class
	if !strings.Contains(got, ".page {") {
		t.Errorf("expected .page class in %q", got)
	}

	// Check for content class
	if !strings.Contains(got, ".content {") {
		t.Errorf("expected .content class in %q", got)
	}

	// Check for text class
	if !strings.Contains(got, ".text {") {
		t.Errorf("expected .text class in %q", got)
	}
}

func TestStrokeStyle(t *testing.T) {
	red := &Color{R: 255, G: 0, B: 0, A: 255}
	stroke := &inline.FixedStroke{
		Paint:     red,
		Thickness: 2,
		LineCap:   inline.LineCapRound,
		LineJoin:  inline.LineJoinRound,
		DashArray: nil,
	}

	s := NewStyleBuilder()
	s.StrokeStyle(stroke)

	got := s.String()
	if !strings.Contains(got, "border-width: 2pt") {
		t.Errorf("expected border-width in %q", got)
	}
	if !strings.Contains(got, "border-style: solid") {
		t.Errorf("expected border-style: solid in %q", got)
	}
}

func TestStrokeStyleDashed(t *testing.T) {
	red := &Color{R: 255, G: 0, B: 0, A: 255}
	stroke := &inline.FixedStroke{
		Paint:     red,
		Thickness: 2,
		DashArray: []layout.Abs{5, 3},
	}

	s := NewStyleBuilder()
	s.StrokeStyle(stroke)

	got := s.String()
	if !strings.Contains(got, "border-style: dashed") {
		t.Errorf("expected border-style: dashed in %q", got)
	}
}

func TestStyleBuilderEmptyValues(t *testing.T) {
	s := NewStyleBuilder()

	// Empty font family should not add property
	s.FontFamily("")
	if s.String() != "" {
		t.Errorf("empty font family should not add property, got %q", s.String())
	}

	// Nil color should not add property
	s.Reset()
	s.Color(nil)
	if s.String() != "" {
		t.Errorf("nil color should not add property, got %q", s.String())
	}

	// Nil background should not add property
	s.Reset()
	s.Background(nil)
	if s.String() != "" {
		t.Errorf("nil background should not add property, got %q", s.String())
	}
}

func TestStylesheetWriteTo(t *testing.T) {
	ss := NewStylesheet()
	style := NewStyleBuilder()
	style.Color(&Color{R: 255, G: 0, B: 0, A: 255})
	ss.AddClass("test", style)

	var buf strings.Builder
	n, err := ss.WriteTo(&buf)

	if err != nil {
		t.Errorf("WriteTo error: %v", err)
	}
	if n == 0 {
		t.Error("WriteTo wrote 0 bytes")
	}
	if !strings.Contains(buf.String(), ".test {") {
		t.Errorf("WriteTo output missing class: %q", buf.String())
	}
}
