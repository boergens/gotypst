package grid

import (
	"math"

	"github.com/boergens/gotypst/layout"
	"github.com/boergens/gotypst/layout/flow"
)

// GridLayouter manages the layout of grid/table content across multiple regions.
type GridLayouter struct {
	// Grid is the grid structure being laid out.
	Grid *Grid
	// Regions provides the available layout regions.
	Regions *flow.Regions
	// RCols holds resolved column widths.
	RCols []layout.Abs
	// Width is the total grid width.
	Width layout.Abs

	// RRows holds resolved row info per region.
	RRows []RowState
	// Current tracks the active region state.
	Current Current
	// Finished holds completed frames.
	Finished []flow.Frame
	// UnbreakableRowsLeft tracks remaining unbreakable rows.
	UnbreakableRowsLeft int

	// RepeatingHeaders holds headers that repeat on each page.
	RepeatingHeaders []Header
	// PendingHeaders holds headers waiting to become repeating.
	PendingHeaders []Header
	// Rowspans tracks active rowspan cells.
	Rowspans []Rowspan
	// FinishedHeaderRows tracks info about completed header rows.
	FinishedHeaderRows []FinishedHeaderRowInfo

	// CellLocators maps cell positions to their locators.
	CellLocators map[Axes[int]]interface{}
	// Styles is the inherited style chain.
	Styles interface{}
	// IsRTL indicates right-to-left layout.
	IsRTL bool
	// RowState tracks the current row state.
	RowState RowState

	// Footer is the grid footer (if any).
	Footer *Footer

	// engine is the layout engine context.
	engine *flow.Engine
}

// NewGridLayouter creates a new GridLayouter for the given grid and regions.
func NewGridLayouter(
	engine *flow.Engine,
	grid *Grid,
	regions *flow.Regions,
	styles interface{},
	isRTL bool,
) *GridLayouter {
	gl := &GridLayouter{
		Grid:         grid,
		Regions:      regions,
		RCols:        make([]layout.Abs, len(grid.Cols)),
		Finished:     make([]flow.Frame, 0),
		CellLocators: make(map[Axes[int]]interface{}),
		Styles:       styles,
		IsRTL:        isRTL,
		RowState:     NewRowState(),
		engine:       engine,
	}

	// Initialize current region state.
	gl.Current = Current{
		RegionIdx: 0,
		Height:    0,
		Row:       0,
		Initial:   0,
	}

	// Initialize per-region row states.
	gl.RRows = []RowState{NewRowState()}

	return gl
}

// Layout performs the complete grid layout.
// Returns the fragment of laid out frames.
func (gl *GridLayouter) Layout() ([]flow.Frame, error) {
	// Phase 1: Measure columns.
	if err := gl.measureColumns(); err != nil {
		return nil, err
	}

	// Phase 2: Layout rows.
	for gl.Current.Row < gl.Grid.RowCount {
		if err := gl.layoutRow(gl.Current.Row); err != nil {
			return nil, err
		}
		gl.Current.Row++
	}

	// Phase 3: Finish final region.
	if err := gl.finishRegion(true); err != nil {
		return nil, err
	}

	return gl.Finished, nil
}

