package eval

// ----------------------------------------------------------------------------
// List Container Elements
// ----------------------------------------------------------------------------

// ListElement represents a bullet list containing list items.
// This is created by grouping consecutive ListItemElements or by the list() function.
type ListElement struct {
	// Items contains the list items.
	Items []*ListItemElement

	// Tight controls spacing between items.
	// If true, uses paragraph leading for spacing; if false, uses paragraph spacing.
	Tight bool

	// Marker is the content used to introduce each item (e.g., "•", "‣", "–").
	// If nil, uses the default marker.
	Marker *Content

	// Indent is the indentation of each item (in points).
	// If nil, uses default (0pt).
	Indent *float64

	// BodyIndent is the space between marker and item body (in points).
	// If nil, uses default (0.5em).
	BodyIndent *float64

	// Spacing is the explicit spacing between items (in points).
	// If nil, uses auto (determined by tight setting).
	Spacing *float64
}

func (*ListElement) IsContentElement() {}

// EnumElement represents an enumerated (numbered) list containing enum items.
// This is created by grouping consecutive EnumItemElements or by the enum() function.
type EnumElement struct {
	// Items contains the enum items.
	Items []*EnumItemElement

	// Tight controls spacing between items.
	// If true, uses paragraph leading for spacing; if false, uses paragraph spacing.
	Tight bool

	// Numbering is the numbering pattern (e.g., "1.", "a)", "I.").
	// If empty, uses the default "1.".
	Numbering string

	// Start is the starting number for the enumeration.
	// If nil, uses auto (determined by first item's number or 1).
	Start *int

	// Full indicates whether to display full numbering for nested enums.
	// For example, "1.1)" instead of "1)".
	Full bool

	// Reversed indicates whether to reverse the numbering direction.
	Reversed bool

	// Indent is the indentation of each item (in points).
	// If nil, uses default (0pt).
	Indent *float64

	// BodyIndent is the space between number and item body (in points).
	// If nil, uses default (0.5em).
	BodyIndent *float64

	// Spacing is the explicit spacing between items (in points).
	// If nil, uses auto (determined by tight setting).
	Spacing *float64

	// NumberAlign is the alignment for numbers.
	// Default is "end + top".
	NumberAlign string
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
