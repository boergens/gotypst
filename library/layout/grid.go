package layout

import (
	"fmt"

	"github.com/boergens/gotypst/library/foundations"
	"github.com/boergens/gotypst/syntax"
)

// GridTrackSizing represents a track sizing specification.
// It can be auto, a length, a fraction, or an array of these.
//
// Reference: typst-reference/crates/typst-library/src/layout/grid/mod.rs
type GridTrackSizing struct {
	// Auto indicates auto-sized tracks.
	Auto bool
	// Length is a fixed length in points (if not Auto or Fr).
	Length *float64
	// Fr is a fractional unit (if not Auto or Length).
	Fr *float64
	// Ratio is a percentage (0.0-1.0) relative to available space.
	Ratio *float64
}

// GridElement represents a grid layout element.
// It arranges its children in a grid with configurable columns and rows.
//
// Reference: typst-reference/crates/typst-library/src/layout/grid/mod.rs
type GridElement struct {
	// Columns defines the column track sizes.
	Columns []GridTrackSizing
	// Rows defines the row track sizes.
	Rows []GridTrackSizing
	// ColumnGutter is the gap between columns (in points).
	ColumnGutter *float64
	// RowGutter is the gap between rows (in points).
	RowGutter *float64
	// Inset is the cell padding.
	Inset foundations.Value
	// Align is the cell alignment.
	Align foundations.Value
	// Fill is the cell background fill.
	Fill foundations.Value
	// Stroke is the cell stroke.
	Stroke foundations.Value
	// Children contains the grid cells.
	Children []foundations.Content
}

func (*GridElement) IsContentElement() {}

// GridDef is the registered element definition for grid.
// Note: Grid uses custom parsing due to complex track sizing and gutter shorthand.
var GridDef *foundations.ElementDef

func init() {
	// Register with manual FuncInfo since we need custom parsing
	GridDef = &foundations.ElementDef{
		Name: "grid",
		Shorthands: map[string][]string{
			"gutter": {"column-gutter", "row-gutter"},
		},
		ShorthandOrder: []string{"gutter"},
	}
}

// GridFunc creates the grid element function.
func GridFunc() *foundations.Func {
	name := "grid"
	return &foundations.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: foundations.NativeFunc{
			Func: gridNative,
			Info: &foundations.FuncInfo{
				Name: "grid",
				Params: []foundations.ParamInfo{
					{Name: "columns", Type: foundations.TypeDyn, Default: foundations.Auto, Named: true},
					{Name: "rows", Type: foundations.TypeDyn, Default: foundations.Auto, Named: true},
					{Name: "gutter", Type: foundations.TypeDyn, Default: foundations.None, Named: true},
					{Name: "column-gutter", Type: foundations.TypeDyn, Default: foundations.None, Named: true},
					{Name: "row-gutter", Type: foundations.TypeDyn, Default: foundations.None, Named: true},
					{Name: "inset", Type: foundations.TypeDyn, Default: foundations.None, Named: true},
					{Name: "align", Type: foundations.TypeDyn, Default: foundations.Auto, Named: true},
					{Name: "fill", Type: foundations.TypeDyn, Default: foundations.None, Named: true},
					{Name: "stroke", Type: foundations.TypeDyn, Default: foundations.None, Named: true},
					{Name: "children", Type: foundations.TypeContent, Named: false, Variadic: true},
				},
			},
		},
	}
}

