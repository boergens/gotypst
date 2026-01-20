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
type StackElement struct {
	// Dir is the stacking direction (ltr, rtl, ttb, btt).
	Dir StackDirection
	// Spacing is the spacing between children (in points).
	// If nil, uses default spacing (0pt).
	Spacing *float64
	// Children contains the content elements to stack.
	Children []Content
}

func (*StackElement) IsContentElement() {}

// StackFunc creates the stack element function.
func StackFunc() *Func {
	name := "stack"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: stackNative,
			Info: &FuncInfo{
				Name: "stack",
				Params: []ParamInfo{
					{Name: "dir", Type: TypeStr, Default: Str("ttb"), Named: true},
					{Name: "spacing", Type: TypeLength, Default: None, Named: true},
					{Name: "children", Type: TypeContent, Named: false, Variadic: true},
				},
			},
		},
	}
}

// stackNative implements the stack() function.
// Creates a StackElement with the given direction and children.
//
// Arguments:
//   - dir (named, str, default: "ttb"): The stacking direction (ltr, rtl, ttb, btt)
//   - spacing (named, length, default: none): The spacing between children
//   - children (positional, variadic, content): The content elements to stack
func stackNative(vm *Vm, args *Args) (Value, error) {
	// Get optional dir argument (default: "ttb")
	dir := StackTTB
	if dirArg := args.Find("dir"); dirArg != nil {
		if !IsNone(dirArg.V) && !IsAuto(dirArg.V) {
			dirStr, ok := AsStr(dirArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "string",
					Got:      dirArg.V.Type().String(),
					Span:     dirArg.Span,
				}
			}
			// Validate direction
			switch dirStr {
			case "ltr":
				dir = StackLTR
			case "rtl":
				dir = StackRTL
			case "ttb":
				dir = StackTTB
			case "btt":
				dir = StackBTT
			default:
				return nil, &TypeMismatchError{
					Expected: "\"ltr\", \"rtl\", \"ttb\", or \"btt\"",
					Got:      "\"" + dirStr + "\"",
					Span:     dirArg.Span,
				}
			}
		}
	}

	// Get optional spacing argument
	var spacing *float64
	if spacingArg := args.Find("spacing"); spacingArg != nil {
		if !IsNone(spacingArg.V) && !IsAuto(spacingArg.V) {
			if lv, ok := spacingArg.V.(LengthValue); ok {
				s := lv.Length.Points
				spacing = &s
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      spacingArg.V.Type().String(),
					Span:     spacingArg.Span,
				}
			}
		}
	}

	// Collect remaining positional arguments as children
	var children []Content
	for {
		childArg := args.Eat()
		if childArg == nil {
			break
		}

		if cv, ok := childArg.V.(ContentValue); ok {
			children = append(children, cv.Content)
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

	// Create the StackElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{&StackElement{
			Dir:      dir,
			Spacing:  spacing,
			Children: children,
		}},
	}}, nil
}

// Alignment2D represents a 2D alignment value (horizontal and vertical).
type Alignment2D struct {
	// Horizontal alignment (left, center, right, or none for not specified).
	Horizontal *string
	// Vertical alignment (top, horizon, bottom, or none for not specified).
	Vertical *string
}

// AlignElement represents an alignment container element.
// It positions its content according to the specified alignment.
type AlignElement struct {
	// Alignment is the 2D alignment specification.
	Alignment Alignment2D
	// Body is the content to align.
	Body Content
}

func (*AlignElement) IsContentElement() {}

// AlignFunc creates the align element function.
func AlignFunc() *Func {
	name := "align"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: alignNative,
			Info: &FuncInfo{
				Name: "align",
				Params: []ParamInfo{
					{Name: "alignment", Type: TypeStr, Named: false},
					{Name: "body", Type: TypeContent, Named: false},
				},
			},
		},
	}
}

// alignNative implements the align() function.
// Creates an AlignElement to position content.
//
// Arguments:
//   - alignment (positional, alignment): The alignment specification
//   - body (positional, content): The content to align
func alignNative(vm *Vm, args *Args) (Value, error) {
	// Get required alignment argument
	alignArg, err := args.Expect("alignment")
	if err != nil {
		return nil, err
	}

	alignment, err := parseAlignment(alignArg.V, alignArg.Span)
	if err != nil {
		return nil, err
	}

	// Get required body argument
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

	// Create the AlignElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{&AlignElement{
			Alignment: alignment,
			Body:      body,
		}},
	}}, nil
}

