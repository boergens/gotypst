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
// Table Element
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
					{Name: "fill", Type: TypeContent, Default: None, Named: true},
					{Name: "stroke", Type: TypeContent, Default: None, Named: true},
					{Name: "inset", Type: TypeLength, Default: None, Named: true},
					{Name: "children", Type: TypeContent, Named: false, Variadic: true},
				},
			},
		},
	}
}

// tableNative implements the table() function.
// Creates a TableElement with the given columns and cells.
//
// Arguments:
//   - columns (named, array, default: auto): Column sizing specifications
//   - rows (named, array, default: auto): Row sizing specifications
//   - gutter (named, length, default: none): Spacing between all cells
//   - column-gutter (named, length, default: auto): Column spacing (overrides gutter)
//   - row-gutter (named, length, default: auto): Row spacing (overrides gutter)
//   - align (named, alignment, default: auto): Default cell alignment
//   - fill (named, content, default: none): Default cell fill
//   - stroke (named, content, default: none): Default cell stroke
//   - inset (named, length, default: none): Default cell padding
//   - children (positional, variadic, content): The cells
func tableNative(vm *Vm, args *Args) (Value, error) {
	// Get optional columns argument
	var columns []TableSizing
	if columnsArg := args.Find("columns"); columnsArg != nil {
		if !IsAuto(columnsArg.V) && !IsNone(columnsArg.V) {
			cols, err := parseTableSizings(columnsArg.V, columnsArg.Span)
			if err != nil {
				return nil, err
			}
			columns = cols
		}
	}

	// Get optional rows argument
	var rows []TableSizing
	if rowsArg := args.Find("rows"); rowsArg != nil {
		if !IsAuto(rowsArg.V) && !IsNone(rowsArg.V) {
			rs, err := parseTableSizings(rowsArg.V, rowsArg.Span)
			if err != nil {
				return nil, err
			}
			rows = rs
		}
	}

	// Get gutter (default for both column and row gutter)
	var gutter *float64
	if gutterArg := args.Find("gutter"); gutterArg != nil {
		if !IsNone(gutterArg.V) && !IsAuto(gutterArg.V) {
			if lv, ok := gutterArg.V.(LengthValue); ok {
				g := lv.Length.Points
				gutter = &g
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      gutterArg.V.Type().String(),
					Span:     gutterArg.Span,
				}
			}
		}
	}

	// Get column-gutter (overrides gutter)
	var columnGutter *float64
	if cgArg := args.Find("column-gutter"); cgArg != nil {
		if !IsNone(cgArg.V) && !IsAuto(cgArg.V) {
			if lv, ok := cgArg.V.(LengthValue); ok {
				cg := lv.Length.Points
				columnGutter = &cg
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      cgArg.V.Type().String(),
					Span:     cgArg.Span,
				}
			}
		}
	}
	// If column-gutter not set but gutter is, use gutter
	if columnGutter == nil && gutter != nil {
		columnGutter = gutter
	}

	// Get row-gutter (overrides gutter)
	var rowGutter *float64
	if rgArg := args.Find("row-gutter"); rgArg != nil {
		if !IsNone(rgArg.V) && !IsAuto(rgArg.V) {
			if lv, ok := rgArg.V.(LengthValue); ok {
				rg := lv.Length.Points
				rowGutter = &rg
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      rgArg.V.Type().String(),
					Span:     rgArg.Span,
				}
			}
		}
	}
	// If row-gutter not set but gutter is, use gutter
	if rowGutter == nil && gutter != nil {
		rowGutter = gutter
	}

	// Get optional align argument
	var align *Alignment2D
	if alignArg := args.Find("align"); alignArg != nil {
		if !IsAuto(alignArg.V) && !IsNone(alignArg.V) {
			a, err := parseAlignment(alignArg.V, alignArg.Span)
			if err != nil {
				return nil, err
			}
			align = &a
		}
	}

	// Get optional fill argument
	var fill *Content
	if fillArg := args.Find("fill"); fillArg != nil {
		if !IsNone(fillArg.V) && !IsAuto(fillArg.V) {
			if cv, ok := fillArg.V.(ContentValue); ok {
				fill = &cv.Content
			} else {
				return nil, &TypeMismatchError{
					Expected: "content or none",
					Got:      fillArg.V.Type().String(),
					Span:     fillArg.Span,
				}
			}
		}
	}

	// Get optional stroke argument
	var stroke *Content
	if strokeArg := args.Find("stroke"); strokeArg != nil {
		if !IsNone(strokeArg.V) && !IsAuto(strokeArg.V) {
			if cv, ok := strokeArg.V.(ContentValue); ok {
				stroke = &cv.Content
			} else {
				return nil, &TypeMismatchError{
					Expected: "content or none",
					Got:      strokeArg.V.Type().String(),
					Span:     strokeArg.Span,
				}
			}
		}
	}

	// Get optional inset argument
	var inset *float64
	if insetArg := args.Find("inset"); insetArg != nil {
		if !IsNone(insetArg.V) && !IsAuto(insetArg.V) {
			if lv, ok := insetArg.V.(LengthValue); ok {
				i := lv.Length.Points
				inset = &i
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      insetArg.V.Type().String(),
					Span:     insetArg.Span,
				}
			}
		}
	}

	// Collect remaining positional arguments as cells
	var cells []*TableCellElement
	for {
		childArg := args.Eat()
		if childArg == nil {
			break
		}

		if cv, ok := childArg.V.(ContentValue); ok {
			// Check if the content contains TableCellElements
			hasCells := false
			for _, elem := range cv.Content.Elements {
				if cell, ok := elem.(*TableCellElement); ok {
					cells = append(cells, cell)
					hasCells = true
				}
			}
			// If no cells found, treat the entire content as a cell
			if !hasCells && len(cv.Content.Elements) > 0 {
				cells = append(cells, &TableCellElement{
					Content: cv.Content,
					X:       -1, // Auto-placement
					Y:       -1,
				})
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

	// Create the TableElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{&TableElement{
			Columns:      columns,
			Rows:         rows,
			Cells:        cells,
			Align:        align,
			Fill:         fill,
			Stroke:       stroke,
			Inset:        inset,
			ColumnGutter: columnGutter,
			RowGutter:    rowGutter,
		}},
	}}, nil
}

