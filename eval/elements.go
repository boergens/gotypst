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
					{Name: "columns", Type: TypeArray, Default: Auto, Named: true},
					{Name: "rows", Type: TypeArray, Default: Auto, Named: true},
					{Name: "gutter", Type: TypeRelative, Default: Auto, Named: true},
					{Name: "column-gutter", Type: TypeRelative, Default: Auto, Named: true},
					{Name: "row-gutter", Type: TypeRelative, Default: Auto, Named: true},
					{Name: "align", Type: TypeArray, Default: Auto, Named: true},
					{Name: "inset", Type: TypeRelative, Default: None, Named: true},
					{Name: "fill", Type: TypeColor, Default: None, Named: true},
					{Name: "stroke", Type: TypeLength, Default: None, Named: true},
					{Name: "children", Type: TypeContent, Variadic: true, Named: false},
				},
			},
		},
	}
}

// gridNative implements the grid() function.
// Creates a GridElement with the specified layout parameters.
//
// Arguments:
//   - columns (named, array or int or auto): Column track sizes
//   - rows (named, array or int or auto): Row track sizes
//   - gutter (named, relative or auto): Gutter between cells
//   - column-gutter (named, relative or auto): Column gutter override
//   - row-gutter (named, relative or auto): Row gutter override
//   - align (named, array or alignment or auto): Cell alignment
//   - inset (named, relative or dict): Cell insets
//   - fill (named, color or array or none): Cell backgrounds
//   - stroke (named, stroke or array or none): Cell borders
//   - children (variadic, content): Grid cell contents
func gridNative(vm *Vm, args *Args) (Value, error) {
	elem := &GridElement{}

	// Get optional columns argument
	if colsArg := args.Find("columns"); colsArg != nil {
		if !IsAuto(colsArg.V) && !IsNone(colsArg.V) {
			tracks, err := parseTrackSizes(colsArg)
			if err != nil {
				return nil, err
			}
			elem.Columns = tracks
		}
	}

	// Get optional rows argument
	if rowsArg := args.Find("rows"); rowsArg != nil {
		if !IsAuto(rowsArg.V) && !IsNone(rowsArg.V) {
			tracks, err := parseTrackSizes(rowsArg)
			if err != nil {
				return nil, err
			}
			elem.Rows = tracks
		}
	}

	// Get optional gutter argument (sets both column-gutter and row-gutter)
	if gutterArg := args.Find("gutter"); gutterArg != nil {
		if !IsAuto(gutterArg.V) && !IsNone(gutterArg.V) {
			gutter, err := parseGutter(gutterArg)
			if err != nil {
				return nil, err
			}
			elem.ColumnGutter = gutter
			elem.RowGutter = gutter
		}
	}

	// Get optional column-gutter argument (overrides gutter)
	if cgArg := args.Find("column-gutter"); cgArg != nil {
		if !IsAuto(cgArg.V) && !IsNone(cgArg.V) {
			gutter, err := parseGutter(cgArg)
			if err != nil {
				return nil, err
			}
			elem.ColumnGutter = gutter
		}
	}

	// Get optional row-gutter argument (overrides gutter)
	if rgArg := args.Find("row-gutter"); rgArg != nil {
		if !IsAuto(rgArg.V) && !IsNone(rgArg.V) {
			gutter, err := parseGutter(rgArg)
			if err != nil {
				return nil, err
			}
			elem.RowGutter = gutter
		}
	}

	// Get optional align argument
	if alignArg := args.Find("align"); alignArg != nil {
		if !IsAuto(alignArg.V) && !IsNone(alignArg.V) {
			align, err := parseAlignment(alignArg)
			if err != nil {
				return nil, err
			}
			elem.Align = align
		}
	}

	// Get optional inset argument
	if insetArg := args.Find("inset"); insetArg != nil {
		if !IsAuto(insetArg.V) && !IsNone(insetArg.V) {
			inset, err := parseInset(insetArg)
			if err != nil {
				return nil, err
			}
			elem.Inset = inset
		}
	}

	// Get optional fill argument
	if fillArg := args.Find("fill"); fillArg != nil {
		if !IsNone(fillArg.V) {
			fill, err := parseFill(fillArg)
			if err != nil {
				return nil, err
			}
			elem.Fill = fill
		}
	}

	// Get optional stroke argument
	if strokeArg := args.Find("stroke"); strokeArg != nil {
		if !IsNone(strokeArg.V) {
			stroke, err := parseStroke(strokeArg)
			if err != nil {
				return nil, err
			}
			elem.Stroke = stroke
		}
	}

	// Collect variadic children
	for _, child := range args.All() {
		if cv, ok := child.V.(ContentValue); ok {
			elem.Children = append(elem.Children, cv.Content)
		} else {
			// Convert non-content values to content
			elem.Children = append(elem.Children, Content{
				Elements: []ContentElement{&TextElement{Text: child.V.Display().String()}},
			})
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

// GridElement represents a grid layout with rows and columns.
type GridElement struct {
	// Columns specifies the column track sizes.
	// Each track can be auto, a length, a fraction, or a relative value.
	Columns []TrackSize
	// Rows specifies the row track sizes.
	Rows []TrackSize
	// ColumnGutter is the gutter between columns.
	ColumnGutter []TrackSize
	// RowGutter is the gutter between rows.
	RowGutter []TrackSize
	// Align specifies the alignment for cells.
	Align []Alignment
	// Inset specifies the padding inside cells.
	Inset *Inset
	// Fill specifies the background fill for cells.
	Fill []FillSpec
	// Stroke specifies the border stroke for cells.
	Stroke []StrokeSpec
	// Children contains the grid cell contents.
	Children []Content
}

func (*GridElement) isContentElement() {}

// TrackSize represents a grid track size (column or row width/height).
type TrackSize struct {
	// Auto indicates automatic sizing.
	Auto bool
	// Length is an absolute length in points.
	Length *float64
	// Fraction is a fraction of remaining space (e.g., 1fr).
	Fraction *float64
	// Relative is a relative length (percentage + absolute).
	Relative *Relative
}

// Alignment represents horizontal and/or vertical alignment.
type Alignment struct {
	// Horizontal alignment: "left", "center", "right", "start", "end".
	Horizontal string
	// Vertical alignment: "top", "horizon", "bottom".
	Vertical string
}

// Inset represents cell padding.
type Inset struct {
	// All is the inset for all sides (if uniform).
	All *float64
	// Top, Right, Bottom, Left are individual side insets.
	Top, Right, Bottom, Left *float64
}

// FillSpec represents a fill specification (color, gradient, etc.).
type FillSpec struct {
	// Color is a solid color fill.
	Color *Color
	// None indicates no fill.
	None bool
}

// StrokeSpec represents a stroke specification.
type StrokeSpec struct {
	// Thickness is the stroke width in points.
	Thickness *float64
	// Color is the stroke color.
	Color *Color
	// None indicates no stroke.
	None bool
}

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
					{Name: "count", Type: TypeInt, Default: Auto, Named: false},
					{Name: "gutter", Type: TypeRelative, Default: Auto, Named: true},
					{Name: "body", Type: TypeContent, Variadic: true, Named: false},
				},
			},
		},
	}
}

