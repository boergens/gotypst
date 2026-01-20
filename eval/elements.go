package eval

import (
	"github.com/boergens/gotypst/syntax"
)

// ----------------------------------------------------------------------------
// Element Functions
// ----------------------------------------------------------------------------

// This file contains element constructor functions that can be called from
// Typst code to create content elements programmatically.
// For example: raw("print('hello')", lang: "python")

// RawFunc creates the raw element function.
func RawFunc() *Func {
	name := "raw"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: rawNative,
			Info: &FuncInfo{
				Name: "raw",
				Params: []ParamInfo{
					{Name: "text", Type: TypeStr, Named: false},
					{Name: "block", Type: TypeBool, Default: False, Named: true},
					{Name: "lang", Type: TypeStr, Default: None, Named: true},
				},
			},
		},
	}
}

// rawNative implements the raw() function.
// Creates a RawElement from the given text, with optional language and block parameters.
//
// Arguments:
//   - text (positional, str): The raw text content
//   - block (named, bool, default: false): Whether this is a block-level element
//   - lang (named, str or none, default: none): The syntax highlighting language
func rawNative(vm *Vm, args *Args) (Value, error) {
	// Get required text argument (can be positional or named)
	textArg := args.Find("text")
	if textArg == nil {
		textArgSpanned, err := args.Expect("text")
		if err != nil {
			return nil, err
		}
		textArg = &textArgSpanned
	}

	text, ok := AsStr(textArg.V)
	if !ok {
		return nil, &TypeMismatchError{
			Expected: "string",
			Got:      textArg.V.Type().String(),
			Span:     textArg.Span,
		}
	}

	// Get optional block argument (default: false)
	block := false
	if blockArg := args.Find("block"); blockArg != nil {
		blockVal, ok := AsBool(blockArg.V)
		if !ok {
			return nil, &TypeMismatchError{
				Expected: "bool",
				Got:      blockArg.V.Type().String(),
				Span:     blockArg.Span,
			}
		}
		block = blockVal
	}

	// Get optional lang argument (default: none/empty string)
	lang := ""
	if langArg := args.Find("lang"); langArg != nil {
		if !IsNone(langArg.V) {
			langStr, ok := AsStr(langArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "string or none",
					Got:      langArg.V.Type().String(),
					Span:     langArg.Span,
				}
			}
			lang = langStr
		}
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Create the RawElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{&RawElement{
			Text:  text,
			Lang:  lang,
			Block: block,
		}},
	}}, nil
}

// ParFunc creates the par (paragraph) element function.
func ParFunc() *Func {
	name := "par"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: parNative,
			Info: &FuncInfo{
				Name: "par",
				Params: []ParamInfo{
					{Name: "body", Type: TypeContent, Named: false},
					{Name: "leading", Type: TypeLength, Default: Auto, Named: true},
					{Name: "justify", Type: TypeBool, Default: Auto, Named: true},
					{Name: "linebreaks", Type: TypeStr, Default: Auto, Named: true},
					{Name: "first-line-indent", Type: TypeLength, Default: None, Named: true},
					{Name: "hanging-indent", Type: TypeLength, Default: None, Named: true},
				},
			},
		},
	}
}

