package eval

import (
	"github.com/boergens/gotypst/syntax"
)

// ----------------------------------------------------------------------------
// Table Element
// ----------------------------------------------------------------------------

// TableElement represents a table with rows and columns of cells.
// Tables support automatic layout, alignment, and styling.
type TableElement struct {
	// Columns specifies the column sizing (auto, lengths, fractions, or int for count).
	// Can be an array of sizing specs or a single int for uniform auto columns.
	Columns []TableSizing

	// Rows specifies the row sizing (auto, lengths, fractions).
	// Can be an array of sizing specs or nil for automatic rows.
	Rows []TableSizing

	// Gutter is the spacing between all cells (overridden by ColumnGutter/RowGutter).
	Gutter *float64

	// ColumnGutter is the horizontal spacing between columns (in points).
	ColumnGutter *float64

	// RowGutter is the vertical spacing between rows (in points).
	RowGutter *float64

	// Align is the default alignment for cells.
	// Can be a single alignment or a function mapping (x, y) -> alignment.
	Align *Alignment2D

	// Fill is the default background fill for cells.
	// Can be a color, gradient, or function mapping (x, y) -> fill.
	Fill interface{}

	// Stroke is the default stroke for grid lines.
	Stroke *TableStroke

	// Inset is the padding inside cells (in points).
	Inset *float64

	// Children contains the table cells/content in row-major order.
	Children []Content
}

func (*TableElement) IsContentElement() {}

// TableSizing represents a column or row sizing specification.
type TableSizing struct {
	// Auto indicates the track should fit its content.
	Auto bool
	// Abs is a fixed length in points (if not auto or fractional).
	Abs float64
	// Ratio is a relative component (0.0-1.0).
	Ratio float64
	// Fr is a fractional unit (like CSS grid fr).
	Fr float64
}

// TableStroke represents stroke styling for table lines.
type TableStroke struct {
	// Paint is the stroke color.
	Paint interface{}
	// Thickness is the stroke thickness in points.
	Thickness float64
}

// ----------------------------------------------------------------------------
// Table Cell Element
// ----------------------------------------------------------------------------

// TableCellElement represents an individual cell in a table.
// Cells can span multiple columns/rows and have their own styling.
type TableCellElement struct {
	// Body is the cell's content.
	Body Content

	// X is the explicit column position (0-based, nil for auto).
	X *int

	// Y is the explicit row position (0-based, nil for auto).
	Y *int

	// Colspan is the number of columns this cell spans (default 1).
	Colspan int

	// Rowspan is the number of rows this cell spans (default 1).
	Rowspan int

	// Fill overrides the table's default fill for this cell.
	Fill interface{}

	// Align overrides the table's default alignment for this cell.
	Align *Alignment2D

	// Inset overrides the table's default inset for this cell.
	Inset *float64

	// Stroke overrides the table's default stroke for this cell's borders.
	Stroke *TableStroke

	// Breakable indicates if this cell can break across page boundaries.
	Breakable *bool
}

func (*TableCellElement) IsContentElement() {}

// ----------------------------------------------------------------------------
// Table Function
// ----------------------------------------------------------------------------

// TableFunc creates the table element function.
func TableFunc() *Func {
	name := "table"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: tableNative,
			Info: &FuncInfo{
				Name: "table",
				Params: []ParamInfo{
					{Name: "columns", Type: TypeArray, Default: Auto, Named: true},
					{Name: "rows", Type: TypeArray, Default: Auto, Named: true},
					{Name: "gutter", Type: TypeLength, Default: None, Named: true},
					{Name: "column-gutter", Type: TypeLength, Default: Auto, Named: true},
					{Name: "row-gutter", Type: TypeLength, Default: Auto, Named: true},
					{Name: "align", Type: TypeStr, Default: Auto, Named: true},
					{Name: "fill", Type: TypeColor, Default: None, Named: true},
					{Name: "stroke", Type: TypeDyn, Default: None, Named: true},
					{Name: "inset", Type: TypeLength, Default: None, Named: true},
					{Name: "children", Type: TypeContent, Named: false, Variadic: true},
				},
			},
		},
	}
}

