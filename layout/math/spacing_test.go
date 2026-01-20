package math

import (
	"testing"

	"github.com/boergens/gotypst/layout"
)

func TestSpaceTypeAmount(t *testing.T) {
	tests := []struct {
		space SpaceType
		want  layout.Em
	}{
		{SpaceNone, 0},
		{SpaceThin, layout.Em(3.0 / 18.0)},
		{SpaceMedium, layout.Em(4.0 / 18.0)},
		{SpaceThick, layout.Em(5.0 / 18.0)},
	}

	for _, tt := range tests {
		if got := tt.space.Amount(); got != tt.want {
			t.Errorf("SpaceType(%d).Amount() = %v, want %v", tt.space, got, tt.want)
		}
	}
}

func TestGetSpacing(t *testing.T) {
	tests := []struct {
		left, right MathClass
		style       MathStyle
		want        SpaceType
	}{
		// Display style tests (full spacing)
		{ClassOrd, ClassOp, StyleDisplay, SpaceThin},
		{ClassOrd, ClassBin, StyleDisplay, SpaceMedium},
		{ClassOrd, ClassRel, StyleDisplay, SpaceThick},
		{ClassOpen, ClassOrd, StyleDisplay, SpaceNone},
		{ClassOrd, ClassClose, StyleDisplay, SpaceNone},

		// Binary operator spacing
		{ClassOrd, ClassBin, StyleDisplay, SpaceMedium},
		{ClassBin, ClassOrd, StyleDisplay, SpaceMedium},

		// Relation spacing
		{ClassOrd, ClassRel, StyleDisplay, SpaceThick},
		{ClassRel, ClassOrd, StyleDisplay, SpaceThick},

		// Script style (reduced spacing)
		{ClassOrd, ClassBin, StyleScript, SpaceThin},   // medium -> thin
		{ClassOrd, ClassRel, StyleScript, SpaceNone},   // thick -> none
		{ClassOrd, ClassOp, StyleScript, SpaceThin},    // thin stays thin

		// ScriptScript style (also reduced)
		{ClassOrd, ClassBin, StyleScriptScript, SpaceThin},
		{ClassOrd, ClassRel, StyleScriptScript, SpaceNone},
	}

	for _, tt := range tests {
		got := GetSpacing(tt.left, tt.right, tt.style)
		if got != tt.want {
			t.Errorf("GetSpacing(%s, %s, %v) = %v, want %v",
				tt.left, tt.right, tt.style, got, tt.want)
		}
	}
}

func TestGetSpacingAbs(t *testing.T) {
	fontSize := layout.Abs(12)

	// Thick space at 12pt = 5/18 * 12 = 10/3 ≈ 3.33pt
	got := GetSpacingAbs(ClassOrd, ClassRel, StyleDisplay, fontSize)
	want := SpaceThick.Amount().At(fontSize)

	if got != want {
		t.Errorf("GetSpacingAbs(Ord, Rel, Display, 12) = %v, want %v", got, want)
	}
}

func TestNewSpaces(t *testing.T) {
	fontSize := layout.Abs(18) // 1em = 18pt makes calculations easy

	thin := NewThinSpace(fontSize)
	if thin.Amount != 3 { // 3/18 * 18 = 3
		t.Errorf("NewThinSpace(18).Amount = %v, want 3", thin.Amount)
	}

	medium := NewMediumSpace(fontSize)
	if medium.Amount != 4 { // 4/18 * 18 = 4
		t.Errorf("NewMediumSpace(18).Amount = %v, want 4", medium.Amount)
	}

	thick := NewThickSpace(fontSize)
	if thick.Amount != 5 { // 5/18 * 18 = 5
		t.Errorf("NewThickSpace(18).Amount = %v, want 5", thick.Amount)
	}

	custom := NewSpace(7.5)
	if custom.Amount != 7.5 {
		t.Errorf("NewSpace(7.5).Amount = %v, want 7.5", custom.Amount)
	}
}

func TestInsertSpacing(t *testing.T) {
	fontSize := layout.Abs(18)

	// Create fragments: x + y (Ord Bin Ord)
	fragments := []MathFragment{
		&GlyphFragment{MathClass: ClassOrd, Glyphs: []MathGlyph{{Advance: 8}}},
		&GlyphFragment{MathClass: ClassBin, Glyphs: []MathGlyph{{Advance: 6}}},
		&GlyphFragment{MathClass: ClassOrd, Glyphs: []MathGlyph{{Advance: 8}}},
	}

	result := InsertSpacing(fragments, StyleDisplay, fontSize)

	// Should have: Ord, space, Bin, space, Ord = 5 items
	if len(result) != 5 {
		t.Errorf("InsertSpacing: len = %d, want 5", len(result))
	}

	// Check spacing was inserted correctly
	if _, ok := result[1].(*SpaceFragment); !ok {
		t.Errorf("result[1] should be SpaceFragment")
	}
	if _, ok := result[3].(*SpaceFragment); !ok {
		t.Errorf("result[3] should be SpaceFragment")
	}
}

func TestInsertSpacingEmpty(t *testing.T) {
	result := InsertSpacing(nil, StyleDisplay, 12)
	if len(result) != 0 {
		t.Errorf("InsertSpacing(nil) should return empty slice")
	}

	result = InsertSpacing([]MathFragment{}, StyleDisplay, 12)
	if len(result) != 0 {
		t.Errorf("InsertSpacing([]) should return empty slice")
	}
}

func TestInsertSpacingSingle(t *testing.T) {
	fragments := []MathFragment{
		&GlyphFragment{MathClass: ClassOrd, Glyphs: []MathGlyph{{Advance: 8}}},
	}

	result := InsertSpacing(fragments, StyleDisplay, 12)

	if len(result) != 1 {
		t.Errorf("InsertSpacing with single fragment should return 1 item, got %d", len(result))
	}
}

func TestInsertSpacingNoSpace(t *testing.T) {
	// Open followed by Ord has no spacing
	fragments := []MathFragment{
		&GlyphFragment{MathClass: ClassOpen, Glyphs: []MathGlyph{{Advance: 4}}},
		&GlyphFragment{MathClass: ClassOrd, Glyphs: []MathGlyph{{Advance: 8}}},
	}

	result := InsertSpacing(fragments, StyleDisplay, 18)

	// Should have: Open, Ord = 2 items (no space between)
	if len(result) != 2 {
		t.Errorf("InsertSpacing: len = %d, want 2", len(result))
	}
}

func TestMathStyleScaledSize(t *testing.T) {
	tests := []struct {
		style MathStyle
		want  float64
	}{
		{StyleDisplay, 1.0},
		{StyleText, 1.0},
		{StyleScript, 0.7},
		{StyleScriptScript, 0.5},
	}

	for _, tt := range tests {
		if got := tt.style.ScaledSize(); got != tt.want {
			t.Errorf("MathStyle(%d).ScaledSize() = %v, want %v", tt.style, got, tt.want)
		}
	}
}

func TestGetSpacingInvalidClass(t *testing.T) {
	// Test with out-of-range class values
	got := GetSpacing(MathClass(-1), ClassOrd, StyleDisplay)
	if got != SpaceNone {
		t.Errorf("GetSpacing(-1, Ord) = %v, want SpaceNone", got)
	}

	got = GetSpacing(ClassOrd, MathClass(100), StyleDisplay)
	if got != SpaceNone {
		t.Errorf("GetSpacing(Ord, 100) = %v, want SpaceNone", got)
	}
}