// measureColumns resolves all column widths.
// This is a three-phase algorithm:
// 1. Resolve fixed/relative columns
// 2. Measure auto columns by laying out their cells
// 3. Distribute remaining space to fractional columns
func (gl *GridLayouter) measureColumns() error {
	available := gl.Regions.Size.Width

	// Track fractional columns for final distribution.
	var frCols []int
	var totalFr layout.Fr

	// Track auto columns for measurement.
	var autoCols []int

	// Phase 1: Resolve fixed and relative columns, collect auto/fr columns.
	for i, sizing := range gl.Grid.Cols {
		switch s := sizing.(type) {
		case SizingRel:
			gl.RCols[i] = s.RelativeTo(available)
		case SizingAuto:
			autoCols = append(autoCols, i)
			gl.RCols[i] = 0 // Will be measured
		case SizingFr:
			frCols = append(frCols, i)
			totalFr += s.Fr
			gl.RCols[i] = 0 // Will be distributed
		}
	}

	// Phase 2: Measure auto columns.
	if err := gl.measureAutoColumns(autoCols); err != nil {
		return err
	}

	// Calculate remaining space after fixed and auto columns.
	var usedWidth layout.Abs
	for _, w := range gl.RCols {
		usedWidth += w
	}
	remaining := available - usedWidth

	// Phase 3: Distribute remaining space to fractional columns.
	if len(frCols) > 0 && remaining > 0 && totalFr > 0 {
		for _, i := range frCols {
			fr := gl.Grid.Cols[i].(SizingFr).Fr
			gl.RCols[i] = layout.Abs(float64(remaining) * float64(fr) / float64(totalFr))
		}
	} else if remaining < 0 && len(autoCols) > 0 {
		// Fair-share shrinking of auto columns.
		gl.shrinkAutoColumns(autoCols, -remaining)
	}

	// Calculate total width.
	gl.Width = 0
	for _, w := range gl.RCols {
		gl.Width += w
	}

	return nil
}

// measureAutoColumns measures the natural width of auto columns.
func (gl *GridLayouter) measureAutoColumns(autoCols []int) error {
	if len(autoCols) == 0 {
		return nil
	}

	// Build a set of auto columns for quick lookup.
	autoColSet := make(map[int]bool, len(autoCols))
	for _, col := range autoCols {
		autoColSet[col] = true
	}

	// For each auto column, find the maximum width needed by any cell.
	for _, col := range autoCols {
		var maxWidth layout.Abs

		for row := 0; row < gl.Grid.RowCount; row++ {
			cell := gl.Grid.CellAt(col, row)
			if cell == nil {
				continue
			}

			// Skip cells that span multiple columns - they're handled separately.
			if cell.Colspan > 1 {
				continue
			}

			// Measure the cell's natural width.
			width, err := gl.measureCellWidth(cell)
			if err != nil {
				return err
			}

			if width > maxWidth {
				maxWidth = width
			}
		}

		gl.RCols[col] = maxWidth
	}

	// Handle colspan cells: distribute their width requirements across spanned columns.
	if err := gl.distributeColspanWidths(autoColSet); err != nil {
		return err
	}

	return nil
}

// distributeColspanWidths handles cells that span multiple columns.
// If a colspan cell's minimum width exceeds the sum of its spanned columns,
// the extra width is distributed among auto columns in the span.
func (gl *GridLayouter) distributeColspanWidths(autoColSet map[int]bool) error {
	for row := 0; row < gl.Grid.RowCount; row++ {
		for col := 0; col < gl.Grid.ColCount; col++ {
			cell := gl.Grid.CellAt(col, row)
			if cell == nil || cell.Colspan <= 1 {
				continue
			}

			// Only process the cell once (at its origin).
			if cell.X != col || cell.Y != row {
				continue
			}

			// Measure the cell's minimum width.
			cellWidth, err := gl.measureCellWidth(cell)
			if err != nil {
				return err
			}

			// Calculate current total width of spanned columns.
			var spannedWidth layout.Abs
			var autoColsInSpan []int
			for c := col; c < col+cell.Colspan && c < gl.Grid.ColCount; c++ {
				spannedWidth += gl.RCols[c]
				if autoColSet[c] {
					autoColsInSpan = append(autoColsInSpan, c)
				}
			}

			// If the cell needs more width than currently available, distribute the excess.
			if cellWidth > spannedWidth && len(autoColsInSpan) > 0 {
				excess := cellWidth - spannedWidth
				perCol := excess / layout.Abs(len(autoColsInSpan))
				remainder := excess - perCol*layout.Abs(len(autoColsInSpan))

				for i, c := range autoColsInSpan {
					gl.RCols[c] += perCol
					// Give remainder to the first column.
					if i == 0 {
						gl.RCols[c] += remainder
					}
				}
			}
		}
	}
	return nil
}

// measureCellWidth measures the natural width of a cell.
func (gl *GridLayouter) measureCellWidth(cell *Cell) (layout.Abs, error) {
	// For now, return a default width.
	// TODO: Actually layout the cell to measure its natural width.
	// This requires introspection of the cell's content.
	return 72, nil // Default to 1 inch
}