// tableNative implements the table() function.
// Creates a TableElement with the given configuration and cells.
//
// Arguments:
//   - columns (named, array or int, default: auto): Column sizing specs
//   - rows (named, array, default: auto): Row sizing specs
//   - gutter (named, length, default: none): Spacing between cells
//   - column-gutter (named, length, default: auto): Horizontal cell spacing
//   - row-gutter (named, length, default: auto): Vertical cell spacing
//   - align (named, alignment, default: auto): Default cell alignment
//   - fill (named, color, default: none): Default cell background
//   - stroke (named, stroke or none, default: none): Grid line styling
//   - inset (named, length, default: none): Cell padding
//   - children (positional, variadic, content): Cell contents
func tableNative(vm *Vm, args *Args) (Value, error) {
	elem := &TableElement{}

	// Parse columns argument
	if colsArg := args.Find("columns"); colsArg != nil {
		if !IsAuto(colsArg.V) && !IsNone(colsArg.V) {
			cols, err := parseTableSizing(colsArg.V, colsArg.Span)
			if err != nil {
				return nil, err
			}
			elem.Columns = cols
		}
	}

	// Parse rows argument
	if rowsArg := args.Find("rows"); rowsArg != nil {
		if !IsAuto(rowsArg.V) && !IsNone(rowsArg.V) {
			rows, err := parseTableSizing(rowsArg.V, rowsArg.Span)
			if err != nil {
				return nil, err
			}
			elem.Rows = rows
		}
	}

	// Parse gutter argument
	if gutterArg := args.Find("gutter"); gutterArg != nil {
		if !IsNone(gutterArg.V) && !IsAuto(gutterArg.V) {
			if lv, ok := gutterArg.V.(LengthValue); ok {
				gutter := lv.Length.Points
				elem.Gutter = &gutter
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      gutterArg.V.Type().String(),
					Span:     gutterArg.Span,
				}
			}
		}
	}

	// Parse column-gutter argument
	if cgArg := args.Find("column-gutter"); cgArg != nil {
		if !IsNone(cgArg.V) && !IsAuto(cgArg.V) {
			if lv, ok := cgArg.V.(LengthValue); ok {
				cg := lv.Length.Points
				elem.ColumnGutter = &cg
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      cgArg.V.Type().String(),
					Span:     cgArg.Span,
				}
			}
		}
	}

	// Parse row-gutter argument
	if rgArg := args.Find("row-gutter"); rgArg != nil {
		if !IsNone(rgArg.V) && !IsAuto(rgArg.V) {
			if lv, ok := rgArg.V.(LengthValue); ok {
				rg := lv.Length.Points
				elem.RowGutter = &rg
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      rgArg.V.Type().String(),
					Span:     rgArg.Span,
				}
			}
		}
	}

	// Parse align argument
	if alignArg := args.Find("align"); alignArg != nil {
		if !IsNone(alignArg.V) && !IsAuto(alignArg.V) {
			alignment, err := parseAlignment(alignArg.V, alignArg.Span)
			if err != nil {
				return nil, err
			}
			elem.Align = &alignment
		}
	}

	// Parse fill argument
	if fillArg := args.Find("fill"); fillArg != nil {
		if !IsNone(fillArg.V) {
			elem.Fill = fillArg.V
		}
	}

	// Parse stroke argument
	if strokeArg := args.Find("stroke"); strokeArg != nil {
		if !IsNone(strokeArg.V) {
			stroke, err := parseTableStroke(strokeArg.V, strokeArg.Span)
			if err != nil {
				return nil, err
			}
			elem.Stroke = stroke
		}
	}

	// Parse inset argument
	if insetArg := args.Find("inset"); insetArg != nil {
		if !IsNone(insetArg.V) && !IsAuto(insetArg.V) {
			if lv, ok := insetArg.V.(LengthValue); ok {
				inset := lv.Length.Points
				elem.Inset = &inset
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      insetArg.V.Type().String(),
					Span:     insetArg.Span,
				}
			}
		}
	}

	// Collect remaining positional arguments as children (cell contents)
	for {
		childArg := args.Eat()
		if childArg == nil {
			break
		}

		if cv, ok := childArg.V.(ContentValue); ok {
			elem.Children = append(elem.Children, cv.Content)
		} else {
			return nil, &TypeMismatchError{
				Expected: "content",
				Got:      childArg.V.Type().String(),
				Span:     childArg.Span,
			}
		}
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Create the TableElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}