// columnsNative implements the columns() function.
// Creates a ColumnsElement for multi-column text layout.
//
// Arguments:
//   - count (positional, int or auto): Number of columns
//   - gutter (named, relative or auto): Space between columns
//   - body (variadic, content): Content to flow into columns
func columnsNative(vm *Vm, args *Args) (Value, error) {
	elem := &ColumnsElement{}

	// Get count argument (first positional or named)
	countArg := args.Find("count")
	if countArg == nil {
		// Try to get as positional
		if next := args.Take(); next != nil {
			// Check if it looks like count (int or auto) vs body (content)
			if _, isInt := next.V.(IntValue); isInt {
				countArg = args.Eat()
			} else if IsAuto(next.V) {
				countArg = args.Eat()
			}
		}
	}

	if countArg != nil {
		if !IsAuto(countArg.V) {
			if count, ok := AsInt(countArg.V); ok {
				c := int(count)
				elem.Count = &c
			} else {
				return nil, &TypeMismatchError{
					Expected: "int or auto",
					Got:      countArg.V.Type().String(),
					Span:     countArg.Span,
				}
			}
		}
	}

	// Get optional gutter argument
	if gutterArg := args.Find("gutter"); gutterArg != nil {
		if !IsAuto(gutterArg.V) && !IsNone(gutterArg.V) {
			if lv, ok := gutterArg.V.(LengthValue); ok {
				gutter := lv.Length.Points
				elem.Gutter = &gutter
			} else if rv, ok := gutterArg.V.(RelativeValue); ok {
				gutter := rv.Relative.Abs.Points
				elem.Gutter = &gutter
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or relative",
					Got:      gutterArg.V.Type().String(),
					Span:     gutterArg.Span,
				}
			}
		}
	}

	// Collect variadic body content
	var body Content
	for _, child := range args.All() {
		if cv, ok := child.V.(ContentValue); ok {
			body.Elements = append(body.Elements, cv.Content.Elements...)
		} else {
			// Convert non-content values to content
			body.Elements = append(body.Elements, &TextElement{Text: child.V.Display().String()})
		}
	}
	elem.Body = body

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}

