package eval

import (
	"github.com/boergens/gotypst/syntax"
)

// ----------------------------------------------------------------------------
// Introspection Elements
// ----------------------------------------------------------------------------
// These elements support document introspection and navigation.
// Reference: typst-reference/crates/typst-library/src/introspection/

// ----------------------------------------------------------------------------
// Tag Element
// ----------------------------------------------------------------------------

// Tag represents a start or end marker for content in the document.
// Tags are used to track element locations after layout for introspection.
// Matches Rust: typst-library/src/introspection/tag.rs
type Tag struct {
	// Kind is "start" or "end".
	Kind string
	// Location uniquely identifies the tagged element.
	Location TagLocation
	// Elem is the element being tagged (for start tags).
	Elem ContentElement
	// Key is a hash key for end tags.
	Key uint64
	// Flags contains tag flags.
	Flags TagFlags
}

// TagFlags contains flags for tag behavior.
type TagFlags struct {
	// Introspectable indicates the element can be found via introspection.
	Introspectable bool
	// Tagged indicates the element is semantically tagged (for accessibility).
	Tagged bool
}

// Any returns true if any flag is set.
func (f TagFlags) Any() bool {
	return f.Introspectable || f.Tagged
}

// TagLocation uniquely identifies an element in the document.
// Matches Rust: typst-library/src/introspection/location.rs
// Named TagLocation to avoid conflict with eval.Location for page positions.
type TagLocation struct {
	// Hash is the unique hash for this location.
	Hash uint64
	// Variant differentiates multiple elements with the same hash.
	Variant uint32
}

// TagElem represents a tag element in content.
// Tags mark the start and end of elements for introspection.
type TagElem struct {
	Tag Tag
}

func (*TagElem) IsContentElement() {}

// NewStartTag creates a start tag for an element.
func NewStartTag(elem ContentElement, loc TagLocation, flags TagFlags) *TagElem {
	return &TagElem{
		Tag: Tag{
			Kind:     "start",
			Location: loc,
			Elem:     elem,
			Flags:    flags,
		},
	}
}

// NewEndTag creates an end tag for an element.
func NewEndTag(loc TagLocation, key uint64, flags TagFlags) *TagElem {
	return &TagElem{
		Tag: Tag{
			Kind:     "end",
			Location: loc,
			Key:      key,
			Flags:    flags,
		},
	}
}

// ----------------------------------------------------------------------------
// Sequence Element
// ----------------------------------------------------------------------------

// SequenceElem represents a sequence of content elements.
// This is used to group multiple elements without adding structure.
// Matches Rust: typst-library/src/foundations/content.rs SequenceElem
type SequenceElem struct {
	// Children contains the sequence elements.
	Children []ContentElement
}

func (*SequenceElem) IsContentElement() {}

// NewSequence creates a new sequence element.
func NewSequence(children []ContentElement) *SequenceElem {
	return &SequenceElem{Children: children}
}

// ----------------------------------------------------------------------------
// H and V Elements (Spacing)
// ----------------------------------------------------------------------------

// HElem represents horizontal spacing.
// Matches Rust: typst-library/src/layout/spacing.rs HElem
type HElem struct {
	// Amount is the spacing amount.
	Amount Spacing
	// Weak indicates if this is weak spacing that collapses.
	Weak bool
}

func (*HElem) IsContentElement() {}

// VElem represents vertical spacing.
// Matches Rust: typst-library/src/layout/spacing.rs VElem
type VElem struct {
	// Amount is the spacing amount.
	Amount Spacing
	// Weak indicates if this is weak spacing that collapses.
	Weak bool
	// Attach indicates if this spacing attaches to previous element.
	Attach bool
}

func (*VElem) IsContentElement() {}

// Spacing represents a spacing amount (absolute or fractional).
type Spacing struct {
	// Abs is the absolute spacing in points (if not fractional).
	Abs float64
	// Fr is the fractional spacing (if fractional).
	Fr float64
	// IsFractional indicates if this is fractional spacing.
	IsFractional bool
}

// IsFractional returns true if this is fractional spacing.
func (s Spacing) IsFrac() bool {
	return s.IsFractional
}

// ----------------------------------------------------------------------------
// Page and Pagebreak Elements
// ----------------------------------------------------------------------------

