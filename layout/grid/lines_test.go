package grid

import (
	"testing"

	"github.com/boergens/gotypst/layout"
)

func TestStrokePriority(t *testing.T) {
	// Verify priority ordering
	if GridStroke >= CellStroke {
		t.Errorf("GridStroke should be lower than CellStroke")
	}
	if CellStroke >= ExplicitLine {
		t.Errorf("CellStroke should be lower than ExplicitLine")
	}
}

func TestNewStroke(t *testing.T) {
	stroke := NewStroke(Black, 1.0)
	if stroke.Thickness != 1.0 {
		t.Errorf("expected thickness 1.0, got %v", stroke.Thickness)
	}
	if stroke.Cap != LineCapButt {
		t.Errorf("expected LineCapButt, got %v", stroke.Cap)
	}
	if stroke.Join != LineJoinMiter {
		t.Errorf("expected LineJoinMiter, got %v", stroke.Join)
	}
	if stroke.MiterLimit != 4.0 {
		t.Errorf("expected miter limit 4.0, got %v", stroke.MiterLimit)
	}
}

func TestStrokeWithMethods(t *testing.T) {
	stroke := NewStroke(Black, 1.0)

	// Test WithCap
	roundCap := stroke.WithCap(LineCapRound)
	if roundCap.Cap != LineCapRound {
		t.Errorf("expected LineCapRound, got %v", roundCap.Cap)
	}
	if stroke.Cap != LineCapButt {
		t.Errorf("original stroke should not be modified")
	}

	// Test WithJoin
	bevelJoin := stroke.WithJoin(LineJoinBevel)
	if bevelJoin.Join != LineJoinBevel {
		t.Errorf("expected LineJoinBevel, got %v", bevelJoin.Join)
	}

	// Test WithDash
	dash := NewDash(2.0, 1.0)
	dashed := stroke.WithDash(dash)
	if dashed.Dash != dash {
		t.Errorf("expected dash pattern to be set")
	}
}

func TestCellSpanContains(t *testing.T) {
	cell := CellSpan{X: 1, Y: 1, Colspan: 2, Rowspan: 2}

	tests := []struct {
		x, y     int
		expected bool
	}{
		{0, 0, false},
		{1, 1, true},
		{2, 1, true},
		{1, 2, true},
		{2, 2, true},
		{3, 1, false},
		{1, 3, false},
	}

	for _, tt := range tests {
		if got := cell.Contains(tt.x, tt.y); got != tt.expected {
			t.Errorf("Contains(%d, %d) = %v, want %v", tt.x, tt.y, got, tt.expected)
		}
	}
}

func TestCellSpanBlocksHLine(t *testing.T) {
	// Cell spanning rows 1-2 (indices 1 and 2)
	cell := CellSpan{X: 0, Y: 1, Colspan: 1, Rowspan: 2}

	tests := []struct {
		row      int
		expected bool
	}{
		{0, false}, // Line before cell
		{1, false}, // Line at cell start
		{2, true},  // Line through middle of rowspan
		{3, false}, // Line at cell end
		{4, false}, // Line after cell
	}

	for _, tt := range tests {
		if got := cell.BlocksHLine(tt.row); got != tt.expected {
			t.Errorf("BlocksHLine(%d) = %v, want %v", tt.row, got, tt.expected)
		}
	}
}

func TestCellSpanBlocksVLine(t *testing.T) {
	// Cell spanning columns 1-2 (indices 1 and 2)
	cell := CellSpan{X: 1, Y: 0, Colspan: 2, Rowspan: 1}

	tests := []struct {
		col      int
		expected bool
	}{
		{0, false}, // Line before cell
		{1, false}, // Line at cell start
		{2, true},  // Line through middle of colspan
		{3, false}, // Line at cell end
		{4, false}, // Line after cell
	}

	for _, tt := range tests {
		if got := cell.BlocksVLine(tt.col); got != tt.expected {
			t.Errorf("BlocksVLine(%d) = %v, want %v", tt.col, got, tt.expected)
		}
	}
}