// ----------------------------------------------------------------------------
// Table Cell Function
// ----------------------------------------------------------------------------

// TableCellFunc creates the table.cell element function.
func TableCellFunc() *Func {
	name := "cell"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: tableCellNative,
			Info: &FuncInfo{
				Name: "cell",
				Params: []ParamInfo{
					{Name: "body", Type: TypeContent, Named: false},
					{Name: "x", Type: TypeInt, Default: Auto, Named: true},
					{Name: "y", Type: TypeInt, Default: Auto, Named: true},
					{Name: "colspan", Type: TypeInt, Default: IntValue(1), Named: true},
					{Name: "rowspan", Type: TypeInt, Default: IntValue(1), Named: true},
					{Name: "fill", Type: TypeColor, Default: Auto, Named: true},
					{Name: "align", Type: TypeStr, Default: Auto, Named: true},
					{Name: "inset", Type: TypeLength, Default: Auto, Named: true},
					{Name: "stroke", Type: TypeDyn, Default: Auto, Named: true},
					{Name: "breakable", Type: TypeBool, Default: Auto, Named: true},
				},
			},
		},
	}
}

// tableCellNative implements the table.cell() function.
// Creates a TableCellElement with the given configuration.
//
// Arguments:
//   - body (positional, content): The cell's content
//   - x (named, int, default: auto): Explicit column position
//   - y (named, int, default: auto): Explicit row position
//   - colspan (named, int, default: 1): Number of columns to span
//   - rowspan (named, int, default: 1): Number of rows to span
//   - fill (named, color, default: auto): Cell background override
//   - align (named, alignment, default: auto): Cell alignment override
//   - inset (named, length, default: auto): Cell padding override
//   - stroke (named, stroke, default: auto): Cell border override
//   - breakable (named, bool, default: auto): Can break across pages
func tableCellNative(vm *Vm, args *Args) (Value, error) {
	// Get required body argument
	bodyArg := args.Find("body")
	if bodyArg == nil {
		bodyArgSpanned, err := args.Expect("body")
		if err != nil {
			return nil, err
		}
		bodyArg = &bodyArgSpanned
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

	elem := &TableCellElement{
		Body:    body,
		Colspan: 1,
		Rowspan: 1,
	}

	// Parse x argument (explicit column position)
	if xArg := args.Find("x"); xArg != nil {
		if !IsAuto(xArg.V) && !IsNone(xArg.V) {
			if iv, ok := xArg.V.(IntValue); ok {
				x := int(iv)
				elem.X = &x
			} else {
				return nil, &TypeMismatchError{
					Expected: "int or auto",
					Got:      xArg.V.Type().String(),
					Span:     xArg.Span,
				}
			}
		}
	}

	// Parse y argument (explicit row position)
	if yArg := args.Find("y"); yArg != nil {
		if !IsAuto(yArg.V) && !IsNone(yArg.V) {
			if iv, ok := yArg.V.(IntValue); ok {
				y := int(iv)
				elem.Y = &y
			} else {
				return nil, &TypeMismatchError{
					Expected: "int or auto",
					Got:      yArg.V.Type().String(),
					Span:     yArg.Span,
				}
			}
		}
	}

	// Parse colspan argument
	if colspanArg := args.Find("colspan"); colspanArg != nil {
		if !IsAuto(colspanArg.V) && !IsNone(colspanArg.V) {
			if iv, ok := colspanArg.V.(IntValue); ok {
				elem.Colspan = int(iv)
				if elem.Colspan < 1 {
					elem.Colspan = 1
				}
			} else {
				return nil, &TypeMismatchError{
					Expected: "int",
					Got:      colspanArg.V.Type().String(),
					Span:     colspanArg.Span,
				}
			}
		}
	}

	// Parse rowspan argument
	if rowspanArg := args.Find("rowspan"); rowspanArg != nil {
		if !IsAuto(rowspanArg.V) && !IsNone(rowspanArg.V) {
			if iv, ok := rowspanArg.V.(IntValue); ok {
				elem.Rowspan = int(iv)
				if elem.Rowspan < 1 {
					elem.Rowspan = 1
				}
			} else {
				return nil, &TypeMismatchError{
					Expected: "int",
					Got:      rowspanArg.V.Type().String(),
					Span:     rowspanArg.Span,
				}
			}
		}
	}

	// Parse fill argument
	if fillArg := args.Find("fill"); fillArg != nil {
		if !IsAuto(fillArg.V) && !IsNone(fillArg.V) {
			elem.Fill = fillArg.V
		}
	}

	// Parse align argument
	if alignArg := args.Find("align"); alignArg != nil {
		if !IsAuto(alignArg.V) && !IsNone(alignArg.V) {
			alignment, err := parseAlignment(alignArg.V, alignArg.Span)
			if err != nil {
				return nil, err
			}
			elem.Align = &alignment
		}
	}

	// Parse inset argument
	if insetArg := args.Find("inset"); insetArg != nil {
		if !IsAuto(insetArg.V) && !IsNone(insetArg.V) {
			if lv, ok := insetArg.V.(LengthValue); ok {
				inset := lv.Length.Points
				elem.Inset = &inset
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      insetArg.V.Type().String(),
					Span:     insetArg.Span,
				}
			}
		}
	}

	// Parse stroke argument
	if strokeArg := args.Find("stroke"); strokeArg != nil {
		if !IsAuto(strokeArg.V) && !IsNone(strokeArg.V) {
			stroke, err := parseTableStroke(strokeArg.V, strokeArg.Span)
			if err != nil {
				return nil, err
			}
			elem.Stroke = stroke
		}
	}

	// Parse breakable argument
	if breakableArg := args.Find("breakable"); breakableArg != nil {
		if !IsAuto(breakableArg.V) && !IsNone(breakableArg.V) {
			if bv, ok := AsBool(breakableArg.V); ok {
				elem.Breakable = &bv
			} else {
				return nil, &TypeMismatchError{
					Expected: "bool or auto",
					Got:      breakableArg.V.Type().String(),
					Span:     breakableArg.Span,
				}
			}
		}
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Create the TableCellElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}

// ----------------------------------------------------------------------------
// Helper Functions
// ----------------------------------------------------------------------------

// parseTableSizing parses sizing specifications for columns or rows.
// Supports: int (count), array of sizing specs, auto, lengths, fractions.
func parseTableSizing(v Value, span syntax.Span) ([]TableSizing, error) {
	// Single int means that many auto columns
	if iv, ok := v.(IntValue); ok {
		count := int(iv)
		if count < 1 {
			count = 1
		}
		result := make([]TableSizing, count)
		for i := range result {
			result[i] = TableSizing{Auto: true}
		}
		return result, nil
	}

	// Array of sizing specs
	if arr, ok := AsArray(v); ok {
		result := make([]TableSizing, len(arr))
		for i, item := range arr {
			sizing, err := parseSingleSizing(item, span)
			if err != nil {
				return nil, err
			}
			result[i] = sizing
		}
		return result, nil
	}

	// Single value (treat as one-element array)
	sizing, err := parseSingleSizing(v, span)
	if err != nil {
		return nil, err
	}
	return []TableSizing{sizing}, nil
}

// parseSingleSizing parses a single sizing value.
func parseSingleSizing(v Value, span syntax.Span) (TableSizing, error) {
	// Auto
	if IsAuto(v) {
		return TableSizing{Auto: true}, nil
	}

	// Length value (absolute or relative)
	if lv, ok := v.(LengthValue); ok {
		return TableSizing{Abs: lv.Length.Points}, nil
	}

	// Ratio value (percentage)
	if rv, ok := v.(RatioValue); ok {
		return TableSizing{Ratio: rv.Ratio.Value}, nil
	}

	// Fraction value (fr units)
	if fv, ok := v.(FractionValue); ok {
		return TableSizing{Fr: fv.Fraction.Value}, nil
	}

	// Relative value (length + ratio)
	if rv, ok := v.(RelativeValue); ok {
		return TableSizing{
			Abs:   rv.Relative.Abs.Points,
			Ratio: rv.Relative.Rel.Value,
		}, nil
	}

	return TableSizing{Auto: true}, nil
}

// parseTableStroke parses a stroke specification for table lines.
func parseTableStroke(v Value, span syntax.Span) (*TableStroke, error) {
	// Color value - treat as stroke paint with default thickness
	if cv, ok := v.(ColorValue); ok {
		return &TableStroke{
			Paint:     cv,
			Thickness: 1.0, // default 1pt
		}, nil
	}

	// Length value - treat as stroke thickness with default color
	if lv, ok := v.(LengthValue); ok {
		return &TableStroke{
			Paint:     nil, // default black
			Thickness: lv.Length.Points,
		}, nil
	}

	// Dictionary with paint and/or thickness
	if dv, ok := v.(DictValue); ok {
		stroke := &TableStroke{Thickness: 1.0}

		if paint, ok := dv.Get("paint"); ok {
			stroke.Paint = paint
		}
		if thickness, ok := dv.Get("thickness"); ok {
			if lv, ok := thickness.(LengthValue); ok {
				stroke.Thickness = lv.Length.Points
			}
		}

		return stroke, nil
	}

	return &TableStroke{Thickness: 1.0}, nil
}

// ----------------------------------------------------------------------------
// Table Header and Footer Elements
// ----------------------------------------------------------------------------

// TableHeaderElement represents a repeating table header.
type TableHeaderElement struct {
	// Children contains the header cells.
	Children []Content
	// Repeat indicates whether the header repeats on each page.
	Repeat bool
}

func (*TableHeaderElement) IsContentElement() {}

// TableFooterElement represents a table footer.
type TableFooterElement struct {
	// Children contains the footer cells.
	Children []Content
	// Repeat indicates whether the footer repeats on each page.
	Repeat bool
}

func (*TableFooterElement) IsContentElement() {}

// TableHlineElement represents a horizontal line in a table.
type TableHlineElement struct {
	// Y is the row position (before which row, 0-based).
	Y *int
	// Start is the starting column (0-based).
	Start int
	// End is the ending column (exclusive, nil for all remaining).
	End *int
	// Stroke is the line styling.
	Stroke *TableStroke
	// Position is "start" or "end" within the row.
	Position string
}

func (*TableHlineElement) IsContentElement() {}

// TableVlineElement represents a vertical line in a table.
type TableVlineElement struct {
	// X is the column position (before which column, 0-based).
	X *int
	// Start is the starting row (0-based).
	Start int
	// End is the ending row (exclusive, nil for all remaining).
	End *int
	// Stroke is the line styling.
	Stroke *TableStroke
	// Position is "start" or "end" within the column.
	Position string
}

func (*TableVlineElement) IsContentElement() {}

// ----------------------------------------------------------------------------
// Table Header Function
// ----------------------------------------------------------------------------

// TableHeaderFunc creates the table.header element function.
func TableHeaderFunc() *Func {
	name := "header"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: tableHeaderNative,
			Info: &FuncInfo{
				Name: "header",
				Params: []ParamInfo{
					{Name: "repeat", Type: TypeBool, Default: True, Named: true},
					{Name: "children", Type: TypeContent, Named: false, Variadic: true},
				},
			},
		},
	}
}

