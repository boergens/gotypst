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
// Sides Type
// ----------------------------------------------------------------------------

// Sides represents a 4-sided value (used for padding, inset, outset, radius).
// Values are in points. A nil pointer means "not specified" (use default/auto).
type Sides struct {
	Left   *float64
	Top    *float64
	Right  *float64
	Bottom *float64
}

// parseSidesValue parses a sides value from arguments.
// Accepts: length (uniform), or dictionary with left/top/right/bottom/x/y/rest keys.
func parseSidesValue(v Value, span syntax.Span) (*Sides, error) {
	if IsNone(v) || IsAuto(v) {
		return nil, nil
	}

	// Handle uniform length value
	if lv, ok := v.(LengthValue); ok {
		pts := lv.Length.Points
		return &Sides{Left: &pts, Top: &pts, Right: &pts, Bottom: &pts}, nil
	}

	// Handle relative length value
	if rv, ok := v.(RelativeValue); ok {
		pts := rv.Relative.Abs.Points
		return &Sides{Left: &pts, Top: &pts, Right: &pts, Bottom: &pts}, nil
	}

	// Handle dictionary with side keys
	if dict, ok := v.(DictValue); ok {
		sides := &Sides{}

		// Helper to extract length from dict
		extractLength := func(key string) (*float64, error) {
			if val, exists := dict.Get(key); exists {
				if IsNone(val) || IsAuto(val) {
					return nil, nil
				}
				if lv, ok := val.(LengthValue); ok {
					pts := lv.Length.Points
					return &pts, nil
				}
				if rv, ok := val.(RelativeValue); ok {
					pts := rv.Relative.Abs.Points
					return &pts, nil
				}
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      val.Type().String(),
					Span:     span,
				}
			}
			return nil, nil
		}

		// Check for rest first (default for unspecified sides)
		restVal, err := extractLength("rest")
		if err != nil {
			return nil, err
		}
		if restVal != nil {
			sides.Left = restVal
			sides.Top = restVal
			sides.Right = restVal
			sides.Bottom = restVal
		}

		// Check for x (left + right)
		xVal, err := extractLength("x")
		if err != nil {
			return nil, err
		}
		if xVal != nil {
			sides.Left = xVal
			sides.Right = xVal
		}

		// Check for y (top + bottom)
		yVal, err := extractLength("y")
		if err != nil {
			return nil, err
		}
		if yVal != nil {
			sides.Top = yVal
			sides.Bottom = yVal
		}

		// Individual sides override x/y/rest
		if leftVal, err := extractLength("left"); err != nil {
			return nil, err
		} else if leftVal != nil {
			sides.Left = leftVal
		}

		if topVal, err := extractLength("top"); err != nil {
			return nil, err
		} else if topVal != nil {
			sides.Top = topVal
		}

		if rightVal, err := extractLength("right"); err != nil {
			return nil, err
		} else if rightVal != nil {
			sides.Right = rightVal
		}

		if bottomVal, err := extractLength("bottom"); err != nil {
			return nil, err
		} else if bottomVal != nil {
			sides.Bottom = bottomVal
		}

		return sides, nil
	}

	return nil, &TypeMismatchError{
		Expected: "length, dictionary, or auto",
		Got:      v.Type().String(),
		Span:     span,
	}
}

// ----------------------------------------------------------------------------
// Box Element
// ----------------------------------------------------------------------------

// BoxElement represents an inline-level container.
// It wraps content in a box with optional dimensions, fill, stroke, and padding.
type BoxElement struct {
	// Body is the content inside the box.
	Body Content
	// Width is the box width (nil = auto).
	Width *float64
	// Height is the box height (nil = auto).
	Height *float64
	// Baseline is the baseline shift (nil = default).
	Baseline *float64
	// Fill is the background color (nil = none).
	Fill *ColorValue
	// Stroke is the border stroke width (nil = none).
	Stroke *float64
	// Radius is the corner radius.
	Radius *Sides
	// Inset is the inner padding.
	Inset *Sides
	// Outset is the outer padding.
	Outset *Sides
	// Clip indicates whether to clip content to the box.
	Clip bool
}

