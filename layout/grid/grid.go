// Package grid provides grid and table layout for GoTypst.
//
// This package is a Go translation of typst-layout/src/grid/ from the original
// Typst compiler. It handles grid and table layout with support for:
//   - Cell positioning and alignment
//   - Cell spanning (colspan and rowspan)
//   - Grid lines and strokes
//   - Multi-region (multi-page) support
//   - Repeating headers and footers
package grid

import (
	"github.com/boergens/gotypst/layout"
)

// Grid represents a grid or table element to be laid out.
// Both #grid() and #table() use this structure, with table having
// additional default styling.
type Grid struct {
	// Tracks defines the column and row track sizing.
	Tracks Tracks

	// Cells contains all cells in the grid.
	// Cells are stored in row-major order but may have gaps
	// due to spanning cells occupying multiple positions.
	Cells []*Cell

	// Gutter defines the spacing between tracks.
	Gutter Gutter

	// Fill is the default background fill for cells.
	Fill *layout.Color

	// Stroke is the default stroke for grid lines.
	Stroke *layout.Stroke

	// Align is the default alignment for cell content.
	Align layout.Alignment

	// Header specifies which rows are repeated headers.
	Header *HeaderFooter

	// Footer specifies which rows are repeated footers.
	Footer *HeaderFooter

	// RowCount is the number of rows in the grid.
	// This is computed from cells and track definitions.
	RowCount int
}

// Tracks defines the sizing for column and row tracks.
type Tracks struct {
	// Columns defines the column track sizes.
	Columns []TrackSize

	// Rows defines the row track sizes.
	Rows []TrackSize
}

// TrackSize represents the size specification for a track (column or row).
type TrackSize interface {
	isTrackSize()
}

// AutoTrack represents an auto-sized track.
// The track will be sized to fit its content.
type AutoTrack struct{}

func (AutoTrack) isTrackSize() {}

// FixedTrack represents a fixed-size track.
type FixedTrack struct {
	Size layout.Abs
}

func (FixedTrack) isTrackSize() {}

// RelativeTrack represents a track sized relative to the container.
type RelativeTrack struct {
	Ratio layout.Ratio
}

func (RelativeTrack) isTrackSize() {}

// FrTrack represents a fractional track that shares remaining space.
type FrTrack struct {
	Fr layout.Fr
}

func (FrTrack) isTrackSize() {}

// Gutter defines spacing between tracks.
type Gutter struct {
	// Column is the space between columns.
	Column layout.Abs
	// Row is the space between rows.
	Row layout.Abs
}

// HeaderFooter defines repeating header or footer rows.
type HeaderFooter struct {
	// Start is the first row index (0-based).
	Start int
	// End is the row index after the last row (exclusive).
	End int
	// Repeat indicates whether to repeat on every page.
	Repeat bool
}

// Cell represents a single cell in the grid.
type Cell struct {
	// Content is the cell's content to be laid out.
	// This is a placeholder - actual content type depends on the content model.
	Content interface{}

	// X is the column index (0-based).
	X int
	// Y is the row index (0-based).
	Y int

	// Colspan is the number of columns this cell spans.
	Colspan int
	// Rowspan is the number of rows this cell spans.
	Rowspan int

	// Fill overrides the grid's default fill for this cell.
	Fill *layout.Color

	// Stroke overrides the grid's default stroke for this cell.
	Stroke *CellStroke

	// Align overrides the grid's default alignment for this cell.
	Align *layout.Alignment

	// Inset defines padding inside the cell.
	Inset layout.Sides[layout.Abs]

	// Breakable indicates whether this cell can break across pages.
	Breakable bool
}

// NewCell creates a new cell at the given position.
func NewCell(x, y int, content interface{}) *Cell {
	return &Cell{
		Content: content,
		X:       x,
		Y:       y,
		Colspan: 1,
		Rowspan: 1,
	}
}

// EndX returns the column index after the last column this cell occupies.
func (c *Cell) EndX() int {
	return c.X + c.Colspan
}