// tableHeaderNative implements the table.header() function.
func tableHeaderNative(vm *Vm, args *Args) (Value, error) {
	elem := &TableHeaderElement{
		Repeat: true, // default to repeating
	}

	// Parse repeat argument
	if repeatArg := args.Find("repeat"); repeatArg != nil {
		if !IsAuto(repeatArg.V) && !IsNone(repeatArg.V) {
			if bv, ok := AsBool(repeatArg.V); ok {
				elem.Repeat = bv
			} else {
				return nil, &TypeMismatchError{
					Expected: "bool",
					Got:      repeatArg.V.Type().String(),
					Span:     repeatArg.Span,
				}
			}
		}
	}

	// Collect children
	for {
		childArg := args.Eat()
		if childArg == nil {
			break
		}

		if cv, ok := childArg.V.(ContentValue); ok {
			elem.Children = append(elem.Children, cv.Content)
		} else {
			return nil, &TypeMismatchError{
				Expected: "content",
				Got:      childArg.V.Type().String(),
				Span:     childArg.Span,
			}
		}
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}

// ----------------------------------------------------------------------------
// Table Footer Function
// ----------------------------------------------------------------------------

// TableFooterFunc creates the table.footer element function.
func TableFooterFunc() *Func {
	name := "footer"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: tableFooterNative,
			Info: &FuncInfo{
				Name: "footer",
				Params: []ParamInfo{
					{Name: "repeat", Type: TypeBool, Default: True, Named: true},
					{Name: "children", Type: TypeContent, Named: false, Variadic: true},
				},
			},
		},
	}
}

