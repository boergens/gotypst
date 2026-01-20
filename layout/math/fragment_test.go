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
