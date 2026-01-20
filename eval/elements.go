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
// Page Element
// ----------------------------------------------------------------------------

// PageFunc creates the page element function.
func PageFunc() *Func {
	name := "page"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: pageNative,
			Info: &FuncInfo{
				Name: "page",
				Params: []ParamInfo{
					{Name: "body", Type: TypeContent, Default: None, Named: false},
					{Name: "width", Type: TypeLength, Default: Auto, Named: true},
					{Name: "height", Type: TypeLength, Default: Auto, Named: true},
					{Name: "flipped", Type: TypeBool, Default: False, Named: true},
					{Name: "margin", Type: TypeDyn, Default: Auto, Named: true},
					{Name: "fill", Type: TypeColor, Default: None, Named: true},
					{Name: "numbering", Type: TypeStr, Default: None, Named: true},
					{Name: "number-align", Type: TypeDyn, Default: Auto, Named: true},
					{Name: "header", Type: TypeContent, Default: None, Named: true},
					{Name: "footer", Type: TypeContent, Default: None, Named: true},
					{Name: "header-ascent", Type: TypeLength, Default: Auto, Named: true},
					{Name: "footer-descent", Type: TypeLength, Default: Auto, Named: true},
					{Name: "background", Type: TypeContent, Default: None, Named: true},
					{Name: "foreground", Type: TypeContent, Default: None, Named: true},
					{Name: "columns", Type: TypeInt, Default: Int(1), Named: true},
					{Name: "binding", Type: TypeDyn, Default: Auto, Named: true},
				},
			},
		},
	}
}

