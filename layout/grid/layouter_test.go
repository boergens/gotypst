package grid

import (
	"testing"

	"github.com/boergens/gotypst/layout"
	"github.com/boergens/gotypst/layout/flow"
)

func TestNewGridLayouter(t *testing.T) {
	grid := &Grid{
		Cols:     []Sizing{SizingAuto{}, SizingAuto{}},
		Rows:     []Sizing{SizingAuto{}, SizingAuto{}},
		Entries:  make([]Entry, 4),
		ColCount: 2,
		RowCount: 2,
	}

	regions := &flow.Regions{
		Size:   layout.Size{Width: 200, Height: 300},
		Full:   layout.Size{Width: 200, Height: 300},
		Expand: flow.Axes[bool]{X: false, Y: false},
	}

	gl := NewGridLayouter(nil, grid, regions, nil, false)

	if gl == nil {
		t.Fatal("expected non-nil GridLayouter")
	}

	if gl.Grid != grid {
		t.Error("expected GridLayouter to reference the input grid")
	}

	if len(gl.RCols) != 2 {
		t.Errorf("expected 2 column widths, got %d", len(gl.RCols))
	}

	if gl.Current.RegionIdx != 0 {
		t.Errorf("expected initial region index 0, got %d", gl.Current.RegionIdx)
	}
}

func TestMeasureColumns_Fixed(t *testing.T) {
	grid := &Grid{
		Cols: []Sizing{
			SizingRel{Abs: 50},
			SizingRel{Abs: 100},
		},
		Rows:     []Sizing{SizingAuto{}},
		Entries:  make([]Entry, 2),
		ColCount: 2,
		RowCount: 1,
	}

	regions := &flow.Regions{
		Size:   layout.Size{Width: 200, Height: 300},
		Full:   layout.Size{Width: 200, Height: 300},
		Expand: flow.Axes[bool]{X: false, Y: false},
	}

	gl := NewGridLayouter(nil, grid, regions, nil, false)
	err := gl.measureColumns()
	if err != nil {
		t.Fatalf("measureColumns failed: %v", err)
	}

	if gl.RCols[0] != 50 {
		t.Errorf("expected column 0 width 50, got %v", gl.RCols[0])
	}

	if gl.RCols[1] != 100 {
		t.Errorf("expected column 1 width 100, got %v", gl.RCols[1])
	}

	if gl.Width != 150 {
		t.Errorf("expected total width 150, got %v", gl.Width)
	}
}

func TestMeasureColumns_Relative(t *testing.T) {
	grid := &Grid{
		Cols: []Sizing{
			SizingRel{Ratio: 0.25}, // 25% of 200 = 50
			SizingRel{Ratio: 0.50}, // 50% of 200 = 100
		},
		Rows:     []Sizing{SizingAuto{}},
		Entries:  make([]Entry, 2),
		ColCount: 2,
		RowCount: 1,
	}

	regions := &flow.Regions{
		Size:   layout.Size{Width: 200, Height: 300},
		Full:   layout.Size{Width: 200, Height: 300},
		Expand: flow.Axes[bool]{X: false, Y: false},
	}

	gl := NewGridLayouter(nil, grid, regions, nil, false)
	err := gl.measureColumns()
	if err != nil {
		t.Fatalf("measureColumns failed: %v", err)
	}

	if gl.RCols[0] != 50 {
		t.Errorf("expected column 0 width 50, got %v", gl.RCols[0])
	}

	if gl.RCols[1] != 100 {
		t.Errorf("expected column 1 width 100, got %v", gl.RCols[1])
	}
}

func TestMeasureColumns_Fractional(t *testing.T) {
	grid := &Grid{
		Cols: []Sizing{
			SizingRel{Abs: 50},   // Fixed 50
			SizingFr{Fr: 1},      // 1fr
			SizingFr{Fr: 2},      // 2fr
		},
		Rows:     []Sizing{SizingAuto{}},
		Entries:  make([]Entry, 3),
		ColCount: 3,
		RowCount: 1,
	}

	regions := &flow.Regions{
		Size:   layout.Size{Width: 200, Height: 300},
		Full:   layout.Size{Width: 200, Height: 300},
		Expand: flow.Axes[bool]{X: false, Y: false},
	}

	gl := NewGridLayouter(nil, grid, regions, nil, false)
	err := gl.measureColumns()
	if err != nil {
		t.Fatalf("measureColumns failed: %v", err)
	}

	if gl.RCols[0] != 50 {
		t.Errorf("expected column 0 width 50, got %v", gl.RCols[0])
	}

	// Remaining 150 is split: 1fr = 50, 2fr = 100
	if gl.RCols[1] != 50 {
		t.Errorf("expected column 1 width 50, got %v", gl.RCols[1])
	}

	if gl.RCols[2] != 100 {
		t.Errorf("expected column 2 width 100, got %v", gl.RCols[2])
	}
}

