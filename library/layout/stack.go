package layout

import (
	"github.com/boergens/gotypst/library/foundations"
	"github.com/boergens/gotypst/syntax"
)

// StackDirection represents the direction for stack layout.
type StackDirection string

const (
	// StackLTR arranges children from left to right.
	StackLTR StackDirection = "ltr"
	// StackRTL arranges children from right to left.
	StackRTL StackDirection = "rtl"
	// StackTTB arranges children from top to bottom.
	StackTTB StackDirection = "ttb"
	// StackBTT arranges children from bottom to top.
	StackBTT StackDirection = "btt"
)

// StackElement represents a stack layout element.
// It arranges its children along an axis with optional spacing.
//
// Reference: typst-reference/crates/typst-library/src/layout/stack.rs
type StackElement struct {
	// Dir is the stacking direction (ltr, rtl, ttb, btt).
	Dir StackDirection
	// Spacing is the spacing between children (in points).
	// If nil, uses default spacing (0pt).
	Spacing *float64
	// Children contains the content elements to stack.
	Children []foundations.Content
}

func (*StackElement) IsContentElement() {}

// StackFunc creates the stack element function.
func StackFunc() *foundations.Func {
	name := "stack"
	return &foundations.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: foundations.NativeFunc{
			Func: stackNative,
			Info: &foundations.FuncInfo{
				Name: "stack",
				Params: []foundations.ParamInfo{
					{Name: "dir", Type: foundations.TypeStr, Default: foundations.Str("ttb"), Named: true},
					{Name: "spacing", Type: foundations.TypeLength, Default: foundations.None, Named: true},
					{Name: "children", Type: foundations.TypeContent, Named: false, Variadic: true},
				},
			},
		},
	}
}

// stackNative implements the stack() function.
func stackNative(engine foundations.Engine, context foundations.Context, args *foundations.Args) (foundations.Value, error) {
	dir := StackTTB
	if dirArg := args.Find("dir"); dirArg != nil {
		if !foundations.IsNone(dirArg.V) && !foundations.IsAuto(dirArg.V) {
			dirStr, ok := foundations.AsStr(dirArg.V)
			if !ok {
				return nil, &foundations.TypeMismatchError{
					Expected: "string",
					Got:      dirArg.V.Type().String(),
					Span:     dirArg.Span,
				}
			}
			switch dirStr {
			case "ltr":
				dir = StackLTR
			case "rtl":
				dir = StackRTL
			case "ttb":
				dir = StackTTB
			case "btt":
				dir = StackBTT
			default:
				return nil, &foundations.TypeMismatchError{
					Expected: "\"ltr\", \"rtl\", \"ttb\", or \"btt\"",
					Got:      "\"" + dirStr + "\"",
					Span:     dirArg.Span,
				}
			}
		}
	}

	var spacing *float64
	if spacingArg := args.Find("spacing"); spacingArg != nil {
		if !foundations.IsNone(spacingArg.V) && !foundations.IsAuto(spacingArg.V) {
			if lv, ok := spacingArg.V.(foundations.LengthValue); ok {
				s := lv.Length.Points
				spacing = &s
			} else {
				return nil, &foundations.TypeMismatchError{
					Expected: "length or none",
					Got:      spacingArg.V.Type().String(),
					Span:     spacingArg.Span,
				}
			}
		}
	}

	var children []foundations.Content
	for {
		childArg := args.Eat()
		if childArg == nil {
			break
		}
		if cv, ok := childArg.V.(foundations.ContentValue); ok {
			children = append(children, cv.Content)
		} else {
			return nil, &foundations.TypeMismatchError{
				Expected: "content",
				Got:      childArg.V.Type().String(),
				Span:     childArg.Span,
			}
		}
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	return foundations.ContentValue{Content: foundations.Content{
		Elements: []foundations.ContentElement{&StackElement{
			Dir:      dir,
			Spacing:  spacing,
			Children: children,
		}},
	}}, nil
}