// parseTableSizings parses an array of sizing values for table columns/rows.
func parseTableSizings(v Value, span syntax.Span) ([]TableSizing, error) {
	// Handle integer as column count (creates that many auto columns)
	if intVal, ok := AsInt(v); ok {
		count := int(intVal)
		if count < 1 {
			return nil, &ConstructorError{
				Message: "column count must be at least 1",
				Span:    span,
			}
		}
		sizings := make([]TableSizing, count)
		for i := range sizings {
			sizings[i] = TableSizing{Auto: true}
		}
		return sizings, nil
	}

	// Handle array of sizing values
	arr, ok := v.(ArrayValue)
	if !ok {
		return nil, &TypeMismatchError{
			Expected: "array or integer",
			Got:      v.Type().String(),
			Span:     span,
		}
	}

	var sizings []TableSizing
	for _, item := range arr {
		sizing, err := parseTableSizing(item, span)
		if err != nil {
			return nil, err
		}
		sizings = append(sizings, sizing)
	}
	return sizings, nil
}

// parseTableSizing parses a single sizing value.
func parseTableSizing(v Value, span syntax.Span) (TableSizing, error) {
	// Handle "auto"
	if IsAuto(v) {
		return TableSizing{Auto: true}, nil
	}

	// Handle length
	if lv, ok := v.(LengthValue); ok {
		return TableSizing{Points: lv.Length.Points}, nil
	}

	// Handle fraction
	if fv, ok := v.(FractionValue); ok {
		return TableSizing{Fr: fv.Fraction.Value}, nil
	}

	// Handle relative value (percentage)
	if rv, ok := v.(RelativeValue); ok {
		// Convert relative to fractional units (100% = 1fr)
		return TableSizing{Fr: rv.Relative.Rel.Value}, nil
	}

	return TableSizing{}, &TypeMismatchError{
		Expected: "auto, length, or fraction",
		Got:      v.Type().String(),
		Span:     span,
	}
}