// shrinkAutoColumns applies fair-share shrinking to auto columns.
// This is used when total column widths exceed available space.
func (gl *GridLayouter) shrinkAutoColumns(autoCols []int, excess layout.Abs) {
	if len(autoCols) == 0 || excess <= 0 {
		return
	}

	// Collect column widths for shrinking.
	type colWidth struct {
		idx   int
		width layout.Abs
	}
	cols := make([]colWidth, len(autoCols))
	for i, idx := range autoCols {
		cols[i] = colWidth{idx: idx, width: gl.RCols[idx]}
	}

	remaining := excess

	// Fair-share shrinking algorithm:
	// 1. Calculate fair share = remaining / overlarge_count
	// 2. Find columns below fair share threshold
	// 3. Shrink them to their measured size
	// 4. Redistribute remaining to overlarge columns
	for remaining > 0 && len(cols) > 0 {
		fairShare := remaining / layout.Abs(len(cols))

		// Find the minimum shrink amount.
		var minShrink layout.Abs = math.MaxFloat64
		for _, c := range cols {
			if c.width < minShrink {
				minShrink = c.width
			}
		}

		if minShrink >= fairShare {
			// All columns can shrink by fair share.
			for _, c := range cols {
				gl.RCols[c.idx] -= fairShare
			}
			break
		}

		// Shrink smallest columns completely and redistribute.
		var newCols []colWidth
		for _, c := range cols {
			if c.width <= minShrink {
				shrink := c.width
				if shrink > remaining {
					shrink = remaining
				}
				gl.RCols[c.idx] -= shrink
				remaining -= shrink
			} else {
				newCols = append(newCols, c)
			}
		}
		cols = newCols
	}
}

// layoutRow lays out a single row.
func (gl *GridLayouter) layoutRow(y int) error {
	// Check if this is a gutter row.
	isGutter := gl.Grid.HasGutter && y%2 == 1

	if isGutter {
		return gl.layoutGutterRow(y)
	}

	// Determine the row sizing.
	sizing := gl.Grid.Rows[y]

	switch s := sizing.(type) {
	case SizingAuto:
		return gl.layoutAutoRow(y)
	case SizingRel:
		return gl.layoutRelativeRow(y, s)
	case SizingFr:
		// Fractional rows are deferred to region end.
		return gl.deferFractionalRow(y, s)
	default:
		// Default to auto sizing.
		return gl.layoutAutoRow(y)
	}
}

// layoutGutterRow lays out a gutter row.
func (gl *GridLayouter) layoutGutterRow(y int) error {
	height := gl.Grid.RowGutter

	// Check if there's room in the current region.
	if !gl.fitsInRegion(height) {
		if err := gl.finishRegion(false); err != nil {
			return err
		}
	}

	// Record the gutter row.
	gl.RRows[gl.Current.RegionIdx].Heights[y] = height
	gl.RRows[gl.Current.RegionIdx].IsGutter[y] = true
	gl.Current.Height += height

	return nil
}

// layoutAutoRow lays out an auto-sized row.
func (gl *GridLayouter) layoutAutoRow(y int) error {
	// Measure the row height by laying out all cells.
	height, err := gl.measureRowHeight(y)
	if err != nil {
		return err
	}

	// Check if the row fits in the current region.
	if !gl.fitsInRegion(height) {
		// Can we break here?
		if gl.canBreakBefore(y) {
			if err := gl.finishRegion(false); err != nil {
				return err
			}
		}
		// Otherwise we have to force it.
	}

	// Layout the row's cells.
	if err := gl.layoutRowCells(y, height); err != nil {
		return err
	}

	// Record the row.
	gl.RRows[gl.Current.RegionIdx].Heights[y] = height
	gl.Current.Height += height

	return nil
}

