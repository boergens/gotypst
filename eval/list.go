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
