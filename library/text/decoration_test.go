package text

import (
	"testing"

	"github.com/boergens/gotypst/layout/inline"
)

func TestNewUnderline(t *testing.T) {
	u := NewUnderline()

	if !u.Evade {
		t.Error("Underline.Evade should default to true")
	}
	if u.Background {
		t.Error("Underline.Background should default to false")
	}
	if u.Stroke != nil {
		t.Error("Underline.Stroke should default to nil")
	}
	if u.Offset != nil {
		t.Error("Underline.Offset should default to nil")
	}
}

func TestUnderlineBuilders(t *testing.T) {
	offset := inline.Abs(2)
	stroke := NewStroke(Red, inline.Abs(1))

	u := NewUnderline().
		WithStroke(stroke).
		WithOffset(offset).
		WithEvade(false).
		WithBackground(true)

	if u.Stroke != stroke {
		t.Error("Underline.Stroke not set correctly")
	}
	if u.Offset == nil || *u.Offset != offset {
		t.Error("Underline.Offset not set correctly")
	}
	if u.Evade {
		t.Error("Underline.Evade should be false")
	}
	if !u.Background {
		t.Error("Underline.Background should be true")
	}
}

func TestUnderlineToDecoLine(t *testing.T) {
	u := NewUnderline().WithEvade(true).WithBackground(false)

	decoLine := u.ToDecoLine(Black, inline.Abs(12))

	underlineDeco, ok := decoLine.(*inline.UnderlineDeco)
	if !ok {
		t.Fatal("ToDecoLine should return *inline.UnderlineDeco")
	}

	if !underlineDeco.Evade {
		t.Error("UnderlineDeco.Evade should be true")
	}
	if underlineDeco.Background {
		t.Error("UnderlineDeco.Background should be false")
	}
}

func TestNewStrikethrough(t *testing.T) {
	s := NewStrikethrough()

	if s.Background {
		t.Error("Strikethrough.Background should default to false")
	}
	if s.Stroke != nil {
		t.Error("Strikethrough.Stroke should default to nil")
	}
	if s.Offset != nil {
		t.Error("Strikethrough.Offset should default to nil")
	}
}

func TestStrikethroughBuilders(t *testing.T) {
	offset := inline.Abs(0.5)
	stroke := NewStroke(Blue, inline.Abs(0.8))

	s := NewStrikethrough().
		WithStroke(stroke).
		WithOffset(offset).
		WithBackground(true)

	if s.Stroke != stroke {
		t.Error("Strikethrough.Stroke not set correctly")
	}
	if s.Offset == nil || *s.Offset != offset {
		t.Error("Strikethrough.Offset not set correctly")
	}
	if !s.Background {
		t.Error("Strikethrough.Background should be true")
	}
}

func TestStrikethroughToDecoLine(t *testing.T) {
	s := NewStrikethrough().WithBackground(true)

	decoLine := s.ToDecoLine(Black, inline.Abs(12))

	strikeDeco, ok := decoLine.(*inline.StrikethroughDeco)
	if !ok {
		t.Fatal("ToDecoLine should return *inline.StrikethroughDeco")
	}

	if !strikeDeco.Background {
		t.Error("StrikethroughDeco.Background should be true")
	}
}

func TestNewOverline(t *testing.T) {
	o := NewOverline()

	if !o.Evade {
		t.Error("Overline.Evade should default to true")
	}
	if o.Background {
		t.Error("Overline.Background should default to false")
	}
	if o.Stroke != nil {
		t.Error("Overline.Stroke should default to nil")
	}
	if o.Offset != nil {
		t.Error("Overline.Offset should default to nil")
	}
}

func TestOverlineBuilders(t *testing.T) {
	offset := inline.Abs(1.5)
	stroke := NewStroke(Green, inline.Abs(0.5))

	o := NewOverline().
		WithStroke(stroke).
		WithOffset(offset).
		WithEvade(false).
		WithBackground(true)

	if o.Stroke != stroke {
		t.Error("Overline.Stroke not set correctly")
	}
	if o.Offset == nil || *o.Offset != offset {
		t.Error("Overline.Offset not set correctly")
	}
	if o.Evade {
		t.Error("Overline.Evade should be false")
	}
	if !o.Background {
		t.Error("Overline.Background should be true")
	}
}

func TestOverlineToDecoLine(t *testing.T) {
	o := NewOverline().WithEvade(true).WithBackground(false)

	decoLine := o.ToDecoLine(Black, inline.Abs(12))

	overlineDeco, ok := decoLine.(*inline.OverlineDeco)
	if !ok {
		t.Fatal("ToDecoLine should return *inline.OverlineDeco")
	}

	if !overlineDeco.Evade {
		t.Error("OverlineDeco.Evade should be true")
	}
	if overlineDeco.Background {
		t.Error("OverlineDeco.Background should be false")
	}
}

func TestDecorationExtent(t *testing.T) {
	fontSize := inline.Abs(12)
	extent := DecorationExtent(fontSize)

	// Should be 2% of font size
	expected := inline.Abs(0.24) // 12 * 0.02 = 0.24

	if extent != expected {
		t.Errorf("DecorationExtent(%v) = %v, want %v", fontSize, extent, expected)
	}
}

func TestDecorationInterface(t *testing.T) {
	// Verify all decoration types implement Decoration interface
	var _ Decoration = (*Underline)(nil)
	var _ Decoration = (*Strikethrough)(nil)
	var _ Decoration = (*Overline)(nil)
}

func TestDecorationWithStroke(t *testing.T) {
	stroke := NewStroke(Red, inline.Abs(2))

	u := NewUnderline().WithStroke(stroke)
	decoLine := u.ToDecoLine(Black, inline.Abs(12))

	underlineDeco := decoLine.(*inline.UnderlineDeco)
	if underlineDeco.Stroke == nil {
		t.Error("UnderlineDeco.Stroke should not be nil when stroke is provided")
	}
	if underlineDeco.Stroke.Thickness != inline.Abs(2) {
		t.Errorf("Stroke thickness = %v, want 2", underlineDeco.Stroke.Thickness)
	}
}