func TestRowspanTracker(t *testing.T) {
	tracker := NewRowspanTracker()

	rs := Rowspan{
		X:            0,
		Y:            0,
		RowspanCount: 3,
		DX:           0,
		DY:           0,
		FirstRegion:  0,
		Heights:      []layout.Abs{100},
	}

	tracker.Register(rs)

	if len(tracker.Rowspans) != 1 {
		t.Errorf("expected 1 rowspan, got %d", len(tracker.Rowspans))
	}

	// Test ActiveAt
	active := tracker.ActiveAt(0)
	if len(active) != 1 {
		t.Errorf("expected 1 active rowspan at row 0, got %d", len(active))
	}

	active = tracker.ActiveAt(2)
	if len(active) != 1 {
		t.Errorf("expected 1 active rowspan at row 2, got %d", len(active))
	}

	active = tracker.ActiveAt(3)
	if len(active) != 0 {
		t.Errorf("expected 0 active rowspans at row 3, got %d", len(active))
	}
}

func TestRowspan_TotalHeight(t *testing.T) {
	rs := Rowspan{
		Heights: []layout.Abs{50, 75, 25},
	}

	total := rs.TotalHeight()
	if total != 150 {
		t.Errorf("expected total height 150, got %v", total)
	}
}

func TestLineGenerator_HorizontalLines(t *testing.T) {
	stroke := &Stroke{Thickness: 1}
	grid := &Grid{
		ColCount: 2,
		RowCount: 2,
		Stroke:   stroke,
	}

	rcols := []layout.Abs{50, 100}
	rowHeights := map[int]layout.Abs{
		0: 30,
		1: 40,
	}

	lg := NewLineGenerator(grid, rcols, rowHeights, false)
	segments := lg.GenerateHorizontalLines()

	// Should have 3 horizontal lines: top, between rows, bottom
	if len(segments) != 3 {
		t.Errorf("expected 3 horizontal line segments, got %d", len(segments))
	}

	// Check positions
	if segments[0].Offset != 0 {
		t.Errorf("expected first line at y=0, got %v", segments[0].Offset)
	}

	if segments[1].Offset != 30 {
		t.Errorf("expected second line at y=30, got %v", segments[1].Offset)
	}

	if segments[2].Offset != 70 {
		t.Errorf("expected third line at y=70, got %v", segments[2].Offset)
	}
}

func TestHeaderManager(t *testing.T) {
	hm := NewHeaderManager()

	header := Header{
		StartY: 0,
		EndY:   2,
		Level:  1,
	}

	hm.AddPendingHeader(header)

	if !hm.HasPendingHeaders() {
		t.Error("expected pending headers")
	}

	if hm.HasRepeatingHeaders() {
		t.Error("expected no repeating headers yet")
	}

	// Promote headers
	hm.PromotePendingHeaders()

	if hm.HasPendingHeaders() {
		t.Error("expected no pending headers after promotion")
	}

	if !hm.HasRepeatingHeaders() {
		t.Error("expected repeating headers after promotion")
	}
}

func TestOrphanDetector(t *testing.T) {
	od := NewOrphanDetector()

	od.AddHeaderHeight(50)

	if !od.IsOrphaned() {
		t.Error("expected orphaned when only header height")
	}

	od.AddContentHeight(10)

	if od.IsOrphaned() {
		t.Error("expected not orphaned when content exists")
	}

	od.Reset()

	if od.HeaderHeight != 0 || od.ContentHeight != 0 {
		t.Error("expected reset to zero")
	}
}

func TestGrid_EntryAt(t *testing.T) {
	cell := &Cell{X: 0, Y: 0, Colspan: 1, Rowspan: 1}
	grid := &Grid{
		Entries: []Entry{
			EntryCell{Cell: cell},
			nil,
			nil,
			nil,
		},
		ColCount: 2,
		RowCount: 2,
	}

	// Valid position
	entry := grid.EntryAt(0, 0)
	if entry == nil {
		t.Error("expected non-nil entry at (0,0)")
	}

	// Out of bounds
	entry = grid.EntryAt(-1, 0)
	if entry != nil {
		t.Error("expected nil for negative x")
	}

	entry = grid.EntryAt(0, 5)
	if entry != nil {
		t.Error("expected nil for out of bounds y")
	}
}

