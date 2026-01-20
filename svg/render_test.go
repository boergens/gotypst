package svg

import (
	"strings"
	"testing"

	"github.com/boergens/gotypst/layout"
	"github.com/boergens/gotypst/layout/inline"
	"github.com/boergens/gotypst/layout/pages"
	"github.com/boergens/gotypst/library/visualize"
)

func TestRenderer_RenderPage_Empty(t *testing.T) {
	r := NewRenderer()

	page := &pages.Page{
		Frame: pages.Frame{
			Size: layout.Size{Width: 100, Height: 200},
		},
	}

	svg := r.RenderPage(page)

	if !strings.Contains(svg, `<svg`) {
		t.Error("missing svg opening tag")
	}
	if !strings.Contains(svg, `</svg>`) {
		t.Error("missing svg closing tag")
	}
	if !strings.Contains(svg, `width="100"`) {
		t.Error("missing width attribute")
	}
	if !strings.Contains(svg, `height="200"`) {
		t.Error("missing height attribute")
	}
	if !strings.Contains(svg, `viewBox="0 0 100 200"`) {
		t.Error("missing viewBox attribute")
	}
}

func TestRenderer_RenderPage_WithBackground(t *testing.T) {
	r := NewRenderer()

	page := &pages.Page{
		Frame: pages.Frame{
			Size: layout.Size{Width: 100, Height: 200},
		},
		Fill: &pages.Paint{
			Color: &pages.Color{R: 255, G: 0, B: 0, A: 255},
		},
	}

	svg := r.RenderPage(page)

	if !strings.Contains(svg, `fill="#ff0000"`) {
		t.Errorf("expected red background, got: %s", svg)
	}
	if !strings.Contains(svg, `<rect`) {
		t.Error("missing background rect")
	}
}

func TestRenderer_RenderPage_WithText(t *testing.T) {
	r := NewRenderer()

	page := &pages.Page{
		Frame: pages.Frame{
			Size: layout.Size{Width: 100, Height: 200},
			Items: []pages.PositionedItem{
				{
					Pos: layout.Point{X: 10, Y: 20},
					Item: pages.TextItem{
						Text:     "Hello",
						FontSize: 12,
					},
				},
			},
		},
	}

	svg := r.RenderPage(page)

	if !strings.Contains(svg, `<text`) {
		t.Error("missing text element")
	}
	if !strings.Contains(svg, `x="10"`) {
		t.Error("missing x position")
	}
	if !strings.Contains(svg, `>Hello</text>`) {
		t.Error("missing text content")
	}
}

