package eval

// ----------------------------------------------------------------------------
// Table Element
// Matches Rust: typst-library/src/model/table.rs
// ----------------------------------------------------------------------------

// TableElement represents a table with cells arranged in a grid.
// Cells flow left-to-right, top-to-bottom.
type TableElement struct {
	// Columns specifies the column sizing. Each element can be:
	// - IntValue: treated as auto-sized columns (count)
	// - LengthValue: fixed width
	// - RelativeValue: percentage of container
	// - FractionValue: fractional unit
	// - ArrayValue: array of the above
	Columns Value
	// Rows specifies the row sizing (same format as Columns).
	Rows Value
	// Gutter is shorthand for column-gutter and row-gutter.
	Gutter Value
	// ColumnGutter specifies gaps between columns.
	ColumnGutter Value
	// RowGutter specifies gaps between rows.
	RowGutter Value
	// Inset is the padding inside cells (default: 5pt).
	Inset Value
	// Align specifies cell content alignment.
	Align Value
	// Fill is the background fill for cells.
	Fill Value
	// Stroke is the border stroke for cells (default: 1pt + black).
	Stroke Value
	// Children contains the table cell contents and explicit cells.
	Children []TableChild
}

func (*TableElement) IsContentElement() {}

// TableChild represents an item in the table's children.
// Can be content, cell, header, footer, hline, or vline.
// Matches Rust's TableChild enum.
type TableChild struct {
	// Content is set for plain content children.
	Content *Content
	// Cell is set for explicit table.cell() children.
	Cell *TableCellElement
	// Header is set for table.header() children.
	Header *TableHeaderElement
	// Footer is set for table.footer() children.
	Footer *TableFooterElement
	// HLine is set for table.hline() children.
	HLine *TableHLineElement
	// VLine is set for table.vline() children.
	VLine *TableVLineElement
}

// TableCellElement represents an explicit table cell with position/span overrides.
// Matches Rust's TableCell struct.
type TableCellElement struct {
	// Body is the cell's content.
	Body Content
	// X is the column position (0-indexed). If nil, auto-positioned.
	X *int
	// Y is the row position (0-indexed). If nil, auto-positioned.
	Y *int
	// Colspan is the number of columns this cell spans (default: 1).
	Colspan int
	// Rowspan is the number of rows this cell spans (default: 1).
	Rowspan int
	// Inset overrides the table's inset for this cell.
	Inset Value
	// Align overrides the table's alignment for this cell.
	Align Value
	// Fill overrides the table's fill for this cell.
	Fill Value
	// Stroke overrides the table's stroke for this cell.
	Stroke Value
	// Breakable controls whether rows can break across pages.
	Breakable Value
}

func (*TableCellElement) IsContentElement() {}

// TableHeaderElement represents a repeatable table header.
// Matches Rust's TableHeader struct.
type TableHeaderElement struct {
	// Repeat indicates whether this header should repeat across pages (default: true).
	Repeat bool
	// Level is the header level (must be at least 1, default: 1).
	Level int
	// Children contains the cells and lines within the header.
	Children []TableItem
}

func (*TableHeaderElement) IsContentElement() {}

// TableFooterElement represents a repeatable table footer.
// Matches Rust's TableFooter struct.
type TableFooterElement struct {
	// Repeat indicates whether this footer should repeat across pages (default: true).
	Repeat bool
	// Children contains the cells and lines within the footer.
	Children []TableItem
}

func (*TableFooterElement) IsContentElement() {}

// TableHLineElement represents a horizontal line in the table.
// Matches Rust's TableHLine struct.
type TableHLineElement struct {
	// Y is the row above which the line is placed (0-indexed). Nil means auto.
	Y *int
	// Start is the column at which the line starts (0-indexed, inclusive).
	Start int
	// End is the column before which the line ends (0-indexed, exclusive). Nil means to end of table.
	End *int
	// Stroke is the line's stroke. Nil means use default stroke.
	Stroke Value
	// Position is where the line is placed: "top" or "bottom" relative to the row.
	Position string
}

func (*TableHLineElement) IsContentElement() {}

// TableVLineElement represents a vertical line in the table.
// Matches Rust's TableVLine struct.
type TableVLineElement struct {
	// X is the column before which the line is placed (0-indexed). Nil means auto.
	X *int
	// Start is the row at which the line starts (0-indexed, inclusive).
	Start int
	// End is the row on top of which the line ends (0-indexed, exclusive). Nil means to end of table.
	End *int
	// Stroke is the line's stroke. Nil means use default stroke.
	Stroke Value
	// Position is where the line is placed: "start" or "end" relative to the column.
	Position string
}

func (*TableVLineElement) IsContentElement() {}

// TableItem represents a cell or line within a header or footer.
// Matches Rust's TableItem enum.
type TableItem struct {
	Cell  *TableCellElement
	HLine *TableHLineElement
	VLine *TableVLineElement
}