func TestAlignContentInCell_Horizontal(t *testing.T) {
	tests := []struct {
		name      string
		cellW     layout.Abs
		contentW  layout.Abs
		alignX    flow.FixedAlignment
		expectedX layout.Abs
	}{
		{"start alignment", 100, 60, flow.FixedAlignStart, 0},
		{"center alignment", 100, 60, flow.FixedAlignCenter, 20},
		{"end alignment", 100, 60, flow.FixedAlignEnd, 40},
		{"content equals cell", 100, 100, flow.FixedAlignCenter, 0},
		{"content larger than cell", 100, 120, flow.FixedAlignEnd, 0}, // No negative offset
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			align := flow.Axes[flow.FixedAlignment]{X: tt.alignX, Y: flow.FixedAlignStart}
			pos := AlignContentInCell(tt.cellW, 100, tt.contentW, 50, align)
			if pos.X != tt.expectedX {
				t.Errorf("expected X offset %v, got %v", tt.expectedX, pos.X)
			}
		})
	}
}

func TestAlignContentInCell_Vertical(t *testing.T) {
	tests := []struct {
		name      string
		cellH     layout.Abs
		contentH  layout.Abs
		alignY    flow.FixedAlignment
		expectedY layout.Abs
	}{
		{"top alignment", 100, 40, flow.FixedAlignStart, 0},
		{"middle alignment", 100, 40, flow.FixedAlignCenter, 30},
		{"bottom alignment", 100, 40, flow.FixedAlignEnd, 60},
		{"content equals cell", 100, 100, flow.FixedAlignCenter, 0},
		{"content larger than cell", 100, 120, flow.FixedAlignEnd, 0}, // No negative offset
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			align := flow.Axes[flow.FixedAlignment]{X: flow.FixedAlignStart, Y: tt.alignY}
			pos := AlignContentInCell(100, tt.cellH, 50, tt.contentH, align)
			if pos.Y != tt.expectedY {
				t.Errorf("expected Y offset %v, got %v", tt.expectedY, pos.Y)
			}
		})
	}
}

func TestAlignContentInCell_Combined(t *testing.T) {
	// Test bottom-right alignment
	align := flow.Axes[flow.FixedAlignment]{
		X: flow.FixedAlignEnd,
		Y: flow.FixedAlignEnd,
	}
	pos := AlignContentInCell(100, 80, 40, 30, align)

	if pos.X != 60 {
		t.Errorf("expected X offset 60 for right alignment, got %v", pos.X)
	}
	if pos.Y != 50 {
		t.Errorf("expected Y offset 50 for bottom alignment, got %v", pos.Y)
	}

	// Test center-center alignment
	align = flow.Axes[flow.FixedAlignment]{
		X: flow.FixedAlignCenter,
		Y: flow.FixedAlignCenter,
	}
	pos = AlignContentInCell(100, 80, 40, 30, align)

	if pos.X != 30 {
		t.Errorf("expected X offset 30 for center alignment, got %v", pos.X)
	}
	if pos.Y != 25 {
		t.Errorf("expected Y offset 25 for middle alignment, got %v", pos.Y)
	}
}

func TestCellAlignment_StoredInLayout(t *testing.T) {
	// Create a cell with specific alignment
	cell := &Cell{
		X:       0,
		Y:       0,
		Colspan: 1,
		Rowspan: 1,
		Align: flow.Axes[flow.FixedAlignment]{
			X: flow.FixedAlignCenter,
			Y: flow.FixedAlignEnd,
		},
	}

	grid := &Grid{
		Cols:     []Sizing{SizingRel{Abs: 100}},
		Rows:     []Sizing{SizingRel{Abs: 50}},
		Entries:  []Entry{EntryCell{Cell: cell}},
		ColCount: 1,
		RowCount: 1,
	}

	regions := &flow.Regions{
		Size:   layout.Size{Width: 200, Height: 300},
		Full:   layout.Size{Width: 200, Height: 300},
		Expand: flow.Axes[bool]{X: false, Y: false},
	}

	gl := NewGridLayouter(nil, grid, regions, nil, false)

	// Layout the cell
	err := gl.layoutCell(cell, 0, 0, 50)
	if err != nil {
		t.Fatalf("layoutCell failed: %v", err)
	}

	// Check that the alignment was stored
	locator := gl.CellLocators[Axes[int]{X: 0, Y: 0}]
	if locator == nil {
		t.Fatal("expected cell locator to be stored")
	}

	cellLayout, ok := locator.(*CellLayout)
	if !ok {
		t.Fatal("expected CellLayout type")
	}

	if cellLayout.Align.X != flow.FixedAlignCenter {
		t.Errorf("expected X alignment to be Center, got %v", cellLayout.Align.X)
	}
	if cellLayout.Align.Y != flow.FixedAlignEnd {
		t.Errorf("expected Y alignment to be End, got %v", cellLayout.Align.Y)
	}
}

