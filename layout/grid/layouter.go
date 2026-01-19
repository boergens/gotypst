// Layouter implements the grid layout algorithm.
//
// The layout algorithm proceeds in three phases:
// 1. Column Measurement - resolve all column widths
// 2. Row Layout - lay out each row, handling page breaks
// 3. Region Finalization - complete each output frame

package grid

import (
	"github.com/boergens/gotypst/layout"
)

// Layouter manages the state of a grid layout operation.
type Layouter struct {
	// Grid is the grid being laid out.
	Grid *Grid

	// Regions contains the available layout regions.
	Regions *layout.Regions

	// Resolved column widths.
	ResolvedCols []layout.Abs

	// Row states for tracking layout progress.
	RowStates []RowState

	// Current region state.
	Current RegionState

	// Laid out rows in the current region.
	LaidRows []LayoutRow

	// Rowspan tracking.
	Rowspans *RowspanTracker

	// Output frames.
	Frames []*layout.Frame

	// RTL indicates right-to-left layout.
	RTL bool
}

// RowState tracks the state of a row during layout.
type RowState struct {
	// Height is the resolved height of this row.
	Height layout.Abs

	// Y is the y-position of this row in the current region.
	Y layout.Abs

	// InProgress indicates layout is still in progress.
	InProgress bool

	// Completed indicates the row has been fully laid out.
	Completed bool
}

// RegionState tracks state for the current output region.
type RegionState struct {
	// Y is the current y-position in the region.
	Y layout.Abs

	// InitialHeaderHeight is the height of initial headers.
	InitialHeaderHeight layout.Abs

	// RepeatingHeaderHeight is the height of repeating headers.
	RepeatingHeaderHeight layout.Abs

	// PendingHeaderHeight is height of headers not yet placed.
	PendingHeaderHeight layout.Abs

	// FooterHeight is the reserved height for footers.
	FooterHeight layout.Abs

	// RegionIndex is the index of the current region.
	RegionIndex int
}

// LayoutRow represents a row that has been laid out.
type LayoutRow struct {
	// Frame contains the row's content.
	Frame *layout.Frame

	// Y is the row's y-position in the current region.
	Y layout.Abs

	// Height is the row's height.
	Height layout.Abs

	// IsGutter indicates this is a gutter row.
	IsGutter bool
}

// NewLayouter creates a new grid layouter.
func NewLayouter(grid *Grid, regions *layout.Regions) *Layouter {
	return &Layouter{
		Grid:       grid,
		Regions:    regions,
		RowStates:  make([]RowState, grid.RowCount),
		Rowspans:   NewRowspanTracker(),
		LaidRows:   nil,
		Frames:     nil,
		RTL:        false,
	}
}

// Layout performs the complete grid layout.
func (l *Layouter) Layout() (layout.Fragment, error) {
	// Phase 1: Measure columns
	if err := l.measureColumns(); err != nil {
		return nil, err
	}

	// Phase 2: Layout rows
	for y := 0; y < l.Grid.RowCount; y++ {
		if err := l.layoutRow(y); err != nil {
			return nil, err
		}
	}

	// Phase 3: Finalize last region
	l.finishRegion()

	return layout.Fragment(l.Frames), nil
}