// layoutRelativeRow lays out a row with relative/absolute height.
func (gl *GridLayouter) layoutRelativeRow(y int, sizing SizingRel) error {
	height := sizing.RelativeTo(gl.Regions.Full.Height)

	// Check if it fits.
	if !gl.fitsInRegion(height) {
		if gl.canBreakBefore(y) {
			if err := gl.finishRegion(false); err != nil {
				return err
			}
		}
	}

	// Layout the row's cells.
	if err := gl.layoutRowCells(y, height); err != nil {
		return err
	}

	gl.RRows[gl.Current.RegionIdx].Heights[y] = height
	gl.Current.Height += height

	return nil
}

// deferFractionalRow defers a fractional row to region finalization.
func (gl *GridLayouter) deferFractionalRow(y int, sizing SizingFr) error {
	// Fractional rows get their size from remaining space at region end.
	// For now, record a placeholder.
	gl.RRows[gl.Current.RegionIdx].Heights[y] = 0
	// The actual height will be determined in finishRegion.
	return nil
}

// measureRowHeight measures the natural height of a row.
func (gl *GridLayouter) measureRowHeight(y int) (layout.Abs, error) {
	var maxHeight layout.Abs

	for x := 0; x < gl.Grid.ColCount; x++ {
		cell := gl.Grid.CellAt(x, y)
		if cell == nil {
			continue
		}

		// Skip cells that start in an earlier row (rowspan continuation).
		if cell.Y != y {
			continue
		}

		// For single-row cells, measure height directly.
		if cell.Rowspan == 1 {
			height, err := gl.measureCellHeight(cell, x)
			if err != nil {
				return 0, err
			}
			if height > maxHeight {
				maxHeight = height
			}
		}
		// Multi-row cells are handled by rowspan tracking.
	}

	// Check for rowspans that end at this row and need more height.
	extraHeight := gl.checkRowspanHeightRequirements(y)
	if extraHeight > 0 {
		maxHeight += extraHeight
	}

	// Minimum row height.
	if maxHeight < 10 {
		maxHeight = 10
	}

	return maxHeight, nil
}

// checkRowspanHeightRequirements checks if any rowspans ending at this row
// need additional height to accommodate their content.
// Returns the extra height needed for this row.
func (gl *GridLayouter) checkRowspanHeightRequirements(y int) layout.Abs {
	var extraHeight layout.Abs

	for i := range gl.Rowspans {
		rs := &gl.Rowspans[i]
		endRow := rs.Y + rs.RowspanCount - 1

		// Skip rowspans that don't end at this row.
		if endRow != y {
			continue
		}

		// Get the cell to measure its required height.
		cell := gl.Grid.CellAt(rs.X, rs.Y)
		if cell == nil {
			continue
		}

		// Measure the cell's natural height.
		cellHeight, err := gl.measureCellHeight(cell, rs.X)
		if err != nil {
			continue
		}

		// Calculate the total height of all rows in the span.
		var spannedHeight layout.Abs
		for row := rs.Y; row < y; row++ {
			if h, ok := gl.RRows[gl.Current.RegionIdx].Heights[row]; ok {
				spannedHeight += h
			}
		}

		// If the cell needs more height, the excess goes to this final row.
		if cellHeight > spannedHeight {
			excess := cellHeight - spannedHeight
			if excess > extraHeight {
				extraHeight = excess
			}
		}
	}

	return extraHeight
}

// measureCellHeight measures the natural height of a cell.
func (gl *GridLayouter) measureCellHeight(cell *Cell, x int) (layout.Abs, error) {
	// Calculate the available width for this cell.
	width := layout.Abs(0)
	for col := x; col < x+cell.Colspan && col < gl.Grid.ColCount; col++ {
		width += gl.RCols[col]
	}

	// TODO: Actually layout the cell to measure its height.
	// For now, return a default height.
	return 20, nil
}

// layoutRowCells lays out all cells in a row.
func (gl *GridLayouter) layoutRowCells(y int, height layout.Abs) error {
	// Calculate the Y position for this row.
	rowY := gl.Current.Height

	// Layout each cell in the row.
	dx := layout.Abs(0)
	for x := 0; x < gl.Grid.ColCount; x++ {
		cell := gl.Grid.CellAt(x, y)
		if cell != nil && cell.X == x && cell.Y == y {
			// This is the start of a cell.
			if err := gl.layoutCell(cell, dx, rowY, height); err != nil {
				return err
			}
		}

		dx += gl.RCols[x]
	}

	return nil
}

