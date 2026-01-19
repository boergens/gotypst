package grid

import (
	"testing"

	"github.com/boergens/gotypst/layout"
)

func TestNewGrid(t *testing.T) {
	columns := []TrackSize{
		FixedTrack{Size: 100 * layout.Pt},
		AutoTrack{},
		FrTrack{Fr: 1},
	}

	g := NewGrid(columns)

	if g.ColCount() != 3 {
		t.Errorf("expected 3 columns, got %d", g.ColCount())
	}

	if g.RowCount != 0 {
		t.Errorf("expected 0 rows initially, got %d", g.RowCount)
	}
}

func TestNewCell(t *testing.T) {
	cell := NewCell(2, 3, "test content")

	if cell.X != 2 {
		t.Errorf("expected X=2, got %d", cell.X)
	}
	if cell.Y != 3 {
		t.Errorf("expected Y=3, got %d", cell.Y)
	}
	if cell.Colspan != 1 {
		t.Errorf("expected Colspan=1, got %d", cell.Colspan)
	}
	if cell.Rowspan != 1 {
		t.Errorf("expected Rowspan=1, got %d", cell.Rowspan)
	}
}

func TestCellSpanning(t *testing.T) {
	cell := NewCell(1, 1, nil)
	cell.Colspan = 2
	cell.Rowspan = 3

	if cell.EndX() != 3 {
		t.Errorf("expected EndX=3, got %d", cell.EndX())
	}
	if cell.EndY() != 4 {
		t.Errorf("expected EndY=4, got %d", cell.EndY())
	}

	// Test Contains
	testCases := []struct {
		x, y     int
		expected bool
	}{
		{0, 0, false}, // Before cell
		{1, 1, true},  // Cell origin
		{2, 2, true},  // Inside span
		{2, 3, true},  // Last row of span
		{3, 1, false}, // After colspan
		{1, 4, false}, // After rowspan
	}

	for _, tc := range testCases {
		result := cell.Contains(tc.x, tc.y)
		if result != tc.expected {
			t.Errorf("Contains(%d, %d) = %v, expected %v", tc.x, tc.y, result, tc.expected)
		}
	}
}

func TestGridAddCell(t *testing.T) {
	g := NewGrid([]TrackSize{AutoTrack{}, AutoTrack{}, AutoTrack{}})

	cell1 := NewCell(0, 0, "A")
	cell2 := NewCell(1, 0, "B")
	cell3 := NewCell(0, 1, "C")

	g.AddCell(cell1)
	g.AddCell(cell2)
	g.AddCell(cell3)

	if len(g.Cells) != 3 {
		t.Errorf("expected 3 cells, got %d", len(g.Cells))
	}

	if g.RowCount != 2 {
		t.Errorf("expected RowCount=2, got %d", g.RowCount)
	}
}

func TestGridCellAt(t *testing.T) {
	g := NewGrid([]TrackSize{AutoTrack{}, AutoTrack{}, AutoTrack{}})

	cell1 := NewCell(0, 0, "A")
	cell2 := NewCell(1, 0, "B")
	cell2.Colspan = 2 // Spans columns 1 and 2

	g.AddCell(cell1)
	g.AddCell(cell2)

	// Test cell lookup
	if c := g.CellAt(0, 0); c != cell1 {
		t.Error("CellAt(0, 0) should return cell1")
	}
	if c := g.CellAt(1, 0); c != cell2 {
		t.Error("CellAt(1, 0) should return cell2")
	}
	if c := g.CellAt(2, 0); c != cell2 {
		t.Error("CellAt(2, 0) should return cell2 (due to colspan)")
	}
	if c := g.CellAt(0, 1); c != nil {
		t.Error("CellAt(0, 1) should return nil")
	}
}

func TestGridTrackSizes(t *testing.T) {
	g := NewGrid([]TrackSize{
		FixedTrack{Size: 50 * layout.Pt},
		RelativeTrack{Ratio: 0.25},
	})
	g.Tracks.Rows = []TrackSize{
		AutoTrack{},
		FrTrack{Fr: 2},
	}

	// Test column access
	if _, ok := g.ColumnAt(0).(FixedTrack); !ok {
		t.Error("Column 0 should be FixedTrack")
	}
	if _, ok := g.ColumnAt(1).(RelativeTrack); !ok {
		t.Error("Column 1 should be RelativeTrack")
	}
	if _, ok := g.ColumnAt(99).(AutoTrack); !ok {
		t.Error("Out of range column should be AutoTrack")
	}

	// Test row access
	if !g.IsAutoRow(0) {
		t.Error("Row 0 should be auto")
	}
	if !g.IsFrRow(1) {
		t.Error("Row 1 should be fractional")
	}
}

func TestGridGutter(t *testing.T) {
	g := NewGrid([]TrackSize{AutoTrack{}, AutoTrack{}})

	if g.HasGutter() {
		t.Error("Grid should not have gutter by default")
	}

	g.Gutter = Gutter{
		Column: 10 * layout.Pt,
		Row:    5 * layout.Pt,
	}

	if !g.HasGutter() {
		t.Error("Grid should have gutter after setting")
	}
}

