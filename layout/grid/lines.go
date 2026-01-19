// Lines handles rendering of grid lines and strokes.
//
// Grid lines can be specified at multiple levels:
// - Grid default stroke
// - Per-cell stroke overrides
// - Horizontal lines (between rows)
// - Vertical lines (between columns)

package grid

import (
	"github.com/boergens/gotypst/layout"
)

// LineSpec specifies how to render a grid line.
type LineSpec struct {
	// Stroke is the line style.
	Stroke *layout.Stroke

	// Start is the starting position.
	Start layout.Point

	// End is the ending position.
	End layout.Point
}

// GridLines contains all lines to be rendered for a grid.
type GridLines struct {
	// Horizontal contains horizontal lines (between rows).
	Horizontal []LineSpec

	// Vertical contains vertical lines (between columns).
	Vertical []LineSpec
}

// NewGridLines creates a new GridLines structure.
func NewGridLines() *GridLines {
	return &GridLines{
		Horizontal: nil,
		Vertical:   nil,
	}
}

// AddHorizontal adds a horizontal line.
func (g *GridLines) AddHorizontal(stroke *layout.Stroke, y, x1, x2 layout.Abs) {
	g.Horizontal = append(g.Horizontal, LineSpec{
		Stroke: stroke,
		Start:  layout.Point{X: x1, Y: y},
		End:    layout.Point{X: x2, Y: y},
	})
}

// AddVertical adds a vertical line.
func (g *GridLines) AddVertical(stroke *layout.Stroke, x, y1, y2 layout.Abs) {
	g.Vertical = append(g.Vertical, LineSpec{
		Stroke: stroke,
		Start:  layout.Point{X: x, Y: y1},
		End:    layout.Point{X: x, Y: y2},
	})
}

// addGridLinesToFrame adds all grid lines to a frame.
// This is called from the layouter after all rows are laid out.
func addGridLinesToFrame(l *Layouter, frame *layout.Frame) {
	lines := computeGridLines(l)

	// Add horizontal lines
	for _, line := range lines.Horizontal {
		addLineToFrame(frame, line)
	}

	// Add vertical lines
	for _, line := range lines.Vertical {
		addLineToFrame(frame, line)
	}
}

// computeGridLines computes all grid lines based on layout state.
func computeGridLines(l *Layouter) *GridLines {
	lines := NewGridLines()
	stroke := l.Grid.Stroke

	if stroke == nil {
		return lines
	}

	// Calculate total width
	var totalWidth layout.Abs
	for _, w := range l.ResolvedCols {
		totalWidth += w
	}
	if l.Grid.ColCount() > 1 && l.Grid.HasGutter() {
		totalWidth += l.Grid.Gutter.Column * layout.Abs(l.Grid.ColCount()-1)
	}

	// Calculate total height from laid rows
	var totalHeight layout.Abs
	for _, row := range l.LaidRows {
		if row.Y+row.Height > totalHeight {
			totalHeight = row.Y + row.Height
		}
	}

	// Add outer border - top
	lines.AddHorizontal(stroke, 0, 0, totalWidth)

	// Add outer border - bottom
	lines.AddHorizontal(stroke, totalHeight, 0, totalWidth)

	// Add outer border - left
	lines.AddVertical(stroke, 0, 0, totalHeight)

	// Add outer border - right
	lines.AddVertical(stroke, totalWidth, 0, totalHeight)

	// Add internal horizontal lines
	var y layout.Abs
	for i, row := range l.LaidRows {
		if i > 0 {
			// Check if we should draw a line here
			// (don't draw through gutter areas for spanning cells)
			if shouldDrawHorizontalLine(l, i) {
				lines.AddHorizontal(stroke, y, 0, totalWidth)
			}
		}
		y = row.Y + row.Height
	}

	// Add internal vertical lines
	var x layout.Abs
	for col := 0; col < l.Grid.ColCount(); col++ {
		if col > 0 {
			// Check if we should draw a line here
			if shouldDrawVerticalLine(l, col) {
				lines.AddVertical(stroke, x, 0, totalHeight)
			}
		}
		x += l.ResolvedCols[col]
		if col < l.Grid.ColCount()-1 && l.Grid.HasGutter() {
			x += l.Grid.Gutter.Column
		}
	}

	return lines
}

// shouldDrawHorizontalLine checks if a horizontal line should be drawn
// between row rowIdx-1 and rowIdx.
func shouldDrawHorizontalLine(l *Layouter, rowIdx int) bool {
	if rowIdx <= 0 || rowIdx >= len(l.LaidRows) {
		return false
	}

	// Don't draw lines through spanning cells
	// Check each column for cells that span this boundary
	for col := 0; col < l.Grid.ColCount(); col++ {
		// Look for cells that span from before to after this row
		for _, cell := range l.Grid.Cells {
			if cell.X <= col && col < cell.EndX() {
				// This cell covers this column
				if cell.Y < rowIdx && rowIdx < cell.EndY() {
					// Cell spans this boundary - might need partial line
					// For now, we draw the line but it will appear over the cell
					// A more sophisticated implementation would clip the line
				}
			}
		}
	}

	return true
}

// shouldDrawVerticalLine checks if a vertical line should be drawn
// between column colIdx-1 and colIdx.
func shouldDrawVerticalLine(l *Layouter, colIdx int) bool {
	if colIdx <= 0 || colIdx >= l.Grid.ColCount() {
		return false
	}

	// Similar to horizontal, check for spanning cells
	return true
}

