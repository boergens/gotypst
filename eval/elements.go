package eval

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"

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
// Columns Element
// ----------------------------------------------------------------------------

// ColumnsElement represents a multi-column layout element.
// It arranges its body content into multiple columns.
type ColumnsElement struct {
	// Count is the number of columns.
	// If nil, defaults to 2.
	Count *int
	// Gutter is the gap between columns (in points).
	// If nil, defaults to 4% of page width.
	Gutter *float64
	// Body is the content to arrange in columns.
	Body Content
}

func (*ColumnsElement) IsContentElement() {}

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
					{Name: "count", Type: TypeInt, Default: Int(2), Named: false},
					{Name: "gutter", Type: TypeRelative, Default: None, Named: true},
					{Name: "body", Type: TypeContent, Named: false},
				},
			},
		},
	}
}

// columnsNative implements the columns() function.
// Creates a ColumnsElement with the given column count, gutter, and body.
//
// Arguments:
//   - count (positional, int, default: 2): The number of columns
//   - gutter (named, relative, default: none): The gap between columns
//   - body (positional, content): The content to arrange in columns
func columnsNative(vm *Vm, args *Args) (Value, error) {
	// Get optional count argument (default: 2)
	count := 2
	countArg := args.Find("count")
	if countArg == nil {
		// Try to get positional count if it's an integer
		if peeked := args.Take(); peeked != nil {
			if _, ok := AsInt(peeked.V); ok {
				countArgSpanned, _ := args.Expect("count")
				countArg = &countArgSpanned
			}
		}
	}
	if countArg != nil {
		if !IsNone(countArg.V) && !IsAuto(countArg.V) {
			countVal, ok := AsInt(countArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "integer",
					Got:      countArg.V.Type().String(),
					Span:     countArg.Span,
				}
			}
			count = int(countVal)
			if count < 1 {
				return nil, &ConstructorError{
					Message: "column count must be at least 1",
					Span:    countArg.Span,
				}
			}
		}
	}

	// Get optional gutter argument
	var gutter *float64
	if gutterArg := args.Find("gutter"); gutterArg != nil {
		if !IsNone(gutterArg.V) && !IsAuto(gutterArg.V) {
			switch g := gutterArg.V.(type) {
			case LengthValue:
				gutter = &g.Length.Points
			case RelativeValue:
				// For relative, we store the absolute part for now
				// Full relative support would need layout context
				gutter = &g.Relative.Abs.Points
			case RatioValue:
				// Convert ratio to a representative value
				// (full conversion needs layout context)
				pts := g.Ratio.Value * 100 // Scale for visibility
				gutter = &pts
			default:
				return nil, &TypeMismatchError{
					Expected: "length or relative",
					Got:      gutterArg.V.Type().String(),
					Span:     gutterArg.Span,
				}
			}
		}
	}

	// Get required body argument (positional)
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

	// Create the ColumnsElement wrapped in ContentValue
	elem := &ColumnsElement{
		Gutter: gutter,
		Body:   body,
	}
	if count != 2 {
		elem.Count = &count
	}

	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}

// ----------------------------------------------------------------------------
// Box Element
// ----------------------------------------------------------------------------

// BoxElement represents an inline box container element.
// It can size its content, apply fills/strokes, and clip overflow.
type BoxElement struct {
	// Width of the box (in points). If nil, auto-sizes to content.
	Width *float64
	// Height of the box (in points). If nil, auto-sizes to content.
	Height *float64
	// Baseline position (in points from bottom). If nil, uses content baseline.
	Baseline *float64
	// Fill color for the background. If nil, no fill.
	Fill Value
	// Stroke for the border. Can be length, color, or stroke dict. If nil, no stroke.
	Stroke Value
	// Radius for rounded corners. Can be single value or dictionary.
	Radius Value
	// Inset padding inside the box.
	Inset Value
	// Outset expansion outside the box.
	Outset Value
	// Whether to clip content that overflows the box.
	Clip bool
	// Body is the content inside the box.
	Body Content
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
					{Name: "stroke", Type: TypeDyn, Default: None, Named: true},
					{Name: "radius", Type: TypeDyn, Default: None, Named: true},
					{Name: "inset", Type: TypeDyn, Default: None, Named: true},
					{Name: "outset", Type: TypeDyn, Default: None, Named: true},
					{Name: "clip", Type: TypeBool, Default: False, Named: true},
				},
			},
		},
	}
}

