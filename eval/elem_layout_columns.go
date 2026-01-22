package eval

import (
	"github.com/boergens/gotypst/library/foundations"
	"github.com/boergens/gotypst/library/layout"
	"github.com/boergens/gotypst/syntax"
)

// Re-export ColumnsElement for backwards compatibility.
type ColumnsElement = layout.ColumnsElement

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
func columnsNative(engine foundations.Engine, context foundations.Context, args *Args) (Value, error) {
	count := 2
	countArg := args.Find("count")
	if countArg == nil {
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

	var gutter *float64
	if gutterArg := args.Find("gutter"); gutterArg != nil {
		if !IsNone(gutterArg.V) && !IsAuto(gutterArg.V) {
			switch g := gutterArg.V.(type) {
			case LengthValue:
				gutter = &g.Length.Points
			case RelativeValue:
				gutter = &g.Relative.Abs.Points
			case RatioValue:
				pts := g.Ratio.Value * 100
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

	elem := &layout.ColumnsElement{
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