// parseAlignment parses an alignment value from a Value.
// Supports: left, center, right, top, horizon, bottom, or 2D combinations.
func parseAlignment(v Value, span syntax.Span) (Alignment2D, error) {
	// Handle string alignment values
	if s, ok := AsStr(v); ok {
		return parseAlignmentString(s, span)
	}

	// Handle alignment value types (for when we have proper alignment types)
	// For now, return an error for unsupported types
	return Alignment2D{}, &TypeMismatchError{
		Expected: "alignment",
		Got:      v.Type().String(),
		Span:     span,
	}
}

// parseAlignmentString parses an alignment from a string.
func parseAlignmentString(s string, span syntax.Span) (Alignment2D, error) {
	var result Alignment2D

	switch s {
	case "left":
		h := "left"
		result.Horizontal = &h
	case "center":
		h := "center"
		result.Horizontal = &h
	case "right":
		h := "right"
		result.Horizontal = &h
	case "top":
		v := "top"
		result.Vertical = &v
	case "horizon":
		v := "horizon"
		result.Vertical = &v
	case "bottom":
		v := "bottom"
		result.Vertical = &v
	case "start":
		h := "start"
		result.Horizontal = &h
	case "end":
		h := "end"
		result.Horizontal = &h
	default:
		return Alignment2D{}, &TypeMismatchError{
			Expected: "\"left\", \"center\", \"right\", \"top\", \"horizon\", \"bottom\", \"start\", or \"end\"",
			Got:      "\"" + s + "\"",
			Span:     span,
		}
	}

	return result, nil
}

// ----------------------------------------------------------------------------
// Heading Element
// ----------------------------------------------------------------------------

// HeadingFunc creates the heading element function.
func HeadingFunc() *Func {
	name := "heading"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: headingNative,
			Info: &FuncInfo{
				Name: "heading",
				Params: []ParamInfo{
					{Name: "body", Type: TypeContent, Named: false},
					{Name: "level", Type: TypeInt, Default: Int(1), Named: true},
					{Name: "depth", Type: TypeInt, Default: None, Named: true},
					{Name: "offset", Type: TypeInt, Default: Int(0), Named: true},
					{Name: "numbering", Type: TypeStr, Default: None, Named: true},
					{Name: "supplement", Type: TypeContent, Default: Auto, Named: true},
					{Name: "outlined", Type: TypeBool, Default: True, Named: true},
					{Name: "bookmarked", Type: TypeBool, Default: Auto, Named: true},
				},
			},
		},
	}
}

// headingNative implements the heading() function.
// Creates a HeadingElement from the given content with optional level and numbering.
//
// Arguments:
//   - body (positional, content): The heading content
//   - level (named, int, default: 1): The heading level (1-6)
//   - depth (named, int, default: none): Depth for numbering inheritance
//   - offset (named, int, default: 0): Numbering offset
//   - numbering (named, str or none, default: none): Numbering pattern (e.g., "1.", "1.1", "I.")
//   - supplement (named, content or auto, default: auto): Supplement content for references
//   - outlined (named, bool, default: true): Whether to show in outline
//   - bookmarked (named, bool or auto, default: auto): Whether to bookmark in PDF
func headingNative(vm *Vm, args *Args) (Value, error) {
	// Get required body argument (can be positional or named)
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

	// Get optional level argument (default: 1)
	level := 1
	if levelArg := args.Find("level"); levelArg != nil {
		levelVal, ok := AsInt(levelArg.V)
		if !ok {
			return nil, &TypeMismatchError{
				Expected: "integer",
				Got:      levelArg.V.Type().String(),
				Span:     levelArg.Span,
			}
		}
		level = int(levelVal)
		if level < 1 || level > 6 {
			return nil, &ConstructorError{
				Message: "heading level must be between 1 and 6",
				Span:    levelArg.Span,
			}
		}
	}

	// Get optional numbering argument (default: none)
	var numbering *string
	if numberingArg := args.Find("numbering"); numberingArg != nil {
		if !IsNone(numberingArg.V) {
			if numStr, ok := AsStr(numberingArg.V); ok {
				numbering = &numStr
			} else {
				return nil, &TypeMismatchError{
					Expected: "string or none",
					Got:      numberingArg.V.Type().String(),
					Span:     numberingArg.Span,
				}
			}
		}
	}

	// Get optional supplement argument (default: auto)
	var supplement *Content
	if supplementArg := args.Find("supplement"); supplementArg != nil {
		if !IsAuto(supplementArg.V) && !IsNone(supplementArg.V) {
			if cv, ok := supplementArg.V.(ContentValue); ok {
				supplement = &cv.Content
			} else {
				return nil, &TypeMismatchError{
					Expected: "content or auto",
					Got:      supplementArg.V.Type().String(),
					Span:     supplementArg.Span,
				}
			}
		}
	}

	// Get optional outlined argument (default: true)
	outlined := true
	if outlinedArg := args.Find("outlined"); outlinedArg != nil {
		outlinedVal, ok := AsBool(outlinedArg.V)
		if !ok {
			return nil, &TypeMismatchError{
				Expected: "bool",
				Got:      outlinedArg.V.Type().String(),
				Span:     outlinedArg.Span,
			}
		}
		outlined = outlinedVal
	}

	// Get optional bookmarked argument (default: auto)
	var bookmarked *bool
	if bookmarkedArg := args.Find("bookmarked"); bookmarkedArg != nil {
		if !IsAuto(bookmarkedArg.V) {
			bv, ok := AsBool(bookmarkedArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "bool or auto",
					Got:      bookmarkedArg.V.Type().String(),
					Span:     bookmarkedArg.Span,
				}
			}
			bookmarked = &bv
		}
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Create the HeadingElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{&HeadingElement{
			Level:      level,
			Content:    body,
			Numbering:  numbering,
			Supplement: supplement,
			Outlined:   outlined,
			Bookmarked: bookmarked,
		}},
	}}, nil
}

