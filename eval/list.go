package eval

// ----------------------------------------------------------------------------
// List Container Elements
// ----------------------------------------------------------------------------

// ListElement represents a bullet list containing list items.
// This is created by grouping consecutive ListItemElements or by the list() function.
// Matches Rust: typst-library/src/model/list.rs
type ListElement struct {
	// Items contains the list items.
	// Corresponds to Rust's `children: Vec<Packed<ListItem>>`.
	Items []*ListItemElement

	// Tight indicates whether items should have tight spacing.
	// If true, uses paragraph leading; if false, uses paragraph spacing.
	// Corresponds to Rust's `tight: bool` (default: true).
	Tight *bool

	// Marker is the content to use as the list marker.
	// Can be a single content, array of contents for nested levels, or a function.
	// Corresponds to Rust's `marker: ListMarker` (default: bullet chars).
	Marker *Content

	// Indent is the indentation of each item.
	// Corresponds to Rust's `indent: Length`.
	Indent *float64

	// BodyIndent is the spacing between the marker and the body of each item.
	// Corresponds to Rust's `body_indent: Length` (default: 0.5em).
	BodyIndent *float64

	// Spacing is the spacing between items.
	// If nil (auto), uses paragraph leading for tight, paragraph spacing for wide.
	// Corresponds to Rust's `spacing: Smart<Length>`.
	Spacing *float64

	// Depth is the nesting depth (internal, for marker cycling).
	// Corresponds to Rust's `depth: Depth` (internal).
	Depth int
}

func (*ListElement) IsContentElement() {}

// EnumElement represents an enumerated (numbered) list containing enum items.
// This is created by grouping consecutive EnumItemElements or by the enum() function.
// Matches Rust: typst-library/src/model/enum.rs
type EnumElement struct {
	// Items contains the enumerated items.
	// Corresponds to Rust's `children: Vec<Packed<EnumItem>>`.
	Items []*EnumItemElement

	// Tight indicates whether items should have tight spacing.
	// If true, uses paragraph leading; if false, uses paragraph spacing.
	// Corresponds to Rust's `tight: bool` (default: true).
	Tight *bool

	// Numbering is the numbering pattern (e.g., "1.", "a)", "I.").
	// Corresponds to Rust's `numbering: Numbering` (default: "1.").
	Numbering *string

	// Start is the starting number for the enumeration.
	// Corresponds to Rust's `start: Smart<u64>`.
	Start *int

	// Full indicates whether to display full numbering (e.g., "1.1.1" vs "1").
	// Corresponds to Rust's `full: bool` (default: false).
	Full *bool

	// Reversed indicates whether to reverse the numbering.
	// Corresponds to Rust's `reversed: bool` (default: false).
	Reversed *bool

	// Indent is the indentation of each item.
	// Corresponds to Rust's `indent: Length`.
	Indent *float64

	// BodyIndent is the spacing between the numbering and the body of each item.
	// Corresponds to Rust's `body_indent: Length` (default: 0.5em).
	BodyIndent *float64

	// Spacing is the spacing between items.
	// If nil (auto), uses paragraph leading for tight, paragraph spacing for wide.
	// Corresponds to Rust's `spacing: Smart<Length>`.
	Spacing *float64

	// NumberAlign is the alignment that enum numbers should have.
	// Corresponds to Rust's `number_align: Alignment` (default: end + top).
	NumberAlign *string

	// Parents contains the numbers of parent items (internal, for nested numbering).
	// Corresponds to Rust's `parents: SmallVec<[u64; 4]>` (internal).
	Parents []int
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
