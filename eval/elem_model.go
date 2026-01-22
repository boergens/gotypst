package eval

import (
	"github.com/boergens/gotypst/syntax"
)

// ----------------------------------------------------------------------------
// Model Element Functions
// ----------------------------------------------------------------------------
// This file contains document structure element functions that correspond to
// Typst's model module:
// - par(), parbreak() - paragraph elements
// - heading() - section headings
// - list() - bullet lists
// - enum() - numbered lists
// - link() - hyperlinks
// - table(), table.cell() - table elements
//
// Reference: typst-reference/crates/typst-library/src/model/

// ----------------------------------------------------------------------------
// Par Element
// ----------------------------------------------------------------------------

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
func parNative(engine *Engine, context *Context, args *Args) (Value, error) {
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
func parbreakNative(engine *Engine, context *Context, args *Args) (Value, error) {
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
// Heading Element
// ----------------------------------------------------------------------------

// HeadingFunc creates the heading element function.
// Matches Rust: typst-library/src/model/heading.rs
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
					{Name: "level", Type: TypeInt, Default: Auto, Named: true},
					{Name: "depth", Type: TypeInt, Default: Int(1), Named: true},
					{Name: "offset", Type: TypeInt, Default: Int(0), Named: true},
					{Name: "numbering", Type: TypeStr, Default: None, Named: true},
					{Name: "supplement", Type: TypeContent, Default: Auto, Named: true},
					{Name: "outlined", Type: TypeBool, Default: True, Named: true},
					{Name: "bookmarked", Type: TypeBool, Default: Auto, Named: true},
					{Name: "hanging-indent", Type: TypeLength, Default: Auto, Named: true},
				},
			},
		},
	}
}

// headingNative implements the heading() function.
// Creates a HeadingElement from the given content with optional level and numbering.
// Matches Rust's HeadingElem struct.
//
// Arguments:
//   - body (positional, content): The heading's title
//   - level (named, int or auto, default: auto): Absolute nesting depth (computed from offset+depth if auto)
//   - depth (named, int, default: 1): Relative nesting depth
//   - offset (named, int, default: 0): Starting offset for level computation
//   - numbering (named, str or none, default: none): Numbering pattern (e.g., "1.", "1.a)")
//   - supplement (named, content or auto, default: auto): Supplement for references
//   - outlined (named, bool, default: true): Whether to show in outline
//   - bookmarked (named, bool or auto, default: auto): Whether to bookmark in PDF
//   - hanging-indent (named, length or auto, default: auto): Indent for multi-line headings
func headingNative(engine *Engine, context *Context, args *Args) (Value, error) {
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

	// Get optional level argument (default: auto, computed from offset + depth)
	var level *int
	if levelArg := args.Find("level"); levelArg != nil && !IsAuto(levelArg.V) {
		levelVal, ok := AsInt(levelArg.V)
		if !ok {
			return nil, &TypeMismatchError{
				Expected: "integer or auto",
				Got:      levelArg.V.Type().String(),
				Span:     levelArg.Span,
			}
		}
		l := int(levelVal)
		if l < 1 {
			return nil, &ConstructorError{
				Message: "heading level must be at least 1",
				Span:    levelArg.Span,
			}
		}
		level = &l
	}

	// Get optional depth argument (default: 1)
	depth := 1
	if depthArg := args.Find("depth"); depthArg != nil && !IsAuto(depthArg.V) && !IsNone(depthArg.V) {
		depthVal, ok := AsInt(depthArg.V)
		if !ok {
			return nil, &TypeMismatchError{
				Expected: "integer",
				Got:      depthArg.V.Type().String(),
				Span:     depthArg.Span,
			}
		}
		depth = int(depthVal)
		if depth < 1 {
			return nil, &ConstructorError{
				Message: "heading depth must be at least 1",
				Span:    depthArg.Span,
			}
		}
	}

	// Get optional offset argument (default: 0)
	offset := 0
	if offsetArg := args.Find("offset"); offsetArg != nil && !IsAuto(offsetArg.V) && !IsNone(offsetArg.V) {
		offsetVal, ok := AsInt(offsetArg.V)
		if !ok {
			return nil, &TypeMismatchError{
				Expected: "integer",
				Got:      offsetArg.V.Type().String(),
				Span:     offsetArg.Span,
			}
		}
		offset = int(offsetVal)
		if offset < 0 {
			return nil, &ConstructorError{
				Message: "heading offset must be non-negative",
				Span:    offsetArg.Span,
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

	// Get optional hanging-indent argument (default: auto)
	var hangingIndent *float64
	if hiArg := args.Find("hanging-indent"); hiArg != nil {
		if !IsAuto(hiArg.V) && !IsNone(hiArg.V) {
			if lv, ok := hiArg.V.(LengthValue); ok {
				hi := lv.Length.Points
				hangingIndent = &hi
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
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

	// Create the HeadingElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{&HeadingElement{
			Level:         level,
			Depth:         depth,
			Offset:        offset,
			Content:       body,
			Numbering:     numbering,
			Supplement:    supplement,
			Outlined:      outlined,
			Bookmarked:    bookmarked,
			HangingIndent: hangingIndent,
		}},
	}}, nil
}

// ----------------------------------------------------------------------------
// List Element
// ----------------------------------------------------------------------------

// ListFunc creates the list element function.
// Matches Rust: typst-library/src/model/list.rs
func ListFunc() *Func {
	name := "list"

	// Create scope with list.item method
	listScope := NewScope()
	listScope.DefineFunc("item", ListItemFunc())

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
			Scope: listScope,
		},
	}
}