func (*BoxElement) IsContentElement() {}

// BoxFunc creates the box element function.
func BoxFunc() *Func {
	name := "box"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: boxNative,
			Info: &FuncInfo{
				Name: "box",
				Params: []ParamInfo{
					{Name: "body", Type: TypeContent, Default: None, Named: false},
					{Name: "width", Type: TypeLength, Default: Auto, Named: true},
					{Name: "height", Type: TypeLength, Default: Auto, Named: true},
					{Name: "baseline", Type: TypeLength, Default: None, Named: true},
					{Name: "fill", Type: TypeColor, Default: None, Named: true},
					{Name: "stroke", Type: TypeLength, Default: None, Named: true},
					{Name: "radius", Type: TypeLength, Default: None, Named: true},
					{Name: "inset", Type: TypeLength, Default: None, Named: true},
					{Name: "outset", Type: TypeLength, Default: None, Named: true},
					{Name: "clip", Type: TypeBool, Default: False, Named: true},
				},
			},
		},
	}
}

// boxNative implements the box() function.
// Creates a BoxElement with optional styling properties.
//
// Arguments:
//   - body (positional, content, default: none): The box content
//   - width (named, auto | length, default: auto): Box width
//   - height (named, auto | length, default: auto): Box height
//   - baseline (named, length, default: none): Baseline shift
//   - fill (named, color, default: none): Background fill
//   - stroke (named, length, default: none): Border stroke width
//   - radius (named, length | dictionary, default: none): Corner radius
//   - inset (named, length | dictionary, default: none): Inner padding
//   - outset (named, length | dictionary, default: none): Outer padding
//   - clip (named, bool, default: false): Whether to clip content
func boxNative(vm *Vm, args *Args) (Value, error) {
	elem := &BoxElement{}

	// Get optional body argument
	bodyArg := args.Find("body")
	if bodyArg == nil {
		bodyArg = args.Eat()
	}
	if bodyArg != nil && !IsNone(bodyArg.V) {
		if cv, ok := bodyArg.V.(ContentValue); ok {
			elem.Body = cv.Content
		} else {
			return nil, &TypeMismatchError{
				Expected: "content or none",
				Got:      bodyArg.V.Type().String(),
				Span:     bodyArg.Span,
			}
		}
	}

	// Get optional width argument
	if widthArg := args.Find("width"); widthArg != nil {
		if !IsAuto(widthArg.V) && !IsNone(widthArg.V) {
			if lv, ok := widthArg.V.(LengthValue); ok {
				w := lv.Length.Points
				elem.Width = &w
			} else if rv, ok := widthArg.V.(RelativeValue); ok {
				w := rv.Relative.Abs.Points
				elem.Width = &w
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      widthArg.V.Type().String(),
					Span:     widthArg.Span,
				}
			}
		}
	}

	// Get optional height argument
	if heightArg := args.Find("height"); heightArg != nil {
		if !IsAuto(heightArg.V) && !IsNone(heightArg.V) {
			if lv, ok := heightArg.V.(LengthValue); ok {
				h := lv.Length.Points
				elem.Height = &h
			} else if rv, ok := heightArg.V.(RelativeValue); ok {
				h := rv.Relative.Abs.Points
				elem.Height = &h
			} else if fv, ok := heightArg.V.(FractionValue); ok {
				h := fv.Fraction.Value
				elem.Height = &h
			} else {
				return nil, &TypeMismatchError{
					Expected: "length, fraction, or auto",
					Got:      heightArg.V.Type().String(),
					Span:     heightArg.Span,
				}
			}
		}
	}

	// Get optional baseline argument
	if baselineArg := args.Find("baseline"); baselineArg != nil {
		if !IsAuto(baselineArg.V) && !IsNone(baselineArg.V) {
			if lv, ok := baselineArg.V.(LengthValue); ok {
				b := lv.Length.Points
				elem.Baseline = &b
			} else if rv, ok := baselineArg.V.(RelativeValue); ok {
				b := rv.Relative.Abs.Points
				elem.Baseline = &b
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      baselineArg.V.Type().String(),
					Span:     baselineArg.Span,
				}
			}
		}
	}

	// Get optional fill argument
	if fillArg := args.Find("fill"); fillArg != nil {
		if !IsNone(fillArg.V) {
			if cv, ok := fillArg.V.(ColorValue); ok {
				elem.Fill = &cv
			} else {
				return nil, &TypeMismatchError{
					Expected: "color or none",
					Got:      fillArg.V.Type().String(),
					Span:     fillArg.Span,
				}
			}
		}
	}

	// Get optional stroke argument
	if strokeArg := args.Find("stroke"); strokeArg != nil {
		if !IsNone(strokeArg.V) && !IsAuto(strokeArg.V) {
			if lv, ok := strokeArg.V.(LengthValue); ok {
				s := lv.Length.Points
				elem.Stroke = &s
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      strokeArg.V.Type().String(),
					Span:     strokeArg.Span,
				}
			}
		}
	}

	// Get optional radius argument
	if radiusArg := args.Find("radius"); radiusArg != nil {
		radius, err := parseSidesValue(radiusArg.V, radiusArg.Span)
		if err != nil {
			return nil, err
		}
		elem.Radius = radius
	}

	// Get optional inset argument
	if insetArg := args.Find("inset"); insetArg != nil {
		inset, err := parseSidesValue(insetArg.V, insetArg.Span)
		if err != nil {
			return nil, err
		}
		elem.Inset = inset
	}

	// Get optional outset argument
	if outsetArg := args.Find("outset"); outsetArg != nil {
		outset, err := parseSidesValue(outsetArg.V, outsetArg.Span)
		if err != nil {
			return nil, err
		}
		elem.Outset = outset
	}

	// Get optional clip argument
	if clipArg := args.Find("clip"); clipArg != nil {
		if clipVal, ok := AsBool(clipArg.V); ok {
			elem.Clip = clipVal
		} else {
			return nil, &TypeMismatchError{
				Expected: "bool",
				Got:      clipArg.V.Type().String(),
				Span:     clipArg.Span,
			}
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
// Block Element
// ----------------------------------------------------------------------------

// BlockElement represents a block-level container.
// It wraps content in a block with optional dimensions, fill, stroke, and spacing.
type BlockElement struct {
	// Body is the content inside the block.
	Body Content
	// Width is the block width (nil = auto).
	Width *float64
	// Height is the block height (nil = auto).
	Height *float64
	// Breakable indicates whether the block can break across pages.
	Breakable bool
	// Fill is the background color (nil = none).
	Fill *ColorValue
	// Stroke is the border stroke width (nil = none).
	Stroke *float64
	// Radius is the corner radius.
	Radius *Sides
	// Inset is the inner padding.
	Inset *Sides
	// Outset is the outer padding.
	Outset *Sides
	// Clip indicates whether to clip content to the block.
	Clip bool
	// Above is the spacing above the block (nil = default).
	Above *float64
	// Below is the spacing below the block (nil = default).
	Below *float64
	// Sticky indicates whether the block sticks to adjacent content.
	Sticky bool
}

func (*BlockElement) IsContentElement() {}

// BlockFunc creates the block element function.
func BlockFunc() *Func {
	name := "block"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: blockNative,
			Info: &FuncInfo{
				Name: "block",
				Params: []ParamInfo{
					{Name: "body", Type: TypeContent, Default: None, Named: false},
					{Name: "width", Type: TypeLength, Default: Auto, Named: true},
					{Name: "height", Type: TypeLength, Default: Auto, Named: true},
					{Name: "breakable", Type: TypeBool, Default: True, Named: true},
					{Name: "fill", Type: TypeColor, Default: None, Named: true},
					{Name: "stroke", Type: TypeLength, Default: None, Named: true},
					{Name: "radius", Type: TypeLength, Default: None, Named: true},
					{Name: "inset", Type: TypeLength, Default: None, Named: true},
					{Name: "outset", Type: TypeLength, Default: None, Named: true},
					{Name: "clip", Type: TypeBool, Default: False, Named: true},
					{Name: "above", Type: TypeLength, Default: Auto, Named: true},
					{Name: "below", Type: TypeLength, Default: Auto, Named: true},
					{Name: "sticky", Type: TypeBool, Default: False, Named: true},
				},
			},
		},
	}
}

