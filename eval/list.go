package eval

// ----------------------------------------------------------------------------
// List Container Elements
// ----------------------------------------------------------------------------

// ListElement represents a bullet list containing list items.
// This is created by grouping consecutive ListItemElements or by the list() function.
type ListElement struct {
	// Items contains the list items.
	Items []*ListItemElement
	// Tight indicates whether items should have tight spacing (no paragraph breaks).
	// If nil, defaults to true.
	Tight *bool
	// Marker is the content to use as the list marker.
	// If nil, defaults to "•".
	Marker *Content
}

func (*ListElement) IsContentElement() {}

// EnumElement represents an enumerated (numbered) list containing enum items.
// This is created by grouping consecutive EnumItemElements or by the enum() function.
type EnumElement struct {
	// Items contains the enumerated items.
	Items []*EnumItemElement
	// Tight indicates whether items should have tight spacing (no paragraph breaks).
	// If nil, defaults to true.
	Tight *bool
	// Numbering is the numbering pattern (e.g., "1.", "a)", "I.").
	// If nil, defaults to "1.".
	Numbering *string
	// Start is the starting number for the enumeration.
	// If nil, defaults to 1.
	Start *int
	// Full indicates whether to display full numbering (e.g., "1.1.1" vs "1").
	// If nil, defaults to false.
	Full *bool
}

func (*EnumElement) IsContentElement() {}

// TermsElement represents a terms (definition) list containing term items.
// This is created by grouping consecutive TermItemElements.
type TermsElement struct {
	Items []*TermItemElement
}

func (*TermsElement) IsContentElement() {}

// ----------------------------------------------------------------------------
// Citation Elements
// ----------------------------------------------------------------------------

// CiteElement represents a single citation reference.
// Citations can be grouped together for proper formatting.
type CiteElement struct {
	// Key is the bibliography key being cited.
	Key string

	// Supplement is optional additional text (e.g., page numbers).
	Supplement *Content

	// Form specifies the citation form.
	// Values: "normal", "prose", "year", "author", "full"
	Form string
}

func (*CiteElement) IsContentElement() {}

// CitationGroup represents multiple citations grouped together.
// For example: [1, 2, 3] or (Smith 2020; Jones 2021)
type CitationGroup struct {
	Citations []*CiteElement
}

func (*CitationGroup) IsContentElement() {}

// ----------------------------------------------------------------------------
// Table Elements
// ----------------------------------------------------------------------------

// TableElement represents a table with cells arranged in rows and columns.
// This is created by the table() function.
type TableElement struct {
	// Columns contains the column sizing specifications.
	// Each element can be "auto", a length, or a fraction.
	Columns []TableSizing
	// Rows contains the row sizing specifications (optional).
	Rows []TableSizing
	// Cells contains all the cells in the table, in row-major order.
	Cells []*TableCellElement
	// Align is the default alignment for cells.
	// If nil, cells use their natural alignment.
	Align *Alignment2D
	// Fill is the default fill for cells.
	// If nil, cells have no background.
	Fill *Content
	// Stroke is the default stroke for cell borders.
	// If nil, cells have no border.
	Stroke *Content
	// Inset is the padding inside each cell (in points).
	// If nil, uses default inset.
	Inset *float64
	// ColumnGutter is the spacing between columns (in points).
	ColumnGutter *float64
	// RowGutter is the spacing between rows (in points).
	RowGutter *float64
}

func (*TableElement) IsContentElement() {}

// TableSizing represents a column or row sizing specification.
type TableSizing struct {
	// Auto is true if the size is automatic.
	Auto bool
	// Points is the absolute size in points (if not Auto and not Fr).
	Points float64
	// Fr is the fractional size (if not Auto and not Points).
	Fr float64
}

// TableCellElement represents a single cell in a table.
type TableCellElement struct {
	// Content is the cell's content.
	Content Content
	// X is the column index (0-based), or -1 for auto-placement.
	X int
	// Y is the row index (0-based), or -1 for auto-placement.
	Y int
	// Colspan is the number of columns this cell spans.
	// If 0 or 1, the cell spans a single column.
	Colspan int
	// Rowspan is the number of rows this cell spans.
	// If 0 or 1, the cell spans a single row.
	Rowspan int
	// Align is the cell's alignment (overrides table default).
	Align *Alignment2D
	// Fill is the cell's fill (overrides table default).
	Fill *Content
	// Stroke is the cell's stroke (overrides table default).
	Stroke *Content
	// Inset is the cell's padding (overrides table default).
	Inset *float64
}

func (*TableCellElement) IsContentElement() {}
