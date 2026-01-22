package eval

import (
	"fmt"

	"github.com/boergens/gotypst/library/foundations"
	"github.com/boergens/gotypst/library/layout"
	"github.com/boergens/gotypst/syntax"
)

// GridFunc creates the grid element function.
func GridFunc() *Func {
	name := "grid"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: gridNative,
			Info: &FuncInfo{
				Name: "grid",
				Params: []ParamInfo{
					{Name: "columns", Type: TypeDyn, Default: Auto, Named: true},
					{Name: "rows", Type: TypeDyn, Default: Auto, Named: true},
					{Name: "gutter", Type: TypeDyn, Default: None, Named: true},
					{Name: "column-gutter", Type: TypeDyn, Default: None, Named: true},
					{Name: "row-gutter", Type: TypeDyn, Default: None, Named: true},
					{Name: "inset", Type: TypeDyn, Default: None, Named: true},
					{Name: "align", Type: TypeDyn, Default: Auto, Named: true},
					{Name: "fill", Type: TypeDyn, Default: None, Named: true},
					{Name: "stroke", Type: TypeDyn, Default: None, Named: true},
					{Name: "children", Type: TypeContent, Named: false, Variadic: true},
				},
			},
		},
	}
}

// gridNative implements the grid() function.
// Creates a GridElement with the given configuration and children.
//
// Arguments:
//   - columns (named, sizing or array, default: auto): Column track sizes
//   - rows (named, sizing or array, default: auto): Row track sizes
//   - gutter (named, length, default: none): Gap between cells (sets both)
//   - column-gutter (named, length, default: none): Gap between columns
//   - row-gutter (named, length, default: none): Gap between rows
//   - inset (named, length or sides, default: none): Cell padding
//   - align (named, alignment, default: auto): Cell alignment
//   - fill (named, color or function, default: none): Cell fill
//   - stroke (named, stroke, default: none): Cell stroke
//   - children (positional, variadic, content): Grid cells
func gridNative(engine foundations.Engine, context foundations.Context, args *Args) (Value, error) {
	elem := &layout.GridElement{}

	// Get optional columns argument
	if colArg := args.Find("columns"); colArg != nil {
		if !IsAuto(colArg.V) && !IsNone(colArg.V) {
			cols, err := parseGridTrackSizings(colArg.V, colArg.Span)
			if err != nil {
				return nil, err
			}
			elem.Columns = cols
		}
	}

	// Get optional rows argument
	if rowArg := args.Find("rows"); rowArg != nil {
		if !IsAuto(rowArg.V) && !IsNone(rowArg.V) {
			rows, err := parseGridTrackSizings(rowArg.V, rowArg.Span)
			if err != nil {
				return nil, err
			}
			elem.Rows = rows
		}
	}

	// Get optional gutter argument (sets both column and row gutter)
	var gutterVal *float64
	if gutterArg := args.Find("gutter"); gutterArg != nil {
		if !IsNone(gutterArg.V) && !IsAuto(gutterArg.V) {
			g, err := parseGridLength(gutterArg.V, gutterArg.Span)
			if err != nil {
				return nil, err
			}
			gutterVal = &g
		}
	}

	// Get optional column-gutter argument
	if cgArg := args.Find("column-gutter"); cgArg != nil {
		if !IsNone(cgArg.V) && !IsAuto(cgArg.V) {
			cg, err := parseGridLength(cgArg.V, cgArg.Span)
			if err != nil {
				return nil, err
			}
			elem.ColumnGutter = &cg
		}
	} else if gutterVal != nil {
		elem.ColumnGutter = gutterVal
	}

	// Get optional row-gutter argument
	if rgArg := args.Find("row-gutter"); rgArg != nil {
		if !IsNone(rgArg.V) && !IsAuto(rgArg.V) {
			rg, err := parseGridLength(rgArg.V, rgArg.Span)
			if err != nil {
				return nil, err
			}
			elem.RowGutter = &rg
		}
	} else if gutterVal != nil {
		elem.RowGutter = gutterVal
	}

	// Get optional inset argument
	if insetArg := args.Find("inset"); insetArg != nil {
		if !IsNone(insetArg.V) {
			elem.Inset = insetArg.V
		}
	}

	// Get optional align argument
	if alignArg := args.Find("align"); alignArg != nil {
		if !IsAuto(alignArg.V) && !IsNone(alignArg.V) {
			elem.Align = alignArg.V
		}
	}

	// Get optional fill argument
	if fillArg := args.Find("fill"); fillArg != nil {
		if !IsNone(fillArg.V) {
			elem.Fill = fillArg.V
		}
	}

	// Get optional stroke argument
	if strokeArg := args.Find("stroke"); strokeArg != nil {
		if !IsNone(strokeArg.V) {
			elem.Stroke = strokeArg.V
		}
	}

	// Collect remaining positional arguments as children
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

	// Create the GridElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}

// parseGridTrackSizings parses a value into grid track sizings.
// Accepts: int (auto count), length, fr, ratio, or array of these.
func parseGridTrackSizings(v Value, span syntax.Span) ([]layout.GridTrackSizing, error) {
	// Handle integer (number of auto columns)
	if intVal, ok := AsInt(v); ok {
		count := int(intVal)
		if count < 1 {
			return nil, &ConstructorError{
				Message: "track count must be at least 1",
				Span:    span,
			}
		}
		result := make([]layout.GridTrackSizing, count)
		for i := range result {
			result[i] = layout.GridTrackSizing{Auto: true}
		}
		return result, nil
	}

	// Handle single sizing value
	if sizing, err := parseGridTrackSizing(v, span); err == nil {
		return []layout.GridTrackSizing{sizing}, nil
	}

	// Handle array
	if arr, ok := v.(*ArrayValue); ok {
		result := make([]layout.GridTrackSizing, 0, arr.Len())
		for i, elem := range arr.Items() {
			sizing, err := parseGridTrackSizing(elem, span)
			if err != nil {
				return nil, &ConstructorError{
					Message: fmt.Sprintf("invalid track sizing at index %d: %v", i, err),
					Span:    span,
				}
			}
			result = append(result, sizing)
		}
		return result, nil
	}

	return nil, &TypeMismatchError{
		Expected: "integer, length, fraction, ratio, or array",
		Got:      v.Type().String(),
		Span:     span,
	}
}

// parseGridTrackSizing parses a single track sizing value.
func parseGridTrackSizing(v Value, span syntax.Span) (layout.GridTrackSizing, error) {
	// Check for auto
	if IsAuto(v) {
		return layout.GridTrackSizing{Auto: true}, nil
	}

	// Check for length
	if lv, ok := v.(LengthValue); ok {
		l := lv.Length.Points
		return layout.GridTrackSizing{Length: &l}, nil
	}

	// Check for fraction
	if fv, ok := v.(FractionValue); ok {
		f := fv.Fraction.Value
		return layout.GridTrackSizing{Fr: &f}, nil
	}

	// Check for ratio
	if rv, ok := v.(RatioValue); ok {
		r := rv.Ratio.Value
		return layout.GridTrackSizing{Ratio: &r}, nil
	}

	// Check for relative (combo of absolute and ratio)
	if rv, ok := v.(RelativeValue); ok {
		// Use the absolute part
		l := rv.Relative.Abs.Points
		return layout.GridTrackSizing{Length: &l}, nil
	}

	return layout.GridTrackSizing{}, &TypeMismatchError{
		Expected: "auto, length, fraction, or ratio",
		Got:      v.Type().String(),
		Span:     span,
	}
}

// parseGridLength extracts a length value in points.
func parseGridLength(v Value, span syntax.Span) (float64, error) {
	if lv, ok := v.(LengthValue); ok {
		return lv.Length.Points, nil
	}
	if rv, ok := v.(RelativeValue); ok {
		return rv.Relative.Abs.Points, nil
	}
	return 0, &TypeMismatchError{
		Expected: "length",
		Got:      v.Type().String(),
		Span:     span,
	}
}
