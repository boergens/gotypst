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
	// Count is the number of columns. Defaults to 2.
	Count *int64 `typst:"count,type=int,positional,default=2"`
	// Gutter is the gap between columns.
	// If nil, defaults to 4% of page width.
	Gutter *foundations.Relative `typst:"gutter,type=relative"`
	// Body is the content to arrange in columns.
	Body foundations.Content `typst:"body,positional,required,type=content"`
}

func (*ColumnsElement) IsContentElement() {}

// ColumnsDef is the registered element definition for columns.
var ColumnsDef *foundations.ElementDef

func init() {
	ColumnsDef = foundations.RegisterElement[ColumnsElement]("columns", nil)
}

// CountInt returns the column count as int, defaulting to 2.
func (c *ColumnsElement) CountInt() int {
	if c.Count == nil {
		return 2
	}
	return int(*c.Count)
}

// GutterPts returns the gutter in points, or 0 if not set.
func (c *ColumnsElement) GutterPts() float64 {
	if c.Gutter == nil {
		return 0
	}
	return c.Gutter.Abs.Points
}

// ColumnsFunc creates the columns element function.
func ColumnsFunc() *foundations.Func {
	name := "columns"
	return &foundations.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: foundations.NativeFunc{
			Func: columnsNative,
			Info: ColumnsDef.ToFuncInfo(),
		},
	}
}

// columnsNative implements the columns() function using the generic element parser.
func columnsNative(engine foundations.Engine, context foundations.Context, args *foundations.Args) (foundations.Value, error) {
	elem, err := foundations.ParseElement[ColumnsElement](ColumnsDef, args)
	if err != nil {
		return nil, err
	}

	// Validate count
	if elem.Count != nil && *elem.Count < 1 {
		return nil, &foundations.ConstructorError{
			Message: "column count must be at least 1",
			Span:    args.Span,
		}
	}

	return foundations.ContentValue{Content: foundations.Content{
		Elements: []foundations.ContentElement{elem},
	}}, nil
}