// PageElem represents page configuration.
// Matches Rust: typst-library/src/layout/page.rs PageElem
type PageElem struct {
	// Paper is the paper size name (e.g., "a4", "us-letter").
	Paper *string
	// Width is the page width (if not using paper size).
	Width *float64
	// Height is the page height (if not using paper size).
	Height *float64
	// Margin is the page margin(s).
	Margin Value
	// Header is the page header content.
	Header *Content
	// Footer is the page footer content.
	Footer *Content
	// Background is the page background content.
	Background *Content
	// Foreground is the page foreground content.
	Foreground *Content
	// Fill is the page fill color.
	Fill Value
	// Numbering is the page numbering pattern.
	Numbering *string
	// NumberAlign is the alignment for page numbers.
	NumberAlign *string
	// Columns is the number of columns.
	Columns *int
	// Binding is the binding side ("left" or "right").
	Binding *string
	// Flipped indicates landscape orientation.
	Flipped bool
}

func (*PageElem) IsContentElement() {}

// PagebreakElem represents a page break.
// Matches Rust: typst-library/src/layout/page.rs PagebreakElem
type PagebreakElem struct {
	// Weak indicates a weak pagebreak that may be ignored.
	Weak bool
	// To specifies target page type ("odd" or "even").
	To *string
}

func (*PagebreakElem) IsContentElement() {}

// SharedWeakPagebreak returns a shared weak pagebreak element.
func SharedWeakPagebreak() *PagebreakElem {
	return &PagebreakElem{Weak: true}
}

// SharedBoundaryPagebreak returns a shared boundary pagebreak element.
func SharedBoundaryPagebreak() *PagebreakElem {
	return &PagebreakElem{Weak: true}
}

// ----------------------------------------------------------------------------
// Inline Element
// ----------------------------------------------------------------------------

// InlineElem wraps content as inline.
// This is used to force content to be treated as inline.
// Matches Rust: typst-library/src/layout/inline/mod.rs InlineElem
type InlineElem struct {
	// Body is the inline content.
	Body Content
}

func (*InlineElem) IsContentElement() {}

// ----------------------------------------------------------------------------
// Symbol Element
// ----------------------------------------------------------------------------

// SymbolElem represents a symbol character.
// Symbols are converted to TextElem during non-math realization.
// Matches Rust: typst-library/src/symbols/symbol.rs SymbolElem
type SymbolElem struct {
	// Text is the symbol's text representation.
	Text string
	// Label is the optional label.
	Label *string
}

func (*SymbolElem) IsContentElement() {}

// ----------------------------------------------------------------------------
// Context Element
// ----------------------------------------------------------------------------

// ContextElem provides styling context.
// Matches Rust: typst-library/src/foundations/context.rs ContextElem
type ContextElem struct {
	// Body is the content to style.
	Body Content
	// Func is an optional function to call with context.
	Func *Func
}

func (*ContextElem) IsContentElement() {}

// ----------------------------------------------------------------------------
// Document Element
// ----------------------------------------------------------------------------

// DocumentElem represents the document root.
// Matches Rust: typst-library/src/model/document.rs DocumentElem
type DocumentElem struct {
	// Title is the document title.
	Title *Content
	// Author is the document author(s).
	Author []string
	// Keywords are document keywords.
	Keywords []string
	// Date is the document date.
	Date Value
}

func (*DocumentElem) IsContentElement() {}

// ELEM constants for element type identification.
var (
	TagElemELEM      = Element{Name: "tag"}
	HElemELEM        = Element{Name: "h"}
	VElemELEM        = Element{Name: "v"}
	PageElemELEM     = Element{Name: "page"}
	PagebreakElemELEM = Element{Name: "pagebreak"}
	DocumentElemELEM = Element{Name: "document"}
	ParElemELEM      = Element{Name: "par"}
	TextElemELEM     = Element{Name: "text"}
	AlignElemELEM    = Element{Name: "align"}
)

// ----------------------------------------------------------------------------
// Element Functions
// ----------------------------------------------------------------------------

