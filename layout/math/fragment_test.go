package math

import (
	"testing"

	"github.com/boergens/gotypst/layout"
)

func TestGlyphFragmentDimensions(t *testing.T) {
	glyph := &GlyphFragment{
		FontSize:  12,
		MathClass: ClassOrd,
		Italics:   1.5,
		Glyphs: []MathGlyph{
			{ID: 1, Advance: 8, Ascent: 10, Descent: 3},
			{ID: 2, Advance: 6, Ascent: 9, Descent: 2},
		},
	}

	if got := glyph.Width(); got != 14 {
		t.Errorf("Width() = %v, want 14", got)
	}

	if got := glyph.Ascent(); got != 10 {
		t.Errorf("Ascent() = %v, want 10", got)
	}

	if got := glyph.Descent(); got != 3 {
		t.Errorf("Descent() = %v, want 3", got)
	}

	if got := glyph.Height(); got != 13 {
		t.Errorf("Height() = %v, want 13", got)
	}

	if got := glyph.Class(); got != ClassOrd {
		t.Errorf("Class() = %v, want ClassOrd", got)
	}

	if got := glyph.ItalicsCorrection(); got != 1.5 {
		t.Errorf("ItalicsCorrection() = %v, want 1.5", got)
	}
}

func TestFrameFragmentDimensions(t *testing.T) {
	frame := &FrameFragment{
		Size:      layout.Size{Width: 20, Height: 15},
		Baseline:  10,
		MathClass: ClassInner,
		Italics:   0.5,
	}

	if got := frame.Width(); got != 20 {
		t.Errorf("Width() = %v, want 20", got)
	}

	if got := frame.Height(); got != 15 {
		t.Errorf("Height() = %v, want 15", got)
	}

	if got := frame.Ascent(); got != 10 {
		t.Errorf("Ascent() = %v, want 10", got)
	}

	if got := frame.Descent(); got != 5 {
		t.Errorf("Descent() = %v, want 5", got)
	}

	if got := frame.Class(); got != ClassInner {
		t.Errorf("Class() = %v, want ClassInner", got)
	}
}

func TestSpaceFragment(t *testing.T) {
	space := &SpaceFragment{Amount: 3.5}

	if got := space.Width(); got != 3.5 {
		t.Errorf("Width() = %v, want 3.5", got)
	}

	if got := space.Height(); got != 0 {
		t.Errorf("Height() = %v, want 0", got)
	}

	if got := space.Class(); got != ClassNone {
		t.Errorf("Class() = %v, want ClassNone", got)
	}
}

func TestLinebreakFragment(t *testing.T) {
	lb := &LinebreakFragment{}

	if got := lb.Width(); got != 0 {
		t.Errorf("Width() = %v, want 0", got)
	}

	if got := lb.Class(); got != ClassNone {
		t.Errorf("Class() = %v, want ClassNone", got)
	}
}

func TestAlignFragment(t *testing.T) {
	align := &AlignFragment{}

	if got := align.Width(); got != 0 {
		t.Errorf("Width() = %v, want 0", got)
	}

	if got := align.Class(); got != ClassNone {
		t.Errorf("Class() = %v, want ClassNone", got)
	}
}