// boxNative implements the box() function.
// Creates a BoxElement with optional sizing and styling.
//
// Arguments:
//   - body (positional, content, default: none): The content inside the box
//   - width (named, length, default: auto): Box width
//   - height (named, length, default: auto): Box height
//   - baseline (named, length, default: none): Baseline position from bottom
//   - fill (named, color, default: none): Background fill
//   - stroke (named, various, default: none): Border stroke
//   - radius (named, various, default: none): Corner radius
//   - inset (named, various, default: none): Inner padding
//   - outset (named, various, default: none): Outer expansion
//   - clip (named, bool, default: false): Whether to clip overflow
func boxNative(vm *Vm, args *Args) (Value, error) {
	elem := &BoxElement{}

	// Get optional body argument (positional or named)
	bodyArg := args.Find("body")
	if bodyArg == nil {
		bodyArg2 := args.Eat()
		if bodyArg2 != nil {
			bodyArg = bodyArg2
		}
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
				// Handle relative values (percentages)
				w := rv.Relative.Rel.Value * 100 // Store as percentage for now
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
				h := rv.Relative.Rel.Value * 100
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

	// Get optional baseline argument
	if baselineArg := args.Find("baseline"); baselineArg != nil {
		if !IsAuto(baselineArg.V) && !IsNone(baselineArg.V) {
			if lv, ok := baselineArg.V.(LengthValue); ok {
				b := lv.Length.Points
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
			elem.Fill = fillArg.V
		}
	}

	// Get optional stroke argument
	if strokeArg := args.Find("stroke"); strokeArg != nil {
		if !IsNone(strokeArg.V) {
			elem.Stroke = strokeArg.V
		}
	}

	// Get optional radius argument
	if radiusArg := args.Find("radius"); radiusArg != nil {
		if !IsNone(radiusArg.V) {
			elem.Radius = radiusArg.V
		}
	}

	// Get optional inset argument
	if insetArg := args.Find("inset"); insetArg != nil {
		if !IsNone(insetArg.V) {
			elem.Inset = insetArg.V
		}
	}

	// Get optional outset argument
	if outsetArg := args.Find("outset"); outsetArg != nil {
		if !IsNone(outsetArg.V) {
			elem.Outset = outsetArg.V
		}
	}

	// Get optional clip argument (default: false)
	if clipArg := args.Find("clip"); clipArg != nil {
		if !IsNone(clipArg.V) && !IsAuto(clipArg.V) {
			clipVal, ok := AsBool(clipArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "bool",
					Got:      clipArg.V.Type().String(),
					Span:     clipArg.Span,
				}
			}
			elem.Clip = clipVal
		}
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Create the BoxElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}

// ----------------------------------------------------------------------------
// Block Element
// ----------------------------------------------------------------------------

// BlockElement represents a block-level container element.
// It creates a new block in the document flow with optional sizing and styling.
type BlockElement struct {
	// Width of the block (in points). If nil, auto-sizes.
	Width *float64
	// Height of the block (in points). If nil, auto-sizes.
	Height *float64
	// Whether the block can break across pages.
	Breakable *bool
	// Fill color for the background.
	Fill Value
	// Stroke for the border.
	Stroke Value
	// Radius for rounded corners.
	Radius Value
	// Inset padding inside the block.
	Inset Value
	// Outset expansion outside the block.
	Outset Value
	// Spacing between adjacent blocks.
	Spacing *float64
	// Spacing above this block (overrides Spacing).
	Above *float64
	// Spacing below this block (overrides Spacing).
	Below *float64
	// Whether to clip content that overflows.
	Clip bool
	// Whether the block sticks to the next block.
	Sticky bool
	// Body is the content inside the block.
	Body Content
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
					{Name: "stroke", Type: TypeDyn, Default: None, Named: true},
					{Name: "radius", Type: TypeDyn, Default: None, Named: true},
					{Name: "inset", Type: TypeDyn, Default: None, Named: true},
					{Name: "outset", Type: TypeDyn, Default: None, Named: true},
					{Name: "spacing", Type: TypeLength, Default: None, Named: true},
					{Name: "above", Type: TypeLength, Default: Auto, Named: true},
					{Name: "below", Type: TypeLength, Default: Auto, Named: true},
					{Name: "clip", Type: TypeBool, Default: False, Named: true},
					{Name: "sticky", Type: TypeBool, Default: False, Named: true},
				},
			},
		},
	}
}