// gridNative implements the grid() function.
// Uses custom parsing due to complex track sizing and gutter shorthand.
func gridNative(engine foundations.Engine, context foundations.Context, args *foundations.Args) (foundations.Value, error) {
	elem := &GridElement{}

	// Get optional columns argument
	if colArg := args.Find("columns"); colArg != nil {
		if !foundations.IsAuto(colArg.V) && !foundations.IsNone(colArg.V) {
			cols, err := parseGridTrackSizings(colArg.V, colArg.Span)
			if err != nil {
				return nil, err
			}
			elem.Columns = cols
		}
	}

	// Get optional rows argument
	if rowArg := args.Find("rows"); rowArg != nil {
		if !foundations.IsAuto(rowArg.V) && !foundations.IsNone(rowArg.V) {
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
		if !foundations.IsNone(gutterArg.V) && !foundations.IsAuto(gutterArg.V) {
			g, err := parseGridLength(gutterArg.V, gutterArg.Span)
			if err != nil {
				return nil, err
			}
			gutterVal = &g
		}
	}

	// Get optional column-gutter argument
	if cgArg := args.Find("column-gutter"); cgArg != nil {
		if !foundations.IsNone(cgArg.V) && !foundations.IsAuto(cgArg.V) {
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
		if !foundations.IsNone(rgArg.V) && !foundations.IsAuto(rgArg.V) {
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
		if !foundations.IsNone(insetArg.V) {
			elem.Inset = insetArg.V
		}
	}

	// Get optional align argument
	if alignArg := args.Find("align"); alignArg != nil {
		if !foundations.IsAuto(alignArg.V) && !foundations.IsNone(alignArg.V) {
			elem.Align = alignArg.V
		}
	}

	// Get optional fill argument
	if fillArg := args.Find("fill"); fillArg != nil {
		if !foundations.IsNone(fillArg.V) {
			elem.Fill = fillArg.V
		}
	}

	// Get optional stroke argument
	if strokeArg := args.Find("stroke"); strokeArg != nil {
		if !foundations.IsNone(strokeArg.V) {
			elem.Stroke = strokeArg.V
		}
	}

	// Collect remaining positional arguments as children
	for {
		childArg := args.Eat()
		if childArg == nil {
			break
		}

		if cv, ok := childArg.V.(foundations.ContentValue); ok {
			elem.Children = append(elem.Children, cv.Content)
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
		Elements: []foundations.ContentElement{elem},
	}}, nil
}

// ColumnGutterPts returns the column gutter in points, or 0 if not set.
func (g *GridElement) ColumnGutterPts() float64 {
	if g.ColumnGutter == nil {
		return 0
	}
	return *g.ColumnGutter
}

// RowGutterPts returns the row gutter in points, or 0 if not set.
func (g *GridElement) RowGutterPts() float64 {
	if g.RowGutter == nil {
		return 0
	}
	return *g.RowGutter
}

// parseGridTrackSizings parses a value into grid track sizings.
func parseGridTrackSizings(v foundations.Value, span syntax.Span) ([]GridTrackSizing, error) {
	// Handle integer (number of auto columns)
	if intVal, ok := foundations.AsInt(v); ok {
		count := int(intVal)
		if count < 1 {
			return nil, &foundations.ConstructorError{
				Message: "track count must be at least 1",
				Span:    span,
			}
		}
		result := make([]GridTrackSizing, count)
		for i := range result {
			result[i] = GridTrackSizing{Auto: true}
		}
		return result, nil
	}

	// Handle single sizing value
	if sizing, err := parseGridTrackSizing(v, span); err == nil {
		return []GridTrackSizing{sizing}, nil
	}

	// Handle array
	if arr, ok := v.(*foundations.Array); ok {
		result := make([]GridTrackSizing, 0, arr.Len())
		for i, elem := range arr.Items() {
			sizing, err := parseGridTrackSizing(elem, span)
			if err != nil {
				return nil, &foundations.ConstructorError{
					Message: fmt.Sprintf("invalid track sizing at index %d: %v", i, err),
					Span:    span,
				}
			}
			result = append(result, sizing)
		}
		return result, nil
	}

	return nil, &foundations.TypeMismatchError{
		Expected: "integer, length, fraction, ratio, or array",
		Got:      v.Type().String(),
		Span:     span,
	}
}

// parseGridTrackSizing parses a single track sizing value.
func parseGridTrackSizing(v foundations.Value, span syntax.Span) (GridTrackSizing, error) {
	if foundations.IsAuto(v) {
		return GridTrackSizing{Auto: true}, nil
	}

	if lv, ok := v.(foundations.LengthValue); ok {
		l := lv.Length.Points
		return GridTrackSizing{Length: &l}, nil
	}

	if fv, ok := v.(foundations.FractionValue); ok {
		f := fv.Fraction.Value
		return GridTrackSizing{Fr: &f}, nil
	}

	if rv, ok := v.(foundations.RatioValue); ok {
		r := rv.Ratio.Value
		return GridTrackSizing{Ratio: &r}, nil
	}

	if rv, ok := v.(foundations.RelativeValue); ok {
		l := rv.Relative.Abs.Points
		return GridTrackSizing{Length: &l}, nil
	}

	return GridTrackSizing{}, &foundations.TypeMismatchError{
		Expected: "auto, length, fraction, or ratio",
		Got:      v.Type().String(),
		Span:     span,
	}
}

// parseGridLength extracts a length value in points.
func parseGridLength(v foundations.Value, span syntax.Span) (float64, error) {
	if lv, ok := v.(foundations.LengthValue); ok {
		return lv.Length.Points, nil
	}
	if rv, ok := v.(foundations.RelativeValue); ok {
		return rv.Relative.Abs.Points, nil
	}
	return 0, &foundations.TypeMismatchError{
		Expected: "length",
		Got:      v.Type().String(),
		Span:     span,
	}
}