// tableFooterNative implements the table.footer() function.
func tableFooterNative(vm *Vm, args *Args) (Value, error) {
	elem := &TableFooterElement{
		Repeat: true,
	}

	// Parse repeat argument
	if repeatArg := args.Find("repeat"); repeatArg != nil {
		if !IsAuto(repeatArg.V) && !IsNone(repeatArg.V) {
			if bv, ok := AsBool(repeatArg.V); ok {
				elem.Repeat = bv
			} else {
				return nil, &TypeMismatchError{
					Expected: "bool",
					Got:      repeatArg.V.Type().String(),
					Span:     repeatArg.Span,
				}
			}
		}
	}

	// Collect children
	for {
		childArg := args.Eat()
		if childArg == nil {
			break
		}

		if cv, ok := childArg.V.(ContentValue); ok {
			elem.Children = append(elem.Children, cv.Content)
		} else {
			return nil, &TypeMismatchError{
				Expected: "content",
				Got:      childArg.V.Type().String(),
				Span:     childArg.Span,
			}
		}
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}

// ----------------------------------------------------------------------------
// Table Hline Function
// ----------------------------------------------------------------------------

// TableHlineFunc creates the table.hline element function.
func TableHlineFunc() *Func {
	name := "hline"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: tableHlineNative,
			Info: &FuncInfo{
				Name: "hline",
				Params: []ParamInfo{
					{Name: "y", Type: TypeInt, Default: Auto, Named: true},
					{Name: "start", Type: TypeInt, Default: IntValue(0), Named: true},
					{Name: "end", Type: TypeInt, Default: None, Named: true},
					{Name: "stroke", Type: TypeDyn, Default: None, Named: true},
					{Name: "position", Type: TypeStr, Default: Str("start"), Named: true},
				},
			},
		},
	}
}