// blockNative implements the block() function.
// Creates a BlockElement with optional styling properties.
//
// Arguments:
//   - body (positional, content, default: none): The block content
//   - width (named, auto | length, default: auto): Block width
//   - height (named, auto | length, default: auto): Block height
//   - breakable (named, bool, default: true): Whether content can break across pages
//   - fill (named, color, default: none): Background fill
//   - stroke (named, length, default: none): Border stroke width
//   - radius (named, length | dictionary, default: none): Corner radius
//   - inset (named, length | dictionary, default: none): Inner padding
//   - outset (named, length | dictionary, default: none): Outer padding
//   - clip (named, bool, default: false): Whether to clip content
//   - above (named, length, default: auto): Spacing above
//   - below (named, length, default: auto): Spacing below
//   - sticky (named, bool, default: false): Whether to stick to adjacent content
func blockNative(vm *Vm, args *Args) (Value, error) {
	elem := &BlockElement{
		Breakable: true, // default
	}

	// Get optional body argument
	bodyArg := args.Find("body")
	if bodyArg == nil {
		bodyArg = args.Eat()
	}
	if bodyArg != nil && !IsNone(bodyArg.V) {
		if cv, ok := bodyArg.V.(ContentValue); ok {
			elem.Body = cv.Content
		} else {
			return nil, &TypeMismatchError{
				Expected: "content or none",
				Got:      bodyArg.V.Type().String(),
				Span:     bodyArg.Span,
			}
		}
	}

	// Get optional width argument
	if widthArg := args.Find("width"); widthArg != nil {
		if !IsAuto(widthArg.V) && !IsNone(widthArg.V) {
			if lv, ok := widthArg.V.(LengthValue); ok {
				w := lv.Length.Points
				elem.Width = &w
			} else if rv, ok := widthArg.V.(RelativeValue); ok {
				w := rv.Relative.Abs.Points
				elem.Width = &w
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      widthArg.V.Type().String(),
					Span:     widthArg.Span,
				}
			}
		}
	}

	// Get optional height argument
	if heightArg := args.Find("height"); heightArg != nil {
		if !IsAuto(heightArg.V) && !IsNone(heightArg.V) {
			if lv, ok := heightArg.V.(LengthValue); ok {
				h := lv.Length.Points
				elem.Height = &h
			} else if rv, ok := heightArg.V.(RelativeValue); ok {
				h := rv.Relative.Abs.Points
				elem.Height = &h
			} else if fv, ok := heightArg.V.(FractionValue); ok {
				h := fv.Fraction.Value
				elem.Height = &h
			} else {
				return nil, &TypeMismatchError{
					Expected: "length, fraction, or auto",
					Got:      heightArg.V.Type().String(),
					Span:     heightArg.Span,
				}
			}
		}
	}

	// Get optional breakable argument
	if breakableArg := args.Find("breakable"); breakableArg != nil {
		if breakableVal, ok := AsBool(breakableArg.V); ok {
			elem.Breakable = breakableVal
		} else {
			return nil, &TypeMismatchError{
				Expected: "bool",
				Got:      breakableArg.V.Type().String(),
				Span:     breakableArg.Span,
			}
		}
	}

	// Get optional fill argument
	if fillArg := args.Find("fill"); fillArg != nil {
		if !IsNone(fillArg.V) {
			if cv, ok := fillArg.V.(ColorValue); ok {
				elem.Fill = &cv
			} else {
				return nil, &TypeMismatchError{
					Expected: "color or none",
					Got:      fillArg.V.Type().String(),
					Span:     fillArg.Span,
				}
			}
		}
	}

	// Get optional stroke argument
	if strokeArg := args.Find("stroke"); strokeArg != nil {
		if !IsNone(strokeArg.V) && !IsAuto(strokeArg.V) {
			if lv, ok := strokeArg.V.(LengthValue); ok {
				s := lv.Length.Points
				elem.Stroke = &s
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      strokeArg.V.Type().String(),
					Span:     strokeArg.Span,
				}
			}
		}
	}

	// Get optional radius argument
	if radiusArg := args.Find("radius"); radiusArg != nil {
		radius, err := parseSidesValue(radiusArg.V, radiusArg.Span)
		if err != nil {
			return nil, err
		}
		elem.Radius = radius
	}

	// Get optional inset argument
	if insetArg := args.Find("inset"); insetArg != nil {
		inset, err := parseSidesValue(insetArg.V, insetArg.Span)
		if err != nil {
			return nil, err
		}
		elem.Inset = inset
	}

	// Get optional outset argument
	if outsetArg := args.Find("outset"); outsetArg != nil {
		outset, err := parseSidesValue(outsetArg.V, outsetArg.Span)
		if err != nil {
			return nil, err
		}
		elem.Outset = outset
	}

	// Get optional clip argument
	if clipArg := args.Find("clip"); clipArg != nil {
		if clipVal, ok := AsBool(clipArg.V); ok {
			elem.Clip = clipVal
		} else {
			return nil, &TypeMismatchError{
				Expected: "bool",
				Got:      clipArg.V.Type().String(),
				Span:     clipArg.Span,
			}
		}
	}

	// Get optional above argument
	if aboveArg := args.Find("above"); aboveArg != nil {
		if !IsAuto(aboveArg.V) && !IsNone(aboveArg.V) {
			if lv, ok := aboveArg.V.(LengthValue); ok {
				a := lv.Length.Points
				elem.Above = &a
			} else if rv, ok := aboveArg.V.(RelativeValue); ok {
				a := rv.Relative.Abs.Points
				elem.Above = &a
			} else if fv, ok := aboveArg.V.(FractionValue); ok {
				a := fv.Fraction.Value
				elem.Above = &a
			} else {
				return nil, &TypeMismatchError{
					Expected: "length, fraction, or auto",
					Got:      aboveArg.V.Type().String(),
					Span:     aboveArg.Span,
				}
			}
		}
	}

	// Get optional below argument
	if belowArg := args.Find("below"); belowArg != nil {
		if !IsAuto(belowArg.V) && !IsNone(belowArg.V) {
			if lv, ok := belowArg.V.(LengthValue); ok {
				b := lv.Length.Points
				elem.Below = &b
			} else if rv, ok := belowArg.V.(RelativeValue); ok {
				b := rv.Relative.Abs.Points
				elem.Below = &b
			} else if fv, ok := belowArg.V.(FractionValue); ok {
				b := fv.Fraction.Value
				elem.Below = &b
			} else {
				return nil, &TypeMismatchError{
					Expected: "length, fraction, or auto",
					Got:      belowArg.V.Type().String(),
					Span:     belowArg.Span,
				}
			}
		}
	}

	// Get optional sticky argument
	if stickyArg := args.Find("sticky"); stickyArg != nil {
		if stickyVal, ok := AsBool(stickyArg.V); ok {
			elem.Sticky = stickyVal
		} else {
			return nil, &TypeMismatchError{
				Expected: "bool",
				Got:      stickyArg.V.Type().String(),
				Span:     stickyArg.Span,
			}
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
// Pad Element
// ----------------------------------------------------------------------------

// PadElement represents a padding container.
// It adds spacing around its content.
type PadElement struct {
	// Body is the content to pad.
	Body Content
	// Left is the left padding (nil = 0).
	Left *float64
	// Top is the top padding (nil = 0).
	Top *float64
	// Right is the right padding (nil = 0).
	Right *float64
	// Bottom is the bottom padding (nil = 0).
	Bottom *float64
}

func (*PadElement) IsContentElement() {}

// PadFunc creates the pad element function.
func PadFunc() *Func {
	name := "pad"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: padNative,
			Info: &FuncInfo{
				Name: "pad",
				Params: []ParamInfo{
					{Name: "body", Type: TypeContent, Named: false},
					{Name: "left", Type: TypeLength, Default: None, Named: true},
					{Name: "top", Type: TypeLength, Default: None, Named: true},
					{Name: "right", Type: TypeLength, Default: None, Named: true},
					{Name: "bottom", Type: TypeLength, Default: None, Named: true},
					{Name: "x", Type: TypeLength, Default: None, Named: true},
					{Name: "y", Type: TypeLength, Default: None, Named: true},
					{Name: "rest", Type: TypeLength, Default: None, Named: true},
				},
			},
		},
	}
}