// listNative implements the list() function.
// Creates a ListElement containing the given items.
// Matches Rust's ListElem struct.
//
// Arguments:
//   - tight (named, bool, default: true): Whether items have tight spacing
//   - marker (named, content, default: none): Custom marker content
//   - indent (named, length, default: none): Indentation of each item
//   - body-indent (named, length, default: 0.5em): Spacing between marker and body
//   - spacing (named, length or auto, default: auto): Spacing between items
//   - children (positional, variadic, content): The list items
func listNative(engine *Engine, context *Context, args *Args) (Value, error) {
	elem := &ListElement{}

	// Get optional tight argument (default: true)
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
			elem.Tight = &tightVal
		}
	}

	// Get optional marker argument
	if markerArg := args.Find("marker"); markerArg != nil {
		if !IsNone(markerArg.V) && !IsAuto(markerArg.V) {
			if cv, ok := markerArg.V.(ContentValue); ok {
				elem.Marker = &cv.Content
			} else if s, ok := AsStr(markerArg.V); ok {
				// Allow string markers for convenience
				c := Content{Elements: []ContentElement{&TextElement{Text: s}}}
				elem.Marker = &c
			} else {
				return nil, &TypeMismatchError{
					Expected: "content or string",
					Got:      markerArg.V.Type().String(),
					Span:     markerArg.Span,
				}
			}
		}
	}

	// Get optional indent argument
	if indentArg := args.Find("indent"); indentArg != nil {
		if !IsAuto(indentArg.V) && !IsNone(indentArg.V) {
			if lv, ok := indentArg.V.(LengthValue); ok {
				elem.Indent = &lv.Length.Points
			} else {
				return nil, &TypeMismatchError{
					Expected: "length",
					Got:      indentArg.V.Type().String(),
					Span:     indentArg.Span,
				}
			}
		}
	}

	// Get optional body-indent argument
	if biArg := args.Find("body-indent"); biArg != nil {
		if !IsAuto(biArg.V) && !IsNone(biArg.V) {
			if lv, ok := biArg.V.(LengthValue); ok {
				elem.BodyIndent = &lv.Length.Points
			} else {
				return nil, &TypeMismatchError{
					Expected: "length",
					Got:      biArg.V.Type().String(),
					Span:     biArg.Span,
				}
			}
		}
	}

	// Get optional spacing argument
	if spacingArg := args.Find("spacing"); spacingArg != nil {
		if !IsAuto(spacingArg.V) && !IsNone(spacingArg.V) {
			if lv, ok := spacingArg.V.(LengthValue); ok {
				elem.Spacing = &lv.Length.Points
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      spacingArg.V.Type().String(),
					Span:     spacingArg.Span,
				}
			}
		}
	}

	// Collect remaining positional arguments as list items
	for {
		childArg := args.Eat()
		if childArg == nil {
			break
		}

		if cv, ok := childArg.V.(ContentValue); ok {
			// Check if the content contains ListItemElements, otherwise wrap as item
			hasListItems := false
			for _, e := range cv.Content.Elements {
				if item, ok := e.(*ListItemElement); ok {
					elem.Items = append(elem.Items, item)
					hasListItems = true
				}
			}
			// If no list items found, treat the entire content as a single item
			if !hasListItems && len(cv.Content.Elements) > 0 {
				elem.Items = append(elem.Items, &ListItemElement{Content: cv.Content})
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
		Elements: []ContentElement{elem},
	}}, nil
}