// blockNative implements the block() function.
// Creates a BlockElement with optional sizing and styling.
//
// Arguments:
//   - body (positional, content, default: none): The content inside the block
//   - width (named, length, default: auto): Block width
//   - height (named, length, default: auto): Block height
//   - breakable (named, bool, default: true): Whether block can break across pages
//   - fill (named, color, default: none): Background fill
//   - stroke (named, various, default: none): Border stroke
//   - radius (named, various, default: none): Corner radius
//   - inset (named, various, default: none): Inner padding
//   - outset (named, various, default: none): Outer expansion
//   - spacing (named, length, default: 1.2em): Spacing between blocks
//   - above (named, length, default: auto): Spacing above this block
//   - below (named, length, default: auto): Spacing below this block
//   - clip (named, bool, default: false): Whether to clip overflow
//   - sticky (named, bool, default: false): Whether to stick to next block
func blockNative(vm *Vm, args *Args) (Value, error) {
	elem := &BlockElement{}

	// Get optional body argument (positional or named)
	bodyArg := args.Find("body")
	if bodyArg == nil {
		bodyArg2 := args.Eat()
		if bodyArg2 != nil {
			bodyArg = bodyArg2
		}
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
				w := rv.Relative.Rel.Value * 100
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
				h := rv.Relative.Rel.Value * 100
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

	// Get optional breakable argument (default: true)
	if breakableArg := args.Find("breakable"); breakableArg != nil {
		if !IsNone(breakableArg.V) && !IsAuto(breakableArg.V) {
			breakVal, ok := AsBool(breakableArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "bool",
					Got:      breakableArg.V.Type().String(),
					Span:     breakableArg.Span,
				}
			}
			elem.Breakable = &breakVal
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

	// Get optional radius argument
	if radiusArg := args.Find("radius"); radiusArg != nil {
		if !IsNone(radiusArg.V) {
			elem.Radius = radiusArg.V
		}
	}

	// Get optional inset argument
	if insetArg := args.Find("inset"); insetArg != nil {
		if !IsNone(insetArg.V) {
			elem.Inset = insetArg.V
		}
	}

	// Get optional outset argument
	if outsetArg := args.Find("outset"); outsetArg != nil {
		if !IsNone(outsetArg.V) {
			elem.Outset = outsetArg.V
		}
	}

	// Get optional spacing argument
	if spacingArg := args.Find("spacing"); spacingArg != nil {
		if !IsAuto(spacingArg.V) && !IsNone(spacingArg.V) {
			if lv, ok := spacingArg.V.(LengthValue); ok {
				s := lv.Length.Points
				elem.Spacing = &s
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      spacingArg.V.Type().String(),
					Span:     spacingArg.Span,
				}
			}
		}
	}

	// Get optional above argument
	if aboveArg := args.Find("above"); aboveArg != nil {
		if !IsAuto(aboveArg.V) && !IsNone(aboveArg.V) {
			if lv, ok := aboveArg.V.(LengthValue); ok {
				a := lv.Length.Points
				elem.Above = &a
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
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
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      belowArg.V.Type().String(),
					Span:     belowArg.Span,
				}
			}
		}
	}

	// Get optional clip argument (default: false)
	if clipArg := args.Find("clip"); clipArg != nil {
		if !IsNone(clipArg.V) && !IsAuto(clipArg.V) {
			clipVal, ok := AsBool(clipArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "bool",
					Got:      clipArg.V.Type().String(),
					Span:     clipArg.Span,
				}
			}
			elem.Clip = clipVal
		}
	}

	// Get optional sticky argument (default: false)
	if stickyArg := args.Find("sticky"); stickyArg != nil {
		if !IsNone(stickyArg.V) && !IsAuto(stickyArg.V) {
			stickyVal, ok := AsBool(stickyArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "bool",
					Got:      stickyArg.V.Type().String(),
					Span:     stickyArg.Span,
				}
			}
			elem.Sticky = stickyVal
		}
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Create the BlockElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}

// ----------------------------------------------------------------------------
// Pad Element
// ----------------------------------------------------------------------------