func TestCellsInRow(t *testing.T) {
	g := NewGrid([]TrackSize{AutoTrack{}, AutoTrack{}, AutoTrack{}})

	g.AddCell(NewCell(0, 0, "A"))
	g.AddCell(NewCell(1, 0, "B"))
	g.AddCell(NewCell(0, 1, "C"))
	g.AddCell(NewCell(2, 1, "D"))

	row0 := g.CellsInRow(0)
	if len(row0) != 2 {
		t.Errorf("expected 2 cells in row 0, got %d", len(row0))
	}

	row1 := g.CellsInRow(1)
	if len(row1) != 2 {
		t.Errorf("expected 2 cells in row 1, got %d", len(row1))
	}

	row2 := g.CellsInRow(2)
	if len(row2) != 0 {
		t.Errorf("expected 0 cells in row 2, got %d", len(row2))
	}
}

func TestCellsInColumn(t *testing.T) {
	g := NewGrid([]TrackSize{AutoTrack{}, AutoTrack{}, AutoTrack{}})

	g.AddCell(NewCell(0, 0, "A"))
	g.AddCell(NewCell(0, 1, "B"))
	g.AddCell(NewCell(1, 0, "C"))

	col0 := g.CellsInColumn(0)
	if len(col0) != 2 {
		t.Errorf("expected 2 cells in column 0, got %d", len(col0))
	}

	col1 := g.CellsInColumn(1)
	if len(col1) != 1 {
		t.Errorf("expected 1 cell in column 1, got %d", len(col1))
	}

	col2 := g.CellsInColumn(2)
	if len(col2) != 0 {
		t.Errorf("expected 0 cells in column 2, got %d", len(col2))
	}
}

func TestLayouterColumnMeasurement(t *testing.T) {
	g := NewGrid([]TrackSize{
		FixedTrack{Size: 100 * layout.Pt},
		RelativeTrack{Ratio: 0.2},
		FrTrack{Fr: 1},
	})

	regions := &layout.Regions{
		Current: layout.Region{
			Size: layout.Size{Width: 500 * layout.Pt, Height: 800 * layout.Pt},
			Full: true,
		},
	}

	layouter := NewLayouter(g, regions)
	if err := layouter.measureColumns(); err != nil {
		t.Fatalf("measureColumns failed: %v", err)
	}

	// Column 0: Fixed 100pt
	if layouter.ResolvedCols[0] != 100*layout.Pt {
		t.Errorf("Column 0: expected 100pt, got %v", layouter.ResolvedCols[0])
	}

	// Column 1: 20% of 500pt = 100pt
	if layouter.ResolvedCols[1] != 100*layout.Pt {
		t.Errorf("Column 1: expected 100pt, got %v", layouter.ResolvedCols[1])
	}

	// Column 2: Remaining 300pt (500 - 100 - 100)
	if layouter.ResolvedCols[2] != 300*layout.Pt {
		t.Errorf("Column 2: expected 300pt, got %v", layouter.ResolvedCols[2])
	}
}

func TestLayouterWithGutter(t *testing.T) {
	g := NewGrid([]TrackSize{
		FrTrack{Fr: 1},
		FrTrack{Fr: 1},
	})
	g.Gutter = Gutter{Column: 20 * layout.Pt}

	regions := &layout.Regions{
		Current: layout.Region{
			Size: layout.Size{Width: 220 * layout.Pt, Height: 800 * layout.Pt},
			Full: true,
		},
	}

	layouter := NewLayouter(g, regions)
	if err := layouter.measureColumns(); err != nil {
		t.Fatalf("measureColumns failed: %v", err)
	}

	// With 20pt gutter, available is 200pt, split evenly = 100pt each
	expected := 100 * layout.Pt
	if layouter.ResolvedCols[0] != expected {
		t.Errorf("Column 0: expected %v, got %v", expected, layouter.ResolvedCols[0])
	}
	if layouter.ResolvedCols[1] != expected {
		t.Errorf("Column 1: expected %v, got %v", expected, layouter.ResolvedCols[1])
	}
}

func TestBasicGridLayout(t *testing.T) {
	g := NewGrid([]TrackSize{
		FixedTrack{Size: 100 * layout.Pt},
		FixedTrack{Size: 100 * layout.Pt},
	})

	g.AddCell(NewCell(0, 0, "A"))
	g.AddCell(NewCell(1, 0, "B"))
	g.AddCell(NewCell(0, 1, "C"))
	g.AddCell(NewCell(1, 1, "D"))

	regions := &layout.Regions{
		Current: layout.Region{
			Size: layout.Size{Width: 300 * layout.Pt, Height: 800 * layout.Pt},
			Full: true,
		},
	}

	fragment, err := LayoutGrid(g, regions)
	if err != nil {
		t.Fatalf("LayoutGrid failed: %v", err)
	}

	if len(fragment) != 1 {
		t.Errorf("expected 1 frame, got %d", len(fragment))
	}

	frame := fragment[0]
	if frame.Size.Width != 300*layout.Pt {
		t.Errorf("expected frame width 300pt, got %v", frame.Size.Width)
	}
}

