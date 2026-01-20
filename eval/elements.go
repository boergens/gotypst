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
// Stack Element
// ----------------------------------------------------------------------------

// StackDir represents the stacking direction.
type StackDir int

const (
	// StackDirLTR stacks left-to-right (horizontal).
	StackDirLTR StackDir = iota
	// StackDirRTL stacks right-to-left (horizontal).
	StackDirRTL
	// StackDirTTB stacks top-to-bottom (vertical, default).
	StackDirTTB
	// StackDirBTT stacks bottom-to-top (vertical).
	StackDirBTT
)

// String returns the string representation of the direction.
func (d StackDir) String() string {
	switch d {
	case StackDirLTR:
		return "ltr"
	case StackDirRTL:
		return "rtl"
	case StackDirTTB:
		return "ttb"
	case StackDirBTT:
		return "btt"
	default:
		return "ttb"
	}
}

// StackElement represents a stack layout container.
// It arranges its children linearly in a specified direction.
type StackElement struct {
	// Dir is the stacking direction (ltr, rtl, ttb, btt).
	// Default is ttb (top-to-bottom).
	Dir StackDir
	// Spacing is the space between children (in points).
	// If nil, no spacing is applied.
	Spacing *float64
	// Children is the content to stack.
	Children []Content
}

func (*StackElement) isContentElement() {}

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
					{Name: "dir", Type: TypeStr, Default: None, Named: true},
					{Name: "spacing", Type: TypeLength, Default: None, Named: true},
					{Name: "children", Type: TypeContent, Variadic: true, Named: false},
				},
			},
		},
	}
}

// stackNative implements the stack() function.
// Creates a StackElement for linear layout of children.
//
// Arguments:
//   - dir (named, str, default: "ttb"): Direction - "ltr", "rtl", "ttb", or "btt"
//   - spacing (named, length, default: none): Spacing between children
//   - ..children (positional, content): Content to stack
func stackNative(vm *Vm, args *Args) (Value, error) {
	// Create element with defaults
	elem := &StackElement{
		Dir: StackDirTTB, // default top-to-bottom
	}

	// Get optional dir argument
	if dirArg := args.Find("dir"); dirArg != nil {
		if !IsNone(dirArg.V) {
			dirStr, ok := AsStr(dirArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "string or none",
					Got:      dirArg.V.Type().String(),
					Span:     dirArg.Span,
				}
			}
			switch dirStr {
			case "ltr":
				elem.Dir = StackDirLTR
			case "rtl":
				elem.Dir = StackDirRTL
			case "ttb":
				elem.Dir = StackDirTTB
			case "btt":
				elem.Dir = StackDirBTT
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
	if spacingArg := args.Find("spacing"); spacingArg != nil {
		if !IsNone(spacingArg.V) {
			if lv, ok := spacingArg.V.(LengthValue); ok {
				spacing := lv.Length.Points
				elem.Spacing = &spacing
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      spacingArg.V.Type().String(),
					Span:     spacingArg.Span,
				}
			}
		}
	}

	// Get all remaining positional arguments as children
	childArgs := args.All()
	for _, childArg := range childArgs {
		if cv, ok := childArg.V.(ContentValue); ok {
			elem.Children = append(elem.Children, cv.Content)
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
		Elements: []ContentElement{elem},
	}}, nil
}

// ----------------------------------------------------------------------------
// Align Element
// ----------------------------------------------------------------------------

// HAlignment represents horizontal alignment.
type HAlignment int

const (
	// HAlignStart aligns to the start (left in LTR, right in RTL).
	HAlignStart HAlignment = iota
	// HAlignCenter centers horizontally.
	HAlignCenter
	// HAlignEnd aligns to the end (right in LTR, left in RTL).
	HAlignEnd
	// HAlignLeft always aligns left.
	HAlignLeft
	// HAlignRight always aligns right.
	HAlignRight
)

// String returns the string representation of horizontal alignment.
func (a HAlignment) String() string {
	switch a {
	case HAlignStart:
		return "start"
	case HAlignCenter:
		return "center"
	case HAlignEnd:
		return "end"
	case HAlignLeft:
		return "left"
	case HAlignRight:
		return "right"
	default:
		return "start"
	}
}

// VAlignment represents vertical alignment.
type VAlignment int

const (
	// VAlignTop aligns to the top.
	VAlignTop VAlignment = iota
	// VAlignHorizon centers vertically.
	VAlignHorizon
	// VAlignBottom aligns to the bottom.
	VAlignBottom
)

// String returns the string representation of vertical alignment.
func (a VAlignment) String() string {
	switch a {
	case VAlignTop:
		return "top"
	case VAlignHorizon:
		return "horizon"
	case VAlignBottom:
		return "bottom"
	default:
		return "top"
	}
}

// AlignElement represents an alignment container.
// It positions its content according to the specified alignment.
type AlignElement struct {
	// Body is the content to align.
	Body Content
	// Horizontal is the horizontal alignment.
	// If nil, defaults to start.
	Horizontal *HAlignment
	// Vertical is the vertical alignment.
	// If nil, defaults to top.
	Vertical *VAlignment
}

func (*AlignElement) isContentElement() {}

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
					{Name: "alignment", Type: TypeStr, Default: None, Named: false},
					{Name: "body", Type: TypeContent, Named: false},
				},
			},
		},
	}
}