// PadElement represents a padding container element.
// It adds spacing around its content.
type PadElement struct {
	// Left padding (in points).
	Left *float64
	// Top padding (in points).
	Top *float64
	// Right padding (in points).
	Right *float64
	// Bottom padding (in points).
	Bottom *float64
	// Body is the content to pad.
	Body Content
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
// Creates a PadElement with padding around content.
//
// Arguments:
//   - body (positional, content): The content to pad
//   - left (named, length, default: 0pt): Left padding
//   - top (named, length, default: 0pt): Top padding
//   - right (named, length, default: 0pt): Right padding
//   - bottom (named, length, default: 0pt): Bottom padding
//   - x (named, length, default: 0pt): Horizontal padding (sets left and right)
//   - y (named, length, default: 0pt): Vertical padding (sets top and bottom)
//   - rest (named, length, default: 0pt): Padding for all sides
func padNative(vm *Vm, args *Args) (Value, error) {
	elem := &PadElement{}

	// Get rest argument first (applies to all sides)
	if restArg := args.Find("rest"); restArg != nil {
		if !IsNone(restArg.V) && !IsAuto(restArg.V) {
			if lv, ok := restArg.V.(LengthValue); ok {
				r := lv.Length.Points
				elem.Left = &r
				elem.Top = &r
				elem.Right = &r
				elem.Bottom = &r
			} else {
				return nil, &TypeMismatchError{
					Expected: "length",
					Got:      restArg.V.Type().String(),
					Span:     restArg.Span,
				}
			}
		}
	}

	// Get x argument (sets left and right)
	if xArg := args.Find("x"); xArg != nil {
		if !IsNone(xArg.V) && !IsAuto(xArg.V) {
			if lv, ok := xArg.V.(LengthValue); ok {
				x := lv.Length.Points
				elem.Left = &x
				elem.Right = &x
			} else {
				return nil, &TypeMismatchError{
					Expected: "length",
					Got:      xArg.V.Type().String(),
					Span:     xArg.Span,
				}
			}
		}
	}

	// Get y argument (sets top and bottom)
	if yArg := args.Find("y"); yArg != nil {
		if !IsNone(yArg.V) && !IsAuto(yArg.V) {
			if lv, ok := yArg.V.(LengthValue); ok {
				y := lv.Length.Points
				elem.Top = &y
				elem.Bottom = &y
			} else {
				return nil, &TypeMismatchError{
					Expected: "length",
					Got:      yArg.V.Type().String(),
					Span:     yArg.Span,
				}
			}
		}
	}

	// Get individual side arguments (override shorthands)
	if leftArg := args.Find("left"); leftArg != nil {
		if !IsNone(leftArg.V) && !IsAuto(leftArg.V) {
			if lv, ok := leftArg.V.(LengthValue); ok {
				l := lv.Length.Points
				elem.Left = &l
			} else {
				return nil, &TypeMismatchError{
					Expected: "length",
					Got:      leftArg.V.Type().String(),
					Span:     leftArg.Span,
				}
			}
		}
	}

	if topArg := args.Find("top"); topArg != nil {
		if !IsNone(topArg.V) && !IsAuto(topArg.V) {
			if lv, ok := topArg.V.(LengthValue); ok {
				t := lv.Length.Points
				elem.Top = &t
			} else {
				return nil, &TypeMismatchError{
					Expected: "length",
					Got:      topArg.V.Type().String(),
					Span:     topArg.Span,
				}
			}
		}
	}

	if rightArg := args.Find("right"); rightArg != nil {
		if !IsNone(rightArg.V) && !IsAuto(rightArg.V) {
			if lv, ok := rightArg.V.(LengthValue); ok {
				r := lv.Length.Points
				elem.Right = &r
			} else {
				return nil, &TypeMismatchError{
					Expected: "length",
					Got:      rightArg.V.Type().String(),
					Span:     rightArg.Span,
				}
			}
		}
	}

	if bottomArg := args.Find("bottom"); bottomArg != nil {
		if !IsNone(bottomArg.V) && !IsAuto(bottomArg.V) {
			if lv, ok := bottomArg.V.(LengthValue); ok {
				b := lv.Length.Points
				elem.Bottom = &b
			} else {
				return nil, &TypeMismatchError{
					Expected: "length",
					Got:      bottomArg.V.Type().String(),
					Span:     bottomArg.Span,
				}
			}
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

	// Create the PadElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}

// LinkFunc creates the link element function.
func LinkFunc() *Func {
	name := "link"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: linkNative,
			Info: &FuncInfo{
				Name: "link",
				Params: []ParamInfo{
					{Name: "dest", Type: TypeStr, Named: false},
					{Name: "body", Type: TypeContent, Default: None, Named: false},
				},
			},
		},
	}
}

// linkNative implements the link() function.
// Creates a LinkElement for hyperlinks.
//
// Arguments:
//   - dest (positional, str): The destination URL or label
//   - body (positional, content, default: none): The content to display (defaults to the URL)
func linkNative(vm *Vm, args *Args) (Value, error) {
	// Get required dest argument (positional)
	destArg := args.Find("dest")
	if destArg == nil {
		destArgSpanned, err := args.Expect("dest")
		if err != nil {
			return nil, err
		}
		destArg = &destArgSpanned
	}

	dest, ok := AsStr(destArg.V)
	if !ok {
		return nil, &TypeMismatchError{
			Expected: "string",
			Got:      destArg.V.Type().String(),
			Span:     destArg.Span,
		}
	}

	elem := &LinkElement{URL: dest}

	// Get optional body argument (positional or named)
	bodyArg := args.Find("body")
	if bodyArg == nil {
		bodyArg = args.Eat()
	}
	if bodyArg != nil && !IsNone(bodyArg.V) {
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

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Create the LinkElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}

// ----------------------------------------------------------------------------
// Table Element
// ----------------------------------------------------------------------------

// TableFunc creates the table element function.
func TableFunc() *Func {
	name := "table"

	// Create scope with table.cell method
	tableScope := NewScope()
	tableScope.DefineFunc("cell", TableCellFunc())

	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: tableNative,
			Info: &FuncInfo{
				Name: "table",
				Params: []ParamInfo{
					{Name: "columns", Type: TypeDyn, Default: None, Named: true},
					{Name: "rows", Type: TypeDyn, Default: None, Named: true},
					{Name: "gutter", Type: TypeDyn, Default: None, Named: true},
					{Name: "column-gutter", Type: TypeDyn, Default: None, Named: true},
					{Name: "row-gutter", Type: TypeDyn, Default: None, Named: true},
					{Name: "inset", Type: TypeDyn, Default: None, Named: true},
					{Name: "align", Type: TypeDyn, Default: None, Named: true},
					{Name: "fill", Type: TypeDyn, Default: None, Named: true},
					{Name: "stroke", Type: TypeDyn, Default: None, Named: true},
					{Name: "children", Type: TypeContent, Variadic: true},
				},
			},
			Scope: tableScope,
		},
	}
}

// tableNative implements the table() function.
// Creates a TableElement with cells arranged in a grid.
//
// Arguments:
//   - columns (named): Column sizing (int for count, or array of track sizes)
//   - rows (named): Row sizing (same format as columns)
//   - gutter (named): Shorthand for column-gutter and row-gutter
//   - column-gutter (named): Gaps between columns
//   - row-gutter (named): Gaps between rows
//   - inset (named): Cell padding (default: 5pt)
//   - align (named): Cell content alignment
//   - fill (named): Cell background fill
//   - stroke (named): Cell border stroke (default: 1pt + black)
//   - children (positional, variadic): Table cells
func tableNative(vm *Vm, args *Args) (Value, error) {
	elem := &TableElement{}

	// Get columns argument (required for determining grid structure)
	if columnsArg := args.Find("columns"); columnsArg != nil && !IsNone(columnsArg.V) {
		elem.Columns = columnsArg.V
	}

	// Get rows argument
	if rowsArg := args.Find("rows"); rowsArg != nil && !IsNone(rowsArg.V) {
		elem.Rows = rowsArg.V
	}

	// Get gutter argument
	if gutterArg := args.Find("gutter"); gutterArg != nil && !IsNone(gutterArg.V) {
		elem.Gutter = gutterArg.V
	}

	// Get column-gutter argument
	if colGutterArg := args.Find("column-gutter"); colGutterArg != nil && !IsNone(colGutterArg.V) {
		elem.ColumnGutter = colGutterArg.V
	}

	// Get row-gutter argument
	if rowGutterArg := args.Find("row-gutter"); rowGutterArg != nil && !IsNone(rowGutterArg.V) {
		elem.RowGutter = rowGutterArg.V
	}

	// Get inset argument
	if insetArg := args.Find("inset"); insetArg != nil && !IsNone(insetArg.V) {
		elem.Inset = insetArg.V
	}

	// Get align argument
	if alignArg := args.Find("align"); alignArg != nil && !IsNone(alignArg.V) {
		elem.Align = alignArg.V
	}

	// Get fill argument
	if fillArg := args.Find("fill"); fillArg != nil && !IsNone(fillArg.V) {
		elem.Fill = fillArg.V
	}

	// Get stroke argument
	if strokeArg := args.Find("stroke"); strokeArg != nil && !IsNone(strokeArg.V) {
		elem.Stroke = strokeArg.V
	}

	// Collect cell children (variadic positional arguments)
	for {
		child := args.Eat()
		if child == nil {
			break
		}

		// Check if it's a TableCellElement
		if cv, ok := child.V.(ContentValue); ok {
			if len(cv.Content.Elements) == 1 {
				if cell, ok := cv.Content.Elements[0].(*TableCellElement); ok {
					elem.Children = append(elem.Children, TableChild{Cell: cell})
					continue
				}
			}
		}

		// Convert child to content
		var content Content
		switch v := child.V.(type) {
		case ContentValue:
			content = v.Content
		case StrValue:
			content = Content{Elements: []ContentElement{&TextElement{Text: string(v)}}}
		default:
			// Try to display other values as text
			content = Content{Elements: []ContentElement{&TextElement{Text: v.Display().String()}}}
		}
		elem.Children = append(elem.Children, TableChild{Content: &content})
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
					{Name: "inset", Type: TypeDyn, Default: None, Named: true},
					{Name: "align", Type: TypeDyn, Default: None, Named: true},
					{Name: "fill", Type: TypeDyn, Default: None, Named: true},
					{Name: "stroke", Type: TypeDyn, Default: None, Named: true},
					{Name: "breakable", Type: TypeBool, Default: Auto, Named: true},
				},
			},
		},
	}
}