func TestRowspanTracker(t *testing.T) {
	tracker := NewRowspanTracker()

	cell := NewCell(0, 1, nil)
	cell.Rowspan = 3

	state := tracker.Start(cell)

	if len(tracker.Active) != 1 {
		t.Errorf("expected 1 active rowspan, got %d", len(tracker.Active))
	}

	// Test ActiveAt
	active := tracker.ActiveAt(2) // Row 2 is within span (1-3)
	if len(active) != 1 {
		t.Errorf("expected 1 active at row 2, got %d", len(active))
	}

	active = tracker.ActiveAt(0) // Row 0 is before span
	if len(active) != 0 {
		t.Errorf("expected 0 active at row 0, got %d", len(active))
	}

	// Test CompletedAt
	completed := tracker.CompletedAt(3) // Row 3 is last row of span
	if len(completed) != 1 {
		t.Errorf("expected 1 completed at row 3, got %d", len(completed))
	}

	// Test Remove
	tracker.Remove(state)
	if len(tracker.Active) != 0 {
		t.Errorf("expected 0 active after remove, got %d", len(tracker.Active))
	}
}

func TestRowspanStateAllocateHeight(t *testing.T) {
	cell := NewCell(0, 0, nil)
	cell.Rowspan = 3

	state := &RowspanState{Cell: cell}

	state.AllocateHeight(30 * layout.Pt)
	state.AllocateHeight(40 * layout.Pt)

	if state.AllocatedHeight != 70*layout.Pt {
		t.Errorf("expected allocated 70pt, got %v", state.AllocatedHeight)
	}

	remaining := state.RemainingHeight(100 * layout.Pt)
	if remaining != 30*layout.Pt {
		t.Errorf("expected remaining 30pt, got %v", remaining)
	}
}

func TestUnbreakableGroup(t *testing.T) {
	cell := NewCell(0, 2, nil)
	cell.Rowspan = 4 // Spans rows 2-5

	unbreakable := NewUnbreakable(cell, 2)

	if unbreakable.StartY() != 2 {
		t.Errorf("expected StartY=2, got %d", unbreakable.StartY())
	}
	if unbreakable.EndY() != 6 {
		t.Errorf("expected EndY=6, got %d", unbreakable.EndY())
	}

	// Test Contains
	if !unbreakable.Contains(3) {
		t.Error("should contain row 3")
	}
	if unbreakable.Contains(6) {
		t.Error("should not contain row 6")
	}
}

func TestGridLines(t *testing.T) {
	lines := NewGridLines()

	stroke := &layout.Stroke{
		Paint:     &layout.Color{R: 0, G: 0, B: 0, A: 255},
		Thickness: 1,
	}

	lines.AddHorizontal(stroke, 0, 0, 200*layout.Pt)
	lines.AddVertical(stroke, 100*layout.Pt, 0, 100*layout.Pt)

	if len(lines.Horizontal) != 1 {
		t.Errorf("expected 1 horizontal line, got %d", len(lines.Horizontal))
	}
	if len(lines.Vertical) != 1 {
		t.Errorf("expected 1 vertical line, got %d", len(lines.Vertical))
	}

	hLine := lines.Horizontal[0]
	if hLine.Start.Y != 0 || hLine.End.X != 200*layout.Pt {
		t.Error("horizontal line coordinates incorrect")
	}

	vLine := lines.Vertical[0]
	if vLine.Start.X != 100*layout.Pt || vLine.End.Y != 100*layout.Pt {
		t.Error("vertical line coordinates incorrect")
	}
}

func TestCellBorders(t *testing.T) {
	g := NewGrid([]TrackSize{AutoTrack{}})
	g.Stroke = &layout.Stroke{
		Paint:     &layout.Color{R: 0, G: 0, B: 0, A: 255},
		Thickness: 1,
	}

	regions := &layout.Regions{
		Current: layout.Region{
			Size: layout.Size{Width: 100 * layout.Pt, Height: 100 * layout.Pt},
		},
	}

	layouter := NewLayouter(g, regions)

	// Cell without override
	cell := NewCell(0, 0, nil)
	borders := computeCellBorders(layouter, cell)

	if borders.Top == nil {
		t.Error("expected top border from grid default")
	}
	if borders.Top.Thickness != 1 {
		t.Errorf("expected thickness 1, got %v", borders.Top.Thickness)
	}

	// Cell with override
	cell2 := NewCell(0, 1, nil)
	cell2.Stroke = &CellStroke{
		Top: &layout.Stroke{
			Paint:     &layout.Color{R: 255, G: 0, B: 0, A: 255},
			Thickness: 2,
		},
	}
	borders2 := computeCellBorders(layouter, cell2)

	if borders2.Top.Thickness != 2 {
		t.Errorf("expected overridden thickness 2, got %v", borders2.Top.Thickness)
	}
	if borders2.Bottom.Thickness != 1 {
		t.Error("expected bottom to use grid default")
	}
}
