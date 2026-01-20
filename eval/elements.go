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
// List Element Functions
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
					{Name: "indent", Type: TypeLength, Default: None, Named: true},
					{Name: "body-indent", Type: TypeLength, Default: None, Named: true},
					{Name: "spacing", Type: TypeLength, Default: Auto, Named: true},
					{Name: "children", Type: TypeContent, Named: false, Variadic: true},
				},
			},
		},
	}
}

// listNative implements the list() function.
// Creates a ListElement with the given items and styling options.
//
// Arguments:
//   - tight (named, bool, default: true): Control spacing between items
//   - marker (named, content, default: none): Custom marker content
//   - indent (named, length, default: none): Indentation of items
//   - body-indent (named, length, default: none): Space between marker and body
//   - spacing (named, length, default: auto): Spacing between items
//   - children (positional, variadic, content): The list items
func listNative(vm *Vm, args *Args) (Value, error) {
	// Get optional tight argument (default: true)
	tight := true
	if tightArg := args.Find("tight"); tightArg != nil {
		if !IsAuto(tightArg.V) && !IsNone(tightArg.V) {
			if tv, ok := AsBool(tightArg.V); ok {
				tight = tv
			} else {
				return nil, &TypeMismatchError{
					Expected: "bool",
					Got:      tightArg.V.Type().String(),
					Span:     tightArg.Span,
				}
			}
		}
	}

	// Get optional marker argument
	var marker *Content
	if markerArg := args.Find("marker"); markerArg != nil {
		if !IsNone(markerArg.V) && !IsAuto(markerArg.V) {
			if cv, ok := markerArg.V.(ContentValue); ok {
				marker = &cv.Content
			} else {
				return nil, &TypeMismatchError{
					Expected: "content or none",
					Got:      markerArg.V.Type().String(),
					Span:     markerArg.Span,
				}
			}
		}
	}

	// Get optional indent argument
	var indent *float64
	if indentArg := args.Find("indent"); indentArg != nil {
		if !IsNone(indentArg.V) && !IsAuto(indentArg.V) {
			if lv, ok := indentArg.V.(LengthValue); ok {
				i := lv.Length.Points
				indent = &i
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      indentArg.V.Type().String(),
					Span:     indentArg.Span,
				}
			}
		}
	}

	// Get optional body-indent argument
	var bodyIndent *float64
	if biArg := args.Find("body-indent"); biArg != nil {
		if !IsNone(biArg.V) && !IsAuto(biArg.V) {
			if lv, ok := biArg.V.(LengthValue); ok {
				bi := lv.Length.Points
				bodyIndent = &bi
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      biArg.V.Type().String(),
					Span:     biArg.Span,
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
					Expected: "length or auto",
					Got:      spacingArg.V.Type().String(),
					Span:     spacingArg.Span,
				}
			}
		}
	}

	// Collect remaining positional arguments as children
	var items []*ListItemElement
	for {
		childArg := args.Eat()
		if childArg == nil {
			break
		}

		if cv, ok := childArg.V.(ContentValue); ok {
			// Wrap the content in a ListItemElement
			items = append(items, &ListItemElement{Content: cv.Content})
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
			Items:      items,
			Tight:      tight,
			Marker:     marker,
			Indent:     indent,
			BodyIndent: bodyIndent,
			Spacing:    spacing,
		}},
	}}, nil
}

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
					{Name: "numbering", Type: TypeStr, Default: Str("1."), Named: true},
					{Name: "start", Type: TypeInt, Default: Auto, Named: true},
					{Name: "full", Type: TypeBool, Default: False, Named: true},
					{Name: "reversed", Type: TypeBool, Default: False, Named: true},
					{Name: "indent", Type: TypeLength, Default: None, Named: true},
					{Name: "body-indent", Type: TypeLength, Default: None, Named: true},
					{Name: "spacing", Type: TypeLength, Default: Auto, Named: true},
					{Name: "number-align", Type: TypeStr, Default: None, Named: true},
					{Name: "children", Type: TypeContent, Named: false, Variadic: true},
				},
			},
		},
	}
}