// parNative implements the par() function.
// Creates a ParagraphElement with optional styling properties.
//
// Arguments:
//   - body (positional, content): The paragraph content
//   - leading (named, length, default: auto): Spacing between lines
//   - justify (named, bool, default: auto): Whether to justify text
//   - linebreaks (named, str, default: auto): Line breaking algorithm ("simple" or "optimized")
//   - first-line-indent (named, length, default: none): Indent for first line
//   - hanging-indent (named, length, default: none): Indent for subsequent lines
func parNative(vm *Vm, args *Args) (Value, error) {
	// Get required body argument (positional)
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

	// Create element with defaults
	elem := &ParagraphElement{
		Body: body,
	}

	// Get optional leading argument
	if leadingArg := args.Find("leading"); leadingArg != nil {
		if !IsAuto(leadingArg.V) && !IsNone(leadingArg.V) {
			if lv, ok := leadingArg.V.(LengthValue); ok {
				leading := lv.Length.Points
				elem.Leading = &leading
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      leadingArg.V.Type().String(),
					Span:     leadingArg.Span,
				}
			}
		}
	}

	// Get optional justify argument
	if justifyArg := args.Find("justify"); justifyArg != nil {
		if !IsAuto(justifyArg.V) && !IsNone(justifyArg.V) {
			if jv, ok := AsBool(justifyArg.V); ok {
				elem.Justify = &jv
			} else {
				return nil, &TypeMismatchError{
					Expected: "bool or auto",
					Got:      justifyArg.V.Type().String(),
					Span:     justifyArg.Span,
				}
			}
		}
	}

	// Get optional linebreaks argument
	if linebreaksArg := args.Find("linebreaks"); linebreaksArg != nil {
		if !IsAuto(linebreaksArg.V) && !IsNone(linebreaksArg.V) {
			if lbs, ok := AsStr(linebreaksArg.V); ok {
				// Validate linebreaks value
				if lbs != "simple" && lbs != "optimized" {
					return nil, &TypeMismatchError{
						Expected: "\"simple\" or \"optimized\"",
						Got:      "\"" + lbs + "\"",
						Span:     linebreaksArg.Span,
					}
				}
				elem.Linebreaks = &lbs
			} else {
				return nil, &TypeMismatchError{
					Expected: "str or auto",
					Got:      linebreaksArg.V.Type().String(),
					Span:     linebreaksArg.Span,
				}
			}
		}
	}

	// Get optional first-line-indent argument
	if fliArg := args.Find("first-line-indent"); fliArg != nil {
		if !IsAuto(fliArg.V) && !IsNone(fliArg.V) {
			if fv, ok := fliArg.V.(LengthValue); ok {
				fli := fv.Length.Points
				elem.FirstLineIndent = &fli
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      fliArg.V.Type().String(),
					Span:     fliArg.Span,
				}
			}
		}
	}

	// Get optional hanging-indent argument
	if hiArg := args.Find("hanging-indent"); hiArg != nil {
		if !IsAuto(hiArg.V) && !IsNone(hiArg.V) {
			if hv, ok := hiArg.V.(LengthValue); ok {
				hi := hv.Length.Points
				elem.HangingIndent = &hi
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      hiArg.V.Type().String(),
					Span:     hiArg.Span,
				}
			}
		}
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Create the ParagraphElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}

// ParbreakFunc creates the parbreak element function.
func ParbreakFunc() *Func {
	name := "parbreak"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: parbreakNative,
			Info: &FuncInfo{
				Name:   "parbreak",
				Params: []ParamInfo{},
			},
		},
	}
}

// parbreakNative implements the parbreak() function.
// Creates a ParbreakElement to separate paragraphs.
func parbreakNative(vm *Vm, args *Args) (Value, error) {
	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Create the ParbreakElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{&ParbreakElement{}},
	}}, nil
}

// ----------------------------------------------------------------------------
// Grid Element
// ----------------------------------------------------------------------------

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
					{Name: "columns", Type: TypeArray, Default: None, Named: true},
					{Name: "rows", Type: TypeArray, Default: None, Named: true},
					{Name: "gutter", Type: TypeRelative, Default: None, Named: true},
					{Name: "column-gutter", Type: TypeRelative, Default: None, Named: true},
					{Name: "row-gutter", Type: TypeRelative, Default: None, Named: true},
					{Name: "align", Type: TypeStr, Default: Auto, Named: true},
					{Name: "children", Type: TypeContent, Named: false, Variadic: true},
				},
			},
		},
	}
}

// parseTrackSizing converts a Value to a TrackSizing.
func parseTrackSizing(v Value, span syntax.Span) (*TrackSizing, error) {
	switch val := v.(type) {
	case AutoValue:
		return &TrackSizing{Auto: true}, nil
	case LengthValue:
		pts := val.Length.Points
		return &TrackSizing{Length: &pts}, nil
	case FractionValue:
		fr := val.Fraction.Value
		return &TrackSizing{Fr: &fr}, nil
	case RatioValue:
		rel := val.Ratio.Value
		return &TrackSizing{Relative: &rel}, nil
	case RelativeValue:
		// For relative values with both abs and rel components, prefer the rel part
		// This is a simplification; full implementation would handle both
		if val.Relative.Rel.Value != 0 {
			rel := val.Relative.Rel.Value
			return &TrackSizing{Relative: &rel}, nil
		}
		pts := val.Relative.Abs.Points
		return &TrackSizing{Length: &pts}, nil
	case IntValue:
		// Allow int to be interpreted as column count (shorthand)
		// In Typst, columns: 2 means 2 auto columns
		return &TrackSizing{Auto: true}, nil
	default:
		return nil, &TypeMismatchError{
			Expected: "auto, length, fr, or relative",
			Got:      v.Type().String(),
			Span:     span,
		}
	}
}