func TestMeasureColumns_Auto(t *testing.T) {
	// Create cells with string content for auto-sizing
	cell1 := &Cell{X: 0, Y: 0, Colspan: 1, Rowspan: 1, Body: "Hi"}         // Short text
	cell2 := &Cell{X: 1, Y: 0, Colspan: 1, Rowspan: 1, Body: "Hello World"} // Longer text

	grid := &Grid{
		Cols: []Sizing{
			SizingAuto{}, // Auto column 1
			SizingAuto{}, // Auto column 2
		},
		Rows: []Sizing{SizingAuto{}},
		Entries: []Entry{
			EntryCell{Cell: cell1},
			EntryCell{Cell: cell2},
		},
		ColCount: 2,
		RowCount: 1,
	}

	regions := &flow.Regions{
		Size:   layout.Size{Width: 500, Height: 300},
		Full:   layout.Size{Width: 500, Height: 300},
		Expand: flow.Axes[bool]{X: false, Y: false},
	}

	gl := NewGridLayouter(nil, grid, regions, nil, false)
	err := gl.measureColumns()
	if err != nil {
		t.Fatalf("measureColumns failed: %v", err)
	}

	// Column 0 should be narrower than column 1 (shorter text)
	if gl.RCols[0] >= gl.RCols[1] {
		t.Errorf("expected column 0 (%v) to be narrower than column 1 (%v)", gl.RCols[0], gl.RCols[1])
	}

	// Both columns should have positive width
	if gl.RCols[0] <= 0 {
		t.Errorf("expected column 0 to have positive width, got %v", gl.RCols[0])
	}
	if gl.RCols[1] <= 0 {
		t.Errorf("expected column 1 to have positive width, got %v", gl.RCols[1])
	}
}

func TestMeasureColumns_AutoWithNilBody(t *testing.T) {
	// Create cell with nil body
	cell := &Cell{X: 0, Y: 0, Colspan: 1, Rowspan: 1, Body: nil}

	grid := &Grid{
		Cols:     []Sizing{SizingAuto{}},
		Rows:     []Sizing{SizingAuto{}},
		Entries:  []Entry{EntryCell{Cell: cell}},
		ColCount: 1,
		RowCount: 1,
	}

	regions := &flow.Regions{
		Size:   layout.Size{Width: 200, Height: 300},
		Full:   layout.Size{Width: 200, Height: 300},
		Expand: flow.Axes[bool]{X: false, Y: false},
	}

	gl := NewGridLayouter(nil, grid, regions, nil, false)
	err := gl.measureColumns()
	if err != nil {
		t.Fatalf("measureColumns failed: %v", err)
	}

	// Column with nil body should have 0 width
	if gl.RCols[0] != 0 {
		t.Errorf("expected column 0 width 0 for nil body, got %v", gl.RCols[0])
	}
}

func TestMeasureColumns_AutoWithMeasurable(t *testing.T) {
	// Create a cell with a measurable body
	measurableContent := &TextContent{
		Text:            "Test content",
		FontSize:        12,
		ApproxCharWidth: 6,
	}
	cell := &Cell{X: 0, Y: 0, Colspan: 1, Rowspan: 1, Body: measurableContent}

	grid := &Grid{
		Cols:     []Sizing{SizingAuto{}},
		Rows:     []Sizing{SizingAuto{}},
		Entries:  []Entry{EntryCell{Cell: cell}},
		ColCount: 1,
		RowCount: 1,
	}

	regions := &flow.Regions{
		Size:   layout.Size{Width: 200, Height: 300},
		Full:   layout.Size{Width: 200, Height: 300},
		Expand: flow.Axes[bool]{X: false, Y: false},
	}

	gl := NewGridLayouter(nil, grid, regions, nil, false)
	err := gl.measureColumns()
	if err != nil {
		t.Fatalf("measureColumns failed: %v", err)
	}

	// Width should match the measurable's reported width
	expectedWidth := measurableContent.MeasureWidth()
	if gl.RCols[0] != expectedWidth {
		t.Errorf("expected column 0 width %v, got %v", expectedWidth, gl.RCols[0])
	}
}

