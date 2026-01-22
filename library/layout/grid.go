package layout

import (
	"github.com/boergens/gotypst/library/foundations"
)

// GridTrackSizing represents a track sizing specification.
// It can be auto, a length, a fraction, or an array of these.
//
// Reference: typst-reference/crates/typst-library/src/layout/grid/mod.rs
type GridTrackSizing struct {
	// Auto indicates auto-sized tracks.
	Auto bool
	// Length is a fixed length in points (if not Auto or Fr).
	Length *float64
	// Fr is a fractional unit (if not Auto or Length).
	Fr *float64
	// Ratio is a percentage (0.0-1.0) relative to available space.
	Ratio *float64
}

// GridElement represents a grid layout element.
// It arranges its children in a grid with configurable columns and rows.
//
// Reference: typst-reference/crates/typst-library/src/layout/grid/mod.rs
type GridElement struct {
	// Columns defines the column track sizes.
	// Can be a number (for auto columns) or an array of sizes.
	Columns []GridTrackSizing
	// Rows defines the row track sizes.
	Rows []GridTrackSizing
	// ColumnGutter is the gap between columns (in points).
	ColumnGutter *float64
	// RowGutter is the gap between rows (in points).
	RowGutter *float64
	// Inset is the cell padding (in points).
	Inset foundations.Value
	// Align is the cell alignment.
	Align foundations.Value
	// Fill is the cell background fill.
	Fill foundations.Value
	// Stroke is the cell stroke.
	Stroke foundations.Value
	// Children contains the grid cells.
	Children []foundations.Content
}

func (*GridElement) IsContentElement() {}