// EndY returns the row index after the last row this cell occupies.
func (c *Cell) EndY() int {
	return c.Y + c.Rowspan
}

// Contains returns true if the cell covers the given position.
func (c *Cell) Contains(x, y int) bool {
	return x >= c.X && x < c.EndX() && y >= c.Y && y < c.EndY()
}

// CellStroke defines stroke styling for cell borders.
// Each side can be independently styled.
type CellStroke struct {
	Left   *layout.Stroke
	Top    *layout.Stroke
	Right  *layout.Stroke
	Bottom *layout.Stroke
}

// CellPosition represents a position in the grid.
type CellPosition struct {
	X int
	Y int
}

// NewGrid creates a new grid with the given column tracks.
func NewGrid(columns []TrackSize) *Grid {
	return &Grid{
		Tracks: Tracks{
			Columns: columns,
			Rows:    nil, // Auto rows by default
		},
		Cells: nil,
	}
}

// AddCell adds a cell to the grid.
func (g *Grid) AddCell(cell *Cell) {
	g.Cells = append(g.Cells, cell)
	// Update row count if necessary
	if cell.EndY() > g.RowCount {
		g.RowCount = cell.EndY()
	}
}

// CellAt returns the cell at the given position, or nil if empty.
// This handles spanning cells - it returns the cell that covers the position.
func (g *Grid) CellAt(x, y int) *Cell {
	for _, cell := range g.Cells {
		if cell.Contains(x, y) {
			return cell
		}
	}
	return nil
}

// ColCount returns the number of columns in the grid.
func (g *Grid) ColCount() int {
	return len(g.Tracks.Columns)
}

// ColumnAt returns the column track size at index, or auto if out of range.
func (g *Grid) ColumnAt(x int) TrackSize {
	if x < len(g.Tracks.Columns) {
		return g.Tracks.Columns[x]
	}
	return AutoTrack{}
}

// RowAt returns the row track size at index, or auto if out of range.
func (g *Grid) RowAt(y int) TrackSize {
	if y < len(g.Tracks.Rows) {
		return g.Tracks.Rows[y]
	}
	return AutoTrack{}
}

// IsAutoRow returns true if the row at y is auto-sized.
func (g *Grid) IsAutoRow(y int) bool {
	_, ok := g.RowAt(y).(AutoTrack)
	return ok
}

// IsFrRow returns true if the row at y is fractional.
func (g *Grid) IsFrRow(y int) bool {
	_, ok := g.RowAt(y).(FrTrack)
	return ok
}

// HasGutter returns true if the grid has any gutter spacing.
func (g *Grid) HasGutter() bool {
	return g.Gutter.Column > 0 || g.Gutter.Row > 0
}

// CellsInRow returns all cells that start in the given row.
func (g *Grid) CellsInRow(y int) []*Cell {
	var cells []*Cell
	for _, cell := range g.Cells {
		if cell.Y == y {
			cells = append(cells, cell)
		}
	}
	return cells
}

// CellsInColumn returns all cells that start in the given column.
func (g *Grid) CellsInColumn(x int) []*Cell {
	var cells []*Cell
	for _, cell := range g.Cells {
		if cell.X == x {
			cells = append(cells, cell)
		}
	}
	return cells
}

// Entry points for layout.

// LayoutGrid lays out a grid element.
// This is the main entry point for grid layout.
func LayoutGrid(grid *Grid, regions *layout.Regions) (layout.Fragment, error) {
	layouter := NewLayouter(grid, regions)
	return layouter.Layout()
}

// LayoutTable lays out a table element.
// Tables use the same underlying layouter with different default styling.
func LayoutTable(grid *Grid, regions *layout.Regions) (layout.Fragment, error) {
	// Table has default styling applied
	if grid.Stroke == nil {
		// Default table stroke: 1pt black lines
		grid.Stroke = &layout.Stroke{
			Paint:     &layout.Color{R: 0, G: 0, B: 0, A: 255},
			Thickness: 1.0,
		}
	}
	return LayoutGrid(grid, regions)
}