// measureColumns resolves all column widths.
func (l *Layouter) measureColumns() error {
	region := l.Regions.First()
	availableWidth := region.Size.Width

	// Account for gutter between columns
	numCols := l.Grid.ColCount()
	if numCols > 1 && l.Grid.HasGutter() {
		availableWidth -= l.Grid.Gutter.Column * layout.Abs(numCols-1)
	}

	l.ResolvedCols = make([]layout.Abs, numCols)

	// First pass: resolve fixed and relative columns
	var totalFixed layout.Abs
	var totalFr layout.Fr
	autoIndices := []int{}

	for i, track := range l.Grid.Tracks.Columns {
		switch t := track.(type) {
		case FixedTrack:
			l.ResolvedCols[i] = t.Size
			totalFixed += t.Size
		case RelativeTrack:
			width := t.Ratio.Resolve(availableWidth)
			l.ResolvedCols[i] = width
			totalFixed += width
		case FrTrack:
			totalFr += t.Fr
			// Will be resolved later
		case AutoTrack:
			autoIndices = append(autoIndices, i)
		}
	}

	// Second pass: measure auto columns
	for _, i := range autoIndices {
		width := l.measureAutoColumn(i)
		l.ResolvedCols[i] = width
		totalFixed += width
	}

	// Third pass: distribute remaining space to fractional columns
	remaining := availableWidth - totalFixed
	if remaining > 0 && totalFr > 0 {
		perFr := remaining / layout.Abs(totalFr)
		for i, track := range l.Grid.Tracks.Columns {
			if fr, ok := track.(FrTrack); ok {
				l.ResolvedCols[i] = layout.Abs(fr.Fr) * perFr
			}
		}
	}

	// Handle shrinking if content doesn't fit
	if totalFixed > availableWidth {
		l.shrinkColumns(availableWidth, totalFixed)
	}

	return nil
}

// measureAutoColumn measures the natural width of an auto column.
func (l *Layouter) measureAutoColumn(col int) layout.Abs {
	var maxWidth layout.Abs

	// Find the widest content in this column
	for _, cell := range l.Grid.CellsInColumn(col) {
		if cell.Colspan == 1 {
			// For non-spanning cells, measure the content
			width := l.measureCellWidth(cell)
			if width > maxWidth {
				maxWidth = width
			}
		}
	}

	// TODO: Handle spanning cells that may require minimum width

	return maxWidth
}

// measureCellWidth measures the natural width of a cell's content.
func (l *Layouter) measureCellWidth(cell *Cell) layout.Abs {
	// Placeholder: actual implementation would layout the cell content
	// and return its natural width.
	// For now, return a reasonable default.
	if cell.Content == nil {
		return 0
	}

	// This would be replaced with actual content measurement
	// based on the content model implementation.
	return 50 * layout.Pt // Default width
}

// shrinkColumns applies fair-share shrinking when columns exceed available space.
func (l *Layouter) shrinkColumns(available, total layout.Abs) {
	if total <= 0 {
		return
	}

	// Simple proportional shrinking
	scale := float64(available) / float64(total)
	for i := range l.ResolvedCols {
		l.ResolvedCols[i] = layout.Abs(float64(l.ResolvedCols[i]) * scale)
	}
}

// layoutRow lays out a single row.
func (l *Layouter) layoutRow(y int) error {
	// Check if we need a region break
	if l.needsRegionBreak(y) {
		l.finishRegion()
		l.startNewRegion()
	}

	// Get row height specification
	track := l.Grid.RowAt(y)

	switch t := track.(type) {
	case AutoTrack:
		return l.layoutAutoRow(y)
	case FixedTrack:
		return l.layoutFixedRow(y, t.Size)
	case RelativeTrack:
		region := l.Regions.First()
		height := t.Ratio.Resolve(region.Size.Height)
		return l.layoutFixedRow(y, height)
	case FrTrack:
		// Fractional rows are deferred until region finalization
		l.RowStates[y].InProgress = true
		return nil
	default:
		return l.layoutAutoRow(y)
	}
}

// layoutAutoRow lays out an auto-height row.
func (l *Layouter) layoutAutoRow(y int) error {
	cells := l.Grid.CellsInRow(y)
	rowHeight := l.measureRowHeight(y, cells)

	frame := l.createRowFrame(y, rowHeight, cells)

	l.LaidRows = append(l.LaidRows, LayoutRow{
		Frame:  frame,
		Y:      l.Current.Y,
		Height: rowHeight,
	})

	l.Current.Y += rowHeight
	l.RowStates[y].Height = rowHeight
	l.RowStates[y].Completed = true

	// Add gutter after row (except last row)
	if y < l.Grid.RowCount-1 && l.Grid.HasGutter() {
		l.Current.Y += l.Grid.Gutter.Row
	}

	return nil
}