// tableCellNative implements the table.cell() function.
// Creates a TableCellElement with optional position/span overrides.
//
// Arguments:
//   - body (positional, content): The cell's content
//   - x (named, int): Column position (0-indexed, auto by default)
//   - y (named, int): Row position (0-indexed, auto by default)
//   - colspan (named, int): Number of columns to span (default: 1)
//   - rowspan (named, int): Number of rows to span (default: 1)
//   - inset (named): Cell padding override
//   - align (named): Cell alignment override
//   - fill (named): Cell background override
//   - stroke (named): Cell border override
//   - breakable (named, bool): Whether rows can break across pages
func tableCellNative(vm *Vm, args *Args) (Value, error) {
	elem := &TableCellElement{
		Colspan: 1,
		Rowspan: 1,
	}

	// Get required body argument (positional or named)
	bodyArg := args.Find("body")
	if bodyArg == nil {
		bodyArg2 := args.Eat()
		if bodyArg2 != nil {
			bodyArg = bodyArg2
		}
	}
	if bodyArg != nil {
		switch v := bodyArg.V.(type) {
		case ContentValue:
			elem.Body = v.Content
		case StrValue:
			elem.Body = Content{Elements: []ContentElement{&TextElement{Text: string(v)}}}
		default:
			elem.Body = Content{Elements: []ContentElement{&TextElement{Text: v.Display().String()}}}
		}
	}

	// Get optional x argument
	if xArg := args.Find("x"); xArg != nil && !IsAuto(xArg.V) && !IsNone(xArg.V) {
		if x, ok := AsInt(xArg.V); ok {
			xInt := int(x)
			elem.X = &xInt
		} else {
			return nil, &TypeMismatchError{
				Expected: "int or auto",
				Got:      xArg.V.Type().String(),
				Span:     xArg.Span,
			}
		}
	}

	// Get optional y argument
	if yArg := args.Find("y"); yArg != nil && !IsAuto(yArg.V) && !IsNone(yArg.V) {
		if y, ok := AsInt(yArg.V); ok {
			yInt := int(y)
			elem.Y = &yInt
		} else {
			return nil, &TypeMismatchError{
				Expected: "int or auto",
				Got:      yArg.V.Type().String(),
				Span:     yArg.Span,
			}
		}
	}

	// Get optional colspan argument
	if colspanArg := args.Find("colspan"); colspanArg != nil && !IsNone(colspanArg.V) {
		if colspan, ok := AsInt(colspanArg.V); ok {
			if colspan < 1 {
				return nil, &InvalidArgumentError{
					Message: "colspan must be at least 1",
					Span:    colspanArg.Span,
				}
			}
			elem.Colspan = int(colspan)
		} else {
			return nil, &TypeMismatchError{
				Expected: "int",
				Got:      colspanArg.V.Type().String(),
				Span:     colspanArg.Span,
			}
		}
	}

	// Get optional rowspan argument
	if rowspanArg := args.Find("rowspan"); rowspanArg != nil && !IsNone(rowspanArg.V) {
		if rowspan, ok := AsInt(rowspanArg.V); ok {
			if rowspan < 1 {
				return nil, &InvalidArgumentError{
					Message: "rowspan must be at least 1",
					Span:    rowspanArg.Span,
				}
			}
			elem.Rowspan = int(rowspan)
		} else {
			return nil, &TypeMismatchError{
				Expected: "int",
				Got:      rowspanArg.V.Type().String(),
				Span:     rowspanArg.Span,
			}
		}
	}

	// Get optional inset argument
	if insetArg := args.Find("inset"); insetArg != nil && !IsNone(insetArg.V) {
		elem.Inset = insetArg.V
	}

	// Get optional align argument
	if alignArg := args.Find("align"); alignArg != nil && !IsNone(alignArg.V) {
		elem.Align = alignArg.V
	}

	// Get optional fill argument
	if fillArg := args.Find("fill"); fillArg != nil && !IsNone(fillArg.V) {
		elem.Fill = fillArg.V
	}

	// Get optional stroke argument
	if strokeArg := args.Find("stroke"); strokeArg != nil && !IsNone(strokeArg.V) {
		elem.Stroke = strokeArg.V
	}

	// Get optional breakable argument
	if breakableArg := args.Find("breakable"); breakableArg != nil && !IsAuto(breakableArg.V) && !IsNone(breakableArg.V) {
		elem.Breakable = breakableArg.V
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
// Image Element
// ----------------------------------------------------------------------------

// ImageFit represents how an image fits within its container.
type ImageFit string

const (
	// ImageFitContain scales the image to fit within the bounds while preserving aspect ratio.
	ImageFitContain ImageFit = "contain"
	// ImageFitCover scales the image to cover the bounds while preserving aspect ratio.
	ImageFitCover ImageFit = "cover"
	// ImageFitStretch stretches the image to fill the bounds exactly.
	ImageFitStretch ImageFit = "stretch"
)

// ImageElement represents an embedded image element.
type ImageElement struct {
	// Path is the source path of the image file.
	Path string
	// Data is the raw image bytes (loaded from Path).
	Data []byte
	// Width is the rendered width (in points). If nil, auto-sizes.
	Width *float64
	// Height is the rendered height (in points). If nil, auto-sizes.
	Height *float64
	// Fit specifies how the image fits within its bounds.
	Fit ImageFit
	// Alt is the alt text for accessibility.
	Alt *string
	// NaturalWidth is the natural width of the image in pixels.
	NaturalWidth int
	// NaturalHeight is the natural height of the image in pixels.
	NaturalHeight int
}

func (*ImageElement) IsContentElement() {}

// ImageFunc creates the image element function.
func ImageFunc() *Func {
	name := "image"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: imageNative,
			Info: &FuncInfo{
				Name: "image",
				Params: []ParamInfo{
					{Name: "path", Type: TypeStr, Named: false},
					{Name: "width", Type: TypeLength, Default: Auto, Named: true},
					{Name: "height", Type: TypeLength, Default: Auto, Named: true},
					{Name: "fit", Type: TypeStr, Default: Str("contain"), Named: true},
					{Name: "alt", Type: TypeStr, Default: None, Named: true},
				},
			},
		},
	}
}