// ----------------------------------------------------------------------------
// Table Element
// ----------------------------------------------------------------------------

// TrackSize represents a column or row track size.
// It can be auto, a length, a ratio, a fraction, or an array of such.
type TrackSize struct {
	// Auto means the track sizes itself to fit its content.
	Auto bool
	// Length is an absolute length (in points) if specified.
	Length *float64
	// Ratio is a percentage of available space if specified.
	Ratio *float64
	// Fraction is a flexible fraction (fr units) if specified.
	Fraction *float64
}

// TableCellElement represents a single cell in a table.
type TableCellElement struct {
	// Body is the content of the cell.
	Body Content
	// X is the column position (0-indexed). If nil, auto-placed.
	X *int
	// Y is the row position (0-indexed). If nil, auto-placed.
	Y *int
	// Colspan is the number of columns this cell spans (default: 1).
	Colspan int
	// Rowspan is the number of rows this cell spans (default: 1).
	Rowspan int
	// Fill is the cell background color (if any).
	Fill *Color
	// Align is the cell-specific alignment (if any).
	Align *Alignment2D
	// Inset is the cell padding (in points). If nil, uses table default.
	Inset *float64
	// Stroke is the cell border stroke (if any).
	Stroke *Stroke
}

func (*TableCellElement) IsContentElement() {}

// Stroke represents a stroke style for borders.
type Stroke struct {
	// Thickness is the stroke width in points.
	Thickness float64
	// Color is the stroke color.
	Color Color
	// Dash is the dash pattern (nil for solid).
	Dash []float64
}

// TableElement represents a table layout element.
type TableElement struct {
	// Columns defines the column track sizes.
	// Can be a single value (number of columns) or specific sizes.
	Columns []TrackSize
	// Rows defines the row track sizes (optional, auto-sized if nil).
	Rows []TrackSize
	// Gutter is the default spacing between cells.
	Gutter *float64
	// ColumnGutter is the spacing between columns (overrides Gutter).
	ColumnGutter *float64
	// RowGutter is the spacing between rows (overrides Gutter).
	RowGutter *float64
	// Fill is the default cell background.
	Fill *Color
	// Align is the default cell alignment.
	Align *Alignment2D
	// Stroke is the table border stroke.
	Stroke *Stroke
	// Inset is the default cell padding (in points).
	Inset *float64
	// Children contains the table cells and content.
	Children []TableChild
}

// TableChild represents either a TableCellElement or plain Content in a table.
type TableChild struct {
	// Cell is set if this is an explicit table.cell().
	Cell *TableCellElement
	// Content is set if this is plain content (implicitly wrapped in a cell).
	Content *Content
}

func (*TableElement) IsContentElement() {}

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
					{Name: "fill", Type: TypeColor, Default: None, Named: true},
					{Name: "align", Type: TypeStr, Default: Auto, Named: true},
					{Name: "stroke", Type: TypeLength, Default: Auto, Named: true},
					{Name: "inset", Type: TypeLength, Default: Auto, Named: true},
					{Name: "children", Type: TypeContent, Named: false, Variadic: true},
				},
			},
		},
	}
}