// addLineToFrame adds a line specification to a frame.
func addLineToFrame(frame *layout.Frame, line LineSpec) {
	if line.Stroke == nil {
		return
	}

	lineShape := &layout.LineShape{
		Start: line.Start,
		End:   line.End,
	}

	item := &layout.ShapeItem{
		Shape:  lineShape,
		Stroke: line.Stroke,
	}

	frame.Push(layout.Point{}, item)
}

// CellBorders computes the borders for a specific cell.
type CellBorders struct {
	Top    *layout.Stroke
	Bottom *layout.Stroke
	Left   *layout.Stroke
	Right  *layout.Stroke
}

// computeCellBorders computes the borders for a cell.
func computeCellBorders(l *Layouter, cell *Cell) CellBorders {
	borders := CellBorders{}

	// Start with grid default
	defaultStroke := l.Grid.Stroke

	// Apply cell-specific overrides
	if cell.Stroke != nil {
		if cell.Stroke.Top != nil {
			borders.Top = cell.Stroke.Top
		} else {
			borders.Top = defaultStroke
		}
		if cell.Stroke.Bottom != nil {
			borders.Bottom = cell.Stroke.Bottom
		} else {
			borders.Bottom = defaultStroke
		}
		if cell.Stroke.Left != nil {
			borders.Left = cell.Stroke.Left
		} else {
			borders.Left = defaultStroke
		}
		if cell.Stroke.Right != nil {
			borders.Right = cell.Stroke.Right
		} else {
			borders.Right = defaultStroke
		}
	} else {
		borders.Top = defaultStroke
		borders.Bottom = defaultStroke
		borders.Left = defaultStroke
		borders.Right = defaultStroke
	}

	// Check if borders should be hidden due to adjacent cells
	// Top border: hidden if cell is not in first row and previous row has spanning cell
	if cell.Y > 0 {
		// Check for cells spanning from above
	}

	// Bottom border: hidden if cell has rowspan and this isn't the last row
	// ... similar logic for other borders

	return borders
}

// addCellBorders adds borders to a cell frame.
func addCellBorders(frame *layout.Frame, borders CellBorders, size layout.Size) {
	// Top border
	if borders.Top != nil {
		frame.Push(layout.Point{}, &layout.ShapeItem{
			Shape: &layout.LineShape{
				Start: layout.Point{X: 0, Y: 0},
				End:   layout.Point{X: size.Width, Y: 0},
			},
			Stroke: borders.Top,
		})
	}

	// Bottom border
	if borders.Bottom != nil {
		frame.Push(layout.Point{}, &layout.ShapeItem{
			Shape: &layout.LineShape{
				Start: layout.Point{X: 0, Y: size.Height},
				End:   layout.Point{X: size.Width, Y: size.Height},
			},
			Stroke: borders.Bottom,
		})
	}

	// Left border
	if borders.Left != nil {
		frame.Push(layout.Point{}, &layout.ShapeItem{
			Shape: &layout.LineShape{
				Start: layout.Point{X: 0, Y: 0},
				End:   layout.Point{X: 0, Y: size.Height},
			},
			Stroke: borders.Left,
		})
	}

	// Right border
	if borders.Right != nil {
		frame.Push(layout.Point{}, &layout.ShapeItem{
			Shape: &layout.LineShape{
				Start: layout.Point{X: size.Width, Y: 0},
				End:   layout.Point{X: size.Width, Y: size.Height},
			},
			Stroke: borders.Right,
		})
	}
}

// GutterLines handles drawing lines around gutter areas.
// When a grid has gutter, lines are drawn on both sides of the gutter.
type GutterLines struct {
	// ColumnGutters contains lines around column gutters.
	ColumnGutters []GutterLine

	// RowGutters contains lines around row gutters.
	RowGutters []GutterLine
}

// GutterLine represents a line at a gutter boundary.
type GutterLine struct {
	// Position is the position of the line.
	Position layout.Abs

	// Start is the start of the line along the perpendicular axis.
	Start layout.Abs

	// End is the end of the line along the perpendicular axis.
	End layout.Abs

	// Stroke is the line style.
	Stroke *layout.Stroke
}

// computeGutterLines computes lines for gutter boundaries.
func computeGutterLines(l *Layouter) *GutterLines {
	if !l.Grid.HasGutter() {
		return nil
	}

	gl := &GutterLines{}
	stroke := l.Grid.Stroke

	if stroke == nil {
		return gl
	}

	// Calculate total height
	var totalHeight layout.Abs
	for _, row := range l.LaidRows {
		if row.Y+row.Height > totalHeight {
			totalHeight = row.Y + row.Height
		}
	}

	// Add lines around column gutters
	var x layout.Abs
	for col := 0; col < l.Grid.ColCount()-1; col++ {
		x += l.ResolvedCols[col]
		// Line before gutter
		gl.ColumnGutters = append(gl.ColumnGutters, GutterLine{
			Position: x,
			Start:    0,
			End:      totalHeight,
			Stroke:   stroke,
		})
		x += l.Grid.Gutter.Column
		// Line after gutter
		gl.ColumnGutters = append(gl.ColumnGutters, GutterLine{
			Position: x,
			Start:    0,
			End:      totalHeight,
			Stroke:   stroke,
		})
	}

	// Similar for row gutters...

	return gl
}
