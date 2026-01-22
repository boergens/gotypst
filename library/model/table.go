// Table element types for Typst.
// Translated from typst-library/src/model/table.rs

package model

import "github.com/boergens/gotypst/library/foundations"

// TableElem represents a table with cells arranged in a grid.
// Cells flow left-to-right, top-to-bottom.
type TableElem struct {
	// Columns specifies the column sizing.
	Columns foundations.Value
	// Rows specifies the row sizing.
	Rows foundations.Value
	// Gutter is shorthand for column-gutter and row-gutter.
	Gutter foundations.Value
	// ColumnGutter specifies gaps between columns.
	ColumnGutter foundations.Value
	// RowGutter specifies gaps between rows.
	RowGutter foundations.Value
	// Inset is the padding inside cells (default: 5pt).
	Inset foundations.Value
	// Align specifies cell content alignment.
	Align foundations.Value
	// Fill is the background fill for cells.
	Fill foundations.Value
	// Stroke is the border stroke for cells (default: 1pt + black).
	Stroke foundations.Value
	// Children contains the table cell contents and explicit cells.
	Children []TableChild
}

func (*TableElem) IsContentElement() {}

// TableChild represents an item in the table's children.
// Corresponds to Rust's TableChild enum.
type TableChild struct {
	// Content is set for plain content children.
	Content *foundations.Content
	// Cell is set for explicit table.cell() children.
	Cell *TableCellElem
	// Header is set for table.header() children.
	Header *TableHeaderElem
	// Footer is set for table.footer() children.
	Footer *TableFooterElem
	// HLine is set for table.hline() children.
	HLine *TableHLineElem
	// VLine is set for table.vline() children.
	VLine *TableVLineElem
}

// TableCellElem represents an explicit table cell with position/span overrides.
// Corresponds to Rust's TableCell struct.
type TableCellElem struct {
	// Body is the cell's content.
	Body foundations.Content
	// X is the column position (0-indexed). If nil, auto-positioned.
	X *int
	// Y is the row position (0-indexed). If nil, auto-positioned.
	Y *int
	// Colspan is the number of columns this cell spans (default: 1).
	Colspan int
	// Rowspan is the number of rows this cell spans (default: 1).
	Rowspan int
	// Inset overrides the table's inset for this cell.
	Inset foundations.Value
	// Align overrides the table's alignment for this cell.
	Align foundations.Value
	// Fill overrides the table's fill for this cell.
	Fill foundations.Value
	// Stroke overrides the table's stroke for this cell.
	Stroke foundations.Value
	// Breakable controls whether rows can break across pages.
	Breakable foundations.Value
}

func (*TableCellElem) IsContentElement() {}

// TableHeaderElem represents a repeatable table header.
// Corresponds to Rust's TableHeader struct.
type TableHeaderElem struct {
	// Repeat indicates whether this header should repeat across pages (default: true).
	Repeat bool
	// Level is the header level (must be at least 1, default: 1).
	Level int
	// Children contains the cells and lines within the header.
	Children []TableItem
}

func (*TableHeaderElem) IsContentElement() {}

// TableFooterElem represents a repeatable table footer.
// Corresponds to Rust's TableFooter struct.
type TableFooterElem struct {
	// Repeat indicates whether this footer should repeat across pages (default: true).
	Repeat bool
	// Children contains the cells and lines within the footer.
	Children []TableItem
}

func (*TableFooterElem) IsContentElement() {}

// TableHLineElem represents a horizontal line in the table.
// Corresponds to Rust's TableHLine struct.
type TableHLineElem struct {
	// Y is the row above which the line is placed (0-indexed). Nil means auto.
	Y *int
	// Start is the column at which the line starts (0-indexed, inclusive).
	Start int
	// End is the column before which the line ends (0-indexed, exclusive). Nil means to end.
	End *int
	// Stroke is the line's stroke. Nil means use default stroke.
	Stroke foundations.Value
	// Position is where the line is placed: "top" or "bottom" relative to the row.
	Position string
}

func (*TableHLineElem) IsContentElement() {}

// TableVLineElem represents a vertical line in the table.
// Corresponds to Rust's TableVLine struct.
type TableVLineElem struct {
	// X is the column before which the line is placed (0-indexed). Nil means auto.
	X *int
	// Start is the row at which the line starts (0-indexed, inclusive).
	Start int
	// End is the row on top of which the line ends (0-indexed, exclusive). Nil means to end.
	End *int
	// Stroke is the line's stroke. Nil means use default stroke.
	Stroke foundations.Value
	// Position is where the line is placed: "start" or "end" relative to the column.
	Position string
}

func (*TableVLineElem) IsContentElement() {}

// TableItem represents a cell or line within a header or footer.
// Corresponds to Rust's TableItem enum.
type TableItem struct {
	Cell  *TableCellElem
	HLine *TableHLineElem
	VLine *TableVLineElem
}
