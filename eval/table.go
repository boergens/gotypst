package eval

// ----------------------------------------------------------------------------
// Table Element
// ----------------------------------------------------------------------------

// TableElement represents a table with cells arranged in a grid.
// Cells flow left-to-right, top-to-bottom.
type TableElement struct {
	// Columns specifies the number of columns in the table.
	Columns int
	// Cells contains the table cell contents, arranged left-to-right, top-to-bottom.
	Cells []Content
}

func (*TableElement) IsContentElement() {}