// tableNative implements the table() function.
// Creates a TableElement with the given columns, rows, and children.
//
// Arguments:
//   - columns (named, int or array, default: auto): Column track sizes
//   - rows (named, int or array, default: auto): Row track sizes
//   - gutter (named, length, default: none): Default cell spacing
//   - column-gutter (named, length, default: auto): Column spacing
//   - row-gutter (named, length, default: auto): Row spacing
//   - fill (named, color, default: none): Default cell background
//   - align (named, alignment, default: auto): Default cell alignment
//   - stroke (named, length or stroke, default: auto): Table border
//   - inset (named, length, default: auto): Default cell padding
//   - children (positional, variadic): Table cells and content
func tableNative(vm *Vm, args *Args) (Value, error) {
	elem := &TableElement{
		Columns: []TrackSize{{Auto: true}}, // Default: 1 auto column
	}

	// Get optional columns argument
	if columnsArg := args.Find("columns"); columnsArg != nil {
		if !IsAuto(columnsArg.V) && !IsNone(columnsArg.V) {
			tracks, err := parseTrackSizes(columnsArg.V, columnsArg.Span)
			if err != nil {
				return nil, err
			}
			elem.Columns = tracks
		}
	}

	// Get optional rows argument
	if rowsArg := args.Find("rows"); rowsArg != nil {
		if !IsAuto(rowsArg.V) && !IsNone(rowsArg.V) {
			tracks, err := parseTrackSizes(rowsArg.V, rowsArg.Span)
			if err != nil {
				return nil, err
			}
			elem.Rows = tracks
		}
	}

	// Get optional gutter argument
	if gutterArg := args.Find("gutter"); gutterArg != nil {
		if !IsAuto(gutterArg.V) && !IsNone(gutterArg.V) {
			if lv, ok := gutterArg.V.(LengthValue); ok {
				g := lv.Length.Points
				elem.Gutter = &g
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      gutterArg.V.Type().String(),
					Span:     gutterArg.Span,
				}
			}
		}
	}

	// Get optional column-gutter argument
	if colGutterArg := args.Find("column-gutter"); colGutterArg != nil {
		if !IsAuto(colGutterArg.V) && !IsNone(colGutterArg.V) {
			if lv, ok := colGutterArg.V.(LengthValue); ok {
				g := lv.Length.Points
				elem.ColumnGutter = &g
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      colGutterArg.V.Type().String(),
					Span:     colGutterArg.Span,
				}
			}
		}
	}

	// Get optional row-gutter argument
	if rowGutterArg := args.Find("row-gutter"); rowGutterArg != nil {
		if !IsAuto(rowGutterArg.V) && !IsNone(rowGutterArg.V) {
			if lv, ok := rowGutterArg.V.(LengthValue); ok {
				g := lv.Length.Points
				elem.RowGutter = &g
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      rowGutterArg.V.Type().String(),
					Span:     rowGutterArg.Span,
				}
			}
		}
	}

	// Get optional fill argument
	if fillArg := args.Find("fill"); fillArg != nil {
		if !IsAuto(fillArg.V) && !IsNone(fillArg.V) {
			if cv, ok := fillArg.V.(ColorValue); ok {
				elem.Fill = &cv.Color
			} else {
				return nil, &TypeMismatchError{
					Expected: "color or none",
					Got:      fillArg.V.Type().String(),
					Span:     fillArg.Span,
				}
			}
		}
	}

	// Get optional align argument
	if alignArg := args.Find("align"); alignArg != nil {
		if !IsAuto(alignArg.V) && !IsNone(alignArg.V) {
			alignment, err := parseAlignment(alignArg.V, alignArg.Span)
			if err != nil {
				return nil, err
			}
			elem.Align = &alignment
		}
	}

	// Get optional stroke argument
	if strokeArg := args.Find("stroke"); strokeArg != nil {
		if !IsAuto(strokeArg.V) && !IsNone(strokeArg.V) {
			stroke, err := parseStroke(strokeArg.V, strokeArg.Span)
			if err != nil {
				return nil, err
			}
			elem.Stroke = stroke
		}
	}

	// Get optional inset argument
	if insetArg := args.Find("inset"); insetArg != nil {
		if !IsAuto(insetArg.V) && !IsNone(insetArg.V) {
			if lv, ok := insetArg.V.(LengthValue); ok {
				i := lv.Length.Points
				elem.Inset = &i
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      insetArg.V.Type().String(),
					Span:     insetArg.Span,
				}
			}
		}
	}

	// Collect remaining positional arguments as children
	for {
		childArg := args.Eat()
		if childArg == nil {
			break
		}

		// Check if this is a table.cell() element
		if cv, ok := childArg.V.(ContentValue); ok {
			if len(cv.Content.Elements) == 1 {
				if cell, ok := cv.Content.Elements[0].(*TableCellElement); ok {
					elem.Children = append(elem.Children, TableChild{Cell: cell})
					continue
				}
			}
			// Plain content - wrap implicitly
			elem.Children = append(elem.Children, TableChild{Content: &cv.Content})
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

// parseTrackSizes parses column or row track sizes from a Value.
func parseTrackSizes(v Value, span syntax.Span) ([]TrackSize, error) {
	// Handle single integer (number of auto columns/rows)
	if i, ok := AsInt(v); ok {
		if i < 1 {
			return nil, &ConstructorError{
				Message: "number of tracks must be at least 1",
				Span:    span,
			}
		}
		tracks := make([]TrackSize, i)
		for j := range tracks {
			tracks[j] = TrackSize{Auto: true}
		}
		return tracks, nil
	}

	// Handle array of track sizes
	if arr, ok := AsArray(v); ok {
		tracks := make([]TrackSize, len(arr))
		for i, elem := range arr {
			track, err := parseTrackSize(elem, span)
			if err != nil {
				return nil, err
			}
			tracks[i] = track
		}
		return tracks, nil
	}

	// Handle single track size
	track, err := parseTrackSize(v, span)
	if err != nil {
		return nil, err
	}
	return []TrackSize{track}, nil
}

// parseTrackSize parses a single track size from a Value.
func parseTrackSize(v Value, span syntax.Span) (TrackSize, error) {
	switch val := v.(type) {
	case AutoValue:
		return TrackSize{Auto: true}, nil
	case LengthValue:
		pts := val.Length.Points
		return TrackSize{Length: &pts}, nil
	case RatioValue:
		ratio := val.Ratio.Value
		return TrackSize{Ratio: &ratio}, nil
	case FractionValue:
		fr := val.Fraction.Value
		return TrackSize{Fraction: &fr}, nil
	default:
		return TrackSize{}, &TypeMismatchError{
			Expected: "auto, length, ratio, or fraction",
			Got:      v.Type().String(),
			Span:     span,
		}
	}
}

// parseStroke parses a stroke from a Value.
func parseStroke(v Value, span syntax.Span) (*Stroke, error) {
	// Handle length as a simple stroke thickness with black color
	if lv, ok := v.(LengthValue); ok {
		return &Stroke{
			Thickness: lv.Length.Points,
			Color:     Color{R: 0, G: 0, B: 0, A: 255},
		}, nil
	}

	// Handle color as a 1pt stroke with that color
	if cv, ok := v.(ColorValue); ok {
		return &Stroke{
			Thickness: 1.0,
			Color:     cv.Color,
		}, nil
	}

	// Handle none for no stroke
	if IsNone(v) {
		return nil, nil
	}

	return nil, &TypeMismatchError{
		Expected: "length, color, or none",
		Got:      v.Type().String(),
		Span:     span,
	}
}

// TableCellFunc creates the table.cell element function.
func TableCellFunc() *Func {
	name := "cell"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: tableCellNative,
			Info: &FuncInfo{
				Name: "table.cell",
				Params: []ParamInfo{
					{Name: "body", Type: TypeContent, Named: false},
					{Name: "x", Type: TypeInt, Default: Auto, Named: true},
					{Name: "y", Type: TypeInt, Default: Auto, Named: true},
					{Name: "colspan", Type: TypeInt, Default: Int(1), Named: true},
					{Name: "rowspan", Type: TypeInt, Default: Int(1), Named: true},
					{Name: "fill", Type: TypeColor, Default: Auto, Named: true},
					{Name: "align", Type: TypeStr, Default: Auto, Named: true},
					{Name: "inset", Type: TypeLength, Default: Auto, Named: true},
					{Name: "stroke", Type: TypeLength, Default: Auto, Named: true},
				},
			},
		},
	}
}