// ListItemFunc creates the list.item element function.
func ListItemFunc() *Func {
	name := "item"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: listItemNative,
			Info: &FuncInfo{
				Name: "list.item",
				Params: []ParamInfo{
					{Name: "body", Type: TypeContent, Named: false},
				},
			},
		},
	}
}

// listItemNative implements the list.item() function.
func listItemNative(engine *Engine, context *Context, args *Args) (Value, error) {
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

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	return ContentValue{Content: Content{
		Elements: []ContentElement{&ListItemElement{Content: body}},
	}}, nil
}

// ----------------------------------------------------------------------------
// Enum Element
// ----------------------------------------------------------------------------

// EnumFunc creates the enum element function.
// Matches Rust: typst-library/src/model/enum.rs
func EnumFunc() *Func {
	name := "enum"

	// Create scope with enum.item method
	enumScope := NewScope()
	enumScope.DefineFunc("item", EnumItemFunc())

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
			Scope: enumScope,
		},
	}
}

// enumNative implements the enum() function.
// Creates an EnumElement containing the given items.
// Matches Rust's EnumElem struct.
//
// Arguments:
//   - tight (named, bool, default: true): Whether items have tight spacing
//   - numbering (named, str, default: "1."): Numbering pattern
//   - start (named, int or auto, default: auto): Starting number
//   - full (named, bool, default: false): Whether to display full numbering
//   - reversed (named, bool, default: false): Whether to reverse numbering
//   - indent (named, length, default: none): Indentation of each item
//   - body-indent (named, length, default: 0.5em): Spacing between number and body
//   - spacing (named, length or auto, default: auto): Spacing between items
//   - number-align (named, str, default: "end + top"): Alignment of numbers
//   - children (positional, variadic, content): The enum items
func enumNative(engine *Engine, context *Context, args *Args) (Value, error) {
	elem := &EnumElement{}

	// Get optional tight argument (default: true)
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
			elem.Tight = &tightVal
		}
	}

	// Get optional numbering argument
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
			elem.Numbering = &numStr
		}
	}

	// Get optional start argument
	if startArg := args.Find("start"); startArg != nil {
		if !IsAuto(startArg.V) && !IsNone(startArg.V) {
			startVal, ok := AsInt(startArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "integer or auto",
					Got:      startArg.V.Type().String(),
					Span:     startArg.Span,
				}
			}
			s := int(startVal)
			elem.Start = &s
		}
	}

	// Get optional full argument (default: false)
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
			elem.Full = &fullVal
		}
	}

	// Get optional reversed argument (default: false)
	if reversedArg := args.Find("reversed"); reversedArg != nil {
		if !IsAuto(reversedArg.V) && !IsNone(reversedArg.V) {
			reversedVal, ok := AsBool(reversedArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "bool",
					Got:      reversedArg.V.Type().String(),
					Span:     reversedArg.Span,
				}
			}
			elem.Reversed = &reversedVal
		}
	}

	// Get optional indent argument
	if indentArg := args.Find("indent"); indentArg != nil {
		if !IsAuto(indentArg.V) && !IsNone(indentArg.V) {
			if lv, ok := indentArg.V.(LengthValue); ok {
				elem.Indent = &lv.Length.Points
			} else {
				return nil, &TypeMismatchError{
					Expected: "length",
					Got:      indentArg.V.Type().String(),
					Span:     indentArg.Span,
				}
			}
		}
	}

	// Get optional body-indent argument
	if biArg := args.Find("body-indent"); biArg != nil {
		if !IsAuto(biArg.V) && !IsNone(biArg.V) {
			if lv, ok := biArg.V.(LengthValue); ok {
				elem.BodyIndent = &lv.Length.Points
			} else {
				return nil, &TypeMismatchError{
					Expected: "length",
					Got:      biArg.V.Type().String(),
					Span:     biArg.Span,
				}
			}
		}
	}

	// Get optional spacing argument
	if spacingArg := args.Find("spacing"); spacingArg != nil {
		if !IsAuto(spacingArg.V) && !IsNone(spacingArg.V) {
			if lv, ok := spacingArg.V.(LengthValue); ok {
				elem.Spacing = &lv.Length.Points
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
	if naArg := args.Find("number-align"); naArg != nil {
		if !IsAuto(naArg.V) && !IsNone(naArg.V) {
			if alignStr, ok := AsStr(naArg.V); ok {
				elem.NumberAlign = &alignStr
			} else {
				return nil, &TypeMismatchError{
					Expected: "alignment string",
					Got:      naArg.V.Type().String(),
					Span:     naArg.Span,
				}
			}
		}
	}

	// Collect remaining positional arguments as enum items
	itemNum := 1
	if elem.Start != nil {
		itemNum = *elem.Start
	}

	for {
		childArg := args.Eat()
		if childArg == nil {
			break
		}

		if cv, ok := childArg.V.(ContentValue); ok {
			// Check if the content contains EnumItemElements, otherwise wrap as item
			hasEnumItems := false
			for _, e := range cv.Content.Elements {
				if item, ok := e.(*EnumItemElement); ok {
					// Preserve existing item numbers if set, otherwise auto-number
					if item.Number == 0 {
						item.Number = itemNum
						itemNum++
					}
					elem.Items = append(elem.Items, item)
					hasEnumItems = true
				}
			}
			// If no enum items found, treat the entire content as a single item
			if !hasEnumItems && len(cv.Content.Elements) > 0 {
				elem.Items = append(elem.Items, &EnumItemElement{
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
		Elements: []ContentElement{elem},
	}}, nil
}

// EnumItemFunc creates the enum.item element function.
func EnumItemFunc() *Func {
	name := "item"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: enumItemNative,
			Info: &FuncInfo{
				Name: "enum.item",
				Params: []ParamInfo{
					{Name: "number", Type: TypeInt, Default: Auto, Named: false},
					{Name: "body", Type: TypeContent, Named: false},
				},
			},
		},
	}
}

// enumItemNative implements the enum.item() function.
func enumItemNative(engine *Engine, context *Context, args *Args) (Value, error) {
	elem := &EnumItemElement{}

	// Get optional number argument (first positional)
	numberArg := args.Eat()
	if numberArg != nil {
		if !IsAuto(numberArg.V) && !IsNone(numberArg.V) {
			if numVal, ok := AsInt(numberArg.V); ok {
				elem.Number = int(numVal)
			} else if cv, ok := numberArg.V.(ContentValue); ok {
				// It's actually the body, not a number
				elem.Content = cv.Content
				// Check for unexpected arguments
				if err := args.Finish(); err != nil {
					return nil, err
				}
				return ContentValue{Content: Content{
					Elements: []ContentElement{elem},
				}}, nil
			} else {
				return nil, &TypeMismatchError{
					Expected: "integer or content",
					Got:      numberArg.V.Type().String(),
					Span:     numberArg.Span,
				}
			}
		}
	}

	// Get required body argument (second positional)
	bodyArg := args.Find("body")
	if bodyArg == nil {
		bodyArg = args.Eat()
	}
	if bodyArg != nil {
		if cv, ok := bodyArg.V.(ContentValue); ok {
			elem.Content = cv.Content
		} else {
			return nil, &TypeMismatchError{
				Expected: "content",
				Got:      bodyArg.V.Type().String(),
				Span:     bodyArg.Span,
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
// Link Element
// ----------------------------------------------------------------------------

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

// stripURLContactScheme strips mailto: or tel: prefix from a URL for display.
// Returns the stripped string if a scheme was found, otherwise returns the original.
func stripURLContactScheme(url string) string {
	if len(url) >= 7 && url[:7] == "mailto:" {
		return url[7:]
	}
	if len(url) >= 4 && url[:4] == "tel:" {
		return url[4:]
	}
	return url
}

// bodyFromURL creates default body content from a URL.
// For mailto: and tel: URLs, the scheme is stripped from the display text.
func bodyFromURL(url string) Content {
	displayText := stripURLContactScheme(url)
	return Content{
		Elements: []ContentElement{&TextElement{Text: displayText}},
	}
}

// linkNative implements the link() function.
// Creates a LinkElement for hyperlinks.
//
// Arguments:
//   - dest (positional, str): The destination URL or label
//   - body (positional, content, default: none): The content to display (defaults to the URL)
func linkNative(engine *Engine, context *Context, args *Args) (Value, error) {
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

	// Validate URL (matches Typst: non-empty, max 8000 chars)
	if dest == "" {
		return nil, &InvalidArgumentError{
			Message: "URL must not be empty",
			Span:    destArg.Span,
		}
	}
	if len(dest) > 8000 {
		return nil, &InvalidArgumentError{
			Message: "URL is too long",
			Span:    destArg.Span,
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
	} else {
		// Default body to URL text with scheme stripping (matches Typst behavior)
		body := bodyFromURL(dest)
		elem.Body = &body
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
// Matches Rust: typst-library/src/model/table.rs
func TableFunc() *Func {
	name := "table"

	// Create scope with table sub-element methods
	tableScope := NewScope()
	tableScope.DefineFunc("cell", TableCellFunc())
	tableScope.DefineFunc("header", TableHeaderFunc())
	tableScope.DefineFunc("footer", TableFooterFunc())
	tableScope.DefineFunc("hline", TableHLineFunc())
	tableScope.DefineFunc("vline", TableVLineFunc())

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
func tableNative(engine *Engine, context *Context, args *Args) (Value, error) {
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

		// Check if it's a table sub-element
		if cv, ok := child.V.(ContentValue); ok {
			if len(cv.Content.Elements) == 1 {
				switch e := cv.Content.Elements[0].(type) {
				case *TableCellElement:
					elem.Children = append(elem.Children, TableChild{Cell: e})
					continue
				case *TableHeaderElement:
					elem.Children = append(elem.Children, TableChild{Header: e})
					continue
				case *TableFooterElement:
					elem.Children = append(elem.Children, TableChild{Footer: e})
					continue
				case *TableHLineElement:
					elem.Children = append(elem.Children, TableChild{HLine: e})
					continue
				case *TableVLineElement:
					elem.Children = append(elem.Children, TableChild{VLine: e})
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
func tableCellNative(engine *Engine, context *Context, args *Args) (Value, error) {
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

// TableHeaderFunc creates the table.header element function.
// Matches Rust: typst-library/src/model/table.rs TableHeader
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
					{Name: "level", Type: TypeInt, Default: Int(1), Named: true},
					{Name: "children", Type: TypeContent, Variadic: true},
				},
			},
		},
	}
}

// tableHeaderNative implements the table.header() function.
func tableHeaderNative(engine *Engine, context *Context, args *Args) (Value, error) {
	elem := &TableHeaderElement{
		Repeat: true,
		Level:  1,
	}

	// Get optional repeat argument (default: true)
	if repeatArg := args.Find("repeat"); repeatArg != nil {
		if !IsAuto(repeatArg.V) && !IsNone(repeatArg.V) {
			if repeatVal, ok := AsBool(repeatArg.V); ok {
				elem.Repeat = repeatVal
			} else {
				return nil, &TypeMismatchError{
					Expected: "bool",
					Got:      repeatArg.V.Type().String(),
					Span:     repeatArg.Span,
				}
			}
		}
	}

	// Get optional level argument (default: 1)
	if levelArg := args.Find("level"); levelArg != nil {
		if !IsAuto(levelArg.V) && !IsNone(levelArg.V) {
			if levelVal, ok := AsInt(levelArg.V); ok {
				if levelVal < 1 {
					return nil, &InvalidArgumentError{
						Message: "header level must be at least 1",
						Span:    levelArg.Span,
					}
				}
				elem.Level = int(levelVal)
			} else {
				return nil, &TypeMismatchError{
					Expected: "integer",
					Got:      levelArg.V.Type().String(),
					Span:     levelArg.Span,
				}
			}
		}
	}

	// Collect children (cells and lines)
	for {
		child := args.Eat()
		if child == nil {
			break
		}

		if cv, ok := child.V.(ContentValue); ok {
			if len(cv.Content.Elements) == 1 {
				switch e := cv.Content.Elements[0].(type) {
				case *TableCellElement:
					elem.Children = append(elem.Children, TableItem{Cell: e})
					continue
				case *TableHLineElement:
					elem.Children = append(elem.Children, TableItem{HLine: e})
					continue
				case *TableVLineElement:
					elem.Children = append(elem.Children, TableItem{VLine: e})
					continue
				}
			}
			// Wrap content in a cell
			elem.Children = append(elem.Children, TableItem{Cell: &TableCellElement{Body: cv.Content, Colspan: 1, Rowspan: 1}})
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

// TableFooterFunc creates the table.footer element function.
// Matches Rust: typst-library/src/model/table.rs TableFooter
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
					{Name: "children", Type: TypeContent, Variadic: true},
				},
			},
		},
	}
}

// tableFooterNative implements the table.footer() function.
func tableFooterNative(engine *Engine, context *Context, args *Args) (Value, error) {
	elem := &TableFooterElement{
		Repeat: true,
	}

	// Get optional repeat argument (default: true)
	if repeatArg := args.Find("repeat"); repeatArg != nil {
		if !IsAuto(repeatArg.V) && !IsNone(repeatArg.V) {
			if repeatVal, ok := AsBool(repeatArg.V); ok {
				elem.Repeat = repeatVal
			} else {
				return nil, &TypeMismatchError{
					Expected: "bool",
					Got:      repeatArg.V.Type().String(),
					Span:     repeatArg.Span,
				}
			}
		}
	}

	// Collect children (cells and lines)
	for {
		child := args.Eat()
		if child == nil {
			break
		}

		if cv, ok := child.V.(ContentValue); ok {
			if len(cv.Content.Elements) == 1 {
				switch e := cv.Content.Elements[0].(type) {
				case *TableCellElement:
					elem.Children = append(elem.Children, TableItem{Cell: e})
					continue
				case *TableHLineElement:
					elem.Children = append(elem.Children, TableItem{HLine: e})
					continue
				case *TableVLineElement:
					elem.Children = append(elem.Children, TableItem{VLine: e})
					continue
				}
			}
			// Wrap content in a cell
			elem.Children = append(elem.Children, TableItem{Cell: &TableCellElement{Body: cv.Content, Colspan: 1, Rowspan: 1}})
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

// TableHLineFunc creates the table.hline element function.
// Matches Rust: typst-library/src/model/table.rs TableHLine
func TableHLineFunc() *Func {
	name := "hline"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: tableHLineNative,
			Info: &FuncInfo{
				Name: "table.hline",
				Params: []ParamInfo{
					{Name: "y", Type: TypeInt, Default: Auto, Named: true},
					{Name: "start", Type: TypeInt, Default: Int(0), Named: true},
					{Name: "end", Type: TypeInt, Default: None, Named: true},
					{Name: "stroke", Type: TypeDyn, Default: None, Named: true},
					{Name: "position", Type: TypeStr, Default: None, Named: true},
				},
			},
		},
	}
}

// tableHLineNative implements the table.hline() function.
func tableHLineNative(engine *Engine, context *Context, args *Args) (Value, error) {
	elem := &TableHLineElement{}

	// Get optional y argument
	if yArg := args.Find("y"); yArg != nil && !IsAuto(yArg.V) && !IsNone(yArg.V) {
		if yVal, ok := AsInt(yArg.V); ok {
			y := int(yVal)
			elem.Y = &y
		} else {
			return nil, &TypeMismatchError{
				Expected: "integer or auto",
				Got:      yArg.V.Type().String(),
				Span:     yArg.Span,
			}
		}
	}

	// Get optional start argument (default: 0)
	if startArg := args.Find("start"); startArg != nil && !IsAuto(startArg.V) && !IsNone(startArg.V) {
		if startVal, ok := AsInt(startArg.V); ok {
			elem.Start = int(startVal)
		} else {
			return nil, &TypeMismatchError{
				Expected: "integer",
				Got:      startArg.V.Type().String(),
				Span:     startArg.Span,
			}
		}
	}

	// Get optional end argument
	if endArg := args.Find("end"); endArg != nil && !IsAuto(endArg.V) && !IsNone(endArg.V) {
		if endVal, ok := AsInt(endArg.V); ok {
			e := int(endVal)
			elem.End = &e
		} else {
			return nil, &TypeMismatchError{
				Expected: "integer or none",
				Got:      endArg.V.Type().String(),
				Span:     endArg.Span,
			}
		}
	}

	// Get optional stroke argument
	if strokeArg := args.Find("stroke"); strokeArg != nil && !IsNone(strokeArg.V) {
		elem.Stroke = strokeArg.V
	}

	// Get optional position argument
	if posArg := args.Find("position"); posArg != nil && !IsAuto(posArg.V) && !IsNone(posArg.V) {
		if posVal, ok := AsStr(posArg.V); ok {
			if posVal != "top" && posVal != "bottom" {
				return nil, &InvalidArgumentError{
					Message: "position must be \"top\" or \"bottom\"",
					Span:    posArg.Span,
				}
			}
			elem.Position = posVal
		} else {
			return nil, &TypeMismatchError{
				Expected: "string",
				Got:      posArg.V.Type().String(),
				Span:     posArg.Span,
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

// TableVLineFunc creates the table.vline element function.
// Matches Rust: typst-library/src/model/table.rs TableVLine
func TableVLineFunc() *Func {
	name := "vline"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: tableVLineNative,
			Info: &FuncInfo{
				Name: "table.vline",
				Params: []ParamInfo{
					{Name: "x", Type: TypeInt, Default: Auto, Named: true},
					{Name: "start", Type: TypeInt, Default: Int(0), Named: true},
					{Name: "end", Type: TypeInt, Default: None, Named: true},
					{Name: "stroke", Type: TypeDyn, Default: None, Named: true},
					{Name: "position", Type: TypeStr, Default: None, Named: true},
				},
			},
		},
	}
}

// tableVLineNative implements the table.vline() function.
func tableVLineNative(engine *Engine, context *Context, args *Args) (Value, error) {
	elem := &TableVLineElement{}

	// Get optional x argument
	if xArg := args.Find("x"); xArg != nil && !IsAuto(xArg.V) && !IsNone(xArg.V) {
		if xVal, ok := AsInt(xArg.V); ok {
			x := int(xVal)
			elem.X = &x
		} else {
			return nil, &TypeMismatchError{
				Expected: "integer or auto",
				Got:      xArg.V.Type().String(),
				Span:     xArg.Span,
			}
		}
	}

	// Get optional start argument (default: 0)
	if startArg := args.Find("start"); startArg != nil && !IsAuto(startArg.V) && !IsNone(startArg.V) {
		if startVal, ok := AsInt(startArg.V); ok {
			elem.Start = int(startVal)
		} else {
			return nil, &TypeMismatchError{
				Expected: "integer",
				Got:      startArg.V.Type().String(),
				Span:     startArg.Span,
			}
		}
	}

	// Get optional end argument
	if endArg := args.Find("end"); endArg != nil && !IsAuto(endArg.V) && !IsNone(endArg.V) {
		if endVal, ok := AsInt(endArg.V); ok {
			e := int(endVal)
			elem.End = &e
		} else {
			return nil, &TypeMismatchError{
				Expected: "integer or none",
				Got:      endArg.V.Type().String(),
				Span:     endArg.Span,
			}
		}
	}

	// Get optional stroke argument
	if strokeArg := args.Find("stroke"); strokeArg != nil && !IsNone(strokeArg.V) {
		elem.Stroke = strokeArg.V
	}

	// Get optional position argument
	if posArg := args.Find("position"); posArg != nil && !IsAuto(posArg.V) && !IsNone(posArg.V) {
		if posVal, ok := AsStr(posArg.V); ok {
			if posVal != "start" && posVal != "end" {
				return nil, &InvalidArgumentError{
					Message: "position must be \"start\" or \"end\"",
					Span:    posArg.Span,
				}
			}
			elem.Position = posVal
		} else {
			return nil, &TypeMismatchError{
				Expected: "string",
				Got:      posArg.V.Type().String(),
				Span:     posArg.Span,
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
