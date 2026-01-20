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
// List Element
// ----------------------------------------------------------------------------

// ListFunc creates the list element function.
func ListFunc() *Func {
	name := "list"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: listNative,
			Info: &FuncInfo{
				Name: "list",
				Params: []ParamInfo{
					{Name: "tight", Type: TypeBool, Default: True, Named: true},
					{Name: "marker", Type: TypeContent, Default: None, Named: true},
					{Name: "children", Type: TypeContent, Named: false, Variadic: true},
				},
			},
		},
	}
}

// listNative implements the list() function.
// Creates a ListElement containing the given items.
//
// Arguments:
//   - tight (named, bool, default: true): Whether items have tight spacing
//   - marker (named, content, default: none): Custom marker content
//   - children (positional, variadic, content): The list items
func listNative(vm *Vm, args *Args) (Value, error) {
	// Get optional tight argument (default: true)
	var tight *bool
	if tightArg := args.Find("tight"); tightArg != nil {
		if !IsAuto(tightArg.V) && !IsNone(tightArg.V) {
			tightVal, ok := AsBool(tightArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "bool",
					Got:      tightArg.V.Type().String(),
					Span:     tightArg.Span,
				}
			}
			tight = &tightVal
		}
	}

	// Get optional marker argument
	var marker *Content
	if markerArg := args.Find("marker"); markerArg != nil {
		if !IsNone(markerArg.V) && !IsAuto(markerArg.V) {
			if cv, ok := markerArg.V.(ContentValue); ok {
				marker = &cv.Content
			} else if s, ok := AsStr(markerArg.V); ok {
				// Allow string markers for convenience
				marker = &Content{
					Elements: []ContentElement{&TextElement{Text: s}},
				}
			} else {
				return nil, &TypeMismatchError{
					Expected: "content or string",
					Got:      markerArg.V.Type().String(),
					Span:     markerArg.Span,
				}
			}
		}
	}

	// Collect remaining positional arguments as list items
	var items []*ListItemElement
	for {
		childArg := args.Eat()
		if childArg == nil {
			break
		}

		if cv, ok := childArg.V.(ContentValue); ok {
			// Check if the content contains ListItemElements, otherwise wrap as item
			hasListItems := false
			for _, elem := range cv.Content.Elements {
				if item, ok := elem.(*ListItemElement); ok {
					items = append(items, item)
					hasListItems = true
				}
			}
			// If no list items found, treat the entire content as a single item
			if !hasListItems && len(cv.Content.Elements) > 0 {
				items = append(items, &ListItemElement{Content: cv.Content})
			}
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

	// Create the ListElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{&ListElement{
			Items:  items,
			Tight:  tight,
			Marker: marker,
		}},
	}}, nil
}

// ----------------------------------------------------------------------------
// Enum Element
// ----------------------------------------------------------------------------

// EnumFunc creates the enum element function.
func EnumFunc() *Func {
	name := "enum"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: enumNative,
			Info: &FuncInfo{
				Name: "enum",
				Params: []ParamInfo{
					{Name: "tight", Type: TypeBool, Default: True, Named: true},
					{Name: "numbering", Type: TypeStr, Default: None, Named: true},
					{Name: "start", Type: TypeInt, Default: Int(1), Named: true},
					{Name: "full", Type: TypeBool, Default: False, Named: true},
					{Name: "children", Type: TypeContent, Named: false, Variadic: true},
				},
			},
		},
	}
}