// layoutCell lays out a single cell at the given position.
func (gl *GridLayouter) layoutCell(cell *Cell, dx, dy, height layout.Abs) error {
	// Calculate the cell's width (accounting for colspan).
	width := layout.Abs(0)
	for col := cell.X; col < cell.X+cell.Colspan && col < gl.Grid.ColCount; col++ {
		width += gl.RCols[col]
	}

	// If this is a multi-row cell, register it as a rowspan.
	if cell.Rowspan > 1 {
		gl.registerRowspan(cell, dx, dy)
	}

	// TODO: Actually layout the cell content into the region.
	// For now, we just track the position.

	// Store the cell's locator for reference.
	gl.CellLocators[Axes[int]{X: cell.X, Y: cell.Y}] = nil

	_ = width // Will be used when actually laying out content

	return nil
}

// registerRowspan registers a multi-row cell for tracking.
func (gl *GridLayouter) registerRowspan(cell *Cell, dx, dy layout.Abs) {
	rowspan := Rowspan{
		X:             cell.X,
		Y:             cell.Y,
		RowspanCount:  cell.Rowspan,
		DX:            dx,
		DY:            dy,
		FirstRegion:   gl.Current.RegionIdx,
		RegionFull:    gl.Regions.Size.Height,
		Heights:       []layout.Abs{gl.Regions.Size.Height - dy},
		IsUnbreakable: !cell.Breakable,
	}
	gl.Rowspans = append(gl.Rowspans, rowspan)
}

// fitsInRegion checks if the given height fits in the current region.
func (gl *GridLayouter) fitsInRegion(height layout.Abs) bool {
	available := gl.Regions.Size.Height - gl.Current.Height
	return available.Fits(height)
}

// canBreakBefore checks if we can break before the given row.
func (gl *GridLayouter) canBreakBefore(y int) bool {
	// Can't break if we have unbreakable rows remaining.
	if gl.UnbreakableRowsLeft > 0 {
		return false
	}

	// Can't break before the first row.
	if y == 0 {
		return false
	}

	// Can't break if this would orphan headers.
	if len(gl.PendingHeaders) > 0 {
		return false
	}

	// Can progress to next region?
	return gl.Regions.MayProgress()
}

// finishRegion completes the current region and starts a new one.
func (gl *GridLayouter) finishRegion(isFinal bool) error {
	// Strip trailing gutter rows.
	gl.stripTrailingGutter()

	// Handle orphan prevention for headers.
	if !isFinal {
		gl.handleOrphanPrevention()
	}

	// Layout footer if present and this is the final region.
	if isFinal && gl.Footer != nil {
		// TODO: Layout footer
	}

	// Size fractional rows from remaining space.
	gl.sizeFractionalRows()

	// Complete pending rowspans.
	gl.completeRowspans()

	// Build the output frame for this region.
	frame := gl.buildRegionFrame()
	gl.Finished = append(gl.Finished, frame)

	if !isFinal {
		// Advance to the next region.
		if err := gl.advanceRegion(); err != nil {
			return err
		}

		// Prepare headers for the next region.
		gl.prepareHeadersForNextRegion()
	}

	return nil
}

// stripTrailingGutter removes trailing gutter rows from the current region.
func (gl *GridLayouter) stripTrailingGutter() {
	rrows := &gl.RRows[gl.Current.RegionIdx]

	// Find the last non-gutter row.
	lastRow := gl.Current.Row - 1
	for lastRow >= 0 {
		if !rrows.IsGutter[lastRow] {
			break
		}
		// Remove this gutter row's height.
		if h, ok := rrows.Heights[lastRow]; ok {
			gl.Current.Height -= h
			delete(rrows.Heights, lastRow)
			delete(rrows.IsGutter, lastRow)
		}
		lastRow--
	}
}

// handleOrphanPrevention removes orphaned headers.
func (gl *GridLayouter) handleOrphanPrevention() {
	// If only headers are in this region, remove them.
	if len(gl.PendingHeaders) > 0 && gl.isOnlyHeaders() {
		gl.PendingHeaders = nil
		gl.Current.Height = gl.Current.Initial
	}
}

