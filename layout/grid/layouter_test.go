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

// Tests for colspan/rowspan handling

func TestGrid_PopulateEntries_Colspan(t *testing.T) {
	// Create a 3x2 grid with a cell spanning 2 columns.
	cell := &Cell{X: 0, Y: 0, Colspan: 2, Rowspan: 1}
	cell2 := &Cell{X: 2, Y: 0, Colspan: 1, Rowspan: 1}
	cell3 := &Cell{X: 0, Y: 1, Colspan: 1, Rowspan: 1}

	grid := NewGrid(3, 2,
		[]Sizing{SizingAuto{}, SizingAuto{}, SizingAuto{}},
		[]Sizing{SizingAuto{}, SizingAuto{}},
		[]*Cell{cell, cell2, cell3},
	)

	// Check that (0,0) has the cell.
	if grid.CellAt(0, 0) != cell {
		t.Error("expected cell at (0,0)")
	}

	// Check that (1,0) is merged into cell.
	entry := grid.EntryAt(1, 0)
	merged, ok := entry.(EntryMerged)
	if !ok {
		t.Errorf("expected EntryMerged at (1,0), got %T", entry)
	} else if merged.Parent != cell {
		t.Error("expected merged entry to point to cell")
	}

	// Check ParentCellAt works for merged positions.
	parent := grid.ParentCellAt(1, 0)
	if parent != cell {
		t.Error("expected ParentCellAt(1,0) to return the spanning cell")
	}

	// Check that (2,0) has cell2.
	if grid.CellAt(2, 0) != cell2 {
		t.Error("expected cell2 at (2,0)")
	}
}

func TestGrid_PopulateEntries_Rowspan(t *testing.T) {
	// Create a 2x3 grid with a cell spanning 2 rows.
	cell := &Cell{X: 0, Y: 0, Colspan: 1, Rowspan: 2}
	cell2 := &Cell{X: 1, Y: 0, Colspan: 1, Rowspan: 1}

	grid := NewGrid(2, 3,
		[]Sizing{SizingAuto{}, SizingAuto{}},
		[]Sizing{SizingAuto{}, SizingAuto{}, SizingAuto{}},
		[]*Cell{cell, cell2},
	)

	// Check that (0,0) has the cell.
	if grid.CellAt(0, 0) != cell {
		t.Error("expected cell at (0,0)")
	}

	// Check that (0,1) is merged into cell.
	entry := grid.EntryAt(0, 1)
	merged, ok := entry.(EntryMerged)
	if !ok {
		t.Errorf("expected EntryMerged at (0,1), got %T", entry)
	} else if merged.Parent != cell {
		t.Error("expected merged entry to point to cell")
	}

	// Check ParentCellAt works for merged positions.
	parent := grid.ParentCellAt(0, 1)
	if parent != cell {
		t.Error("expected ParentCellAt(0,1) to return the spanning cell")
	}
}

func TestGrid_PopulateEntries_ColspanRowspan(t *testing.T) {
	// Create a 3x3 grid with a cell spanning 2x2.
	cell := &Cell{X: 0, Y: 0, Colspan: 2, Rowspan: 2}

	grid := NewGrid(3, 3,
		[]Sizing{SizingAuto{}, SizingAuto{}, SizingAuto{}},
		[]Sizing{SizingAuto{}, SizingAuto{}, SizingAuto{}},
		[]*Cell{cell},
	)

	// Check all 4 positions in the 2x2 span.
	positions := [][2]int{{0, 0}, {1, 0}, {0, 1}, {1, 1}}
	for _, pos := range positions {
		parent := grid.ParentCellAt(pos[0], pos[1])
		if parent != cell {
			t.Errorf("expected ParentCellAt(%d,%d) to return cell", pos[0], pos[1])
		}
	}

	// Check that (2,0) is nil (empty).
	if grid.EntryAt(2, 0) != nil {
		t.Error("expected nil at (2,0)")
	}
}

func TestLineGenerator_HorizontalLines_WithRowspan(t *testing.T) {
	// Create a 2x3 grid with a cell spanning rows 0-1 in column 0.
	cell := &Cell{X: 0, Y: 0, Colspan: 1, Rowspan: 2}
	grid := NewGrid(2, 3,
		[]Sizing{SizingAuto{}, SizingAuto{}},
		[]Sizing{SizingAuto{}, SizingAuto{}, SizingAuto{}},
		[]*Cell{cell},
	)
	grid.Stroke = &Stroke{Thickness: 1}

	rcols := []layout.Abs{50, 50}
	rowHeights := map[int]layout.Abs{
		0: 30,
		1: 30,
		2: 30,
	}

	lg := NewLineGenerator(grid, rcols, rowHeights, false)
	segments := lg.GenerateHorizontalLines()

	// Top line (y=0) should be full width (100).
	// Line between row 0 and 1 (y=30) should be interrupted by the rowspan.
	// Line between row 1 and 2 (y=60) should be interrupted by the rowspan.
	// Bottom line (y=90) should be full width.

	// Count segments - we should have more than 4 due to interruptions.
	// Top: 1 segment (full width)
	// y=30: 1 segment (only column 1, width 50)
	// y=60: 1 segment (full width - rowspan ends here)
	// Bottom: 1 segment (full width)

	foundY30Interrupted := false
	for _, seg := range segments {
		if seg.Offset == 30 && seg.Length == 50 {
			foundY30Interrupted = true
		}
	}

	if !foundY30Interrupted {
		t.Errorf("expected line at y=30 to be interrupted (length 50), segments: %v", segments)
	}
}

func TestLineGenerator_VerticalLines_WithColspan(t *testing.T) {
	// Create a 3x2 grid with a cell spanning columns 0-1 in row 0.
	cell := &Cell{X: 0, Y: 0, Colspan: 2, Rowspan: 1}
	grid := NewGrid(3, 2,
		[]Sizing{SizingAuto{}, SizingAuto{}, SizingAuto{}},
		[]Sizing{SizingAuto{}, SizingAuto{}},
		[]*Cell{cell},
	)
	grid.Stroke = &Stroke{Thickness: 1}

	rcols := []layout.Abs{50, 50, 50}
	rowHeights := map[int]layout.Abs{
		0: 30,
		1: 30,
	}

	lg := NewLineGenerator(grid, rcols, rowHeights, false)
	segments := lg.GenerateVerticalLines()

	// Left line (x=0) should be full height (60).
	// Line between col 0 and 1 (x=50) should be interrupted by the colspan.
	// Line between col 1 and 2 (x=100) should be full height (colspan ends before col 2).
	// Right line (x=150) should be full height.

	// Check that the line at x=50 is interrupted (has length < 60).
	foundX50Interrupted := false
	for _, seg := range segments {
		if seg.Offset == 50 && seg.Length < 60 {
			foundX50Interrupted = true
		}
	}

	if !foundX50Interrupted {
		t.Errorf("expected line at x=50 to be interrupted, segments: %v", segments)
	}
}