// tableCellNative implements the table.cell() function.
// Creates a TableCellElement with the given content and options.
//
// Arguments:
//   - body (positional, content): The cell content
//   - x (named, int, default: auto): Column position (0-indexed)
//   - y (named, int, default: auto): Row position (0-indexed)
//   - colspan (named, int, default: 1): Number of columns to span
//   - rowspan (named, int, default: 1): Number of rows to span
//   - fill (named, color, default: auto): Cell background color
//   - align (named, alignment, default: auto): Cell alignment
//   - inset (named, length, default: auto): Cell padding
//   - stroke (named, stroke, default: auto): Cell border
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

	// Get optional x argument
	if xArg := args.Find("x"); xArg != nil {
		if !IsAuto(xArg.V) && !IsNone(xArg.V) {
			x, ok := AsInt(xArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "int or auto",
					Got:      xArg.V.Type().String(),
					Span:     xArg.Span,
				}
			}
			if x < 0 {
				return nil, &ConstructorError{
					Message: "cell x position must be non-negative",
					Span:    xArg.Span,
				}
			}
			xInt := int(x)
			elem.X = &xInt
		}
	}

	// Get optional y argument
	if yArg := args.Find("y"); yArg != nil {
		if !IsAuto(yArg.V) && !IsNone(yArg.V) {
			y, ok := AsInt(yArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "int or auto",
					Got:      yArg.V.Type().String(),
					Span:     yArg.Span,
				}
			}
			if y < 0 {
				return nil, &ConstructorError{
					Message: "cell y position must be non-negative",
					Span:    yArg.Span,
				}
			}
			yInt := int(y)
			elem.Y = &yInt
		}
	}

	// Get optional colspan argument
	if colspanArg := args.Find("colspan"); colspanArg != nil {
		colspan, ok := AsInt(colspanArg.V)
		if !ok {
			return nil, &TypeMismatchError{
				Expected: "int",
				Got:      colspanArg.V.Type().String(),
				Span:     colspanArg.Span,
			}
		}
		if colspan < 1 {
			return nil, &ConstructorError{
				Message: "colspan must be at least 1",
				Span:    colspanArg.Span,
			}
		}
		elem.Colspan = int(colspan)
	}

	// Get optional rowspan argument
	if rowspanArg := args.Find("rowspan"); rowspanArg != nil {
		rowspan, ok := AsInt(rowspanArg.V)
		if !ok {
			return nil, &TypeMismatchError{
				Expected: "int",
				Got:      rowspanArg.V.Type().String(),
				Span:     rowspanArg.Span,
			}
		}
		if rowspan < 1 {
			return nil, &ConstructorError{
				Message: "rowspan must be at least 1",
				Span:    rowspanArg.Span,
			}
		}
		elem.Rowspan = int(rowspan)
	}

	// Get optional fill argument
	if fillArg := args.Find("fill"); fillArg != nil {
		if !IsAuto(fillArg.V) && !IsNone(fillArg.V) {
			if cv, ok := fillArg.V.(ColorValue); ok {
				elem.Fill = &cv.Color
			} else {
				return nil, &TypeMismatchError{
					Expected: "color or auto",
					Got:      fillArg.V.Type().String(),
					Span:     fillArg.Span,
				}
			}
		}
	}

	// Get optional align argument
	if alignArg := args.Find("align"); alignArg != nil {
		if !IsAuto(alignArg.V) && !IsNone(alignArg.V) {
			alignment, err := parseAlignment(alignArg.V, alignArg.Span)
			if err != nil {
				return nil, err
			}
			elem.Align = &alignment
		}
	}

	// Get optional inset argument
	if insetArg := args.Find("inset"); insetArg != nil {
		if !IsAuto(insetArg.V) && !IsNone(insetArg.V) {
			if lv, ok := insetArg.V.(LengthValue); ok {
				i := lv.Length.Points
				elem.Inset = &i
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      insetArg.V.Type().String(),
					Span:     insetArg.Span,
				}
			}
		}
	}

	// Get optional stroke argument
	if strokeArg := args.Find("stroke"); strokeArg != nil {
		if !IsAuto(strokeArg.V) && !IsNone(strokeArg.V) {
			stroke, err := parseStroke(strokeArg.V, strokeArg.Span)
			if err != nil {
				return nil, err
			}
			elem.Stroke = stroke
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
// Element Methods
// ----------------------------------------------------------------------------

// GetElementMethod returns a method function for an element function.
// For example, GetElementMethod("table", "cell") returns the table.cell function.
// Returns nil if no such method exists.
func GetElementMethod(elementName, methodName string, span syntax.Span) Value {
	switch elementName {
	case "table":
		switch methodName {
		case "cell":
			return FuncValue{Func: TableCellFunc()}
		case "header":
			return FuncValue{Func: TableHeaderFunc()}
		case "footer":
			return FuncValue{Func: TableFooterFunc()}
		case "hline":
			return FuncValue{Func: TableHlineFunc()}
		case "vline":
			return FuncValue{Func: TableVlineFunc()}
		}
	}
	return nil
}

// TableHeaderFunc creates the table.header element function.
func TableHeaderFunc() *Func {
	name := "header"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: tableHeaderNative,
			Info: &FuncInfo{
				Name: "table.header",
				Params: []ParamInfo{
					{Name: "repeat", Type: TypeBool, Default: True, Named: true},
					{Name: "children", Type: TypeContent, Named: false, Variadic: true},
				},
			},
		},
	}
}