// layoutFixedRow lays out a fixed-height row.
func (l *Layouter) layoutFixedRow(y int, height layout.Abs) error {
	cells := l.Grid.CellsInRow(y)
	frame := l.createRowFrame(y, height, cells)

	l.LaidRows = append(l.LaidRows, LayoutRow{
		Frame:  frame,
		Y:      l.Current.Y,
		Height: height,
	})

	l.Current.Y += height
	l.RowStates[y].Height = height
	l.RowStates[y].Completed = true

	// Add gutter after row (except last row)
	if y < l.Grid.RowCount-1 && l.Grid.HasGutter() {
		l.Current.Y += l.Grid.Gutter.Row
	}

	return nil
}

// measureRowHeight determines the height of an auto row.
func (l *Layouter) measureRowHeight(y int, cells []*Cell) layout.Abs {
	var maxHeight layout.Abs

	for _, cell := range cells {
		// Only consider cells that start in this row and don't span
		if cell.Y == y && cell.Rowspan == 1 {
			height := l.measureCellHeight(cell)
			if height > maxHeight {
				maxHeight = height
			}
		}
	}

	// Consider rowspans that complete in this row
	for _, cell := range l.Grid.Cells {
		if cell.Rowspan > 1 && cell.EndY()-1 == y {
			// Calculate how much height this spanning cell needs
			// beyond already allocated rows
			cellHeight := l.measureCellHeight(cell)
			var allocated layout.Abs
			for row := cell.Y; row < y; row++ {
				allocated += l.RowStates[row].Height
				if row < y-1 && l.Grid.HasGutter() {
					allocated += l.Grid.Gutter.Row
				}
			}
			needed := cellHeight - allocated
			if needed > maxHeight {
				maxHeight = needed
			}
		}
	}

	return maxHeight
}

// measureCellHeight measures the natural height of a cell.
func (l *Layouter) measureCellHeight(cell *Cell) layout.Abs {
	// Placeholder: actual implementation would layout the cell content
	// with the resolved column width and return its height.
	if cell.Content == nil {
		return 0
	}

	// Get the available width for this cell
	_ = l.getCellWidth(cell)

	// This would be replaced with actual content measurement
	return 20 * layout.Pt // Default height
}

// getCellWidth returns the total width available for a cell.
func (l *Layouter) getCellWidth(cell *Cell) layout.Abs {
	var width layout.Abs
	for x := cell.X; x < cell.EndX(); x++ {
		width += l.ResolvedCols[x]
		if x < cell.EndX()-1 && l.Grid.HasGutter() {
			width += l.Grid.Gutter.Column
		}
	}
	return width
}

// createRowFrame creates a frame containing all cells in a row.
func (l *Layouter) createRowFrame(y int, height layout.Abs, cells []*Cell) *layout.Frame {
	// Calculate total width
	var totalWidth layout.Abs
	for _, w := range l.ResolvedCols {
		totalWidth += w
	}
	if l.Grid.ColCount() > 1 && l.Grid.HasGutter() {
		totalWidth += l.Grid.Gutter.Column * layout.Abs(l.Grid.ColCount()-1)
	}

	frame := layout.NewFrame(layout.Size{Width: totalWidth, Height: height})

	// Position each cell
	var x layout.Abs
	for col := 0; col < l.Grid.ColCount(); col++ {
		cell := l.Grid.CellAt(col, y)
		if cell != nil && cell.X == col && cell.Y == y {
			// This is the cell's origin, lay it out
			cellFrame := l.layoutCell(cell, height)
			pos := layout.Point{X: x, Y: 0}
			if l.RTL {
				// Mirror for RTL
				pos.X = totalWidth - x - cellFrame.Size.Width
			}
			frame.PushFrame(pos, cellFrame)
		}

		x += l.ResolvedCols[col]
		if col < l.Grid.ColCount()-1 && l.Grid.HasGutter() {
			x += l.Grid.Gutter.Column
		}
	}

	return frame
}