// imageNative implements the image() function.
// Loads an image from the given path and creates an ImageElement.
//
// Arguments:
//   - path (positional, str): The path to the image file
//   - width (named, length, default: auto): Rendered width
//   - height (named, length, default: auto): Rendered height
//   - fit (named, str, default: "contain"): How to fit within bounds ("contain", "cover", "stretch")
//   - alt (named, str, default: none): Alt text for accessibility
func imageNative(vm *Vm, args *Args) (Value, error) {
	// Get required path argument
	pathArg, err := args.Expect("path")
	if err != nil {
		return nil, err
	}

	path, ok := AsStr(pathArg.V)
	if !ok {
		return nil, &TypeMismatchError{
			Expected: "string",
			Got:      pathArg.V.Type().String(),
			Span:     pathArg.Span,
		}
	}

	elem := &ImageElement{
		Path: path,
		Fit:  ImageFitContain,
	}

	// Get optional width argument
	if widthArg := args.Find("width"); widthArg != nil {
		if !IsAuto(widthArg.V) && !IsNone(widthArg.V) {
			if lv, ok := widthArg.V.(LengthValue); ok {
				w := lv.Length.Points
				elem.Width = &w
			} else if rv, ok := widthArg.V.(RelativeValue); ok {
				// Handle relative values (percentages)
				w := rv.Relative.Rel.Value * 100 // Store as percentage
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
				h := rv.Relative.Rel.Value * 100
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

	// Get optional fit argument (default: "contain")
	if fitArg := args.Find("fit"); fitArg != nil {
		if !IsNone(fitArg.V) && !IsAuto(fitArg.V) {
			fitStr, ok := AsStr(fitArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "string",
					Got:      fitArg.V.Type().String(),
					Span:     fitArg.Span,
				}
			}
			// Validate fit value
			switch fitStr {
			case "contain":
				elem.Fit = ImageFitContain
			case "cover":
				elem.Fit = ImageFitCover
			case "stretch":
				elem.Fit = ImageFitStretch
			default:
				return nil, &TypeMismatchError{
					Expected: "\"contain\", \"cover\", or \"stretch\"",
					Got:      "\"" + fitStr + "\"",
					Span:     fitArg.Span,
				}
			}
		}
	}

	// Get optional alt argument
	if altArg := args.Find("alt"); altArg != nil {
		if !IsNone(altArg.V) {
			altStr, ok := AsStr(altArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "string or none",
					Got:      altArg.V.Type().String(),
					Span:     altArg.Span,
				}
			}
			elem.Alt = &altStr
		}
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Load the image file
	data, err := readFileFromWorld(vm, path)
	if err != nil {
		return nil, err
	}
	elem.Data = data

	// Decode image to get natural dimensions
	width, height, err := decodeImageDimensions(data)
	if err != nil {
		return nil, &FileParseError{
			Path:    path,
			Format:  "image",
			Message: err.Error(),
		}
	}
	elem.NaturalWidth = width
	elem.NaturalHeight = height

	// Create the ImageElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}

// decodeImageDimensions decodes an image to get its natural dimensions.
func decodeImageDimensions(data []byte) (width, height int, err error) {
	// Try JPEG first
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xD8 {
		return decodeJPEGDimensions(data)
	}

	// Try PNG
	if len(data) >= 8 &&
		data[0] == 0x89 && data[1] == 'P' && data[2] == 'N' && data[3] == 'G' &&
		data[4] == 0x0D && data[5] == 0x0A && data[6] == 0x1A && data[7] == 0x0A {
		return decodePNGDimensions(data)
	}

	// Fallback to generic image decode
	return decodeGenericDimensions(data)
}

// decodeJPEGDimensions extracts dimensions from JPEG data without full decode.
func decodeJPEGDimensions(data []byte) (width, height int, err error) {
	// Parse JPEG markers to find SOF
	i := 2 // Skip SOI marker
	for i < len(data)-1 {
		if data[i] != 0xFF {
			return 0, 0, fmt.Errorf("invalid JPEG marker")
		}
		marker := data[i+1]
		i += 2

		// Skip padding
		for marker == 0xFF && i < len(data) {
			marker = data[i]
			i++
		}

		// Check for SOF markers (Start of Frame)
		if (marker >= 0xC0 && marker <= 0xC3) || (marker >= 0xC5 && marker <= 0xC7) ||
			(marker >= 0xC9 && marker <= 0xCB) || (marker >= 0xCD && marker <= 0xCF) {
			if i+7 > len(data) {
				return 0, 0, fmt.Errorf("truncated JPEG")
			}
			height = int(data[i+1])<<8 | int(data[i+2])
			width = int(data[i+3])<<8 | int(data[i+4])
			return width, height, nil
		}

		// Skip other markers
		if marker == 0xD8 || marker == 0xD9 || (marker >= 0xD0 && marker <= 0xD7) {
			continue // No length field
		}

		if i+1 >= len(data) {
			return 0, 0, fmt.Errorf("truncated JPEG")
		}
		length := int(data[i])<<8 | int(data[i+1])
		i += length
	}

	return 0, 0, fmt.Errorf("no SOF marker found in JPEG")
}

// decodePNGDimensions extracts dimensions from PNG data without full decode.
func decodePNGDimensions(data []byte) (width, height int, err error) {
	// PNG dimensions are in the IHDR chunk, immediately after the signature
	if len(data) < 24 {
		return 0, 0, fmt.Errorf("truncated PNG")
	}

	// Skip signature (8 bytes), length (4 bytes), "IHDR" (4 bytes)
	// Then read width (4 bytes) and height (4 bytes)
	width = int(data[16])<<24 | int(data[17])<<16 | int(data[18])<<8 | int(data[19])
	height = int(data[20])<<24 | int(data[21])<<16 | int(data[22])<<8 | int(data[23])

	return width, height, nil
}

// decodeGenericDimensions uses Go's image package to decode dimensions.
func decodeGenericDimensions(data []byte) (width, height int, err error) {
	cfg, _, err := decodeImageConfig(data)
	if err != nil {
		return 0, 0, err
	}
	return cfg.Width, cfg.Height, nil
}

// decodeImageConfig decodes image configuration (dimensions) from raw data.
func decodeImageConfig(data []byte) (image.Config, string, error) {
	return image.DecodeConfig(bytes.NewReader(data))
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
	// Register columns element function
	scope.DefineFunc("columns", ColumnsFunc())
	// Register box element function
	scope.DefineFunc("box", BoxFunc())
	// Register block element function
	scope.DefineFunc("block", BlockFunc())
	// Register pad element function
	scope.DefineFunc("pad", PadFunc())
	// Register table element function
	scope.DefineFunc("table", TableFunc())
	// Register link element function
	scope.DefineFunc("link", LinkFunc())
	// Register image element function
	scope.DefineFunc("image", ImageFunc())
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
		"columns":  ColumnsFunc(),
		"box":      BoxFunc(),
		"block":    BlockFunc(),
		"pad":      PadFunc(),
		"table":    TableFunc(),
		"link":     LinkFunc(),
		"image":    ImageFunc(),
	}
}