// TableHeaderElement represents a table header section.
type TableHeaderElement struct {
	// Repeat indicates whether to repeat the header on each page.
	Repeat bool
	// Children contains the header cells.
	Children []TableChild
}

func (*TableHeaderElement) IsContentElement() {}

// tableHeaderNative implements the table.header() function.
func tableHeaderNative(vm *Vm, args *Args) (Value, error) {
	elem := &TableHeaderElement{
		Repeat: true,
	}

	// Get optional repeat argument
	if repeatArg := args.Find("repeat"); repeatArg != nil {
		repeat, ok := AsBool(repeatArg.V)
		if !ok {
			return nil, &TypeMismatchError{
				Expected: "bool",
				Got:      repeatArg.V.Type().String(),
				Span:     repeatArg.Span,
			}
		}
		elem.Repeat = repeat
	}

	// Collect remaining positional arguments as children
	for {
		childArg := args.Eat()
		if childArg == nil {
			break
		}

		if cv, ok := childArg.V.(ContentValue); ok {
			if len(cv.Content.Elements) == 1 {
				if cell, ok := cv.Content.Elements[0].(*TableCellElement); ok {
					elem.Children = append(elem.Children, TableChild{Cell: cell})
					continue
				}
			}
			elem.Children = append(elem.Children, TableChild{Content: &cv.Content})
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

// TableFooterFunc creates the table.footer element function.
func TableFooterFunc() *Func {
	name := "footer"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: tableFooterNative,
			Info: &FuncInfo{
				Name: "table.footer",
				Params: []ParamInfo{
					{Name: "repeat", Type: TypeBool, Default: True, Named: true},
					{Name: "children", Type: TypeContent, Named: false, Variadic: true},
				},
			},
		},
	}
}

// TableFooterElement represents a table footer section.
type TableFooterElement struct {
	// Repeat indicates whether to repeat the footer on each page.
	Repeat bool
	// Children contains the footer cells.
	Children []TableChild
}

func (*TableFooterElement) IsContentElement() {}

// tableFooterNative implements the table.footer() function.
func tableFooterNative(vm *Vm, args *Args) (Value, error) {
	elem := &TableFooterElement{
		Repeat: true,
	}

	// Get optional repeat argument
	if repeatArg := args.Find("repeat"); repeatArg != nil {
		repeat, ok := AsBool(repeatArg.V)
		if !ok {
			return nil, &TypeMismatchError{
				Expected: "bool",
				Got:      repeatArg.V.Type().String(),
				Span:     repeatArg.Span,
			}
		}
		elem.Repeat = repeat
	}

	// Collect remaining positional arguments as children
	for {
		childArg := args.Eat()
		if childArg == nil {
			break
		}

		if cv, ok := childArg.V.(ContentValue); ok {
			if len(cv.Content.Elements) == 1 {
				if cell, ok := cv.Content.Elements[0].(*TableCellElement); ok {
					elem.Children = append(elem.Children, TableChild{Cell: cell})
					continue
				}
			}
			elem.Children = append(elem.Children, TableChild{Content: &cv.Content})
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

// TableHlineFunc creates the table.hline element function.
func TableHlineFunc() *Func {
	name := "hline"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: tableHlineNative,
			Info: &FuncInfo{
				Name: "table.hline",
				Params: []ParamInfo{
					{Name: "y", Type: TypeInt, Default: Auto, Named: true},
					{Name: "start", Type: TypeInt, Default: Int(0), Named: true},
					{Name: "end", Type: TypeInt, Default: None, Named: true},
					{Name: "stroke", Type: TypeLength, Default: Auto, Named: true},
					{Name: "position", Type: TypeStr, Default: Str("top"), Named: true},
				},
			},
		},
	}
}

// TableHlineElement represents a horizontal line in a table.
type TableHlineElement struct {
	// Y is the row position (0-indexed). If nil, auto-placed.
	Y *int
	// Start is the starting column (0-indexed, default: 0).
	Start int
	// End is the ending column (exclusive). If nil, extends to end.
	End *int
	// Stroke is the line stroke.
	Stroke *Stroke
	// Position is "top" or "bottom" relative to the row.
	Position string
}

func (*TableHlineElement) IsContentElement() {}

// tableHlineNative implements the table.hline() function.
func tableHlineNative(vm *Vm, args *Args) (Value, error) {
	elem := &TableHlineElement{
		Start:    0,
		Position: "top",
	}

	// Get optional y argument
	if yArg := args.Find("y"); yArg != nil {
		if !IsAuto(yArg.V) && !IsNone(yArg.V) {
			y, ok := AsInt(yArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "int or auto",
					Got:      yArg.V.Type().String(),
					Span:     yArg.Span,
				}
			}
			yInt := int(y)
			elem.Y = &yInt
		}
	}

	// Get optional start argument
	if startArg := args.Find("start"); startArg != nil {
		start, ok := AsInt(startArg.V)
		if !ok {
			return nil, &TypeMismatchError{
				Expected: "int",
				Got:      startArg.V.Type().String(),
				Span:     startArg.Span,
			}
		}
		elem.Start = int(start)
	}

	// Get optional end argument
	if endArg := args.Find("end"); endArg != nil {
		if !IsNone(endArg.V) {
			end, ok := AsInt(endArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "int or none",
					Got:      endArg.V.Type().String(),
					Span:     endArg.Span,
				}
			}
			endInt := int(end)
			elem.End = &endInt
		}
	}

	// Get optional stroke argument
	if strokeArg := args.Find("stroke"); strokeArg != nil {
		if !IsAuto(strokeArg.V) && !IsNone(strokeArg.V) {
			stroke, err := parseStroke(strokeArg.V, strokeArg.Span)
			if err != nil {
				return nil, err
			}
			elem.Stroke = stroke
		}
	}

	// Get optional position argument
	if posArg := args.Find("position"); posArg != nil {
		pos, ok := AsStr(posArg.V)
		if !ok {
			return nil, &TypeMismatchError{
				Expected: "string",
				Got:      posArg.V.Type().String(),
				Span:     posArg.Span,
			}
		}
		if pos != "top" && pos != "bottom" {
			return nil, &TypeMismatchError{
				Expected: "\"top\" or \"bottom\"",
				Got:      "\"" + pos + "\"",
				Span:     posArg.Span,
			}
		}
		elem.Position = pos
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}

// TableVlineFunc creates the table.vline element function.
func TableVlineFunc() *Func {
	name := "vline"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: tableVlineNative,
			Info: &FuncInfo{
				Name: "table.vline",
				Params: []ParamInfo{
					{Name: "x", Type: TypeInt, Default: Auto, Named: true},
					{Name: "start", Type: TypeInt, Default: Int(0), Named: true},
					{Name: "end", Type: TypeInt, Default: None, Named: true},
					{Name: "stroke", Type: TypeLength, Default: Auto, Named: true},
					{Name: "position", Type: TypeStr, Default: Str("start"), Named: true},
				},
			},
		},
	}
}