func TestMeasureColumns_MixedSizing(t *testing.T) {
	// Mix of auto, fixed, and fractional columns
	cell := &Cell{X: 1, Y: 0, Colspan: 1, Rowspan: 1, Body: "AutoContent"}

	grid := &Grid{
		Cols: []Sizing{
			SizingRel{Abs: 30}, // Fixed 30
			SizingAuto{},       // Auto
			SizingFr{Fr: 1},    // 1fr
		},
		Rows:     []Sizing{SizingAuto{}},
		Entries:  []Entry{nil, EntryCell{Cell: cell}, nil},
		ColCount: 3,
		RowCount: 1,
	}

	regions := &flow.Regions{
		Size:   layout.Size{Width: 200, Height: 300},
		Full:   layout.Size{Width: 200, Height: 300},
		Expand: flow.Axes[bool]{X: false, Y: false},
	}

	gl := NewGridLayouter(nil, grid, regions, nil, false)
	err := gl.measureColumns()
	if err != nil {
		t.Fatalf("measureColumns failed: %v", err)
	}

	// Column 0 should be exactly 30 (fixed)
	if gl.RCols[0] != 30 {
		t.Errorf("expected column 0 width 30, got %v", gl.RCols[0])
	}

	// Column 1 should have auto width based on content
	if gl.RCols[1] <= 0 {
		t.Errorf("expected column 1 to have positive width, got %v", gl.RCols[1])
	}

	// Column 2 should get the remaining space (1fr)
	expectedRemaining := 200 - gl.RCols[0] - gl.RCols[1]
	if !gl.RCols[2].ApproxEq(expectedRemaining) {
		t.Errorf("expected column 2 width ~%v, got %v", expectedRemaining, gl.RCols[2])
	}
}

func TestTextContent_MeasureWidth(t *testing.T) {
	tc := &TextContent{
		Text:            "Hello",
		FontSize:        12,
		ApproxCharWidth: 6,
	}

	width := tc.MeasureWidth()
	expected := layout.Abs(5 * 6) // 5 chars * 6pt per char
	if width != expected {
		t.Errorf("expected width %v, got %v", expected, width)
	}
}

func TestTextContent_MeasureHeight(t *testing.T) {
	tc := &TextContent{
		Text:            "Hello World Test",
		FontSize:        12,
		ApproxCharWidth: 6,
	}

	// With wide width, should be single line
	height := tc.MeasureHeight(200)
	expectedSingleLine := layout.Abs(12 * 1.2) // 12pt * 1.2 line height
	if !height.ApproxEq(expectedSingleLine) {
		t.Errorf("expected height %v for single line, got %v", expectedSingleLine, height)
	}

	// With narrow width, should wrap to multiple lines
	height = tc.MeasureHeight(36) // 36pt = 6 chars per line
	// "Hello World Test" = 16 chars, 6 chars per line = 3 lines
	expectedMultiLine := layout.Abs(3 * 12 * 1.2)
	if !height.ApproxEq(expectedMultiLine) {
		t.Errorf("expected height %v for multiple lines, got %v", expectedMultiLine, height)
	}
}

func TestFrameContent_Measure(t *testing.T) {
	fc := &FrameContent{
		Width:  100,
		Height: 50,
	}

	if fc.MeasureWidth() != 100 {
		t.Errorf("expected width 100, got %v", fc.MeasureWidth())
	}

	if fc.MeasureHeight(200) != 50 {
		t.Errorf("expected height 50, got %v", fc.MeasureHeight(200))
	}
}

func TestMeasuredCell(t *testing.T) {
	mc := &MeasuredCell{
		Body:   "test",
		Width:  80,
		Height: 20,
	}

	if mc.MeasureWidth() != 80 {
		t.Errorf("expected width 80, got %v", mc.MeasureWidth())
	}

	if mc.MeasureHeight(100) != 20 {
		t.Errorf("expected height 20, got %v", mc.MeasureHeight(100))
	}
}