// layoutCell lays out a single cell's content.
func (l *Layouter) layoutCell(cell *Cell, rowHeight layout.Abs) *layout.Frame {
	width := l.getCellWidth(cell)

	// For spanning cells, calculate total height
	height := rowHeight
	if cell.Rowspan > 1 {
		height = l.getCellHeight(cell)
	}

	frame := layout.NewFrame(layout.Size{Width: width, Height: height})

	// Apply cell fill if set
	if fill := l.getCellFill(cell); fill != nil {
		rect := &layout.ShapeItem{
			Shape: &layout.RectShape{Size: frame.Size},
			Fill:  fill,
		}
		frame.Push(layout.Point{}, rect)
	}

	// Layout cell content (placeholder)
	// Actual implementation would call content layout

	return frame
}

// getCellHeight returns the total height for a spanning cell.
func (l *Layouter) getCellHeight(cell *Cell) layout.Abs {
	var height layout.Abs
	for y := cell.Y; y < cell.EndY(); y++ {
		height += l.RowStates[y].Height
		if y < cell.EndY()-1 && l.Grid.HasGutter() {
			height += l.Grid.Gutter.Row
		}
	}
	return height
}

// getCellFill returns the fill color for a cell.
func (l *Layouter) getCellFill(cell *Cell) *layout.Color {
	if cell.Fill != nil {
		return cell.Fill
	}
	return l.Grid.Fill
}

// needsRegionBreak checks if we need to break to a new region before row y.
func (l *Layouter) needsRegionBreak(y int) bool {
	if l.Current.RegionIndex == 0 && l.Current.Y == 0 {
		return false // First row of first region
	}

	region := l.Regions.First()
	availableHeight := region.Size.Height - l.Current.Y - l.Current.FooterHeight

	// Estimate height needed for this row
	estimatedHeight := l.estimateRowHeight(y)

	return estimatedHeight > availableHeight
}

// estimateRowHeight estimates the height needed for a row.
func (l *Layouter) estimateRowHeight(y int) layout.Abs {
	track := l.Grid.RowAt(y)
	switch t := track.(type) {
	case FixedTrack:
		return t.Size
	case RelativeTrack:
		region := l.Regions.First()
		return t.Ratio.Resolve(region.Size.Height)
	default:
		// For auto rows, we need to measure
		return l.measureRowHeight(y, l.Grid.CellsInRow(y))
	}
}

// finishRegion completes the current region and adds it to output.
func (l *Layouter) finishRegion() {
	if len(l.LaidRows) == 0 {
		return
	}

	region := l.Regions.First()
	frame := layout.NewFrame(region.Size)

	// Add all laid out rows
	for _, row := range l.LaidRows {
		frame.PushFrame(layout.Point{X: 0, Y: row.Y}, row.Frame)
	}

	// Add grid lines
	l.addGridLines(frame)

	l.Frames = append(l.Frames, frame)
	l.LaidRows = nil
}

// startNewRegion prepares for a new output region.
func (l *Layouter) startNewRegion() {
	l.Current.RegionIndex++
	l.Current.Y = 0

	// Handle repeating headers
	if l.Grid.Header != nil && l.Grid.Header.Repeat {
		// Re-layout header rows
		for y := l.Grid.Header.Start; y < l.Grid.Header.End; y++ {
			l.layoutRow(y)
		}
	}
}

// addGridLines adds grid line strokes to a frame.
func (l *Layouter) addGridLines(frame *layout.Frame) {
	if l.Grid.Stroke == nil {
		return
	}

	// This will be implemented in lines.go
	addGridLinesToFrame(l, frame)
}
