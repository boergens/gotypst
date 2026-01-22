package eval

import (
	"github.com/boergens/gotypst/library/foundations"
	"github.com/boergens/gotypst/library/layout"
	"github.com/boergens/gotypst/syntax"
)

// AlignFunc creates the align element function.
func AlignFunc() *Func {
	name := "align"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: alignNative,
			Info: &FuncInfo{
				Name: "align",
				Params: []ParamInfo{
					{Name: "alignment", Type: TypeStr, Named: false},
					{Name: "body", Type: TypeContent, Named: false},
				},
			},
		},
	}
}

// alignNative implements the align() function.
func alignNative(engine foundations.Engine, context foundations.Context, args *Args) (Value, error) {
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

	var body Content
	if cv, ok := bodyArg.V.(ContentValue); ok {
		body = cv.Content
	} else {
		return nil, &TypeMismatchError{
			Expected: "content",
			Got:      bodyArg.V.Type().String(),
			Span:     bodyArg.Span,
		}
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	return ContentValue{Content: Content{
		Elements: []ContentElement{&layout.AlignElement{
			Alignment: alignment,
			Body:      body,
		}},
	}}, nil
}

// parseAlignment parses an alignment value from a Value.
func parseAlignment(v Value, span syntax.Span) (layout.Alignment2D, error) {
	if s, ok := AsStr(v); ok {
		return parseAlignmentString(s, span)
	}
	return layout.Alignment2D{}, &TypeMismatchError{
		Expected: "alignment",
		Got:      v.Type().String(),
		Span:     span,
	}
}

// parseAlignmentString parses an alignment from a string.
func parseAlignmentString(s string, span syntax.Span) (layout.Alignment2D, error) {
	var result layout.Alignment2D

	switch s {
	case "left":
		h := layout.HAlignLeft
		result.Horizontal = &h
	case "center":
		h := layout.HAlignCenter
		result.Horizontal = &h
	case "right":
		h := layout.HAlignRight
		result.Horizontal = &h
	case "top":
		v := layout.VAlignTop
		result.Vertical = &v
	case "horizon":
		v := layout.VAlignHorizon
		result.Vertical = &v
	case "bottom":
		v := layout.VAlignBottom
		result.Vertical = &v
	case "start":
		h := layout.HAlignStart
		result.Horizontal = &h
	case "end":
		h := layout.HAlignEnd
		result.Horizontal = &h
	default:
		return layout.Alignment2D{}, &TypeMismatchError{
			Expected: "\"left\", \"center\", \"right\", \"top\", \"horizon\", \"bottom\", \"start\", or \"end\"",
			Got:      "\"" + s + "\"",
			Span:     span,
		}
	}

	return result, nil
}
