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