// ColumnsElement represents multi-column text layout.
type ColumnsElement struct {
	// Count is the number of columns (nil means auto).
	Count *int
	// Gutter is the space between columns in points.
	Gutter *float64
	// Body is the content to flow into columns.
	Body Content
}

func (*ColumnsElement) isContentElement() {}

// ----------------------------------------------------------------------------
// Grid/Columns Helper Functions
// ----------------------------------------------------------------------------

// parseTrackSizes parses track sizes from an argument value.
func parseTrackSizes(arg *syntax.Spanned[Value]) ([]TrackSize, error) {
	switch v := arg.V.(type) {
	case IntValue:
		// Integer means that many auto-sized tracks
		count := int(v)
		tracks := make([]TrackSize, count)
		for i := range tracks {
			tracks[i] = TrackSize{Auto: true}
		}
		return tracks, nil
	case ArrayValue:
		tracks := make([]TrackSize, len(v))
		for i, elem := range v {
			track, err := parseTrackSize(elem, arg.Span)
			if err != nil {
				return nil, err
			}
			tracks[i] = track
		}
		return tracks, nil
	case LengthValue:
		// Single length becomes one track
		pts := v.Length.Points
		return []TrackSize{{Length: &pts}}, nil
	case FractionValue:
		// Single fraction becomes one track
		fr := v.Fraction.Value
		return []TrackSize{{Fraction: &fr}}, nil
	case AutoValue:
		return []TrackSize{{Auto: true}}, nil
	default:
		return nil, &TypeMismatchError{
			Expected: "int, array, length, fraction, or auto",
			Got:      arg.V.Type().String(),
			Span:     arg.Span,
		}
	}
}

// parseTrackSize parses a single track size value.
func parseTrackSize(v Value, span syntax.Span) (TrackSize, error) {
	switch val := v.(type) {
	case AutoValue:
		return TrackSize{Auto: true}, nil
	case LengthValue:
		pts := val.Length.Points
		return TrackSize{Length: &pts}, nil
	case FractionValue:
		fr := val.Fraction.Value
		return TrackSize{Fraction: &fr}, nil
	case RelativeValue:
		return TrackSize{Relative: &val.Relative}, nil
	case IntValue:
		// Integer in an array context means absolute points
		pts := float64(val)
		return TrackSize{Length: &pts}, nil
	default:
		return TrackSize{}, &TypeMismatchError{
			Expected: "auto, length, fraction, or relative",
			Got:      v.Type().String(),
			Span:     span,
		}
	}
}

// parseGutter parses gutter values from an argument.
func parseGutter(arg *syntax.Spanned[Value]) ([]TrackSize, error) {
	switch v := arg.V.(type) {
	case LengthValue:
		pts := v.Length.Points
		return []TrackSize{{Length: &pts}}, nil
	case FractionValue:
		fr := v.Fraction.Value
		return []TrackSize{{Fraction: &fr}}, nil
	case RelativeValue:
		return []TrackSize{{Relative: &v.Relative}}, nil
	case ArrayValue:
		tracks := make([]TrackSize, len(v))
		for i, elem := range v {
			track, err := parseTrackSize(elem, arg.Span)
			if err != nil {
				return nil, err
			}
			tracks[i] = track
		}
		return tracks, nil
	default:
		return nil, &TypeMismatchError{
			Expected: "length, fraction, relative, or array",
			Got:      arg.V.Type().String(),
			Span:     arg.Span,
		}
	}
}

// parseAlignment parses alignment from an argument.
func parseAlignment(arg *syntax.Spanned[Value]) ([]Alignment, error) {
	switch v := arg.V.(type) {
	case StrValue:
		align, err := parseAlignmentString(string(v))
		if err != nil {
			return nil, err
		}
		return []Alignment{align}, nil
	case ArrayValue:
		aligns := make([]Alignment, len(v))
		for i, elem := range v {
			if s, ok := elem.(StrValue); ok {
				align, err := parseAlignmentString(string(s))
				if err != nil {
					return nil, err
				}
				aligns[i] = align
			} else {
				return nil, &TypeMismatchError{
					Expected: "alignment string",
					Got:      elem.Type().String(),
					Span:     arg.Span,
				}
			}
		}
		return aligns, nil
	default:
		return nil, &TypeMismatchError{
			Expected: "alignment or array of alignments",
			Got:      arg.V.Type().String(),
			Span:     arg.Span,
		}
	}
}

