package layout

import (
	"github.com/boergens/gotypst/library/foundations"
	"github.com/boergens/gotypst/syntax"
)

// ColumnsElement represents a multi-column layout element.
// It arranges its body content into multiple columns.
//
// Reference: typst-reference/crates/typst-library/src/layout/columns.rs
type ColumnsElement struct {
	// Count is the number of columns.
	// If nil, defaults to 2.
	Count *int
	// Gutter is the gap between columns (in points).
	// If nil, defaults to 4% of page width.
	Gutter *float64
	// Body is the content to arrange in columns.
	Body foundations.Content
}

func (*ColumnsElement) IsContentElement() {}

// ColumnsFunc creates the columns element function.
func ColumnsFunc() *foundations.Func {
	name := "columns"
	return &foundations.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: foundations.NativeFunc{
			Func: columnsNative,
			Info: &foundations.FuncInfo{
				Name: "columns",
				Params: []foundations.ParamInfo{
					{Name: "count", Type: foundations.TypeInt, Default: foundations.Int(2), Named: false},
					{Name: "gutter", Type: foundations.TypeRelative, Default: foundations.None, Named: true},
					{Name: "body", Type: foundations.TypeContent, Named: false},
				},
			},
		},
	}
}

// columnsNative implements the columns() function.
func columnsNative(engine foundations.Engine, context foundations.Context, args *foundations.Args) (foundations.Value, error) {
	count := 2
	countArg := args.Find("count")
	if countArg == nil {
		if peeked := args.Peek(); peeked != nil {
			if _, ok := foundations.AsInt(peeked.V); ok {
				countArgSpanned, _ := args.Expect("count")
				countArg = &countArgSpanned
			}
		}
	}
	if countArg != nil {
		if !foundations.IsNone(countArg.V) && !foundations.IsAuto(countArg.V) {
			countVal, ok := foundations.AsInt(countArg.V)
			if !ok {
				return nil, &foundations.TypeMismatchError{
					Expected: "integer",
					Got:      countArg.V.Type().String(),
					Span:     countArg.Span,
				}
			}
			count = int(countVal)
			if count < 1 {
				return nil, &foundations.ConstructorError{
					Message: "column count must be at least 1",
					Span:    countArg.Span,
				}
			}
		}
	}

	var gutter *float64
	if gutterArg := args.Find("gutter"); gutterArg != nil {
		if !foundations.IsNone(gutterArg.V) && !foundations.IsAuto(gutterArg.V) {
			switch g := gutterArg.V.(type) {
			case foundations.LengthValue:
				gutter = &g.Length.Points
			case foundations.RelativeValue:
				gutter = &g.Relative.Abs.Points
			case foundations.RatioValue:
				pts := g.Ratio.Value * 100
				gutter = &pts
			default:
				return nil, &foundations.TypeMismatchError{
					Expected: "length or relative",
					Got:      gutterArg.V.Type().String(),
					Span:     gutterArg.Span,
				}
			}
		}
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

	elem := &ColumnsElement{
		Gutter: gutter,
		Body:   body,
	}
	if count != 2 {
		elem.Count = &count
	}

	return foundations.ContentValue{Content: foundations.Content{
		Elements: []foundations.ContentElement{elem},
	}}, nil
}