func TestLineGeneratorBasic(t *testing.T) {
	stroke := NewStroke(Black, 1.0)
	gen := &LineGenerator{
		Cols:          []layout.Abs{100, 100, 100},
		Rows:          []layout.Abs{50, 50},
		DefaultStroke: stroke,
	}

	// Generate all segments
	hSegs, vSegs := gen.GenerateAllSegments()

	// Should have 3 horizontal lines (top, middle, bottom)
	if len(hSegs) != 3 {
		t.Errorf("expected 3 horizontal segments, got %d", len(hSegs))
	}

	// Should have 4 vertical lines (left, 2 middle, right)
	if len(vSegs) != 4 {
		t.Errorf("expected 4 vertical segments, got %d", len(vSegs))
	}

	// Check horizontal segment lengths (should span full width)
	for i, seg := range hSegs {
		if seg.Length != 300 {
			t.Errorf("hSeg[%d] expected length 300, got %v", i, seg.Length)
		}
	}

	// Check vertical segment lengths (should span full height)
	for i, seg := range vSegs {
		if seg.Length != 100 {
			t.Errorf("vSeg[%d] expected length 100, got %v", i, seg.Length)
		}
	}
}

func TestLineGeneratorWithRowspan(t *testing.T) {
	stroke := NewStroke(Black, 1.0)
	gen := &LineGenerator{
		Cols:          []layout.Abs{100, 100},
		Rows:          []layout.Abs{50, 50, 50},
		DefaultStroke: stroke,
		Cells: []CellSpan{
			{X: 0, Y: 0, Colspan: 1, Rowspan: 2}, // Cell spanning rows 0-1
		},
	}

	// The horizontal line at row 1 should be blocked by the rowspan in column 0
	hSegs := gen.GenerateHLineSegments(1)

	// Should have 1 segment (only the right part, column 1)
	if len(hSegs) != 1 {
		t.Errorf("expected 1 horizontal segment, got %d", len(hSegs))
	}

	if len(hSegs) > 0 && hSegs[0].Length != 100 {
		t.Errorf("expected segment length 100, got %v", hSegs[0].Length)
	}
}

func TestLineGeneratorWithColspan(t *testing.T) {
	stroke := NewStroke(Black, 1.0)
	gen := &LineGenerator{
		Cols:          []layout.Abs{100, 100, 100},
		Rows:          []layout.Abs{50, 50},
		DefaultStroke: stroke,
		Cells: []CellSpan{
			{X: 0, Y: 0, Colspan: 2, Rowspan: 1}, // Cell spanning cols 0-1
		},
	}

	// The vertical line at col 1 should be blocked by the colspan in row 0
	vSegs := gen.GenerateVLineSegments(1)

	// Should have 1 segment (only the bottom part, row 1)
	if len(vSegs) != 1 {
		t.Errorf("expected 1 vertical segment, got %d", len(vSegs))
	}

	if len(vSegs) > 0 && vSegs[0].Length != 50 {
		t.Errorf("expected segment length 50, got %v", vSegs[0].Length)
	}
}

func TestLineGeneratorNoStroke(t *testing.T) {
	gen := &LineGenerator{
		Cols:          []layout.Abs{100, 100},
		Rows:          []layout.Abs{50, 50},
		DefaultStroke: nil, // No default stroke
	}

	hSegs, vSegs := gen.GenerateAllSegments()

	if len(hSegs) != 0 {
		t.Errorf("expected no horizontal segments with nil stroke, got %d", len(hSegs))
	}

	if len(vSegs) != 0 {
		t.Errorf("expected no vertical segments with nil stroke, got %d", len(vSegs))
	}
}

