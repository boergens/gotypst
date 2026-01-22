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
	Dir string `typst:"dir,type=str,default=ttb"`
	// Spacing is the spacing between children.
	Spacing *foundations.Length `typst:"spacing,type=length"`
	// Children contains the content elements to stack.
	Children []foundations.Content `typst:"children,positional,variadic"`
}

func (*StackElement) IsContentElement() {}

// StackDef is the registered element definition for stack.
var StackDef *foundations.ElementDef

func init() {
	StackDef = foundations.RegisterElement[StackElement]("stack", nil)
}

// Direction returns the stack direction as a typed enum.
func (s *StackElement) Direction() StackDirection {
	switch s.Dir {
	case "ltr":
		return StackLTR
	case "rtl":
		return StackRTL
	case "btt":
		return StackBTT
	default:
		return StackTTB
	}
}

// SpacingPts returns the spacing in points, or 0 if not set.
func (s *StackElement) SpacingPts() float64 {
	if s.Spacing == nil {
		return 0
	}
	return s.Spacing.Points
}

// StackFunc creates the stack element function.
func StackFunc() *foundations.Func {
	name := "stack"
	return &foundations.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: foundations.NativeFunc{
			Func: stackNative,
			Info: StackDef.ToFuncInfo(),
		},
	}
}

// stackNative implements the stack() function using the generic element parser.
func stackNative(engine foundations.Engine, context foundations.Context, args *foundations.Args) (foundations.Value, error) {
	elem, err := foundations.ParseElement[StackElement](StackDef, args)
	if err != nil {
		return nil, err
	}

	// Validate direction
	switch elem.Dir {
	case "", "ltr", "rtl", "ttb", "btt":
		// Valid
		if elem.Dir == "" {
			elem.Dir = "ttb" // Apply default
		}
	default:
		return nil, &foundations.TypeMismatchError{
			Expected: "\"ltr\", \"rtl\", \"ttb\", or \"btt\"",
			Got:      "\"" + elem.Dir + "\"",
			Span:     args.Span,
		}
	}

	return foundations.ContentValue{Content: foundations.Content{
		Elements: []foundations.ContentElement{elem},
	}}, nil
}
