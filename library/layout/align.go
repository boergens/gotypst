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
	// Alignment is the 2D alignment specification.
	Alignment Alignment2D
	// Body is the content to align.
	Body foundations.Content
}

func (*AlignElement) IsContentElement() {}

// AlignFunc creates the align element function.
func AlignFunc() *foundations.Func {
	name := "align"
	return &foundations.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: foundations.NativeFunc{
			Func: alignNative,
			Info: &foundations.FuncInfo{
				Name: "align",
				Params: []foundations.ParamInfo{
					{Name: "alignment", Type: foundations.TypeStr, Named: false},
					{Name: "body", Type: foundations.TypeContent, Named: false},
				},
			},
		},
	}
}

// alignNative implements the align() function.
func alignNative(engine foundations.Engine, context foundations.Context, args *foundations.Args) (foundations.Value, error) {
	alignArg, err := args.Expect("alignment")
	if err != nil {
		return nil, err
	}

	alignment, err := parseAlignment(alignArg.V, alignArg.Span)
	if err != nil {
		return nil, err
	}

	bodyArg, err := args.Expect("body")
	if err != nil {
		return nil, err
	}

	var body foundations.Content
	if cv, ok := bodyArg.V.(foundations.ContentValue); ok {
		body = cv.Content
	} else {
		return nil, &foundations.TypeMismatchError{
			Expected: "content",
			Got:      bodyArg.V.Type().String(),
			Span:     bodyArg.Span,
		}
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	return foundations.ContentValue{Content: foundations.Content{
		Elements: []foundations.ContentElement{&AlignElement{
			Alignment: alignment,
			Body:      body,
		}},
	}}, nil
}

// parseAlignment parses an alignment value from a Value.
func parseAlignment(v foundations.Value, span syntax.Span) (Alignment2D, error) {
	if s, ok := foundations.AsStr(v); ok {
		return parseAlignmentString(s, span)
	}
	return Alignment2D{}, &foundations.TypeMismatchError{
		Expected: "alignment",
		Got:      v.Type().String(),
		Span:     span,
	}
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