// tableHlineNative implements the table.hline() function.
func tableHlineNative(vm *Vm, args *Args) (Value, error) {
	elem := &TableHlineElement{
		Start:    0,
		Position: "start",
	}

	// Parse y argument
	if yArg := args.Find("y"); yArg != nil {
		if !IsAuto(yArg.V) && !IsNone(yArg.V) {
			if iv, ok := yArg.V.(IntValue); ok {
				y := int(iv)
				elem.Y = &y
			} else {
				return nil, &TypeMismatchError{
					Expected: "int or auto",
					Got:      yArg.V.Type().String(),
					Span:     yArg.Span,
				}
			}
		}
	}

	// Parse start argument
	if startArg := args.Find("start"); startArg != nil {
		if !IsNone(startArg.V) {
			if iv, ok := startArg.V.(IntValue); ok {
				elem.Start = int(iv)
			}
		}
	}

	// Parse end argument
	if endArg := args.Find("end"); endArg != nil {
		if !IsNone(endArg.V) && !IsAuto(endArg.V) {
			if iv, ok := endArg.V.(IntValue); ok {
				end := int(iv)
				elem.End = &end
			}
		}
	}

	// Parse stroke argument
	if strokeArg := args.Find("stroke"); strokeArg != nil {
		if !IsNone(strokeArg.V) {
			stroke, err := parseTableStroke(strokeArg.V, strokeArg.Span)
			if err != nil {
				return nil, err
			}
			elem.Stroke = stroke
		}
	}

	// Parse position argument
	if posArg := args.Find("position"); posArg != nil {
		if !IsNone(posArg.V) {
			if s, ok := AsStr(posArg.V); ok {
				if s != "start" && s != "end" {
					return nil, &TypeMismatchError{
						Expected: "\"start\" or \"end\"",
						Got:      "\"" + s + "\"",
						Span:     posArg.Span,
					}
				}
				elem.Position = s
			}
		}
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}

// ----------------------------------------------------------------------------
// Table Vline Function
// ----------------------------------------------------------------------------

// TableVlineFunc creates the table.vline element function.
func TableVlineFunc() *Func {
	name := "vline"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: tableVlineNative,
			Info: &FuncInfo{
				Name: "vline",
				Params: []ParamInfo{
					{Name: "x", Type: TypeInt, Default: Auto, Named: true},
					{Name: "start", Type: TypeInt, Default: IntValue(0), Named: true},
					{Name: "end", Type: TypeInt, Default: None, Named: true},
					{Name: "stroke", Type: TypeDyn, Default: None, Named: true},
					{Name: "position", Type: TypeStr, Default: Str("start"), Named: true},
				},
			},
		},
	}
}