// padNative implements the pad() function.
// Creates a PadElement with specified padding values.
//
// Arguments:
//   - body (positional, content): The content to pad
//   - left (named, length, default: none): Left padding
//   - top (named, length, default: none): Top padding
//   - right (named, length, default: none): Right padding
//   - bottom (named, length, default: none): Bottom padding
//   - x (named, length, default: none): Horizontal padding (sets left and right)
//   - y (named, length, default: none): Vertical padding (sets top and bottom)
//   - rest (named, length, default: none): Default padding for unspecified sides
func padNative(vm *Vm, args *Args) (Value, error) {
	elem := &PadElement{}

	// Helper to extract length value
	extractLength := func(arg *syntax.Spanned[Value]) (*float64, error) {
		if arg == nil || IsNone(arg.V) || IsAuto(arg.V) {
			return nil, nil
		}
		if lv, ok := arg.V.(LengthValue); ok {
			pts := lv.Length.Points
			return &pts, nil
		}
		if rv, ok := arg.V.(RelativeValue); ok {
			pts := rv.Relative.Abs.Points
			return &pts, nil
		}
		return nil, &TypeMismatchError{
			Expected: "length or none",
			Got:      arg.V.Type().String(),
			Span:     arg.Span,
		}
	}

	// Get required body argument
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

	// Get rest argument first (sets defaults)
	if restArg := args.Find("rest"); restArg != nil {
		restVal, err := extractLength(restArg)
		if err != nil {
			return nil, err
		}
		if restVal != nil {
			elem.Left = restVal
			elem.Top = restVal
			elem.Right = restVal
			elem.Bottom = restVal
		}
	}

	// Get x argument (sets left and right)
	if xArg := args.Find("x"); xArg != nil {
		xVal, err := extractLength(xArg)
		if err != nil {
			return nil, err
		}
		if xVal != nil {
			elem.Left = xVal
			elem.Right = xVal
		}
	}

	// Get y argument (sets top and bottom)
	if yArg := args.Find("y"); yArg != nil {
		yVal, err := extractLength(yArg)
		if err != nil {
			return nil, err
		}
		if yVal != nil {
			elem.Top = yVal
			elem.Bottom = yVal
		}
	}

	// Individual side arguments override x/y/rest
	if leftArg := args.Find("left"); leftArg != nil {
		leftVal, err := extractLength(leftArg)
		if err != nil {
			return nil, err
		}
		if leftVal != nil {
			elem.Left = leftVal
		}
	}

	if topArg := args.Find("top"); topArg != nil {
		topVal, err := extractLength(topArg)
		if err != nil {
			return nil, err
		}
		if topVal != nil {
			elem.Top = topVal
		}
	}

	if rightArg := args.Find("right"); rightArg != nil {
		rightVal, err := extractLength(rightArg)
		if err != nil {
			return nil, err
		}
		if rightVal != nil {
			elem.Right = rightVal
		}
	}

	if bottomArg := args.Find("bottom"); bottomArg != nil {
		bottomVal, err := extractLength(bottomArg)
		if err != nil {
			return nil, err
		}
		if bottomVal != nil {
			elem.Bottom = bottomVal
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
	// Register stack element function
	scope.DefineFunc("stack", StackFunc())
	// Register align element function
	scope.DefineFunc("align", AlignFunc())
	// Register heading element function
	scope.DefineFunc("heading", HeadingFunc())
	// Register box element function
	scope.DefineFunc("box", BoxFunc())
	// Register block element function
	scope.DefineFunc("block", BlockFunc())
	// Register pad element function
	scope.DefineFunc("pad", PadFunc())
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
		"box":      BoxFunc(),
		"block":    BlockFunc(),
		"pad":      PadFunc(),
	}
}