// pageNative implements the page() function.
// Creates a PageElement with the given configuration.
//
// Arguments:
//   - body (positional, content or none): Content for this page (creates page break)
//   - width (named, length, default: auto): Page width
//   - height (named, length, default: auto): Page height
//   - flipped (named, bool, default: false): Swap width and height
//   - margin (named, length or dict, default: auto): Page margins
//   - fill (named, color, default: none): Background fill
//   - numbering (named, str, default: none): Page numbering pattern
//   - number-align (named, alignment, default: auto): Page number alignment
//   - header (named, content, default: none): Header content
//   - footer (named, content, default: none): Footer content
//   - header-ascent (named, length, default: auto): Header ascent space
//   - footer-descent (named, length, default: auto): Footer descent space
//   - background (named, content, default: none): Background content
//   - foreground (named, content, default: none): Foreground content
//   - columns (named, int, default: 1): Number of columns
//   - binding (named, alignment, default: auto): Binding side
func pageNative(vm *Vm, args *Args) (Value, error) {
	elem := &PageElement{}

	// Get optional body argument (positional)
	if bodyArg := args.Find("body"); bodyArg != nil {
		if !IsNone(bodyArg.V) {
			if cv, ok := bodyArg.V.(ContentValue); ok {
				elem.Body = &cv.Content
			} else {
				return nil, &TypeMismatchError{
					Expected: "content or none",
					Got:      bodyArg.V.Type().String(),
					Span:     bodyArg.Span,
				}
			}
		}
	} else if bodyArg := args.Eat(); bodyArg != nil {
		if !IsNone(bodyArg.V) {
			if cv, ok := bodyArg.V.(ContentValue); ok {
				elem.Body = &cv.Content
			} else {
				return nil, &TypeMismatchError{
					Expected: "content or none",
					Got:      bodyArg.V.Type().String(),
					Span:     bodyArg.Span,
				}
			}
		}
	}

	// Get optional width argument
	if widthArg := args.Find("width"); widthArg != nil {
		if !IsAuto(widthArg.V) && !IsNone(widthArg.V) {
			if lv, ok := widthArg.V.(LengthValue); ok {
				w := lv.Length.Points
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
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      heightArg.V.Type().String(),
					Span:     heightArg.Span,
				}
			}
		}
	}

	// Get optional flipped argument
	if flippedArg := args.Find("flipped"); flippedArg != nil {
		if !IsAuto(flippedArg.V) && !IsNone(flippedArg.V) {
			if fv, ok := AsBool(flippedArg.V); ok {
				elem.Flipped = &fv
			} else {
				return nil, &TypeMismatchError{
					Expected: "bool",
					Got:      flippedArg.V.Type().String(),
					Span:     flippedArg.Span,
				}
			}
		}
	}

	// Get optional margin argument (can be length or dict)
	if marginArg := args.Find("margin"); marginArg != nil {
		if !IsAuto(marginArg.V) && !IsNone(marginArg.V) {
			margin, err := parsePageMargin(marginArg.V, marginArg.Span)
			if err != nil {
				return nil, err
			}
			elem.Margin = margin
		}
	}

	// Get optional fill argument
	if fillArg := args.Find("fill"); fillArg != nil {
		if !IsNone(fillArg.V) && !IsAuto(fillArg.V) {
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

	// Get optional numbering argument
	if numberingArg := args.Find("numbering"); numberingArg != nil {
		if !IsNone(numberingArg.V) && !IsAuto(numberingArg.V) {
			if nv, ok := AsStr(numberingArg.V); ok {
				elem.Numbering = &nv
			} else {
				return nil, &TypeMismatchError{
					Expected: "string or none",
					Got:      numberingArg.V.Type().String(),
					Span:     numberingArg.Span,
				}
			}
		}
	}

	// Get optional number-align argument
	if numberAlignArg := args.Find("number-align"); numberAlignArg != nil {
		if !IsAuto(numberAlignArg.V) && !IsNone(numberAlignArg.V) {
			alignment, err := parseAlignment(numberAlignArg.V, numberAlignArg.Span)
			if err != nil {
				return nil, err
			}
			elem.NumberAlign = &alignment
		}
	}

	// Get optional header argument
	if headerArg := args.Find("header"); headerArg != nil {
		if !IsNone(headerArg.V) && !IsAuto(headerArg.V) {
			if cv, ok := headerArg.V.(ContentValue); ok {
				elem.Header = &cv.Content
			} else {
				return nil, &TypeMismatchError{
					Expected: "content or none",
					Got:      headerArg.V.Type().String(),
					Span:     headerArg.Span,
				}
			}
		}
	}

	// Get optional footer argument
	if footerArg := args.Find("footer"); footerArg != nil {
		if !IsNone(footerArg.V) && !IsAuto(footerArg.V) {
			if cv, ok := footerArg.V.(ContentValue); ok {
				elem.Footer = &cv.Content
			} else {
				return nil, &TypeMismatchError{
					Expected: "content or none",
					Got:      footerArg.V.Type().String(),
					Span:     footerArg.Span,
				}
			}
		}
	}

	// Get optional header-ascent argument
	if headerAscentArg := args.Find("header-ascent"); headerAscentArg != nil {
		if !IsAuto(headerAscentArg.V) && !IsNone(headerAscentArg.V) {
			if lv, ok := headerAscentArg.V.(LengthValue); ok {
				ha := lv.Length.Points
				elem.HeaderAscent = &ha
			} else if rv, ok := headerAscentArg.V.(RatioValue); ok {
				// Ratio values are relative to the margin
				ha := rv.Ratio.Value * 100 // Store as percentage for later resolution
				elem.HeaderAscent = &ha
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      headerAscentArg.V.Type().String(),
					Span:     headerAscentArg.Span,
				}
			}
		}
	}

	// Get optional footer-descent argument
	if footerDescentArg := args.Find("footer-descent"); footerDescentArg != nil {
		if !IsAuto(footerDescentArg.V) && !IsNone(footerDescentArg.V) {
			if lv, ok := footerDescentArg.V.(LengthValue); ok {
				fd := lv.Length.Points
				elem.FooterDescent = &fd
			} else if rv, ok := footerDescentArg.V.(RatioValue); ok {
				// Ratio values are relative to the margin
				fd := rv.Ratio.Value * 100 // Store as percentage for later resolution
				elem.FooterDescent = &fd
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      footerDescentArg.V.Type().String(),
					Span:     footerDescentArg.Span,
				}
			}
		}
	}

	// Get optional background argument
	if backgroundArg := args.Find("background"); backgroundArg != nil {
		if !IsNone(backgroundArg.V) && !IsAuto(backgroundArg.V) {
			if cv, ok := backgroundArg.V.(ContentValue); ok {
				elem.Background = &cv.Content
			} else {
				return nil, &TypeMismatchError{
					Expected: "content or none",
					Got:      backgroundArg.V.Type().String(),
					Span:     backgroundArg.Span,
				}
			}
		}
	}

	// Get optional foreground argument
	if foregroundArg := args.Find("foreground"); foregroundArg != nil {
		if !IsNone(foregroundArg.V) && !IsAuto(foregroundArg.V) {
			if cv, ok := foregroundArg.V.(ContentValue); ok {
				elem.Foreground = &cv.Content
			} else {
				return nil, &TypeMismatchError{
					Expected: "content or none",
					Got:      foregroundArg.V.Type().String(),
					Span:     foregroundArg.Span,
				}
			}
		}
	}

	// Get optional columns argument
	if columnsArg := args.Find("columns"); columnsArg != nil {
		if !IsAuto(columnsArg.V) && !IsNone(columnsArg.V) {
			if cv, ok := AsInt(columnsArg.V); ok {
				cols := int(cv)
				if cols < 1 {
					return nil, &ConstructorError{
						Message: "columns must be at least 1",
						Span:    columnsArg.Span,
					}
				}
				elem.Columns = &cols
			} else {
				return nil, &TypeMismatchError{
					Expected: "integer",
					Got:      columnsArg.V.Type().String(),
					Span:     columnsArg.Span,
				}
			}
		}
	}

	// Get optional binding argument
	if bindingArg := args.Find("binding"); bindingArg != nil {
		if !IsAuto(bindingArg.V) && !IsNone(bindingArg.V) {
			if sv, ok := AsStr(bindingArg.V); ok {
				// Validate binding value
				if sv != "left" && sv != "right" {
					return nil, &TypeMismatchError{
						Expected: "\"left\" or \"right\"",
						Got:      "\"" + sv + "\"",
						Span:     bindingArg.Span,
					}
				}
				elem.Binding = &sv
			} else if av, ok := bindingArg.V.(AlignmentValue); ok {
				// Handle alignment values (left, right)
				if av.Alignment.H != nil {
					if *av.Alignment.H < 0 {
						binding := "left"
						elem.Binding = &binding
					} else if *av.Alignment.H > 0 {
						binding := "right"
						elem.Binding = &binding
					}
				}
			} else {
				return nil, &TypeMismatchError{
					Expected: "alignment or auto",
					Got:      bindingArg.V.Type().String(),
					Span:     bindingArg.Span,
				}
			}
		}
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Create the PageElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}