// parseTrackSizingArray converts an array Value to a slice of TrackSizing.
func parseTrackSizingArray(v Value, span syntax.Span) ([]TrackSizing, error) {
	// Handle single value (not array)
	switch val := v.(type) {
	case ArrayValue:
		result := make([]TrackSizing, len(val))
		for i, elem := range val {
			ts, err := parseTrackSizing(elem, span)
			if err != nil {
				return nil, err
			}
			result[i] = *ts
		}
		return result, nil
	case IntValue:
		// columns: 2 means 2 auto columns
		count := int(val)
		result := make([]TrackSizing, count)
		for i := 0; i < count; i++ {
			result[i] = TrackSizing{Auto: true}
		}
		return result, nil
	default:
		// Single value - wrap in array
		ts, err := parseTrackSizing(v, span)
		if err != nil {
			return nil, err
		}
		return []TrackSizing{*ts}, nil
	}
}

// gridNative implements the grid() function.
// Creates a GridElement for grid layout.
//
// Arguments:
//   - columns (named, array, default: none): Column track sizing
//   - rows (named, array, default: none): Row track sizing
//   - gutter (named, relative, default: none): Spacing between all cells
//   - column-gutter (named, relative, default: none): Spacing between columns
//   - row-gutter (named, relative, default: none): Spacing between rows
//   - align (named, alignment, default: auto): Cell content alignment
//   - ..children (variadic, content): Grid cell contents
func gridNative(vm *Vm, args *Args) (Value, error) {
	elem := &GridElement{}

	// Get optional columns argument
	if colArg := args.Find("columns"); colArg != nil {
		if !IsNone(colArg.V) && !IsAuto(colArg.V) {
			cols, err := parseTrackSizingArray(colArg.V, colArg.Span)
			if err != nil {
				return nil, err
			}
			elem.Columns = cols
		}
	}

	// Get optional rows argument
	if rowArg := args.Find("rows"); rowArg != nil {
		if !IsNone(rowArg.V) && !IsAuto(rowArg.V) {
			rows, err := parseTrackSizingArray(rowArg.V, rowArg.Span)
			if err != nil {
				return nil, err
			}
			elem.Rows = rows
		}
	}

	// Get optional gutter argument (applies to both column and row gutter)
	if gutterArg := args.Find("gutter"); gutterArg != nil {
		if !IsNone(gutterArg.V) && !IsAuto(gutterArg.V) {
			gutter, err := parseGutterValue(gutterArg.V, gutterArg.Span)
			if err != nil {
				return nil, err
			}
			elem.ColumnGutter = &gutter
			elem.RowGutter = &gutter
		}
	}

	// Get optional column-gutter argument (overrides gutter for columns)
	if colGutterArg := args.Find("column-gutter"); colGutterArg != nil {
		if !IsNone(colGutterArg.V) && !IsAuto(colGutterArg.V) {
			gutter, err := parseGutterValue(colGutterArg.V, colGutterArg.Span)
			if err != nil {
				return nil, err
			}
			elem.ColumnGutter = &gutter
		}
	}

	// Get optional row-gutter argument (overrides gutter for rows)
	if rowGutterArg := args.Find("row-gutter"); rowGutterArg != nil {
		if !IsNone(rowGutterArg.V) && !IsAuto(rowGutterArg.V) {
			gutter, err := parseGutterValue(rowGutterArg.V, rowGutterArg.Span)
			if err != nil {
				return nil, err
			}
			elem.RowGutter = &gutter
		}
	}

	// Get optional align argument
	if alignArg := args.Find("align"); alignArg != nil {
		if !IsNone(alignArg.V) && !IsAuto(alignArg.V) {
			if alignStr, ok := AsStr(alignArg.V); ok {
				elem.Align = &alignStr
			} else {
				return nil, &TypeMismatchError{
					Expected: "alignment or auto",
					Got:      alignArg.V.Type().String(),
					Span:     alignArg.Span,
				}
			}
		}
	}

	// Collect variadic children (positional arguments)
	for {
		arg := args.Eat()
		if arg == nil {
			break // No more positional arguments
		}
		if cv, ok := arg.V.(ContentValue); ok {
			elem.Children = append(elem.Children, cv.Content)
		} else {
			// Try to convert to content
			elem.Children = append(elem.Children, valueToContent(arg.V))
		}
	}

	// Check for unexpected named arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}

// parseGutterValue converts a Value to a gutter value in points.
func parseGutterValue(v Value, span syntax.Span) (float64, error) {
	switch val := v.(type) {
	case LengthValue:
		return val.Length.Points, nil
	case RatioValue:
		// Convert ratio to points (assuming a base width, simplified)
		return val.Ratio.Value * 100, nil // Placeholder conversion
	case RelativeValue:
		// Prefer absolute part, fallback to relative
		if val.Relative.Abs.Points != 0 {
			return val.Relative.Abs.Points, nil
		}
		return val.Relative.Rel.Value * 100, nil // Placeholder conversion
	case IntValue:
		return float64(val), nil
	case FloatValue:
		return float64(val), nil
	default:
		return 0, &TypeMismatchError{
			Expected: "length or relative",
			Got:      v.Type().String(),
			Span:     span,
		}
	}
}

// ----------------------------------------------------------------------------
// Columns Element
// ----------------------------------------------------------------------------

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
					{Name: "count", Type: TypeInt, Default: Int(2), Named: true},
					{Name: "gutter", Type: TypeRelative, Default: None, Named: true},
					{Name: "body", Type: TypeContent, Named: false},
				},
			},
		},
	}
}