// enumNative implements the enum() function.
// Creates an EnumElement containing the given items.
//
// Arguments:
//   - tight (named, bool, default: true): Whether items have tight spacing
//   - numbering (named, str, default: none): Numbering pattern (e.g., "1.", "a)", "I.")
//   - start (named, int, default: 1): Starting number
//   - full (named, bool, default: false): Whether to display full numbering
//   - children (positional, variadic, content): The enum items
func enumNative(vm *Vm, args *Args) (Value, error) {
	// Get optional tight argument (default: true)
	var tight *bool
	if tightArg := args.Find("tight"); tightArg != nil {
		if !IsAuto(tightArg.V) && !IsNone(tightArg.V) {
			tightVal, ok := AsBool(tightArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "bool",
					Got:      tightArg.V.Type().String(),
					Span:     tightArg.Span,
				}
			}
			tight = &tightVal
		}
	}

	// Get optional numbering argument
	var numbering *string
	if numberingArg := args.Find("numbering"); numberingArg != nil {
		if !IsNone(numberingArg.V) && !IsAuto(numberingArg.V) {
			numStr, ok := AsStr(numberingArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "string",
					Got:      numberingArg.V.Type().String(),
					Span:     numberingArg.Span,
				}
			}
			numbering = &numStr
		}
	}

	// Get optional start argument (default: 1)
	var start *int
	if startArg := args.Find("start"); startArg != nil {
		if !IsAuto(startArg.V) && !IsNone(startArg.V) {
			startVal, ok := AsInt(startArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "integer",
					Got:      startArg.V.Type().String(),
					Span:     startArg.Span,
				}
			}
			s := int(startVal)
			start = &s
		}
	}

	// Get optional full argument (default: false)
	var full *bool
	if fullArg := args.Find("full"); fullArg != nil {
		if !IsAuto(fullArg.V) && !IsNone(fullArg.V) {
			fullVal, ok := AsBool(fullArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "bool",
					Got:      fullArg.V.Type().String(),
					Span:     fullArg.Span,
				}
			}
			full = &fullVal
		}
	}

	// Collect remaining positional arguments as enum items
	var items []*EnumItemElement
	itemNum := 1
	if start != nil {
		itemNum = *start
	}

	for {
		childArg := args.Eat()
		if childArg == nil {
			break
		}

		if cv, ok := childArg.V.(ContentValue); ok {
			// Check if the content contains EnumItemElements, otherwise wrap as item
			hasEnumItems := false
			for _, elem := range cv.Content.Elements {
				if item, ok := elem.(*EnumItemElement); ok {
					// Preserve existing item numbers if set, otherwise auto-number
					if item.Number == 0 {
						item.Number = itemNum
						itemNum++
					}
					items = append(items, item)
					hasEnumItems = true
				}
			}
			// If no enum items found, treat the entire content as a single item
			if !hasEnumItems && len(cv.Content.Elements) > 0 {
				items = append(items, &EnumItemElement{
					Number:  itemNum,
					Content: cv.Content,
				})
				itemNum++
			}
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

	// Create the EnumElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{&EnumElement{
			Items:     items,
			Tight:     tight,
			Numbering: numbering,
			Start:     start,
			Full:      full,
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
					{Name: "body", Type: TypeContent, Named: false},
					{Name: "paper", Type: TypeStr, Default: Auto, Named: true},
					{Name: "width", Type: TypeLength, Default: Auto, Named: true},
					{Name: "height", Type: TypeLength, Default: Auto, Named: true},
					{Name: "flipped", Type: TypeBool, Default: False, Named: true},
					{Name: "margin", Type: TypeLength, Default: Auto, Named: true},
					{Name: "binding", Type: TypeStr, Default: Auto, Named: true},
					{Name: "columns", Type: TypeInt, Default: Int(1), Named: true},
					{Name: "fill", Type: TypeColor, Default: Auto, Named: true},
					{Name: "numbering", Type: TypeStr, Default: None, Named: true},
					{Name: "number-align", Type: TypeStr, Default: Auto, Named: true},
					{Name: "header", Type: TypeContent, Default: Auto, Named: true},
					{Name: "header-ascent", Type: TypeLength, Default: Auto, Named: true},
					{Name: "footer", Type: TypeContent, Default: Auto, Named: true},
					{Name: "footer-descent", Type: TypeLength, Default: Auto, Named: true},
					{Name: "background", Type: TypeContent, Default: None, Named: true},
					{Name: "foreground", Type: TypeContent, Default: None, Named: true},
				},
			},
		},
	}
}