// parsePageMargin parses margin configuration from a value.
// Supports:
//   - Single length: applied to all sides
//   - Dictionary with keys: top, bottom, left, right, inside, outside, x, y, rest
func parsePageMargin(v Value, span syntax.Span) (*PageMargin, error) {
	margin := &PageMargin{}

	// Single length value applies to all sides
	if lv, ok := v.(LengthValue); ok {
		m := lv.Length.Points
		margin.Top = &m
		margin.Bottom = &m
		margin.Left = &m
		margin.Right = &m
		return margin, nil
	}

	// Dictionary with named margins
	if dv, ok := AsDict(v); ok {
		// Extract named margin values
		if val, exists := dv.Get("top"); exists {
			if lv, ok := val.(LengthValue); ok {
				m := lv.Length.Points
				margin.Top = &m
			} else if !IsAuto(val) && !IsNone(val) {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      val.Type().String(),
					Span:     span,
				}
			}
		}

		if val, exists := dv.Get("bottom"); exists {
			if lv, ok := val.(LengthValue); ok {
				m := lv.Length.Points
				margin.Bottom = &m
			} else if !IsAuto(val) && !IsNone(val) {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      val.Type().String(),
					Span:     span,
				}
			}
		}

		if val, exists := dv.Get("left"); exists {
			if lv, ok := val.(LengthValue); ok {
				m := lv.Length.Points
				margin.Left = &m
			} else if !IsAuto(val) && !IsNone(val) {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      val.Type().String(),
					Span:     span,
				}
			}
		}

		if val, exists := dv.Get("right"); exists {
			if lv, ok := val.(LengthValue); ok {
				m := lv.Length.Points
				margin.Right = &m
			} else if !IsAuto(val) && !IsNone(val) {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      val.Type().String(),
					Span:     span,
				}
			}
		}

		if val, exists := dv.Get("inside"); exists {
			if lv, ok := val.(LengthValue); ok {
				m := lv.Length.Points
				margin.Inside = &m
			} else if !IsAuto(val) && !IsNone(val) {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      val.Type().String(),
					Span:     span,
				}
			}
		}

		if val, exists := dv.Get("outside"); exists {
			if lv, ok := val.(LengthValue); ok {
				m := lv.Length.Points
				margin.Outside = &m
			} else if !IsAuto(val) && !IsNone(val) {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      val.Type().String(),
					Span:     span,
				}
			}
		}

		if val, exists := dv.Get("x"); exists {
			if lv, ok := val.(LengthValue); ok {
				m := lv.Length.Points
				margin.X = &m
			} else if !IsAuto(val) && !IsNone(val) {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      val.Type().String(),
					Span:     span,
				}
			}
		}

		if val, exists := dv.Get("y"); exists {
			if lv, ok := val.(LengthValue); ok {
				m := lv.Length.Points
				margin.Y = &m
			} else if !IsAuto(val) && !IsNone(val) {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      val.Type().String(),
					Span:     span,
				}
			}
		}

		if val, exists := dv.Get("rest"); exists {
			if lv, ok := val.(LengthValue); ok {
				m := lv.Length.Points
				margin.Rest = &m
			} else if !IsAuto(val) && !IsNone(val) {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      val.Type().String(),
					Span:     span,
				}
			}
		}

		return margin, nil
	}

	return nil, &TypeMismatchError{
		Expected: "length or dictionary",
		Got:      v.Type().String(),
		Span:     span,
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
	// Register stack element function
	scope.DefineFunc("stack", StackFunc())
	// Register align element function
	scope.DefineFunc("align", AlignFunc())
	// Register heading element function
	scope.DefineFunc("heading", HeadingFunc())
	// Register page element function
	scope.DefineFunc("page", PageFunc())
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
		"page":     PageFunc(),
	}
}
