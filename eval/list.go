package eval

// ----------------------------------------------------------------------------
// List Container Elements
// ----------------------------------------------------------------------------

// ListElement represents a bullet list containing list items.
// This is created by grouping consecutive ListItemElements.
type ListElement struct {
	Items []*ListItemElement
}

func (*ListElement) IsContentElement() {}

// EnumElement represents an enumerated (numbered) list containing enum items.
// This is created by grouping consecutive EnumItemElements.
type EnumElement struct {
	Items []*EnumItemElement
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
