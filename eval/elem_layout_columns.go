package eval

import (
	"github.com/boergens/gotypst/syntax"
)

// ----------------------------------------------------------------------------
// Columns Element
// ----------------------------------------------------------------------------
// Reference: typst-reference/crates/typst-library/src/layout/columns.rs

// ColumnsElement represents a multi-column layout element.
// It arranges its body content into multiple columns.
type ColumnsElement struct {
	// Count is the number of columns.
	// If nil, defaults to 2.
	Count *int
	// Gutter is the gap between columns (in points).
	// If nil, defaults to 4% of page width.
	Gutter *float64
	// Body is the content to arrange in columns.
	Body Content
}

func (*ColumnsElement) IsContentElement() {}

// ColumnsFunc creates the columns element function.
func ColumnsFunc() *Func {
	name := "columns"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: columnsNative,
			Info: &FuncInfo{
				Name: "columns",
				Params: []ParamInfo{
					{Name: "count", Type: TypeInt, Default: Int(2), Named: false},
					{Name: "gutter", Type: TypeRelative, Default: None, Named: true},
					{Name: "body", Type: TypeContent, Named: false},
				},
			},
		},
	}
}

// columnsNative implements the columns() function.
// Creates a ColumnsElement with the given column count, gutter, and body.
//
// Arguments:
//   - count (positional, int, default: 2): The number of columns
//   - gutter (named, relative, default: none): The gap between columns
//   - body (positional, content): The content to arrange in columns
func columnsNative(engine *Engine, context *Context, args *Args) (Value, error) {
	// Get optional count argument (default: 2)
	count := 2
	countArg := args.Find("count")
	if countArg == nil {
		// Try to get positional count if it's an integer
		if peeked := args.Take(); peeked != nil {
			if _, ok := AsInt(peeked.V); ok {
				countArgSpanned, _ := args.Expect("count")
				countArg = &countArgSpanned
			}
		}
	}
	if countArg != nil {
		if !IsNone(countArg.V) && !IsAuto(countArg.V) {
			countVal, ok := AsInt(countArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "integer",
					Got:      countArg.V.Type().String(),
					Span:     countArg.Span,
				}
			}
			count = int(countVal)
			if count < 1 {
				return nil, &ConstructorError{
					Message: "column count must be at least 1",
					Span:    countArg.Span,
				}
			}
		}
	}

	// Get optional gutter argument
	var gutter *float64
	if gutterArg := args.Find("gutter"); gutterArg != nil {
		if !IsNone(gutterArg.V) && !IsAuto(gutterArg.V) {
			switch g := gutterArg.V.(type) {
			case LengthValue:
				gutter = &g.Length.Points
			case RelativeValue:
				// For relative, we store the absolute part for now
				// Full relative support would need layout context
				gutter = &g.Relative.Abs.Points
			case RatioValue:
				// Convert ratio to a representative value
				// (full conversion needs layout context)
				pts := g.Ratio.Value * 100 // Scale for visibility
				gutter = &pts
			default:
				return nil, &TypeMismatchError{
					Expected: "length or relative",
					Got:      gutterArg.V.Type().String(),
					Span:     gutterArg.Span,
				}
			}
		}
	}

	// Get required body argument (positional)
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

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Create the ColumnsElement wrapped in ContentValue
	elem := &ColumnsElement{
		Gutter: gutter,
		Body:   body,
	}
	if count != 2 {
		elem.Count = &count
	}

	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}