// alignNative implements the align() function.
// Creates an AlignElement for content positioning.
//
// Arguments:
//   - alignment (positional, str): Alignment value like "center", "left", "right",
//     "top", "bottom", "horizon", "start", "end", or combinations like "center + horizon"
//   - body (positional, content): Content to align
func alignNative(vm *Vm, args *Args) (Value, error) {
	// Get required alignment argument
	alignArg := args.Find("alignment")
	if alignArg == nil {
		alignArgSpanned, err := args.Expect("alignment")
		if err != nil {
			return nil, err
		}
		alignArg = &alignArgSpanned
	}

	elem := &AlignElement{}

	// Parse alignment value
	alignStr, ok := AsStr(alignArg.V)
	if !ok {
		return nil, &TypeMismatchError{
			Expected: "string",
			Got:      alignArg.V.Type().String(),
			Span:     alignArg.Span,
		}
	}

	// Parse alignment string (can be single value or "h + v" combination)
	if err := parseAlignment(alignStr, elem, alignArg.Span); err != nil {
		return nil, err
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

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Create the AlignElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}

// parseAlignment parses an alignment string and sets the element's alignment fields.
func parseAlignment(s string, elem *AlignElement, span syntax.Span) error {
	// Handle simple single values
	switch s {
	case "start":
		h := HAlignStart
		elem.Horizontal = &h
		return nil
	case "center":
		h := HAlignCenter
		elem.Horizontal = &h
		return nil
	case "end":
		h := HAlignEnd
		elem.Horizontal = &h
		return nil
	case "left":
		h := HAlignLeft
		elem.Horizontal = &h
		return nil
	case "right":
		h := HAlignRight
		elem.Horizontal = &h
		return nil
	case "top":
		v := VAlignTop
		elem.Vertical = &v
		return nil
	case "horizon":
		v := VAlignHorizon
		elem.Vertical = &v
		return nil
	case "bottom":
		v := VAlignBottom
		elem.Vertical = &v
		return nil
	}

	// Handle combinations like "center + horizon" or "left + top"
	// For now, we'll support simple parsing - a more robust parser would handle whitespace variations
	parts := splitAlignment(s)
	if len(parts) == 2 {
		for _, part := range parts {
			switch part {
			case "start":
				h := HAlignStart
				elem.Horizontal = &h
			case "center":
				h := HAlignCenter
				elem.Horizontal = &h
			case "end":
				h := HAlignEnd
				elem.Horizontal = &h
			case "left":
				h := HAlignLeft
				elem.Horizontal = &h
			case "right":
				h := HAlignRight
				elem.Horizontal = &h
			case "top":
				v := VAlignTop
				elem.Vertical = &v
			case "horizon":
				v := VAlignHorizon
				elem.Vertical = &v
			case "bottom":
				v := VAlignBottom
				elem.Vertical = &v
			default:
				return &TypeMismatchError{
					Expected: "valid alignment (start, center, end, left, right, top, horizon, bottom)",
					Got:      "\"" + part + "\"",
					Span:     span,
				}
			}
		}
		return nil
	}

	return &TypeMismatchError{
		Expected: "valid alignment (start, center, end, left, right, top, horizon, bottom)",
		Got:      "\"" + s + "\"",
		Span:     span,
	}
}

// splitAlignment splits an alignment string like "center + horizon" into parts.
func splitAlignment(s string) []string {
	var parts []string
	var current string
	for _, c := range s {
		if c == '+' {
			if trimmed := trimSpaces(current); trimmed != "" {
				parts = append(parts, trimmed)
			}
			current = ""
		} else {
			current += string(c)
		}
	}
	if trimmed := trimSpaces(current); trimmed != "" {
		parts = append(parts, trimmed)
	}
	return parts
}

// trimSpaces removes leading and trailing spaces from a string.
func trimSpaces(s string) string {
	start := 0
	end := len(s)
	for start < end && s[start] == ' ' {
		start++
	}
	for end > start && s[end-1] == ' ' {
		end--
	}
	return s[start:end]
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
	}
}