// TableVlineElement represents a vertical line in a table.
type TableVlineElement struct {
	// X is the column position (0-indexed). If nil, auto-placed.
	X *int
	// Start is the starting row (0-indexed, default: 0).
	Start int
	// End is the ending row (exclusive). If nil, extends to end.
	End *int
	// Stroke is the line stroke.
	Stroke *Stroke
	// Position is "start" or "end" relative to the column.
	Position string
}

func (*TableVlineElement) IsContentElement() {}

// tableVlineNative implements the table.vline() function.
func tableVlineNative(vm *Vm, args *Args) (Value, error) {
	elem := &TableVlineElement{
		Start:    0,
		Position: "start",
	}

	// Get optional x argument
	if xArg := args.Find("x"); xArg != nil {
		if !IsAuto(xArg.V) && !IsNone(xArg.V) {
			x, ok := AsInt(xArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "int or auto",
					Got:      xArg.V.Type().String(),
					Span:     xArg.Span,
				}
			}
			xInt := int(x)
			elem.X = &xInt
		}
	}

	// Get optional start argument
	if startArg := args.Find("start"); startArg != nil {
		start, ok := AsInt(startArg.V)
		if !ok {
			return nil, &TypeMismatchError{
				Expected: "int",
				Got:      startArg.V.Type().String(),
				Span:     startArg.Span,
			}
		}
		elem.Start = int(start)
	}

	// Get optional end argument
	if endArg := args.Find("end"); endArg != nil {
		if !IsNone(endArg.V) {
			end, ok := AsInt(endArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "int or none",
					Got:      endArg.V.Type().String(),
					Span:     endArg.Span,
				}
			}
			endInt := int(end)
			elem.End = &endInt
		}
	}

	// Get optional stroke argument
	if strokeArg := args.Find("stroke"); strokeArg != nil {
		if !IsAuto(strokeArg.V) && !IsNone(strokeArg.V) {
			stroke, err := parseStroke(strokeArg.V, strokeArg.Span)
			if err != nil {
				return nil, err
			}
			elem.Stroke = stroke
		}
	}

	// Get optional position argument
	if posArg := args.Find("position"); posArg != nil {
		pos, ok := AsStr(posArg.V)
		if !ok {
			return nil, &TypeMismatchError{
				Expected: "string",
				Got:      posArg.V.Type().String(),
				Span:     posArg.Span,
			}
		}
		if pos != "start" && pos != "end" {
			return nil, &TypeMismatchError{
				Expected: "\"start\" or \"end\"",
				Got:      "\"" + pos + "\"",
				Span:     posArg.Span,
			}
		}
		elem.Position = pos
	}

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
	// Register stack element function
	scope.DefineFunc("stack", StackFunc())
	// Register align element function
	scope.DefineFunc("align", AlignFunc())
	// Register heading element function
	scope.DefineFunc("heading", HeadingFunc())
	// Register table element function
	scope.DefineFunc("table", TableFunc())
}

// ElementFunctions returns a map of all element function names to their functions.
// This is useful for introspection and testing.
func ElementFunctions() map[string]*Func {
	return map[string]*Func{
		"raw":      RawFunc(),
		"par":      ParFunc(),
		"parbreak": ParbreakFunc(),
		"stack":    StackFunc(),
		"align":    AlignFunc(),
		"heading":  HeadingFunc(),
		"table":    TableFunc(),
	}
}