// enumNative implements the enum() function.
// Creates an EnumElement with the given items and styling options.
//
// Arguments:
//   - tight (named, bool, default: true): Control spacing between items
//   - numbering (named, str, default: "1."): Numbering pattern
//   - start (named, int, default: auto): Starting number
//   - full (named, bool, default: false): Show full numbering for nested enums
//   - reversed (named, bool, default: false): Reverse numbering direction
//   - indent (named, length, default: none): Indentation of items
//   - body-indent (named, length, default: none): Space between number and body
//   - spacing (named, length, default: auto): Spacing between items
//   - number-align (named, alignment, default: end+top): Number alignment
//   - children (positional, variadic, content): The enum items
func enumNative(vm *Vm, args *Args) (Value, error) {
	// Get optional tight argument (default: true)
	tight := true
	if tightArg := args.Find("tight"); tightArg != nil {
		if !IsAuto(tightArg.V) && !IsNone(tightArg.V) {
			if tv, ok := AsBool(tightArg.V); ok {
				tight = tv
			} else {
				return nil, &TypeMismatchError{
					Expected: "bool",
					Got:      tightArg.V.Type().String(),
					Span:     tightArg.Span,
				}
			}
		}
	}

	// Get optional numbering argument (default: "1.")
	numbering := "1."
	if numbArg := args.Find("numbering"); numbArg != nil {
		if !IsNone(numbArg.V) && !IsAuto(numbArg.V) {
			if ns, ok := AsStr(numbArg.V); ok {
				numbering = ns
			} else {
				return nil, &TypeMismatchError{
					Expected: "str",
					Got:      numbArg.V.Type().String(),
					Span:     numbArg.Span,
				}
			}
		}
	}

	// Get optional start argument
	var start *int
	if startArg := args.Find("start"); startArg != nil {
		if !IsNone(startArg.V) && !IsAuto(startArg.V) {
			if iv, ok := AsInt(startArg.V); ok {
				s := int(iv)
				start = &s
			} else {
				return nil, &TypeMismatchError{
					Expected: "int or auto",
					Got:      startArg.V.Type().String(),
					Span:     startArg.Span,
				}
			}
		}
	}

	// Get optional full argument (default: false)
	full := false
	if fullArg := args.Find("full"); fullArg != nil {
		if !IsAuto(fullArg.V) && !IsNone(fullArg.V) {
			if fv, ok := AsBool(fullArg.V); ok {
				full = fv
			} else {
				return nil, &TypeMismatchError{
					Expected: "bool",
					Got:      fullArg.V.Type().String(),
					Span:     fullArg.Span,
				}
			}
		}
	}

	// Get optional reversed argument (default: false)
	reversed := false
	if revArg := args.Find("reversed"); revArg != nil {
		if !IsAuto(revArg.V) && !IsNone(revArg.V) {
			if rv, ok := AsBool(revArg.V); ok {
				reversed = rv
			} else {
				return nil, &TypeMismatchError{
					Expected: "bool",
					Got:      revArg.V.Type().String(),
					Span:     revArg.Span,
				}
			}
		}
	}

	// Get optional indent argument
	var indent *float64
	if indentArg := args.Find("indent"); indentArg != nil {
		if !IsNone(indentArg.V) && !IsAuto(indentArg.V) {
			if lv, ok := indentArg.V.(LengthValue); ok {
				i := lv.Length.Points
				indent = &i
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      indentArg.V.Type().String(),
					Span:     indentArg.Span,
				}
			}
		}
	}

	// Get optional body-indent argument
	var bodyIndent *float64
	if biArg := args.Find("body-indent"); biArg != nil {
		if !IsNone(biArg.V) && !IsAuto(biArg.V) {
			if lv, ok := biArg.V.(LengthValue); ok {
				bi := lv.Length.Points
				bodyIndent = &bi
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      biArg.V.Type().String(),
					Span:     biArg.Span,
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
					Expected: "length or auto",
					Got:      spacingArg.V.Type().String(),
					Span:     spacingArg.Span,
				}
			}
		}
	}

	// Get optional number-align argument
	numberAlign := ""
	if naArg := args.Find("number-align"); naArg != nil {
		if !IsNone(naArg.V) && !IsAuto(naArg.V) {
			if na, ok := AsStr(naArg.V); ok {
				numberAlign = na
			} else {
				return nil, &TypeMismatchError{
					Expected: "alignment or none",
					Got:      naArg.V.Type().String(),
					Span:     naArg.Span,
				}
			}
		}
	}

	// Collect remaining positional arguments as children
	var items []*EnumItemElement
	startNum := 1
	if start != nil {
		startNum = *start
	}
	for i := 0; ; i++ {
		childArg := args.Eat()
		if childArg == nil {
			break
		}

		if cv, ok := childArg.V.(ContentValue); ok {
			// Calculate item number
			num := startNum + i
			if reversed {
				// For reversed, we'll handle this in layout
				// but still assign incrementing numbers for now
				num = startNum + i
			}
			// Wrap the content in an EnumItemElement
			items = append(items, &EnumItemElement{
				Number:  num,
				Content: cv.Content,
			})
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
			Items:       items,
			Tight:       tight,
			Numbering:   numbering,
			Start:       start,
			Full:        full,
			Reversed:    reversed,
			Indent:      indent,
			BodyIndent:  bodyIndent,
			Spacing:     spacing,
			NumberAlign: numberAlign,
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
	// Register list element function
	scope.DefineFunc("list", ListFunc())
	// Register enum element function
	scope.DefineFunc("enum", EnumFunc())
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
		"list":     ListFunc(),
		"enum":     EnumFunc(),
	}
}
