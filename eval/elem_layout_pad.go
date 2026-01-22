package eval

import (
	"github.com/boergens/gotypst/syntax"
)

// ----------------------------------------------------------------------------
// Pad Element
// ----------------------------------------------------------------------------
// Reference: typst-reference/crates/typst-library/src/layout/pad.rs

// PadElement represents a padding container element.
// It adds spacing around its content.
type PadElement struct {
	// Left padding (in points).
	Left *float64
	// Top padding (in points).
	Top *float64
	// Right padding (in points).
	Right *float64
	// Bottom padding (in points).
	Bottom *float64
	// Body is the content to pad.
	Body Content
}

func (*PadElement) IsContentElement() {}

// PadFunc creates the pad element function.
func PadFunc() *Func {
	name := "pad"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: padNative,
			Info: &FuncInfo{
				Name: "pad",
				Params: []ParamInfo{
					{Name: "body", Type: TypeContent, Named: false},
					{Name: "left", Type: TypeLength, Default: None, Named: true},
					{Name: "top", Type: TypeLength, Default: None, Named: true},
					{Name: "right", Type: TypeLength, Default: None, Named: true},
					{Name: "bottom", Type: TypeLength, Default: None, Named: true},
					{Name: "x", Type: TypeLength, Default: None, Named: true},
					{Name: "y", Type: TypeLength, Default: None, Named: true},
					{Name: "rest", Type: TypeLength, Default: None, Named: true},
				},
			},
		},
	}
}

// padNative implements the pad() function.
// Creates a PadElement with padding around content.
//
// Arguments:
//   - body (positional, content): The content to pad
//   - left (named, length, default: 0pt): Left padding
//   - top (named, length, default: 0pt): Top padding
//   - right (named, length, default: 0pt): Right padding
//   - bottom (named, length, default: 0pt): Bottom padding
//   - x (named, length, default: 0pt): Horizontal padding (sets left and right)
//   - y (named, length, default: 0pt): Vertical padding (sets top and bottom)
//   - rest (named, length, default: 0pt): Padding for all sides
func padNative(vm *Vm, args *Args) (Value, error) {
	elem := &PadElement{}

	// Get rest argument first (applies to all sides)
	if restArg := args.Find("rest"); restArg != nil {
		if !IsNone(restArg.V) && !IsAuto(restArg.V) {
			if lv, ok := restArg.V.(LengthValue); ok {
				r := lv.Length.Points
				elem.Left = &r
				elem.Top = &r
				elem.Right = &r
				elem.Bottom = &r
			} else {
				return nil, &TypeMismatchError{
					Expected: "length",
					Got:      restArg.V.Type().String(),
					Span:     restArg.Span,
				}
			}
		}
	}

	// Get x argument (sets left and right)
	if xArg := args.Find("x"); xArg != nil {
		if !IsNone(xArg.V) && !IsAuto(xArg.V) {
			if lv, ok := xArg.V.(LengthValue); ok {
				x := lv.Length.Points
				elem.Left = &x
				elem.Right = &x
			} else {
				return nil, &TypeMismatchError{
					Expected: "length",
					Got:      xArg.V.Type().String(),
					Span:     xArg.Span,
				}
			}
		}
	}

	// Get y argument (sets top and bottom)
	if yArg := args.Find("y"); yArg != nil {
		if !IsNone(yArg.V) && !IsAuto(yArg.V) {
			if lv, ok := yArg.V.(LengthValue); ok {
				y := lv.Length.Points
				elem.Top = &y
				elem.Bottom = &y
			} else {
				return nil, &TypeMismatchError{
					Expected: "length",
					Got:      yArg.V.Type().String(),
					Span:     yArg.Span,
				}
			}
		}
	}

	// Get individual side arguments (override shorthands)
	if leftArg := args.Find("left"); leftArg != nil {
		if !IsNone(leftArg.V) && !IsAuto(leftArg.V) {
			if lv, ok := leftArg.V.(LengthValue); ok {
				l := lv.Length.Points
				elem.Left = &l
			} else {
				return nil, &TypeMismatchError{
					Expected: "length",
					Got:      leftArg.V.Type().String(),
					Span:     leftArg.Span,
				}
			}
		}
	}

	if topArg := args.Find("top"); topArg != nil {
		if !IsNone(topArg.V) && !IsAuto(topArg.V) {
			if lv, ok := topArg.V.(LengthValue); ok {
				t := lv.Length.Points
				elem.Top = &t
			} else {
				return nil, &TypeMismatchError{
					Expected: "length",
					Got:      topArg.V.Type().String(),
					Span:     topArg.Span,
				}
			}
		}
	}

	if rightArg := args.Find("right"); rightArg != nil {
		if !IsNone(rightArg.V) && !IsAuto(rightArg.V) {
			if lv, ok := rightArg.V.(LengthValue); ok {
				r := lv.Length.Points
				elem.Right = &r
			} else {
				return nil, &TypeMismatchError{
					Expected: "length",
					Got:      rightArg.V.Type().String(),
					Span:     rightArg.Span,
				}
			}
		}
	}

	if bottomArg := args.Find("bottom"); bottomArg != nil {
		if !IsNone(bottomArg.V) && !IsAuto(bottomArg.V) {
			if lv, ok := bottomArg.V.(LengthValue); ok {
				b := lv.Length.Points
				elem.Bottom = &b
			} else {
				return nil, &TypeMismatchError{
					Expected: "length",
					Got:      bottomArg.V.Type().String(),
					Span:     bottomArg.Span,
				}
			}
		}
	}

	// Get required body argument (positional)
	bodyArg := args.Find("body")
	if bodyArg == nil {
		bodyArgSpanned, err := args.Expect("body")
		if err != nil {
			return nil, err
		}
		bodyArg = &bodyArgSpanned
	}

	if cv, ok := bodyArg.V.(ContentValue); ok {
		elem.Body = cv.Content
	} else {
		return nil, &TypeMismatchError{
			Expected: "content",
			Got:      bodyArg.V.Type().String(),
			Span:     bodyArg.Span,
		}
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Create the PadElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}