func TestRenderer_RenderPage_WithNestedFrame(t *testing.T) {
	r := NewRenderer()

	page := &pages.Page{
		Frame: pages.Frame{
			Size: layout.Size{Width: 100, Height: 200},
			Items: []pages.PositionedItem{
				{
					Pos: layout.Point{X: 10, Y: 20},
					Item: pages.GroupItem{
						Frame: pages.Frame{
							Size: layout.Size{Width: 50, Height: 50},
							Items: []pages.PositionedItem{
								{
									Pos: layout.Point{X: 5, Y: 5},
									Item: pages.TextItem{
										Text:     "Nested",
										FontSize: 10,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	svg := r.RenderPage(page)

	if !strings.Contains(svg, `<text`) {
		t.Error("missing text element")
	}
	// Position should be 10+5=15 for X
	if !strings.Contains(svg, `x="15"`) {
		t.Errorf("expected x=15, got: %s", svg)
	}
}

func TestColorToSVG(t *testing.T) {
	tests := []struct {
		color *pages.Color
		want  string
	}{
		{&pages.Color{R: 255, G: 0, B: 0, A: 255}, "#ff0000"},
		{&pages.Color{R: 0, G: 255, B: 0, A: 255}, "#00ff00"},
		{&pages.Color{R: 0, G: 0, B: 255, A: 255}, "#0000ff"},
		{&pages.Color{R: 0, G: 0, B: 0, A: 255}, "#000000"},
		{&pages.Color{R: 255, G: 255, B: 255, A: 255}, "#ffffff"},
		{&pages.Color{R: 255, G: 0, B: 0, A: 128}, "rgba(255,0,0,0.502)"},
	}

	for _, tt := range tests {
		got := colorToSVG(tt.color)
		if got != tt.want {
			t.Errorf("colorToSVG(%+v) = %q, want %q", tt.color, got, tt.want)
		}
	}
}

func TestStrokeToSVG(t *testing.T) {
	stroke := &inline.FixedStroke{
		Paint:     &pages.Color{R: 255, G: 0, B: 0, A: 255},
		Thickness: 2,
		LineCap:   inline.LineCapRound,
		LineJoin:  inline.LineJoinBevel,
		DashArray: []layout.Abs{4, 2},
		DashPhase: 1,
	}

	result := strokeToSVG(stroke)

	if !strings.Contains(result, `stroke="#ff0000"`) {
		t.Errorf("missing stroke color, got: %s", result)
	}
	if !strings.Contains(result, `stroke-width="2"`) {
		t.Errorf("missing stroke width, got: %s", result)
	}
	if !strings.Contains(result, `stroke-linecap="round"`) {
		t.Errorf("missing linecap, got: %s", result)
	}
	if !strings.Contains(result, `stroke-linejoin="bevel"`) {
		t.Errorf("missing linejoin, got: %s", result)
	}
	if !strings.Contains(result, `stroke-dasharray="4 2"`) {
		t.Errorf("missing dasharray, got: %s", result)
	}
	if !strings.Contains(result, `stroke-dashoffset="1"`) {
		t.Errorf("missing dashoffset, got: %s", result)
	}
}

func TestEscapeXML(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"<tag>", "&lt;tag&gt;"},
		{"a & b", "a &amp; b"},
		{`"quoted"`, "&quot;quoted&quot;"},
		{"it's", "it&apos;s"},
	}

	for _, tt := range tests {
		got := escapeXML(tt.input)
		if got != tt.want {
			t.Errorf("escapeXML(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSegmentsToSVGPath(t *testing.T) {
	origin := layout.Point{X: 0, Y: 0}

	t.Run("LineSegment", func(t *testing.T) {
		segments := []inline.PathSegment{
			&inline.LineSegment{X0: 0, Y0: 0, X1: 100, Y1: 100},
		}
		result := segmentsToSVGPath(segments, origin)
		if !strings.Contains(result, "M0 0") {
			t.Errorf("expected M0 0, got: %s", result)
		}
		if !strings.Contains(result, "L100 100") {
			t.Errorf("expected L100 100, got: %s", result)
		}
	})

	t.Run("QuadSegment", func(t *testing.T) {
		segments := []inline.PathSegment{
			&inline.QuadSegment{X0: 0, Y0: 0, X1: 50, Y1: 50, X2: 100, Y2: 0},
		}
		result := segmentsToSVGPath(segments, origin)
		if !strings.Contains(result, "M0 0") {
			t.Errorf("expected M0 0, got: %s", result)
		}
		if !strings.Contains(result, "Q50 50 100 0") {
			t.Errorf("expected Q50 50 100 0, got: %s", result)
		}
	})

	t.Run("CubicSegment", func(t *testing.T) {
		segments := []inline.PathSegment{
			&inline.CubicSegment{X0: 0, Y0: 0, X1: 25, Y1: 50, X2: 75, Y2: 50, X3: 100, Y3: 0},
		}
		result := segmentsToSVGPath(segments, origin)
		if !strings.Contains(result, "M0 0") {
			t.Errorf("expected M0 0, got: %s", result)
		}
		if !strings.Contains(result, "C25 50 75 50 100 0") {
			t.Errorf("expected C25 50 75 50 100 0, got: %s", result)
		}
	})

	t.Run("WithOffset", func(t *testing.T) {
		offset := layout.Point{X: 10, Y: 20}
		segments := []inline.PathSegment{
			&inline.LineSegment{X0: 0, Y0: 0, X1: 100, Y1: 100},
		}
		result := segmentsToSVGPath(segments, offset)
		if !strings.Contains(result, "M10 20") {
			t.Errorf("expected M10 20, got: %s", result)
		}
		if !strings.Contains(result, "L110 120") {
			t.Errorf("expected L110 120, got: %s", result)
		}
	})
}

func TestRenderer_DrawPath(t *testing.T) {
	r := NewRenderer()
	var b strings.Builder

	segments := []inline.PathSegment{
		&inline.LineSegment{X0: 0, Y0: 0, X1: 100, Y1: 100},
	}
	stroke := &inline.FixedStroke{
		Paint:     &pages.Color{R: 0, G: 0, B: 0, A: 255},
		Thickness: 1,
	}

	r.DrawPath(&b, segments, layout.Point{X: 0, Y: 0}, stroke, nil)

	result := b.String()
	if !strings.Contains(result, `<path`) {
		t.Error("missing path element")
	}
	if !strings.Contains(result, `d="M0 0 L100 100"`) {
		t.Errorf("unexpected path data, got: %s", result)
	}
	if !strings.Contains(result, `fill="none"`) {
		t.Error("expected fill=none for stroke-only path")
	}
}

func TestRenderer_RenderDocument(t *testing.T) {
	r := NewRenderer()

	doc := &pages.PagedDocument{
		Pages: []pages.Page{
			{
				Frame: pages.Frame{
					Size: layout.Size{Width: 100, Height: 100},
				},
			},
			{
				Frame: pages.Frame{
					Size: layout.Size{Width: 200, Height: 200},
				},
			},
		},
	}

	svgs := r.RenderDocument(doc)

	if len(svgs) != 2 {
		t.Errorf("expected 2 pages, got %d", len(svgs))
	}
	if !strings.Contains(svgs[0], `width="100"`) {
		t.Error("first page should have width=100")
	}
	if !strings.Contains(svgs[1], `width="200"`) {
		t.Error("second page should have width=200")
	}
}

func TestLinearGradientDef(t *testing.T) {
	red := visualize.NewColorRGB(255, 0, 0)
	blue := visualize.NewColorRGB(0, 0, 255)

	gradient := visualize.NewLinearGradientFromColors([]*visualize.Color{red, blue})

	var b strings.Builder
	writeGradientDef(&b, gradient, "grad0")
	result := b.String()

	if !strings.Contains(result, `<linearGradient id="grad0"`) {
		t.Errorf("missing linearGradient element, got: %s", result)
	}
	if !strings.Contains(result, `<stop offset="0.0%"`) {
		t.Errorf("missing first stop, got: %s", result)
	}
	if !strings.Contains(result, `<stop offset="100.0%"`) {
		t.Errorf("missing last stop, got: %s", result)
	}
	if !strings.Contains(result, `stop-color="#ff0000"`) {
		t.Errorf("missing red color, got: %s", result)
	}
	if !strings.Contains(result, `stop-color="#0000ff"`) {
		t.Errorf("missing blue color, got: %s", result)
	}
	if !strings.Contains(result, `</linearGradient>`) {
		t.Errorf("missing closing tag, got: %s", result)
	}
}

func TestRadialGradientDef(t *testing.T) {
	red := visualize.NewColorRGB(255, 0, 0)
	blue := visualize.NewColorRGB(0, 0, 255)

	gradient := visualize.NewRadialGradientFromColors([]*visualize.Color{red, blue})

	var b strings.Builder
	writeGradientDef(&b, gradient, "grad0")
	result := b.String()

	if !strings.Contains(result, `<radialGradient id="grad0"`) {
		t.Errorf("missing radialGradient element, got: %s", result)
	}
	if !strings.Contains(result, `cx="50.0%"`) {
		t.Errorf("missing center x, got: %s", result)
	}
	if !strings.Contains(result, `cy="50.0%"`) {
		t.Errorf("missing center y, got: %s", result)
	}
	if !strings.Contains(result, `</radialGradient>`) {
		t.Errorf("missing closing tag, got: %s", result)
	}
}

func TestFillToSVGWithGradient(t *testing.T) {
	ctx := newRenderContext()

	red := visualize.NewColorRGB(255, 0, 0)
	blue := visualize.NewColorRGB(0, 0, 255)
	gradient := visualize.NewLinearGradientFromColors([]*visualize.Color{red, blue})

	result := fillToSVGWithContext(ctx, gradient)

	if !strings.Contains(result, `fill="url(#grad0)"`) {
		t.Errorf("expected gradient reference, got: %s", result)
	}
	if len(ctx.gradients) != 1 {
		t.Errorf("expected 1 gradient registered, got %d", len(ctx.gradients))
	}
}

func TestStrokeToSVGWithGradient(t *testing.T) {
	ctx := newRenderContext()

	red := visualize.NewColorRGB(255, 0, 0)
	blue := visualize.NewColorRGB(0, 0, 255)
	gradient := visualize.NewLinearGradientFromColors([]*visualize.Color{red, blue})

	stroke := &inline.FixedStroke{
		Paint:     gradient,
		Thickness: 2,
	}

	result := strokeToSVGWithContext(ctx, stroke)

	if !strings.Contains(result, `stroke="url(#grad0)"`) {
		t.Errorf("expected gradient reference, got: %s", result)
	}
	if len(ctx.gradients) != 1 {
		t.Errorf("expected 1 gradient registered, got %d", len(ctx.gradients))
	}
}

func TestVisualizeColorToSVG(t *testing.T) {
	tests := []struct {
		color *visualize.Color
		want  string
	}{
		{visualize.NewColorRGB(255, 0, 0), "#ff0000"},
		{visualize.NewColorRGB(0, 255, 0), "#00ff00"},
		{visualize.NewColorRGB(0, 0, 255), "#0000ff"},
		{visualize.NewColor(255, 0, 0, 128), "rgba(255,0,0,0.502)"},
		{nil, "#000000"},
	}

	for _, tt := range tests {
		got := visualizeColorToSVG(tt.color)
		if got != tt.want {
			t.Errorf("visualizeColorToSVG(%+v) = %q, want %q", tt.color, got, tt.want)
		}
	}
}

func TestNormalizeGradientStops(t *testing.T) {
	red := visualize.NewColorRGB(255, 0, 0)
	green := visualize.NewColorRGB(0, 255, 0)
	blue := visualize.NewColorRGB(0, 0, 255)

	// Stops without explicit offsets
	stops := []visualize.GradientStop{
		visualize.NewGradientStopAuto(red),
		visualize.NewGradientStopAuto(green),
		visualize.NewGradientStopAuto(blue),
	}

	normalized := normalizeGradientStops(stops)

	if len(normalized) != 3 {
		t.Fatalf("expected 3 stops, got %d", len(normalized))
	}

	// First should be 0%
	if normalized[0].Offset == nil || *normalized[0].Offset != 0.0 {
		t.Errorf("first stop offset = %v, want 0.0", normalized[0].Offset)
	}

	// Last should be 100%
	if normalized[2].Offset == nil || *normalized[2].Offset != 1.0 {
		t.Errorf("last stop offset = %v, want 1.0", normalized[2].Offset)
	}

	// Middle should be 50%
	if normalized[1].Offset == nil || *normalized[1].Offset != 0.5 {
		t.Errorf("middle stop offset = %v, want 0.5", normalized[1].Offset)
	}
}

// mockGlyphOutliner is a test implementation of GlyphOutliner.
type mockGlyphOutliner struct {
	unitsPerEm float64
	glyphs     map[uint16][]inline.PathSegment
}

func (m *mockGlyphOutliner) OutlineGlyph(glyphID uint16) ([]inline.PathSegment, bool) {
	segments, ok := m.glyphs[glyphID]
	return segments, ok
}

func (m *mockGlyphOutliner) UnitsPerEm() float64 {
	return m.unitsPerEm
}

func TestRenderTextAsPath_NilOutliner(t *testing.T) {
	r := NewRenderer()
	var b strings.Builder

	// With nil outliner, should fall back to text element
	// We can't easily test this without a real ShapedText, but we can verify no panic
	r.RenderTextAsPath(&b, nil, layout.Point{}, 0, nil, nil)

	if b.Len() != 0 {
		t.Errorf("expected empty output for nil text, got: %s", b.String())
	}
}

func TestRenderTextAsPath_WithMockOutliner(t *testing.T) {
	r := NewRenderer()
	var b strings.Builder

	outliner := &mockGlyphOutliner{
		unitsPerEm: 1000,
		glyphs: map[uint16][]inline.PathSegment{
			1: {
				&inline.LineSegment{X0: 0, Y0: 0, X1: 100, Y1: 0},
				&inline.LineSegment{X0: 100, Y0: 0, X1: 100, Y1: 100},
			},
		},
	}

	// Create a minimal ShapedText for testing
	glyphs := inline.NewGlyphsFromSlice([]inline.ShapedGlyph{
		{
			GlyphID:  1,
			XAdvance: inline.Em(500),
			Size:     12,
		},
	})

	text := &inline.ShapedText{
		Glyphs: glyphs,
	}

	r.RenderTextAsPath(&b, text, layout.Point{X: 10, Y: 20}, 10, outliner, nil)

	result := b.String()

	// Should produce a path element
	if !strings.Contains(result, "<path") {
		t.Errorf("expected path element, got: %s", result)
	}
	if !strings.Contains(result, `fill="#000000"`) {
		t.Errorf("expected default fill, got: %s", result)
	}
}

func TestGlyphOutlinerInterface(t *testing.T) {
	// Verify the interface is correctly defined
	var _ GlyphOutliner = &mockGlyphOutliner{}
}
