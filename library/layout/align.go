package layout

import (
	"github.com/boergens/gotypst/library/foundations"
	"github.com/boergens/gotypst/syntax"
)

// Alignment2D represents a 2D alignment value (horizontal and vertical).
// Reference: typst-reference/crates/typst-library/src/layout/align.rs
type Alignment2D struct {
	// Horizontal alignment (left, center, right, start, end, or nil for not specified).
	Horizontal *HAlignment
	// Vertical alignment (top, horizon, bottom, or nil for not specified).
	Vertical *VAlignment
}

// HAlignment represents horizontal alignment values.
type HAlignment string

const (
	HAlignStart  HAlignment = "start"
	HAlignLeft   HAlignment = "left"
	HAlignCenter HAlignment = "center"
	HAlignRight  HAlignment = "right"
	HAlignEnd    HAlignment = "end"
)

// VAlignment represents vertical alignment values.
type VAlignment string

const (
	VAlignTop     VAlignment = "top"
	VAlignHorizon VAlignment = "horizon"
	VAlignBottom  VAlignment = "bottom"
)

// AlignElement represents an alignment container element.
// It positions its content according to the specified alignment.
//
// Reference: typst-reference/crates/typst-library/src/layout/align.rs
type AlignElement struct {
	// AlignmentStr is the raw alignment string for declarative parsing.
	AlignmentStr string `typst:"alignment,positional,required,type=str"`
	// Body is the content to align.
	Body foundations.Content `typst:"body,positional,required,type=content"`
}

func (*AlignElement) IsContentElement() {}

// AlignDef is the registered element definition for align.
var AlignDef *foundations.ElementDef

func init() {
	AlignDef = foundations.RegisterElement[AlignElement]("align", nil)
}

// Alignment returns the parsed 2D alignment from the string.
func (a *AlignElement) Alignment() Alignment2D {
	result, _ := parseAlignmentString(a.AlignmentStr, syntax.Detached())
	return result
}

// AlignFunc creates the align element function.
func AlignFunc() *foundations.Func {
	name := "align"
	return &foundations.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: foundations.NativeFunc{
			Func: alignNative,
			Info: AlignDef.ToFuncInfo(),
		},
	}
}

// alignNative implements the align() function using the generic element parser.
func alignNative(engine foundations.Engine, context foundations.Context, args *foundations.Args) (foundations.Value, error) {
	elem, err := foundations.ParseElement[AlignElement](AlignDef, args)
	if err != nil {
		return nil, err
	}

	// Validate alignment string
	if _, err := parseAlignmentString(elem.AlignmentStr, args.Span); err != nil {
		return nil, err
	}

	return foundations.ContentValue{Content: foundations.Content{
		Elements: []foundations.ContentElement{elem},
	}}, nil
}

// parseAlignmentString parses an alignment from a string.
func parseAlignmentString(s string, span syntax.Span) (Alignment2D, error) {
	var result Alignment2D

	switch s {
	case "left":
		h := HAlignLeft
		result.Horizontal = &h
	case "center":
		h := HAlignCenter
		result.Horizontal = &h
	case "right":
		h := HAlignRight
		result.Horizontal = &h
	case "top":
		v := VAlignTop
		result.Vertical = &v
	case "horizon":
		v := VAlignHorizon
		result.Vertical = &v
	case "bottom":
		v := VAlignBottom
		result.Vertical = &v
	case "start":
		h := HAlignStart
		result.Horizontal = &h
	case "end":
		h := HAlignEnd
		result.Horizontal = &h
	default:
		return Alignment2D{}, &foundations.TypeMismatchError{
			Expected: "\"left\", \"center\", \"right\", \"top\", \"horizon\", \"bottom\", \"start\", or \"end\"",
			Got:      "\"" + s + "\"",
			Span:     span,
		}
	}

	return result, nil
}