func TestLineGeneratorWithOverrides(t *testing.T) {
	defaultStroke := NewStroke(Black, 1.0)
	thickStroke := NewStroke(Black, 2.0)

	gen := &LineGenerator{
		Cols:          []layout.Abs{100, 100},
		Rows:          []layout.Abs{50, 50},
		DefaultStroke: defaultStroke,
		HLineStrokes:  []*Stroke{thickStroke, nil, thickStroke}, // Top and bottom thick
		VLineStrokes:  []*Stroke{nil, thickStroke, nil},         // Middle vertical thick
	}

	// Check that top line has explicit priority
	hSegs := gen.GenerateHLineSegments(0)
	if len(hSegs) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(hSegs))
	}
	if hSegs[0].Priority != ExplicitLine {
		t.Errorf("expected ExplicitLine priority for overridden line")
	}
	if hSegs[0].Stroke != thickStroke {
		t.Errorf("expected thick stroke for overridden line")
	}

	// Check that middle line has grid priority
	hSegs = gen.GenerateHLineSegments(1)
	if len(hSegs) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(hSegs))
	}
	if hSegs[0].Priority != GridStroke {
		t.Errorf("expected GridStroke priority for default line")
	}
}

func TestLineGeneratorRTL(t *testing.T) {
	stroke := NewStroke(Black, 1.0)
	gen := &LineGenerator{
		Cols:          []layout.Abs{100, 200},
		Rows:          []layout.Abs{50},
		DefaultStroke: stroke,
		IsRTL:         true,
	}

	// In RTL mode, vertical lines should be mirrored
	// Total width is 300
	// LTR: line 0 at x=0, line 1 at x=100, line 2 at x=300
	// RTL: line 0 at x=300, line 1 at x=200, line 2 at x=0

	vSegs := gen.GenerateVLineSegments(0)
	if len(vSegs) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(vSegs))
	}
	if vSegs[0].Offset != 300 {
		t.Errorf("RTL line 0 expected offset 300, got %v", vSegs[0].Offset)
	}

	vSegs = gen.GenerateVLineSegments(1)
	if len(vSegs) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(vSegs))
	}
	if vSegs[0].Offset != 200 {
		t.Errorf("RTL line 1 expected offset 200, got %v", vSegs[0].Offset)
	}

	vSegs = gen.GenerateVLineSegments(2)
	if len(vSegs) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(vSegs))
	}
	if vSegs[0].Offset != 0 {
		t.Errorf("RTL line 2 expected offset 0, got %v", vSegs[0].Offset)
	}
}

func TestLineSegmentBuilder(t *testing.T) {
	stroke := NewStroke(Black, 1.0)
	builder := NewLineSegmentBuilder(10) // Offset of 10

	// Extend with same stroke
	builder.Extend(50, stroke, GridStroke)
	builder.Extend(50, stroke, GridStroke)

	// Gap
	builder.Gap(20)

	// Extend with same stroke again
	builder.Extend(30, stroke, GridStroke)

	builder.Finalize()
	segments := builder.Segments()

	if len(segments) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(segments))
	}

	// First segment: 100 length
	if segments[0].Length != 100 {
		t.Errorf("segment 0 expected length 100, got %v", segments[0].Length)
	}
	if segments[0].Offset != 10 {
		t.Errorf("segment 0 expected offset 10, got %v", segments[0].Offset)
	}

	// Second segment: 30 length (after 20 gap + 100)
	if segments[1].Length != 30 {
		t.Errorf("segment 1 expected length 30, got %v", segments[1].Length)
	}
}