// tableVlineNative implements the table.vline() function.
func tableVlineNative(vm *Vm, args *Args) (Value, error) {
	elem := &TableVlineElement{
		Start:    0,
		Position: "start",
	}

	// Parse x argument
	if xArg := args.Find("x"); xArg != nil {
		if !IsAuto(xArg.V) && !IsNone(xArg.V) {
			if iv, ok := xArg.V.(IntValue); ok {
				x := int(iv)
				elem.X = &x
			} else {
				return nil, &TypeMismatchError{
					Expected: "int or auto",
					Got:      xArg.V.Type().String(),
					Span:     xArg.Span,
				}
			}
		}
	}

	// Parse start argument
	if startArg := args.Find("start"); startArg != nil {
		if !IsNone(startArg.V) {
			if iv, ok := startArg.V.(IntValue); ok {
				elem.Start = int(iv)
			}
		}
	}

	// Parse end argument
	if endArg := args.Find("end"); endArg != nil {
		if !IsNone(endArg.V) && !IsAuto(endArg.V) {
			if iv, ok := endArg.V.(IntValue); ok {
				end := int(iv)
				elem.End = &end
			}
		}
	}

	// Parse stroke argument
	if strokeArg := args.Find("stroke"); strokeArg != nil {
		if !IsNone(strokeArg.V) {
			stroke, err := parseTableStroke(strokeArg.V, strokeArg.Span)
			if err != nil {
				return nil, err
			}
			elem.Stroke = stroke
		}
	}

	// Parse position argument
	if posArg := args.Find("position"); posArg != nil {
		if !IsNone(posArg.V) {
			if s, ok := AsStr(posArg.V); ok {
				if s != "start" && s != "end" {
					return nil, &TypeMismatchError{
						Expected: "\"start\" or \"end\"",
						Got:      "\"" + s + "\"",
						Span:     posArg.Span,
					}
				}
				elem.Position = s
			}
		}
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}