// isOnlyHeaders checks if the current region contains only header content.
func (gl *GridLayouter) isOnlyHeaders() bool {
	// Simple check: if current height equals header heights, only headers present.
	// This is a simplified check.
	return false
}

// sizeFractionalRows sizes fractional rows from remaining space.
func (gl *GridLayouter) sizeFractionalRows() {
	// Collect fractional rows.
	var frRows []int
	var totalFr layout.Fr

	for y := 0; y < gl.Current.Row; y++ {
		if sizing, ok := gl.Grid.Rows[y].(SizingFr); ok {
			frRows = append(frRows, y)
			totalFr += sizing.Fr
		}
	}

	if len(frRows) == 0 || totalFr == 0 {
		return
	}

	// Calculate remaining space.
	remaining := gl.Regions.Size.Height - gl.Current.Height
	if remaining <= 0 {
		return
	}

	// Distribute to fractional rows.
	for _, y := range frRows {
		fr := gl.Grid.Rows[y].(SizingFr).Fr
		height := layout.Abs(float64(remaining) * float64(fr) / float64(totalFr))
		gl.RRows[gl.Current.RegionIdx].Heights[y] = height
		gl.Current.Height += height
	}
}

// completeRowspans finalizes rowspans that end in this region.
func (gl *GridLayouter) completeRowspans() {
	// Filter out completed rowspans.
	var remaining []Rowspan
	for _, rs := range gl.Rowspans {
		endRow := rs.Y + rs.RowspanCount
		if endRow <= gl.Current.Row {
			// Rowspan is complete.
			// TODO: Place the rowspan cell content.
		} else {
			remaining = append(remaining, rs)
		}
	}
	gl.Rowspans = remaining
}

// buildRegionFrame creates the output frame for the current region.
func (gl *GridLayouter) buildRegionFrame() flow.Frame {
	// Calculate the actual height used.
	height := gl.Current.Height

	size := layout.Size{Width: gl.Width, Height: height}
	frame := flow.NewFrame(size)

	// TODO: Add cell frames to the output frame.
	// TODO: Add grid lines.

	return frame
}

// advanceRegion moves to the next region.
func (gl *GridLayouter) advanceRegion() error {
	gl.Current.RegionIdx++

	// Get the next region's height from backlog.
	if len(gl.Regions.Backlog) > 0 {
		gl.Regions.Size.Height = gl.Regions.Backlog[0]
		gl.Regions.Backlog = gl.Regions.Backlog[1:]
	} else if gl.Regions.Last != nil {
		gl.Regions.Size.Height = gl.Regions.Last.Height
		gl.Regions.Last = nil
	}

	// Reset current region state.
	gl.Current.Height = 0
	gl.Current.Initial = 0

	// Add a new row state for this region.
	gl.RRows = append(gl.RRows, NewRowState())

	// Update rowspan heights for the new region.
	for i := range gl.Rowspans {
		gl.Rowspans[i].Heights = append(gl.Rowspans[i].Heights, gl.Regions.Size.Height)
	}

	return nil
}

// prepareHeadersForNextRegion prepares headers to repeat in the next region.
func (gl *GridLayouter) prepareHeadersForNextRegion() {
	// Move pending headers to repeating headers.
	gl.RepeatingHeaders = append(gl.RepeatingHeaders, gl.PendingHeaders...)
	gl.PendingHeaders = nil

	// Layout repeating headers at the start of the new region.
	for _, header := range gl.RepeatingHeaders {
		_ = header // TODO: Re-layout header rows
	}
}

// ColumnWidth returns the resolved width of column x.
func (gl *GridLayouter) ColumnWidth(x int) layout.Abs {
	if x < 0 || x >= len(gl.RCols) {
		return 0
	}
	return gl.RCols[x]
}

// RowHeight returns the resolved height of row y in the current region.
func (gl *GridLayouter) RowHeight(y int) layout.Abs {
	if gl.Current.RegionIdx >= len(gl.RRows) {
		return 0
	}
	return gl.RRows[gl.Current.RegionIdx].Heights[y]
}
