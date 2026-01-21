package eval

// ----------------------------------------------------------------------------
// Table Element
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
// It can be either plain content or an explicit table.cell().
type TableChild struct {
	// Content is set for plain content children.
	Content *Content
	// Cell is set for explicit table.cell() children.
	Cell *TableCellElement
}

// TableCellElement represents an explicit table cell with position/span overrides.
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