func TestLineSegmentBuilderStrokeChange(t *testing.T) {
	stroke1 := NewStroke(Black, 1.0)
	stroke2 := NewStroke(White, 2.0)
	builder := NewLineSegmentBuilder(0)

	builder.Extend(50, stroke1, GridStroke)
	builder.Extend(50, stroke2, GridStroke) // Different stroke
	builder.Extend(30, stroke2, GridStroke)

	builder.Finalize()
	segments := builder.Segments()

	if len(segments) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(segments))
	}

	if segments[0].Length != 50 {
		t.Errorf("segment 0 expected length 50, got %v", segments[0].Length)
	}
	if segments[0].Stroke != stroke1 {
		t.Errorf("segment 0 expected stroke1")
	}

	if segments[1].Length != 80 {
		t.Errorf("segment 1 expected length 80, got %v", segments[1].Length)
	}
	if segments[1].Stroke != stroke2 {
		t.Errorf("segment 1 expected stroke2")
	}
}

func TestLineSegmentBuilderPriorityChange(t *testing.T) {
	stroke := NewStroke(Black, 1.0)
	builder := NewLineSegmentBuilder(0)

	builder.Extend(50, stroke, GridStroke)
	builder.Extend(50, stroke, CellStroke) // Same stroke, different priority

	builder.Finalize()
	segments := builder.Segments()

	if len(segments) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(segments))
	}

	if segments[0].Priority != GridStroke {
		t.Errorf("segment 0 expected GridStroke priority")
	}
	if segments[1].Priority != CellStroke {
		t.Errorf("segment 1 expected CellStroke priority")
	}
}

func TestLineSegmentBuilderNilStroke(t *testing.T) {
	stroke := NewStroke(Black, 1.0)
	builder := NewLineSegmentBuilder(0)

	builder.Extend(50, stroke, GridStroke)
	builder.Extend(50, nil, GridStroke) // nil stroke acts as gap
	builder.Extend(30, stroke, GridStroke)

	builder.Finalize()
	segments := builder.Segments()

	if len(segments) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(segments))
	}

	if segments[0].Length != 50 {
		t.Errorf("segment 0 expected length 50, got %v", segments[0].Length)
	}
	if segments[1].Length != 30 {
		t.Errorf("segment 1 expected length 30, got %v", segments[1].Length)
	}
}

func TestColor(t *testing.T) {
	c := NewColor(255, 128, 64, 200)
	if c.R != 255 || c.G != 128 || c.B != 64 || c.A != 200 {
		t.Errorf("color components incorrect: %v", c)
	}

	opaque := NewRGB(100, 150, 200)
	if opaque.A != 255 {
		t.Errorf("RGB color should have alpha 255, got %d", opaque.A)
	}
}

func TestNewDash(t *testing.T) {
	dash := NewDash(5.0, 3.0)
	if len(dash.Array) != 2 {
		t.Errorf("expected 2 dash values, got %d", len(dash.Array))
	}
	if dash.Array[0] != 5.0 || dash.Array[1] != 3.0 {
		t.Errorf("dash values incorrect: %v", dash.Array)
	}
	if dash.Phase != 0 {
		t.Errorf("expected phase 0, got %v", dash.Phase)
	}
}

func TestLineGeneratorEmptyGrid(t *testing.T) {
	stroke := NewStroke(Black, 1.0)

	// Empty cols
	gen := &LineGenerator{
		Cols:          []layout.Abs{},
		Rows:          []layout.Abs{50},
		DefaultStroke: stroke,
	}
	hSegs, vSegs := gen.GenerateAllSegments()
	if len(hSegs) != 0 {
		t.Errorf("expected no h segments with empty cols, got %d", len(hSegs))
	}
	if len(vSegs) != 0 {
		t.Errorf("expected no v segments with empty cols, got %d", len(vSegs))
	}

	// Empty rows
	gen = &LineGenerator{
		Cols:          []layout.Abs{100},
		Rows:          []layout.Abs{},
		DefaultStroke: stroke,
	}
	hSegs, vSegs = gen.GenerateAllSegments()
	if len(hSegs) != 0 {
		t.Errorf("expected no h segments with empty rows, got %d", len(hSegs))
	}
	if len(vSegs) != 0 {
		t.Errorf("expected no v segments with empty rows, got %d", len(vSegs))
	}
}
