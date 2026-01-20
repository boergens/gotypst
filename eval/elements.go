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
// Layout Element Functions
// ----------------------------------------------------------------------------

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
					{Name: "left", Type: TypeRelative, Default: None, Named: true},
					{Name: "top", Type: TypeRelative, Default: None, Named: true},
					{Name: "right", Type: TypeRelative, Default: None, Named: true},
					{Name: "bottom", Type: TypeRelative, Default: None, Named: true},
					{Name: "x", Type: TypeRelative, Default: None, Named: true},
					{Name: "y", Type: TypeRelative, Default: None, Named: true},
					{Name: "rest", Type: TypeRelative, Default: None, Named: true},
				},
			},
		},
	}
}

// padNative implements the pad() function.
// Creates a PadElement from the given content with optional padding values.
//
// Arguments:
//   - body (positional, content): The content to pad
//   - left (named, relative, default: 0pt): The padding at the left side
//   - top (named, relative, default: 0pt): The padding at the top side
//   - right (named, relative, default: 0pt): The padding at the right side
//   - bottom (named, relative, default: 0pt): The padding at the bottom side
//   - x (named, relative): Shorthand setting both left and right
//   - y (named, relative): Shorthand setting both top and bottom
//   - rest (named, relative): Shorthand setting all four sides
func padNative(vm *Vm, args *Args) (Value, error) {
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

	// Create element with zero padding defaults
	elem := &PadElement{
		Body: body,
	}

	// Helper to parse a relative value from an argument
	parseRelative := func(arg *syntax.Spanned[Value]) (Relative, error) {
		if IsNone(arg.V) || IsAuto(arg.V) {
			return Relative{}, nil
		}
		switch v := arg.V.(type) {
		case LengthValue:
			return Relative{Abs: v.Length}, nil
		case RatioValue:
			return Relative{Rel: v.Ratio}, nil
		case RelativeValue:
			return v.Relative, nil
		default:
			return Relative{}, &TypeMismatchError{
				Expected: "relative length",
				Got:      arg.V.Type().String(),
				Span:     arg.Span,
			}
		}
	}

	// Get optional rest argument (sets all four sides)
	if restArg := args.Find("rest"); restArg != nil {
		rel, err := parseRelative(restArg)
		if err != nil {
			return nil, err
		}
		elem.Left = rel
		elem.Top = rel
		elem.Right = rel
		elem.Bottom = rel
	}

	// Get optional x argument (sets left and right)
	if xArg := args.Find("x"); xArg != nil {
		rel, err := parseRelative(xArg)
		if err != nil {
			return nil, err
		}
		elem.Left = rel
		elem.Right = rel
	}

	// Get optional y argument (sets top and bottom)
	if yArg := args.Find("y"); yArg != nil {
		rel, err := parseRelative(yArg)
		if err != nil {
			return nil, err
		}
		elem.Top = rel
		elem.Bottom = rel
	}

	// Get individual side arguments (override shorthands)
	if leftArg := args.Find("left"); leftArg != nil {
		rel, err := parseRelative(leftArg)
		if err != nil {
			return nil, err
		}
		elem.Left = rel
	}

	if topArg := args.Find("top"); topArg != nil {
		rel, err := parseRelative(topArg)
		if err != nil {
			return nil, err
		}
		elem.Top = rel
	}

	if rightArg := args.Find("right"); rightArg != nil {
		rel, err := parseRelative(rightArg)
		if err != nil {
			return nil, err
		}
		elem.Right = rel
	}

	if bottomArg := args.Find("bottom"); bottomArg != nil {
		rel, err := parseRelative(bottomArg)
		if err != nil {
			return nil, err
		}
		elem.Bottom = rel
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

// PlaceFunc creates the place element function.
func PlaceFunc() *Func {
	name := "place"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: placeNative,
			Info: &FuncInfo{
				Name: "place",
				Params: []ParamInfo{
					{Name: "body", Type: TypeContent, Named: false},
					{Name: "alignment", Type: TypeStr, Default: Auto, Named: true},
					{Name: "scope", Type: TypeStr, Default: Str("column"), Named: true},
					{Name: "float", Type: TypeBool, Default: False, Named: true},
					{Name: "clearance", Type: TypeLength, Default: None, Named: true},
					{Name: "dx", Type: TypeRelative, Default: None, Named: true},
					{Name: "dy", Type: TypeRelative, Default: None, Named: true},
				},
			},
		},
	}
}