// pageNative implements the page() function.
// Creates a PageElement with the given content and page settings.
//
// Arguments:
//   - body (positional, content): The content on this page
//   - paper (named, str, default: auto): Paper size (e.g., "a4", "us-letter")
//   - width (named, length, default: auto): Page width
//   - height (named, length, default: auto): Page height
//   - flipped (named, bool, default: false): Whether to flip width and height
//   - margin (named, length or dict, default: auto): Page margins
//   - binding (named, str, default: auto): Binding side ("left" or "right")
//   - columns (named, int, default: 1): Number of columns
//   - fill (named, color, default: auto): Page background color
//   - numbering (named, str, default: none): Page numbering pattern
//   - number-align (named, alignment, default: auto): Page number alignment
//   - header (named, content, default: auto): Page header
//   - header-ascent (named, length, default: auto): Header ascent
//   - footer (named, content, default: auto): Page footer
//   - footer-descent (named, length, default: auto): Footer descent
//   - background (named, content, default: none): Background content
//   - foreground (named, content, default: none): Foreground content
func pageNative(vm *Vm, args *Args) (Value, error) {
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

	// Create element with defaults
	elem := &PageElement{
		Body: body,
	}

	// Get optional paper argument
	if paperArg := args.Find("paper"); paperArg != nil {
		if !IsAuto(paperArg.V) && !IsNone(paperArg.V) {
			paper, ok := AsStr(paperArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "string or auto",
					Got:      paperArg.V.Type().String(),
					Span:     paperArg.Span,
				}
			}
			// Validate paper size
			if !isValidPaperSize(paper) {
				return nil, &ConstructorError{
					Message: "unknown paper size: " + paper,
					Span:    paperArg.Span,
				}
			}
			elem.Paper = &paper
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
			flipped, ok := AsBool(flippedArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "bool",
					Got:      flippedArg.V.Type().String(),
					Span:     flippedArg.Span,
				}
			}
			elem.Flipped = &flipped
		}
	}

	// Get optional margin argument
	// This can be a single length (uniform margin) or a dictionary (per-side margins)
	if marginArg := args.Find("margin"); marginArg != nil {
		if !IsAuto(marginArg.V) && !IsNone(marginArg.V) {
			switch v := marginArg.V.(type) {
			case LengthValue:
				// Uniform margin
				m := v.Length.Points
				elem.Margin = &m
			case *DictValue:
				// Per-side margins
				if err := parseMarginDict(v, elem, marginArg.Span); err != nil {
					return nil, err
				}
			case DictValue:
				// Per-side margins
				if err := parseMarginDict(&v, elem, marginArg.Span); err != nil {
					return nil, err
				}
			default:
				return nil, &TypeMismatchError{
					Expected: "length or dictionary",
					Got:      marginArg.V.Type().String(),
					Span:     marginArg.Span,
				}
			}
		}
	}

	// Get optional binding argument
	if bindingArg := args.Find("binding"); bindingArg != nil {
		if !IsAuto(bindingArg.V) && !IsNone(bindingArg.V) {
			binding, ok := AsStr(bindingArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "string or auto",
					Got:      bindingArg.V.Type().String(),
					Span:     bindingArg.Span,
				}
			}
			if binding != "left" && binding != "right" {
				return nil, &ConstructorError{
					Message: "binding must be \"left\" or \"right\"",
					Span:    bindingArg.Span,
				}
			}
			elem.Binding = &binding
		}
	}

	// Get optional columns argument
	if columnsArg := args.Find("columns"); columnsArg != nil {
		if !IsAuto(columnsArg.V) && !IsNone(columnsArg.V) {
			columns, ok := AsInt(columnsArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "integer",
					Got:      columnsArg.V.Type().String(),
					Span:     columnsArg.Span,
				}
			}
			if columns < 1 {
				return nil, &ConstructorError{
					Message: "columns must be at least 1",
					Span:    columnsArg.Span,
				}
			}
			c := int(columns)
			elem.Columns = &c
		}
	}

	// Get optional fill argument (color)
	if fillArg := args.Find("fill"); fillArg != nil {
		if !IsAuto(fillArg.V) && !IsNone(fillArg.V) {
			color, err := parseColor(fillArg.V, fillArg.Span)
			if err != nil {
				return nil, err
			}
			elem.Fill = color
		}
	}

	// Get optional numbering argument
	if numberingArg := args.Find("numbering"); numberingArg != nil {
		if !IsNone(numberingArg.V) && !IsAuto(numberingArg.V) {
			numbering, ok := AsStr(numberingArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "string or none",
					Got:      numberingArg.V.Type().String(),
					Span:     numberingArg.Span,
				}
			}
			elem.Numbering = &numbering
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
		if !IsAuto(headerArg.V) && !IsNone(headerArg.V) {
			if cv, ok := headerArg.V.(ContentValue); ok {
				elem.Header = &cv.Content
			} else {
				return nil, &TypeMismatchError{
					Expected: "content or auto",
					Got:      headerArg.V.Type().String(),
					Span:     headerArg.Span,
				}
			}
		}
	}

	// Get optional header-ascent argument
	if headerAscentArg := args.Find("header-ascent"); headerAscentArg != nil {
		if !IsAuto(headerAscentArg.V) && !IsNone(headerAscentArg.V) {
			ascent, err := parseLengthOrRatio(headerAscentArg.V, headerAscentArg.Span)
			if err != nil {
				return nil, err
			}
			elem.HeaderAscent = &ascent
		}
	}

	// Get optional footer argument
	if footerArg := args.Find("footer"); footerArg != nil {
		if !IsAuto(footerArg.V) && !IsNone(footerArg.V) {
			if cv, ok := footerArg.V.(ContentValue); ok {
				elem.Footer = &cv.Content
			} else {
				return nil, &TypeMismatchError{
					Expected: "content or auto",
					Got:      footerArg.V.Type().String(),
					Span:     footerArg.Span,
				}
			}
		}
	}

	// Get optional footer-descent argument
	if footerDescentArg := args.Find("footer-descent"); footerDescentArg != nil {
		if !IsAuto(footerDescentArg.V) && !IsNone(footerDescentArg.V) {
			descent, err := parseLengthOrRatio(footerDescentArg.V, footerDescentArg.Span)
			if err != nil {
				return nil, err
			}
			elem.FooterDescent = &descent
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

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Create the PageElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}

// isValidPaperSize checks if a paper size name is valid.
func isValidPaperSize(paper string) bool {
	validSizes := map[string]bool{
		"a0": true, "a1": true, "a2": true, "a3": true, "a4": true,
		"a5": true, "a6": true, "a7": true, "a8": true, "a9": true, "a10": true, "a11": true,
		"iso-b1": true, "iso-b2": true, "iso-b3": true, "iso-b4": true,
		"iso-b5": true, "iso-b6": true, "iso-b7": true, "iso-b8": true,
		"iso-c3": true, "iso-c4": true, "iso-c5": true, "iso-c6": true, "iso-c7": true, "iso-c8": true,
		"din-d3": true, "din-d4": true, "din-d5": true, "din-d6": true, "din-d7": true, "din-d8": true,
		"sis-g5": true, "sis-e5": true,
		"ansi-a": true, "ansi-b": true, "ansi-c": true, "ansi-d": true, "ansi-e": true,
		"arch-a": true, "arch-b": true, "arch-c": true, "arch-d": true, "arch-e": true, "arch-e1": true,
		"jis-b0": true, "jis-b1": true, "jis-b2": true, "jis-b3": true, "jis-b4": true,
		"jis-b5": true, "jis-b6": true, "jis-b7": true, "jis-b8": true, "jis-b9": true,
		"jis-b10": true, "jis-b11": true,
		"sac-d0": true, "sac-d1": true, "sac-d2": true, "sac-d3": true, "sac-d4": true,
		"sac-d5": true, "sac-d6": true,
		"iso-id-1": true, "iso-id-2": true, "iso-id-3": true,
		"asia-f4": true,
		"jp-shiroku-ban-4": true, "jp-shiroku-ban-5": true, "jp-shiroku-ban-6": true,
		"jp-kiku-4": true, "jp-kiku-5": true,
		"jp-business-card": true,
		"cn-business-card": true,
		"eu-business-card": true,
		"us-business-card": true,
		"jp-hagaki": true,
		"us-letter": true, "us-legal": true, "us-tabloid": true, "us-executive": true,
		"us-foolscap-folio": true, "us-statement": true, "us-ledger": true, "us-oficio": true,
		"us-gov-letter": true, "us-gov-legal": true,
		"presentation-16-9": true, "presentation-4-3": true,
	}
	return validSizes[paper]
}

// parseMarginDict parses a margin dictionary and sets per-side margins on the element.
func parseMarginDict(dict *DictValue, elem *PageElement, span syntax.Span) error {
	for _, key := range dict.Keys() {
		val, _ := dict.Get(key)
		switch key {
		case "left", "x":
			if lv, ok := val.(LengthValue); ok {
				m := lv.Length.Points
				if key == "x" {
					elem.MarginLeft = &m
					elem.MarginRight = &m
				} else {
					elem.MarginLeft = &m
				}
			} else if !IsAuto(val) && !IsNone(val) {
				return &TypeMismatchError{
					Expected: "length or auto",
					Got:      val.Type().String(),
					Span:     span,
				}
			}
		case "right":
			if lv, ok := val.(LengthValue); ok {
				m := lv.Length.Points
				elem.MarginRight = &m
			} else if !IsAuto(val) && !IsNone(val) {
				return &TypeMismatchError{
					Expected: "length or auto",
					Got:      val.Type().String(),
					Span:     span,
				}
			}
		case "top", "y":
			if lv, ok := val.(LengthValue); ok {
				m := lv.Length.Points
				if key == "y" {
					elem.MarginTop = &m
					elem.MarginBottom = &m
				} else {
					elem.MarginTop = &m
				}
			} else if !IsAuto(val) && !IsNone(val) {
				return &TypeMismatchError{
					Expected: "length or auto",
					Got:      val.Type().String(),
					Span:     span,
				}
			}
		case "bottom":
			if lv, ok := val.(LengthValue); ok {
				m := lv.Length.Points
				elem.MarginBottom = &m
			} else if !IsAuto(val) && !IsNone(val) {
				return &TypeMismatchError{
					Expected: "length or auto",
					Got:      val.Type().String(),
					Span:     span,
				}
			}
		case "rest":
			// "rest" sets all unset margins
			if lv, ok := val.(LengthValue); ok {
				m := lv.Length.Points
				if elem.MarginLeft == nil {
					elem.MarginLeft = &m
				}
				if elem.MarginRight == nil {
					elem.MarginRight = &m
				}
				if elem.MarginTop == nil {
					elem.MarginTop = &m
				}
				if elem.MarginBottom == nil {
					elem.MarginBottom = &m
				}
			} else if !IsAuto(val) && !IsNone(val) {
				return &TypeMismatchError{
					Expected: "length or auto",
					Got:      val.Type().String(),
					Span:     span,
				}
			}
		default:
			return &ConstructorError{
				Message: "unknown margin key: " + key,
				Span:    span,
			}
		}
	}
	return nil
}

// parseColor parses a color value.
func parseColor(v Value, span syntax.Span) (*Color, error) {
	// Handle ColorValue type if it exists
	if cv, ok := v.(ColorValue); ok {
		return &Color{
			R: cv.Color.R,
			G: cv.Color.G,
			B: cv.Color.B,
			A: cv.Color.A,
		}, nil
	}

	// Handle string color names
	if s, ok := AsStr(v); ok {
		color := parseColorName(s)
		if color != nil {
			return color, nil
		}
		return nil, &ConstructorError{
			Message: "unknown color: " + s,
			Span:    span,
		}
	}

	return nil, &TypeMismatchError{
		Expected: "color",
		Got:      v.Type().String(),
		Span:     span,
	}
}

// parseColorName parses a color name and returns a Color, or nil if unknown.
func parseColorName(name string) *Color {
	colors := map[string]*Color{
		"black":   {R: 0, G: 0, B: 0, A: 255},
		"white":   {R: 255, G: 255, B: 255, A: 255},
		"red":     {R: 255, G: 0, B: 0, A: 255},
		"green":   {R: 0, G: 255, B: 0, A: 255},
		"blue":    {R: 0, G: 0, B: 255, A: 255},
		"yellow":  {R: 255, G: 255, B: 0, A: 255},
		"cyan":    {R: 0, G: 255, B: 255, A: 255},
		"magenta": {R: 255, G: 0, B: 255, A: 255},
		"gray":    {R: 128, G: 128, B: 128, A: 255},
		"grey":    {R: 128, G: 128, B: 128, A: 255},
	}
	return colors[name]
}

// parseLengthOrRatio parses a length or ratio value and returns points.
// For ratios, returns the ratio as a fraction (0.0-1.0).
func parseLengthOrRatio(v Value, span syntax.Span) (float64, error) {
	switch val := v.(type) {
	case LengthValue:
		return val.Length.Points, nil
	case RatioValue:
		return val.Ratio.Value, nil
	case RelativeValue:
		// For relative values, return the ratio part
		return val.Relative.Rel.Value, nil
	default:
		return 0, &TypeMismatchError{
			Expected: "length or ratio",
			Got:      v.Type().String(),
			Span:     span,
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
	// Register stack element function
	scope.DefineFunc("stack", StackFunc())
	// Register align element function
	scope.DefineFunc("align", AlignFunc())
	// Register heading element function
	scope.DefineFunc("heading", HeadingFunc())
	// Register list element function
	scope.DefineFunc("list", ListFunc())
	// Register enum element function
	scope.DefineFunc("enum", EnumFunc())
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
		"list":     ListFunc(),
		"enum":     EnumFunc(),
		"page":     PageFunc(),
	}
}