// parseAlignmentString parses an alignment string.
func parseAlignmentString(s string) (Alignment, error) {
	switch s {
	case "left", "start":
		return Alignment{Horizontal: "left"}, nil
	case "center":
		return Alignment{Horizontal: "center"}, nil
	case "right", "end":
		return Alignment{Horizontal: "right"}, nil
	case "top":
		return Alignment{Vertical: "top"}, nil
	case "horizon":
		return Alignment{Vertical: "horizon"}, nil
	case "bottom":
		return Alignment{Vertical: "bottom"}, nil
	default:
		// Could be a compound like "left + top"
		return Alignment{Horizontal: s}, nil
	}
}

// parseInset parses inset from an argument.
func parseInset(arg *syntax.Spanned[Value]) (*Inset, error) {
	switch v := arg.V.(type) {
	case LengthValue:
		pts := v.Length.Points
		return &Inset{All: &pts}, nil
	case RelativeValue:
		pts := v.Relative.Abs.Points
		return &Inset{All: &pts}, nil
	case DictValue:
		inset := &Inset{}
		if top, ok := v.Get("top"); ok {
			if lv, ok := top.(LengthValue); ok {
				pts := lv.Length.Points
				inset.Top = &pts
			}
		}
		if right, ok := v.Get("right"); ok {
			if lv, ok := right.(LengthValue); ok {
				pts := lv.Length.Points
				inset.Right = &pts
			}
		}
		if bottom, ok := v.Get("bottom"); ok {
			if lv, ok := bottom.(LengthValue); ok {
				pts := lv.Length.Points
				inset.Bottom = &pts
			}
		}
		if left, ok := v.Get("left"); ok {
			if lv, ok := left.(LengthValue); ok {
				pts := lv.Length.Points
				inset.Left = &pts
			}
		}
		return inset, nil
	default:
		return nil, &TypeMismatchError{
			Expected: "length, relative, or dictionary",
			Got:      arg.V.Type().String(),
			Span:     arg.Span,
		}
	}
}

// parseFill parses fill from an argument.
func parseFill(arg *syntax.Spanned[Value]) ([]FillSpec, error) {
	switch v := arg.V.(type) {
	case ColorValue:
		return []FillSpec{{Color: &v.Color}}, nil
	case ArrayValue:
		fills := make([]FillSpec, len(v))
		for i, elem := range v {
			if c, ok := elem.(ColorValue); ok {
				fills[i] = FillSpec{Color: &c.Color}
			} else if IsNone(elem) {
				fills[i] = FillSpec{None: true}
			} else {
				return nil, &TypeMismatchError{
					Expected: "color or none",
					Got:      elem.Type().String(),
					Span:     arg.Span,
				}
			}
		}
		return fills, nil
	case NoneValue:
		return []FillSpec{{None: true}}, nil
	default:
		return nil, &TypeMismatchError{
			Expected: "color, array, or none",
			Got:      arg.V.Type().String(),
			Span:     arg.Span,
		}
	}
}

// parseStroke parses stroke from an argument.
func parseStroke(arg *syntax.Spanned[Value]) ([]StrokeSpec, error) {
	switch v := arg.V.(type) {
	case LengthValue:
		pts := v.Length.Points
		return []StrokeSpec{{Thickness: &pts}}, nil
	case ColorValue:
		return []StrokeSpec{{Color: &v.Color}}, nil
	case ArrayValue:
		strokes := make([]StrokeSpec, len(v))
		for i, elem := range v {
			if lv, ok := elem.(LengthValue); ok {
				pts := lv.Length.Points
				strokes[i] = StrokeSpec{Thickness: &pts}
			} else if c, ok := elem.(ColorValue); ok {
				strokes[i] = StrokeSpec{Color: &c.Color}
			} else if IsNone(elem) {
				strokes[i] = StrokeSpec{None: true}
			} else {
				return nil, &TypeMismatchError{
					Expected: "length, color, or none",
					Got:      elem.Type().String(),
					Span:     arg.Span,
				}
			}
		}
		return strokes, nil
	case NoneValue:
		return []StrokeSpec{{None: true}}, nil
	default:
		return nil, &TypeMismatchError{
			Expected: "length, color, array, or none",
			Got:      arg.V.Type().String(),
			Span:     arg.Span,
		}
	}
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