// placeNative implements the place() function.
// Creates a PlaceElement for positioning content relative to its parent container.
//
// Arguments:
//   - body (positional, content): The content to place
//   - alignment (named, alignment or auto, default: start): Position in parent container
//   - scope (named, str, default: "column"): Placement scope ("column" or "parent")
//   - float (named, bool, default: false): Enables floating layout
//   - clearance (named, length, default: 1.5em): Spacing for floating elements
//   - dx (named, relative, default: 0pt): Horizontal offset
//   - dy (named, relative, default: 0pt): Vertical offset
func placeNative(vm *Vm, args *Args) (Value, error) {
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
	elem := &PlaceElement{
		Body:      body,
		Scope:     "column",
		Clearance: 18.0, // 1.5em at 12pt font = 18pt
	}

	// Helper to parse a relative value from an argument
	parseRelative := func(arg *syntax.Spanned[Value]) (Relative, error) {
		if IsNone(arg.V) || IsAuto(arg.V) {
			return Relative{}, nil
		}
		switch v := arg.V.(type) {
		case LengthValue:
			return Relative{Abs: v.Length}, nil
		case RatioValue:
			return Relative{Rel: v.Ratio}, nil
		case RelativeValue:
			return v.Relative, nil
		default:
			return Relative{}, &TypeMismatchError{
				Expected: "relative length",
				Got:      arg.V.Type().String(),
				Span:     arg.Span,
			}
		}
	}

	// Get optional alignment argument
	// Alignment can be specified as positional or named
	if alignArg := args.Find("alignment"); alignArg != nil {
		if !IsAuto(alignArg.V) && !IsNone(alignArg.V) {
			// Parse alignment value
			// For now, accept string representations: "left", "right", "center", "top", "bottom", etc.
			if alignStr, ok := AsStr(alignArg.V); ok {
				// Parse 2D alignment from string
				switch alignStr {
				case "left", "start":
					elem.AlignmentX = "start"
				case "right", "end":
					elem.AlignmentX = "end"
				case "center":
					elem.AlignmentX = "center"
				case "top":
					elem.AlignmentY = "top"
				case "bottom":
					elem.AlignmentY = "bottom"
				case "horizon":
					elem.AlignmentY = "horizon"
				default:
					// Could be compound like "top + left"
					elem.AlignmentX = alignStr
				}
			}
		}
	}

	// Get optional scope argument
	if scopeArg := args.Find("scope"); scopeArg != nil {
		if !IsNone(scopeArg.V) {
			scopeStr, ok := AsStr(scopeArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "string",
					Got:      scopeArg.V.Type().String(),
					Span:     scopeArg.Span,
				}
			}
			if scopeStr != "column" && scopeStr != "parent" {
				return nil, &TypeMismatchError{
					Expected: "\"column\" or \"parent\"",
					Got:      "\"" + scopeStr + "\"",
					Span:     scopeArg.Span,
				}
			}
			elem.Scope = scopeStr
		}
	}

	// Get optional float argument
	if floatArg := args.Find("float"); floatArg != nil {
		if !IsNone(floatArg.V) {
			floatVal, ok := AsBool(floatArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "bool",
					Got:      floatArg.V.Type().String(),
					Span:     floatArg.Span,
				}
			}
			elem.Float = floatVal
		}
	}

	// Get optional clearance argument
	if clearanceArg := args.Find("clearance"); clearanceArg != nil {
		if !IsNone(clearanceArg.V) && !IsAuto(clearanceArg.V) {
			if cv, ok := clearanceArg.V.(LengthValue); ok {
				elem.Clearance = cv.Length.Points
			} else {
				return nil, &TypeMismatchError{
					Expected: "length",
					Got:      clearanceArg.V.Type().String(),
					Span:     clearanceArg.Span,
				}
			}
		}
	}

	// Get optional dx argument
	if dxArg := args.Find("dx"); dxArg != nil {
		rel, err := parseRelative(dxArg)
		if err != nil {
			return nil, err
		}
		elem.Dx = rel
	}

	// Get optional dy argument
	if dyArg := args.Find("dy"); dyArg != nil {
		rel, err := parseRelative(dyArg)
		if err != nil {
			return nil, err
		}
		elem.Dy = rel
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Create the PlaceElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}

// PlaceFlushFunc creates the place.flush element function.
func PlaceFlushFunc() *Func {
	name := "flush"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: placeFlushNative,
			Info: &FuncInfo{
				Name:   "flush",
				Params: []ParamInfo{},
			},
		},
	}
}

// placeFlushNative implements the place.flush() function.
// Creates a PlaceFlushElement to signal that floats should be flushed.
func placeFlushNative(vm *Vm, args *Args) (Value, error) {
	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Create the PlaceFlushElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{&PlaceFlushElement{}},
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
	// Register pad element function
	scope.DefineFunc("pad", PadFunc())
	// Register place element function with flush sub-function
	placeFunc := PlaceFunc()
	scope.DefineFunc("place", placeFunc)
}

// ElementFunctions returns a map of all element function names to their functions.
// This is useful for introspection and testing.
func ElementFunctions() map[string]*Func {
	return map[string]*Func{
		"raw":      RawFunc(),
		"par":      ParFunc(),
		"parbreak": ParbreakFunc(),
		"pad":      PadFunc(),
		"place":    PlaceFunc(),
	}
}
