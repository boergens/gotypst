package grid

import (
	"github.com/boergens/gotypst/layout"
	"github.com/boergens/gotypst/layout/flow"
)

// Grid represents the resolved grid structure.
type Grid struct {
	// Cols contains column definitions.
	Cols []Track
	// Rows contains row definitions.
	Rows []Track
	// Cells contains the grid cells in row-major order.
	Cells []Cell
	// ColGutter is the gutter between columns.
	ColGutter Sizing
	// RowGutter is the gutter between rows.
	RowGutter Sizing
	// HasFill indicates if any cell has a fill.
	HasFill bool
	// HasStroke indicates if any cell has a stroke.
	HasStroke bool
}

// Track represents a column or row definition.
type Track struct {
	// Sizing determines the track's size.
	Sizing Sizing
	// Index is the track's position (0-based).
	Index int
	// IsGutter indicates if this is a gutter track.
	IsGutter bool
}

// Sizing represents a track sizing value.
type Sizing interface {
	isSizing()
}

// SizingAuto represents auto-sizing based on content.
type SizingAuto struct{}

func (SizingAuto) isSizing() {}

// SizingFixed represents a fixed absolute size.
type SizingFixed struct {
	Value layout.Abs
}

func (SizingFixed) isSizing() {}

// SizingRelative represents a relative size (percentage of available space).
type SizingRelative struct {
	Ratio float64
}

func (SizingRelative) isSizing() {}

// SizingFractional represents a fractional size.
type SizingFractional struct {
	Fr layout.Fr
}

func (SizingFractional) isSizing() {}

// Cell represents a single grid cell.
type Cell struct {
	// Content is the cell's content.
	Content []flow.Child
	// X is the column index.
	X int
	// Y is the row index.
	Y int
	// Colspan is the number of columns spanned.
	Colspan int
	// Rowspan is the number of rows spanned.
	Rowspan int
	// Fill is the cell's background fill.
	Fill *Paint
	// Stroke contains per-side stroke settings.
	Stroke Sides[*Stroke]
	// BreakableInPlace indicates if this cell can break at its position.
	BreakableInPlace bool
}

// Paint represents a fill color or pattern.
type Paint struct {
	Color *Color
}

// Color represents an RGBA color.
type Color struct {
	R, G, B, A uint8
}

// Stroke represents a stroke style.
type Stroke struct {
	// Paint is the stroke color/pattern.
	Paint Paint
	// Thickness is the stroke width.
	Thickness layout.Abs
	// LineCap specifies line cap style.
	LineCap LineCap
	// LineJoin specifies line join style.
	LineJoin LineJoin
	// DashPattern is the dash pattern (nil for solid).
	DashPattern []layout.Abs
	// DashPhase is the dash offset.
	DashPhase layout.Abs
}

// LineCap specifies how line ends are drawn.
type LineCap int

const (
	LineCapButt LineCap = iota
	LineCapRound
	LineCapSquare
)

// LineJoin specifies how line corners are drawn.
type LineJoin int

const (
	LineJoinMiter LineJoin = iota
	LineJoinRound
	LineJoinBevel
)

// Sides holds values for all four sides.
type Sides[T any] struct {
	Left, Top, Right, Bottom T
}

// Axes holds a pair of values for horizontal (X) and vertical (Y) axes.
type Axes[T any] struct {
	X, Y T
}

// RowState tracks the state of a laid out row.
type RowState struct {
	// Height is the row's resolved height.
	Height layout.Abs
	// Y is the row's vertical position in the region.
	Y layout.Abs
	// IsGutter indicates if this is a gutter row.
	IsGutter bool
}

// Current holds the current region's layout state.
type Current struct {
	// InitialHeaderHeight is the height of headers at region start.
	InitialHeaderHeight layout.Abs
	// RepeatingHeaderHeight is the height of repeating headers.
	RepeatingHeaderHeight layout.Abs
	// PendingHeaderHeight is the height of pending headers.
	PendingHeaderHeight layout.Abs
	// FooterHeight is the height reserved for the footer.
	FooterHeight layout.Abs
	// UsedHeight is the height consumed in the current region.
	UsedHeight layout.Abs
}

// AvailableHeight returns the available height in the current region after
// accounting for headers and footers.
func (c *Current) AvailableHeight(regionHeight layout.Abs) layout.Abs {
	return regionHeight - c.RepeatingHeaderHeight - c.FooterHeight - c.UsedHeight
}