func TestMathClassString(t *testing.T) {
	tests := []struct {
		class MathClass
		want  string
	}{
		{ClassOrd, "Ord"},
		{ClassOp, "Op"},
		{ClassBin, "Bin"},
		{ClassRel, "Rel"},
		{ClassOpen, "Open"},
		{ClassClose, "Close"},
		{ClassPunct, "Punct"},
		{ClassInner, "Inner"},
		{ClassNone, "None"},
		{MathClass(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.class.String(); got != tt.want {
			t.Errorf("MathClass(%d).String() = %v, want %v", tt.class, got, tt.want)
		}
	}
}

func TestMathFragmentInterface(t *testing.T) {
	// Verify all fragment types implement MathFragment
	var _ MathFragment = (*GlyphFragment)(nil)
	var _ MathFragment = (*FrameFragment)(nil)
	var _ MathFragment = (*SpaceFragment)(nil)
	var _ MathFragment = (*LinebreakFragment)(nil)
	var _ MathFragment = (*AlignFragment)(nil)
	var _ MathFragment = (*AccentFragment)(nil)
}

func TestEmptyGlyphFragment(t *testing.T) {
	glyph := &GlyphFragment{
		FontSize:  12,
		MathClass: ClassOrd,
		Glyphs:    []MathGlyph{},
	}

	if got := glyph.Width(); got != 0 {
		t.Errorf("Width() = %v, want 0", got)
	}

	if got := glyph.Ascent(); got != 0 {
		t.Errorf("Ascent() = %v, want 0", got)
	}

	if got := glyph.Descent(); got != 0 {
		t.Errorf("Descent() = %v, want 0", got)
	}
}

func TestFrameFragmentWithItems(t *testing.T) {
	inner := &SpaceFragment{Amount: 5}
	frame := &FrameFragment{
		Size:      layout.Size{Width: 30, Height: 20},
		Baseline:  12,
		MathClass: ClassInner,
		Items: []FrameItem{
			{Pos: layout.Point{X: 0, Y: 0}, Fragment: inner},
		},
	}

	if len(frame.Items) != 1 {
		t.Errorf("len(Items) = %v, want 1", len(frame.Items))
	}

	if frame.Items[0].Fragment.Width() != 5 {
		t.Errorf("Items[0].Fragment.Width() = %v, want 5", frame.Items[0].Fragment.Width())
	}
}

func TestAccentKindString(t *testing.T) {
	tests := []struct {
		kind AccentKind
		want string
	}{
		{AccentHat, "hat"},
		{AccentTilde, "tilde"},
		{AccentBar, "bar"},
		{AccentVec, "vec"},
		{AccentDot, "dot"},
		{AccentDDot, "ddot"},
		{AccentBreve, "breve"},
		{AccentAcute, "acute"},
		{AccentGrave, "grave"},
		{AccentKind(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.kind.String(); got != tt.want {
			t.Errorf("AccentKind(%d).String() = %v, want %v", tt.kind, got, tt.want)
		}
	}
}

func TestAccentKindGlyphID(t *testing.T) {
	tests := []struct {
		kind AccentKind
		want rune
	}{
		{AccentHat, '\u0302'},
		{AccentTilde, '\u0303'},
		{AccentBar, '\u0304'},
		{AccentVec, '\u20D7'},
		{AccentDot, '\u0307'},
		{AccentDDot, '\u0308'},
		{AccentBreve, '\u0306'},
		{AccentAcute, '\u0301'},
		{AccentGrave, '\u0300'},
	}

	for _, tt := range tests {
		if got := tt.kind.AccentGlyphID(); got != tt.want {
			t.Errorf("AccentKind(%d).AccentGlyphID() = %U, want %U", tt.kind, got, tt.want)
		}
	}
}

func TestAccentFragmentDimensions(t *testing.T) {
	// Create base fragment (simulating a letter like 'x')
	base := &GlyphFragment{
		FontSize:  12,
		MathClass: ClassOrd,
		Italics:   0.5,
		Glyphs: []MathGlyph{
			{ID: 1, Advance: 8, Ascent: 8, Descent: 2},
		},
	}

	// Create accent fragment (simulating a hat)
	accent := &GlyphFragment{
		FontSize:  12,
		MathClass: ClassOrd,
		Glyphs: []MathGlyph{
			{ID: 2, Advance: 6, Ascent: 2, Descent: 0},
		},
	}

	accentFrag := &AccentFragment{
		Base:      base,
		Accent:    accent,
		Kind:      AccentHat,
		AccentGap: 1,
	}

	// Width should be max of base and accent widths
	if got := accentFrag.Width(); got != 8 {
		t.Errorf("Width() = %v, want 8 (base width)", got)
	}

	// Ascent = base ascent (8) + gap (1) + accent height (2)
	if got := accentFrag.Ascent(); got != 11 {
		t.Errorf("Ascent() = %v, want 11", got)
	}

	// Descent from base
	if got := accentFrag.Descent(); got != 2 {
		t.Errorf("Descent() = %v, want 2", got)
	}

	// Total height = ascent + descent
	if got := accentFrag.Height(); got != 13 {
		t.Errorf("Height() = %v, want 13", got)
	}

	// Class inherits from base
	if got := accentFrag.Class(); got != ClassOrd {
		t.Errorf("Class() = %v, want ClassOrd", got)
	}

	// Italics correction from base
	if got := accentFrag.ItalicsCorrection(); got != 0.5 {
		t.Errorf("ItalicsCorrection() = %v, want 0.5", got)
	}
}

func TestAccentFragmentCentering(t *testing.T) {
	// Test case: base is wider than accent
	wideBase := &GlyphFragment{
		FontSize:  12,
		MathClass: ClassOrd,
		Glyphs: []MathGlyph{
			{ID: 1, Advance: 20, Ascent: 8, Descent: 2},
		},
	}
	narrowAccent := &GlyphFragment{
		FontSize:  12,
		MathClass: ClassOrd,
		Glyphs: []MathGlyph{
			{ID: 2, Advance: 10, Ascent: 2, Descent: 0},
		},
	}

	accentFrag := &AccentFragment{
		Base:      wideBase,
		Accent:    narrowAccent,
		Kind:      AccentHat,
		AccentGap: 1,
	}

	// Base offset should be 0 (no centering needed)
	if got := accentFrag.BaseOffset(); got != 0 {
		t.Errorf("BaseOffset() = %v, want 0", got)
	}

	// Accent offset should center the accent over the base
	// (20 - 10) / 2 = 5
	if got := accentFrag.AccentOffset(); got != 5 {
		t.Errorf("AccentOffset() = %v, want 5", got)
	}

	// Width should be the base width
	if got := accentFrag.Width(); got != 20 {
		t.Errorf("Width() = %v, want 20", got)
	}
}

func TestAccentFragmentWideAccent(t *testing.T) {
	// Test case: accent is wider than base (e.g., long overline)
	narrowBase := &GlyphFragment{
		FontSize:  12,
		MathClass: ClassOrd,
		Glyphs: []MathGlyph{
			{ID: 1, Advance: 8, Ascent: 8, Descent: 2},
		},
	}
	wideAccent := &GlyphFragment{
		FontSize:  12,
		MathClass: ClassOrd,
		Glyphs: []MathGlyph{
			{ID: 2, Advance: 16, Ascent: 2, Descent: 0},
		},
	}

	accentFrag := &AccentFragment{
		Base:      narrowBase,
		Accent:    wideAccent,
		Kind:      AccentBar,
		AccentGap: 1,
	}

	// Base offset should center the base under the accent
	// (16 - 8) / 2 = 4
	if got := accentFrag.BaseOffset(); got != 4 {
		t.Errorf("BaseOffset() = %v, want 4", got)
	}

	// Accent offset should be 0
	if got := accentFrag.AccentOffset(); got != 0 {
		t.Errorf("AccentOffset() = %v, want 0", got)
	}

	// Width should be the accent width
	if got := accentFrag.Width(); got != 16 {
		t.Errorf("Width() = %v, want 16", got)
	}
}

func TestAccentFragmentY(t *testing.T) {
	base := &GlyphFragment{
		FontSize:  12,
		MathClass: ClassOrd,
		Glyphs: []MathGlyph{
			{ID: 1, Advance: 8, Ascent: 10, Descent: 3},
		},
	}
	accent := &GlyphFragment{
		FontSize:  12,
		MathClass: ClassOrd,
		Glyphs: []MathGlyph{
			{ID: 2, Advance: 6, Ascent: 2, Descent: 0},
		},
	}

	accentFrag := &AccentFragment{
		Base:      base,
		Accent:    accent,
		Kind:      AccentHat,
		AccentGap: 2,
	}

	// AccentY = base ascent + gap = 10 + 2 = 12
	if got := accentFrag.AccentY(); got != 12 {
		t.Errorf("AccentY() = %v, want 12", got)
	}
}
