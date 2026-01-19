package math

import (
	"testing"
)

func TestMathContext_Push(t *testing.T) {
	ctx := &MathContext{
		Fragments: nil,
	}

	space := &SpaceFragment{Amount: 10.0}
	ctx.Push(space)

	if len(ctx.Fragments) != 1 {
		t.Errorf("expected 1 fragment, got %d", len(ctx.Fragments))
	}
}

func TestFragmentsAscent(t *testing.T) {
	frags := []MathFragment{
		&SpaceFragment{Amount: 10.0},
	}

	ascent := FragmentsAscent(frags)
	if ascent != 0 {
		t.Errorf("expected ascent 0, got %f", float64(ascent))
	}
}

func TestFragmentsDescent(t *testing.T) {
	frags := []MathFragment{
		&SpaceFragment{Amount: 10.0},
	}

	descent := FragmentsDescent(frags)
	if descent != 0 {
		t.Errorf("expected descent 0, got %f", float64(descent))
	}
}

func TestRows(t *testing.T) {
	frags := []MathFragment{
		&SpaceFragment{Amount: 5.0},
		&LinebreakFragment{},
		&SpaceFragment{Amount: 10.0},
	}

	rows := Rows(frags)
	if len(rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(rows))
	}
}

func TestIsMultiline(t *testing.T) {
	tests := []struct {
		name     string
		frags    []MathFragment
		expected bool
	}{
		{
			name:     "no linebreak",
			frags:    []MathFragment{&SpaceFragment{Amount: 10.0}},
			expected: false,
		},
		{
			name:     "with linebreak",
			frags:    []MathFragment{&SpaceFragment{Amount: 5.0}, &LinebreakFragment{}, &SpaceFragment{Amount: 10.0}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMultiline(tt.frags)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestNewSoftFrame(t *testing.T) {
	frame := NewSoftFrame(Size{X: 100, Y: 50})

	if frame.Width() != 100 {
		t.Errorf("expected width 100, got %f", float64(frame.Width()))
	}
	if frame.Height() != 50 {
		t.Errorf("expected height 50, got %f", float64(frame.Height()))
	}
}

func TestFrame_SetBaseline(t *testing.T) {
	frame := NewSoftFrame(Size{X: 100, Y: 50})
	frame.SetBaseline(40)

	if frame.Baseline() != 40 {
		t.Errorf("expected baseline 40, got %f", float64(frame.Baseline()))
	}
	if frame.Ascent() != 40 {
		t.Errorf("expected ascent 40, got %f", float64(frame.Ascent()))
	}
	if frame.Descent() != 10 {
		t.Errorf("expected descent 10, got %f", float64(frame.Descent()))
	}
}

func TestAlignments(t *testing.T) {
	rows := [][]MathFragment{
		{&SpaceFragment{Amount: 20.0}},
		{&SpaceFragment{Amount: 30.0}},
	}

	result := Alignments(rows)
	if result.Width != 30 {
		t.Errorf("expected width 30, got %f", float64(result.Width))
	}
}

func TestGetMathClass(t *testing.T) {
	tests := []struct {
		r        rune
		expected Class
	}{
		{'0', Normal},
		{'a', Alphabetic},
		{'+', Binary},
		{'=', Relation},
		{'(', Opening},
		{')', Closing},
		{',', Punctuation},
	}

	for _, tt := range tests {
		result := GetMathClass(tt.r)
		if result != tt.expected {
			t.Errorf("GetMathClass(%q) = %v, want %v", tt.r, result, tt.expected)
		}
	}
}

func TestAbs_Max(t *testing.T) {
	a := Abs(10)
	b := Abs(20)

	if a.Max(b) != 20 {
		t.Errorf("expected 20, got %f", float64(a.Max(b)))
	}
	if b.Max(a) != 20 {
		t.Errorf("expected 20, got %f", float64(b.Max(a)))
	}
}

func TestAbs_Min(t *testing.T) {
	a := Abs(10)
	b := Abs(20)

	if a.Min(b) != 10 {
		t.Errorf("expected 10, got %f", float64(a.Min(b)))
	}
	if b.Min(a) != 10 {
		t.Errorf("expected 10, got %f", float64(b.Min(a)))
	}
}

func TestEm_At(t *testing.T) {
	em := Em(1.5)
	fontSize := Abs(12.0)

	result := em.At(fontSize)
	expected := Abs(18.0)

	if result != expected {
		t.Errorf("expected %f, got %f", float64(expected), float64(result))
	}
}

func TestSize_ToPoint(t *testing.T) {
	size := Size{X: 100, Y: 50}
	point := size.ToPoint()

	if point.X != 100 || point.Y != 50 {
		t.Errorf("expected Point{100, 50}, got Point{%f, %f}", float64(point.X), float64(point.Y))
	}
}

func TestCorner_Inv(t *testing.T) {
	tests := []struct {
		corner   Corner
		expected Corner
	}{
		{CornerTopLeft, CornerBottomRight},
		{CornerTopRight, CornerBottomLeft},
		{CornerBottomRight, CornerTopLeft},
		{CornerBottomLeft, CornerTopRight},
	}

	for _, tt := range tests {
		result := tt.corner.Inv()
		if result != tt.expected {
			t.Errorf("Corner(%d).Inv() = %d, want %d", tt.corner, result, tt.expected)
		}
	}
}