// HFunc creates the h (horizontal spacing) element function.
func HFunc() *Func {
	name := "h"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: hNative,
			Info: &FuncInfo{
				Name: "h",
				Params: []ParamInfo{
					{Name: "amount", Type: TypeLength, Named: false},
					{Name: "weak", Type: TypeBool, Default: False, Named: true},
				},
			},
		},
	}
}

// hNative implements the h() function.
func hNative(vm *Vm, args *Args) (Value, error) {
	// Get required amount argument
	amountArg := args.Find("amount")
	if amountArg == nil {
		amountArgSpanned, err := args.Expect("amount")
		if err != nil {
			return nil, err
		}
		amountArg = &amountArgSpanned
	}

	var spacing Spacing
	switch v := amountArg.V.(type) {
	case LengthValue:
		spacing = Spacing{Abs: v.Length.Points, IsFractional: false}
	case FractionValue:
		spacing = Spacing{Fr: v.Fraction.Value, IsFractional: true}
	case RelativeValue:
		spacing = Spacing{Abs: v.Relative.Abs.Points, IsFractional: false}
	default:
		return nil, &TypeMismatchError{
			Expected: "length or fraction",
			Got:      amountArg.V.Type().String(),
			Span:     amountArg.Span,
		}
	}

	// Get optional weak argument
	weak := false
	if weakArg := args.Find("weak"); weakArg != nil {
		if w, ok := AsBool(weakArg.V); ok {
			weak = w
		}
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	return ContentValue{Content: Content{
		Elements: []ContentElement{&HElem{Amount: spacing, Weak: weak}},
	}}, nil
}

// VFunc creates the v (vertical spacing) element function.
func VFunc() *Func {
	name := "v"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: vNative,
			Info: &FuncInfo{
				Name: "v",
				Params: []ParamInfo{
					{Name: "amount", Type: TypeLength, Named: false},
					{Name: "weak", Type: TypeBool, Default: False, Named: true},
				},
			},
		},
	}
}

// vNative implements the v() function.
func vNative(vm *Vm, args *Args) (Value, error) {
	// Get required amount argument
	amountArg := args.Find("amount")
	if amountArg == nil {
		amountArgSpanned, err := args.Expect("amount")
		if err != nil {
			return nil, err
		}
		amountArg = &amountArgSpanned
	}

	var spacing Spacing
	switch v := amountArg.V.(type) {
	case LengthValue:
		spacing = Spacing{Abs: v.Length.Points, IsFractional: false}
	case FractionValue:
		spacing = Spacing{Fr: v.Fraction.Value, IsFractional: true}
	case RelativeValue:
		spacing = Spacing{Abs: v.Relative.Abs.Points, IsFractional: false}
	default:
		return nil, &TypeMismatchError{
			Expected: "length or fraction",
			Got:      amountArg.V.Type().String(),
			Span:     amountArg.Span,
		}
	}

	// Get optional weak argument
	weak := false
	if weakArg := args.Find("weak"); weakArg != nil {
		if w, ok := AsBool(weakArg.V); ok {
			weak = w
		}
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	return ContentValue{Content: Content{
		Elements: []ContentElement{&VElem{Amount: spacing, Weak: weak}},
	}}, nil
}

// PagebreakFunc creates the pagebreak element function.
func PagebreakFunc() *Func {
	name := "pagebreak"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: pagebreakNative,
			Info: &FuncInfo{
				Name: "pagebreak",
				Params: []ParamInfo{
					{Name: "weak", Type: TypeBool, Default: False, Named: true},
					{Name: "to", Type: TypeStr, Default: None, Named: true},
				},
			},
		},
	}
}

// pagebreakNative implements the pagebreak() function.
func pagebreakNative(vm *Vm, args *Args) (Value, error) {
	elem := &PagebreakElem{}

	// Get optional weak argument
	if weakArg := args.Find("weak"); weakArg != nil {
		if w, ok := AsBool(weakArg.V); ok {
			elem.Weak = w
		}
	}

	// Get optional to argument
	if toArg := args.Find("to"); toArg != nil && !IsNone(toArg.V) {
		if s, ok := AsStr(toArg.V); ok {
			if s != "odd" && s != "even" {
				return nil, &InvalidArgumentError{
					Message: "to must be \"odd\" or \"even\"",
					Span:    toArg.Span,
				}
			}
			elem.To = &s
		}
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}