// TableCellFunc creates the table.cell element function.
func TableCellFunc() *Func {
	name := "table.cell"
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
					{Name: "align", Type: TypeStr, Default: Auto, Named: true},
					{Name: "fill", Type: TypeContent, Default: None, Named: true},
					{Name: "stroke", Type: TypeContent, Default: None, Named: true},
					{Name: "inset", Type: TypeLength, Default: None, Named: true},
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
//   - x (named, int, default: auto): Column index
//   - y (named, int, default: auto): Row index
//   - colspan (named, int, default: 1): Number of columns to span
//   - rowspan (named, int, default: 1): Number of rows to span
//   - align (named, alignment, default: auto): Cell alignment
//   - fill (named, content, default: none): Cell fill
//   - stroke (named, content, default: none): Cell stroke
//   - inset (named, length, default: none): Cell padding
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

	// Get optional x argument (default: auto = -1)
	x := -1
	if xArg := args.Find("x"); xArg != nil {
		if !IsAuto(xArg.V) {
			xVal, ok := AsInt(xArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "integer or auto",
					Got:      xArg.V.Type().String(),
					Span:     xArg.Span,
				}
			}
			x = int(xVal)
		}
	}

	// Get optional y argument (default: auto = -1)
	y := -1
	if yArg := args.Find("y"); yArg != nil {
		if !IsAuto(yArg.V) {
			yVal, ok := AsInt(yArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "integer or auto",
					Got:      yArg.V.Type().String(),
					Span:     yArg.Span,
				}
			}
			y = int(yVal)
		}
	}

	// Get optional colspan argument (default: 1)
	colspan := 1
	if colspanArg := args.Find("colspan"); colspanArg != nil {
		if !IsAuto(colspanArg.V) {
			csVal, ok := AsInt(colspanArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "integer",
					Got:      colspanArg.V.Type().String(),
					Span:     colspanArg.Span,
				}
			}
			colspan = int(csVal)
			if colspan < 1 {
				return nil, &ConstructorError{
					Message: "colspan must be at least 1",
					Span:    colspanArg.Span,
				}
			}
		}
	}

	// Get optional rowspan argument (default: 1)
	rowspan := 1
	if rowspanArg := args.Find("rowspan"); rowspanArg != nil {
		if !IsAuto(rowspanArg.V) {
			rsVal, ok := AsInt(rowspanArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "integer",
					Got:      rowspanArg.V.Type().String(),
					Span:     rowspanArg.Span,
				}
			}
			rowspan = int(rsVal)
			if rowspan < 1 {
				return nil, &ConstructorError{
					Message: "rowspan must be at least 1",
					Span:    rowspanArg.Span,
				}
			}
		}
	}

	// Get optional align argument
	var align *Alignment2D
	if alignArg := args.Find("align"); alignArg != nil {
		if !IsAuto(alignArg.V) && !IsNone(alignArg.V) {
			a, err := parseAlignment(alignArg.V, alignArg.Span)
			if err != nil {
				return nil, err
			}
			align = &a
		}
	}

	// Get optional fill argument
	var fill *Content
	if fillArg := args.Find("fill"); fillArg != nil {
		if !IsNone(fillArg.V) && !IsAuto(fillArg.V) {
			if cv, ok := fillArg.V.(ContentValue); ok {
				fill = &cv.Content
			} else {
				return nil, &TypeMismatchError{
					Expected: "content or none",
					Got:      fillArg.V.Type().String(),
					Span:     fillArg.Span,
				}
			}
		}
	}

	// Get optional stroke argument
	var stroke *Content
	if strokeArg := args.Find("stroke"); strokeArg != nil {
		if !IsNone(strokeArg.V) && !IsAuto(strokeArg.V) {
			if cv, ok := strokeArg.V.(ContentValue); ok {
				stroke = &cv.Content
			} else {
				return nil, &TypeMismatchError{
					Expected: "content or none",
					Got:      strokeArg.V.Type().String(),
					Span:     strokeArg.Span,
				}
			}
		}
	}

	// Get optional inset argument
	var inset *float64
	if insetArg := args.Find("inset"); insetArg != nil {
		if !IsNone(insetArg.V) && !IsAuto(insetArg.V) {
			if lv, ok := insetArg.V.(LengthValue); ok {
				i := lv.Length.Points
				inset = &i
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      insetArg.V.Type().String(),
					Span:     insetArg.Span,
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
		Elements: []ContentElement{&TableCellElement{
			Content: body,
			X:       x,
			Y:       y,
			Colspan: colspan,
			Rowspan: rowspan,
			Align:   align,
			Fill:    fill,
			Stroke:  stroke,
			Inset:   inset,
		}},
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
	// Register list element function
	scope.DefineFunc("list", ListFunc())
	// Register enum element function
	scope.DefineFunc("enum", EnumFunc())
	// Register table element function
	scope.DefineFunc("table", TableFunc())
	// Register table.cell element function
	scope.DefineFunc("table.cell", TableCellFunc())
}

// ElementFunctions returns a map of all element function names to their functions.
// This is useful for introspection and testing.
func ElementFunctions() map[string]*Func {
	return map[string]*Func{
		"raw":        RawFunc(),
		"par":        ParFunc(),
		"parbreak":   ParbreakFunc(),
		"stack":      StackFunc(),
		"align":      AlignFunc(),
		"heading":    HeadingFunc(),
		"list":       ListFunc(),
		"enum":       EnumFunc(),
		"table":      TableFunc(),
		"table.cell": TableCellFunc(),
	}
}
