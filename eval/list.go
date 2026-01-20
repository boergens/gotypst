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

// TableElement represents a table containing cells arranged in a grid.
// This is created by the table() function.
type TableElement struct {
	// Columns specifies the column sizing.
	// Each entry can be: auto (nil), a length in points, a ratio, or a fraction.
	Columns []TableSizing
	// Rows specifies the row sizing (optional).
	// If nil, rows are auto-sized.
	Rows []TableSizing
	// Gutter is the gap between all cells (in points).
	// If nil, uses default (0pt).
	Gutter *float64
	// ColumnGutter overrides gutter for columns.
	ColumnGutter *float64
	// RowGutter overrides gutter for rows.
	RowGutter *float64
	// Fill is the default background fill for cells.
	Fill interface{}
	// Align is the default alignment for cell content.
	Align *Alignment2D
	// Stroke is the default stroke for cell borders.
	Stroke *TableStroke
	// Inset is the default cell padding (in points).
	Inset *float64
	// Children contains the table cells and other content.
	Children []Content
}

func (*TableElement) IsContentElement() {}

// TableSizing represents a column or row sizing specification.
type TableSizing struct {
	// Auto indicates the track should be sized to fit content.
	Auto bool
	// Length is an absolute length in points (if not Auto).
	Length *float64
	// Ratio is a relative ratio (0.0-1.0) of available space.
	Ratio *float64
	// Fraction is a fraction of remaining space (1fr = 1.0).
	Fraction *float64
}

// TableStroke represents stroke styling for table borders.
type TableStroke struct {
	// Paint is the stroke color (RGBA).
	Paint *Color
	// Thickness is the stroke thickness in points.
	Thickness *float64
}

// TableCellElement represents a single cell in a table.
// This is created by table.cell() or implicitly from content.
type TableCellElement struct {
	// Body is the cell content.
	Body Content
	// X is the column index (0-based, optional for explicit positioning).
	X *int
	// Y is the row index (0-based, optional for explicit positioning).
	Y *int
	// Colspan is the number of columns this cell spans.
	// If nil, defaults to 1.
	Colspan *int
	// Rowspan is the number of rows this cell spans.
	// If nil, defaults to 1.
	Rowspan *int
	// Fill overrides the table's default fill for this cell.
	Fill interface{}
	// Align overrides the table's default alignment for this cell.
	Align *Alignment2D
	// Stroke overrides the table's default stroke for this cell.
	Stroke *TableStroke
	// Inset overrides the table's default inset for this cell.
	Inset *float64
	// Breakable indicates if this cell can break across pages.
	Breakable *bool
}

func (*TableCellElement) IsContentElement() {}

// TableHlineElement represents a horizontal line in a table.
type TableHlineElement struct {
	// Y is the row index where the line is placed.
	Y *int
	// Start is the starting column index.
	Start *int
	// End is the ending column index (exclusive).
	End *int
	// Stroke is the line stroke style.
	Stroke *TableStroke
}

func (*TableHlineElement) IsContentElement() {}

// TableVlineElement represents a vertical line in a table.
type TableVlineElement struct {
	// X is the column index where the line is placed.
	X *int
	// Start is the starting row index.
	Start *int
	// End is the ending row index (exclusive).
	End *int
	// Stroke is the line stroke style.
	Stroke *TableStroke
}

func (*TableVlineElement) IsContentElement() {}
