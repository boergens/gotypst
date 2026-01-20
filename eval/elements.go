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
// List Element Functions
// ----------------------------------------------------------------------------

// ListFunc creates the list (bullet list) element function.
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
					{Name: "children", Type: TypeContent, Named: false, Variadic: true},
					{Name: "marker", Type: TypeContent, Default: None, Named: true},
					{Name: "indent", Type: TypeLength, Default: None, Named: true},
					{Name: "body-indent", Type: TypeLength, Default: None, Named: true},
					{Name: "spacing", Type: TypeLength, Default: Auto, Named: true},
					{Name: "tight", Type: TypeBool, Default: True, Named: true},
				},
			},
		},
	}
}

// listNative implements the list() function.
// Creates a ListElement from variadic content items with optional styling.
//
// Arguments:
//   - children (variadic, content): The list items
//   - marker (named, content or none, default: none): Custom bullet marker
//   - indent (named, length or none, default: none): Indent before marker
//   - body-indent (named, length or none, default: none): Indent after marker
//   - spacing (named, length or auto, default: auto): Space between items
//   - tight (named, bool, default: true): Whether to use tight spacing
func listNative(vm *Vm, args *Args) (Value, error) {
	// Create element with defaults
	elem := &ListElement{
		Children: []Content{},
	}

	// Get optional marker argument
	if markerArg := args.Find("marker"); markerArg != nil {
		if !IsNone(markerArg.V) {
			if ms, ok := AsStr(markerArg.V); ok {
				elem.Marker = &ms
			} else if cv, ok := markerArg.V.(ContentValue); ok {
				// Convert content to string representation for marker
				marker := contentToText(cv.Content)
				elem.Marker = &marker
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
	if indentArg := args.Find("indent"); indentArg != nil {
		if !IsNone(indentArg.V) && !IsAuto(indentArg.V) {
			if lv, ok := indentArg.V.(LengthValue); ok {
				indent := lv.Length.Points
				elem.Indent = &indent
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
	if biArg := args.Find("body-indent"); biArg != nil {
		if !IsNone(biArg.V) && !IsAuto(biArg.V) {
			if lv, ok := biArg.V.(LengthValue); ok {
				bi := lv.Length.Points
				elem.BodyIndent = &bi
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
	if spacingArg := args.Find("spacing"); spacingArg != nil {
		if !IsNone(spacingArg.V) && !IsAuto(spacingArg.V) {
			if lv, ok := spacingArg.V.(LengthValue); ok {
				spacing := lv.Length.Points
				elem.Spacing = &spacing
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      spacingArg.V.Type().String(),
					Span:     spacingArg.Span,
				}
			}
		}
	}

	// Get optional tight argument
	if tightArg := args.Find("tight"); tightArg != nil {
		if tv, ok := AsBool(tightArg.V); ok {
			elem.Tight = &tv
		} else {
			return nil, &TypeMismatchError{
				Expected: "bool",
				Got:      tightArg.V.Type().String(),
				Span:     tightArg.Span,
			}
		}
	}

	// Collect variadic children (positional arguments)
	for {
		arg := args.Eat()
		if arg == nil {
			break // No more positional arguments
		}
		if cv, ok := arg.V.(ContentValue); ok {
			elem.Children = append(elem.Children, cv.Content)
		} else {
			return nil, &TypeMismatchError{
				Expected: "content",
				Got:      arg.V.Type().String(),
				Span:     arg.Span,
			}
		}
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Create the ListElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}

// EnumFunc creates the enum (numbered list) element function.
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
					{Name: "children", Type: TypeContent, Named: false, Variadic: true},
					{Name: "numbering", Type: TypeStr, Default: None, Named: true},
					{Name: "start", Type: TypeInt, Default: None, Named: true},
					{Name: "full", Type: TypeBool, Default: None, Named: true},
					{Name: "indent", Type: TypeLength, Default: None, Named: true},
					{Name: "body-indent", Type: TypeLength, Default: None, Named: true},
					{Name: "spacing", Type: TypeLength, Default: Auto, Named: true},
					{Name: "tight", Type: TypeBool, Default: True, Named: true},
				},
			},
		},
	}
}

// enumNative implements the enum() function.
// Creates an EnumElement from variadic content items with optional styling.
//
// Arguments:
//   - children (variadic, content): The enum items
//   - numbering (named, str or none, default: none): Number format pattern
//   - start (named, int or none, default: none): Starting number
//   - full (named, bool or none, default: none): Whether to show full numbering
//   - indent (named, length or none, default: none): Indent before number
//   - body-indent (named, length or none, default: none): Indent after number
//   - spacing (named, length or auto, default: auto): Space between items
//   - tight (named, bool, default: true): Whether to use tight spacing
func enumNative(vm *Vm, args *Args) (Value, error) {
	// Create element with defaults
	elem := &EnumElement{
		Children: []Content{},
	}

	// Get optional numbering argument
	if numberingArg := args.Find("numbering"); numberingArg != nil {
		if !IsNone(numberingArg.V) {
			if ns, ok := AsStr(numberingArg.V); ok {
				elem.Numbering = &ns
			} else {
				return nil, &TypeMismatchError{
					Expected: "string or none",
					Got:      numberingArg.V.Type().String(),
					Span:     numberingArg.Span,
				}
			}
		}
	}

	// Get optional start argument
	if startArg := args.Find("start"); startArg != nil {
		if !IsNone(startArg.V) {
			if si, ok := AsInt(startArg.V); ok {
				startInt := int(si)
				elem.Start = &startInt
			} else {
				return nil, &TypeMismatchError{
					Expected: "int or none",
					Got:      startArg.V.Type().String(),
					Span:     startArg.Span,
				}
			}
		}
	}

	// Get optional full argument
	if fullArg := args.Find("full"); fullArg != nil {
		if !IsNone(fullArg.V) {
			if fv, ok := AsBool(fullArg.V); ok {
				elem.Full = &fv
			} else {
				return nil, &TypeMismatchError{
					Expected: "bool or none",
					Got:      fullArg.V.Type().String(),
					Span:     fullArg.Span,
				}
			}
		}
	}

	// Get optional indent argument
	if indentArg := args.Find("indent"); indentArg != nil {
		if !IsNone(indentArg.V) && !IsAuto(indentArg.V) {
			if lv, ok := indentArg.V.(LengthValue); ok {
				indent := lv.Length.Points
				elem.Indent = &indent
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
	if biArg := args.Find("body-indent"); biArg != nil {
		if !IsNone(biArg.V) && !IsAuto(biArg.V) {
			if lv, ok := biArg.V.(LengthValue); ok {
				bi := lv.Length.Points
				elem.BodyIndent = &bi
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
	if spacingArg := args.Find("spacing"); spacingArg != nil {
		if !IsNone(spacingArg.V) && !IsAuto(spacingArg.V) {
			if lv, ok := spacingArg.V.(LengthValue); ok {
				spacing := lv.Length.Points
				elem.Spacing = &spacing
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      spacingArg.V.Type().String(),
					Span:     spacingArg.Span,
				}
			}
		}
	}

	// Get optional tight argument
	if tightArg := args.Find("tight"); tightArg != nil {
		if tv, ok := AsBool(tightArg.V); ok {
			elem.Tight = &tv
		} else {
			return nil, &TypeMismatchError{
				Expected: "bool",
				Got:      tightArg.V.Type().String(),
				Span:     tightArg.Span,
			}
		}
	}

	// Collect variadic children (positional arguments)
	for {
		arg := args.Eat()
		if arg == nil {
			break // No more positional arguments
		}
		if cv, ok := arg.V.(ContentValue); ok {
			elem.Children = append(elem.Children, cv.Content)
		} else {
			return nil, &TypeMismatchError{
				Expected: "content",
				Got:      arg.V.Type().String(),
				Span:     arg.Span,
			}
		}
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Create the EnumElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}

// TermsFunc creates the terms (description list) element function.
func TermsFunc() *Func {
	name := "terms"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: termsNative,
			Info: &FuncInfo{
				Name: "terms",
				Params: []ParamInfo{
					{Name: "children", Type: TypeContent, Named: false, Variadic: true},
					{Name: "separator", Type: TypeContent, Default: None, Named: true},
					{Name: "indent", Type: TypeLength, Default: None, Named: true},
					{Name: "hanging-indent", Type: TypeLength, Default: None, Named: true},
					{Name: "spacing", Type: TypeLength, Default: Auto, Named: true},
					{Name: "tight", Type: TypeBool, Default: True, Named: true},
				},
			},
		},
	}
}

// termsNative implements the terms() function.
// Creates a TermsElement from variadic term items with optional styling.
//
// Arguments:
//   - children (variadic, content): The term items
//   - separator (named, content or none, default: none): Separator between term and description
//   - indent (named, length or none, default: none): Indent for terms
//   - hanging-indent (named, length or none, default: none): Hanging indent
//   - spacing (named, length or auto, default: auto): Space between items
//   - tight (named, bool, default: true): Whether to use tight spacing
func termsNative(vm *Vm, args *Args) (Value, error) {
	// Create element with defaults
	elem := &TermsElement{
		Children: []TermItemElement{},
	}

	// Get optional separator argument
	if sepArg := args.Find("separator"); sepArg != nil {
		if !IsNone(sepArg.V) {
			if ss, ok := AsStr(sepArg.V); ok {
				elem.Separator = &ss
			} else if cv, ok := sepArg.V.(ContentValue); ok {
				// Convert content to string representation for separator
				sep := contentToText(cv.Content)
				elem.Separator = &sep
			} else {
				return nil, &TypeMismatchError{
					Expected: "content or none",
					Got:      sepArg.V.Type().String(),
					Span:     sepArg.Span,
				}
			}
		}
	}

	// Get optional indent argument
	if indentArg := args.Find("indent"); indentArg != nil {
		if !IsNone(indentArg.V) && !IsAuto(indentArg.V) {
			if lv, ok := indentArg.V.(LengthValue); ok {
				indent := lv.Length.Points
				elem.Indent = &indent
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      indentArg.V.Type().String(),
					Span:     indentArg.Span,
				}
			}
		}
	}

	// Get optional hanging-indent argument
	if hiArg := args.Find("hanging-indent"); hiArg != nil {
		if !IsNone(hiArg.V) && !IsAuto(hiArg.V) {
			if lv, ok := hiArg.V.(LengthValue); ok {
				hi := lv.Length.Points
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

	// Get optional spacing argument
	if spacingArg := args.Find("spacing"); spacingArg != nil {
		if !IsNone(spacingArg.V) && !IsAuto(spacingArg.V) {
			if lv, ok := spacingArg.V.(LengthValue); ok {
				spacing := lv.Length.Points
				elem.Spacing = &spacing
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      spacingArg.V.Type().String(),
					Span:     spacingArg.Span,
				}
			}
		}
	}

	// Get optional tight argument
	if tightArg := args.Find("tight"); tightArg != nil {
		if tv, ok := AsBool(tightArg.V); ok {
			elem.Tight = &tv
		} else {
			return nil, &TypeMismatchError{
				Expected: "bool",
				Got:      tightArg.V.Type().String(),
				Span:     tightArg.Span,
			}
		}
	}

	// Collect variadic children (positional arguments)
	// For terms, children are expected to be TermItemElements embedded in content
	for {
		arg := args.Eat()
		if arg == nil {
			break // No more positional arguments
		}
		if cv, ok := arg.V.(ContentValue); ok {
			// Extract TermItemElements from the content
			for _, e := range cv.Content.Elements {
				if ti, ok := e.(*TermItemElement); ok {
					elem.Children = append(elem.Children, *ti)
				}
			}
		} else {
			return nil, &TypeMismatchError{
				Expected: "content",
				Got:      arg.V.Type().String(),
				Span:     arg.Span,
			}
		}
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Create the TermsElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}

// contentToText extracts text content from a Content value.
// This is a helper for converting content markers/separators to strings.
func contentToText(c Content) string {
	var result string
	for _, elem := range c.Elements {
		if te, ok := elem.(*TextElement); ok {
			result += te.Text
		}
	}
	return result
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
	// Register list element functions
	scope.DefineFunc("list", ListFunc())
	scope.DefineFunc("enum", EnumFunc())
	scope.DefineFunc("terms", TermsFunc())
}

// ElementFunctions returns a map of all element function names to their functions.
// This is useful for introspection and testing.
func ElementFunctions() map[string]*Func {
	return map[string]*Func{
		"raw":      RawFunc(),
		"par":      ParFunc(),
		"parbreak": ParbreakFunc(),
		"list":     ListFunc(),
		"enum":     EnumFunc(),
		"terms":    TermsFunc(),
	}
}