// columnsNative implements the columns() function.
// Creates a ColumnsElement for multi-column layout.
//
// Arguments:
//   - count (named, int, default: 2): Number of columns
//   - gutter (named, relative, default: 4%): Spacing between columns
//   - body (positional, content): Content to layout in columns
func columnsNative(vm *Vm, args *Args) (Value, error) {
	elem := &ColumnsElement{
		Count: 2, // Default to 2 columns
	}

	// Get optional count argument
	if countArg := args.Find("count"); countArg != nil {
		if !IsNone(countArg.V) && !IsAuto(countArg.V) {
			if countVal, ok := countArg.V.(IntValue); ok {
				elem.Count = int(countVal)
				if elem.Count < 1 {
					elem.Count = 1
				}
			} else {
				return nil, &TypeMismatchError{
					Expected: "int",
					Got:      countArg.V.Type().String(),
					Span:     countArg.Span,
				}
			}
		}
	}

	// Get optional gutter argument
	if gutterArg := args.Find("gutter"); gutterArg != nil {
		if !IsNone(gutterArg.V) && !IsAuto(gutterArg.V) {
			gutter, err := parseGutterValue(gutterArg.V, gutterArg.Span)
			if err != nil {
				return nil, err
			}
			elem.Gutter = &gutter
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

	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}

// ----------------------------------------------------------------------------
// Library Registration
// ----------------------------------------------------------------------------

// RegisterElementFunctions registers all element functions in the given scope.
// Call this when setting up the standard library scope.
func RegisterElementFunctions(scope *Scope) {
	// Register raw element function
	scope.DefineFunc("raw", RawFunc())
	// Register paragraph element function
	scope.DefineFunc("par", ParFunc())
	// Register parbreak element function
	scope.DefineFunc("parbreak", ParbreakFunc())
	// Register grid element function
	scope.DefineFunc("grid", GridFunc())
	// Register columns element function
	scope.DefineFunc("columns", ColumnsFunc())
}

// ElementFunctions returns a map of all element function names to their functions.
// This is useful for introspection and testing.
func ElementFunctions() map[string]*Func {
	return map[string]*Func{
		"raw":      RawFunc(),
		"par":      ParFunc(),
		"parbreak": ParbreakFunc(),
		"grid":     GridFunc(),
		"columns":  ColumnsFunc(),
	}
}
