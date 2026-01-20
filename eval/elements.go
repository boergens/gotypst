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

// PageElement represents a page configuration element.
// It sets page properties like size, margins, headers, and footers.
// When used with set rules, it configures the page layout for the document.
type PageElement struct {
	// Width is the page width (in points). If nil, uses default (A4).
	Width *float64
	// Height is the page height (in points). If nil, uses default (A4).
	Height *float64
	// Flipped indicates if width/height should be swapped.
	Flipped *bool
	// Margin is the default margin for all sides (in points).
	// Individual margins override this.
	Margin *float64
	// MarginTop is the top margin (in points).
	MarginTop *float64
	// MarginBottom is the bottom margin (in points).
	MarginBottom *float64
	// MarginLeft is the left margin (in points).
	MarginLeft *float64
	// MarginRight is the right margin (in points).
	MarginRight *float64
	// MarginInside is the inside margin for two-sided documents (in points).
	MarginInside *float64
	// MarginOutside is the outside margin for two-sided documents (in points).
	MarginOutside *float64
	// Header is the header content.
	Header *Content
	// Footer is the footer content.
	Footer *Content
	// HeaderAscent is the ascent of the header from the top of the page (in points).
	HeaderAscent *float64
	// FooterDescent is the descent of the footer from the bottom of the page (in points).
	FooterDescent *float64
	// Background is the background content.
	Background *Content
	// Foreground is the foreground content.
	Foreground *Content
	// Fill is the page background color.
	Fill *Color
	// Numbering is the page numbering pattern.
	Numbering *string
	// NumberAlign is the alignment of the page number.
	NumberAlign *string
	// Binding is the binding side (left or right).
	Binding *string
	// Columns is the number of columns.
	Columns *int
	// Body is the content for this page (when used as a content element).
	Body *Content
}

func (*PageElement) IsContentElement() {}

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
					{Name: "margin", Type: TypeLength, Default: Auto, Named: true},
					{Name: "margin-top", Type: TypeLength, Default: None, Named: true},
					{Name: "margin-bottom", Type: TypeLength, Default: None, Named: true},
					{Name: "margin-left", Type: TypeLength, Default: None, Named: true},
					{Name: "margin-right", Type: TypeLength, Default: None, Named: true},
					{Name: "margin-inside", Type: TypeLength, Default: None, Named: true},
					{Name: "margin-outside", Type: TypeLength, Default: None, Named: true},
					{Name: "header", Type: TypeContent, Default: None, Named: true},
					{Name: "footer", Type: TypeContent, Default: None, Named: true},
					{Name: "header-ascent", Type: TypeLength, Default: Auto, Named: true},
					{Name: "footer-descent", Type: TypeLength, Default: Auto, Named: true},
					{Name: "background", Type: TypeContent, Default: None, Named: true},
					{Name: "foreground", Type: TypeContent, Default: None, Named: true},
					{Name: "fill", Type: TypeColor, Default: None, Named: true},
					{Name: "numbering", Type: TypeStr, Default: None, Named: true},
					{Name: "number-align", Type: TypeStr, Default: None, Named: true},
					{Name: "binding", Type: TypeStr, Default: Auto, Named: true},
					{Name: "columns", Type: TypeInt, Default: Int(1), Named: true},
				},
			},
		},
	}
}

// pageNative implements the page() function.
// Creates a PageElement with the given configuration.
//
// Arguments:
//   - body (positional, content, default: none): Content for this page
//   - width (named, length, default: auto): Page width
//   - height (named, length, default: auto): Page height
//   - flipped (named, bool, default: false): Swap width/height
//   - margin (named, length, default: auto): Default margin for all sides
//   - margin-top (named, length, default: none): Top margin
//   - margin-bottom (named, length, default: none): Bottom margin
//   - margin-left (named, length, default: none): Left margin
//   - margin-right (named, length, default: none): Right margin
//   - margin-inside (named, length, default: none): Inside margin (two-sided)
//   - margin-outside (named, length, default: none): Outside margin (two-sided)
//   - header (named, content, default: none): Header content
//   - footer (named, content, default: none): Footer content
//   - header-ascent (named, length, default: auto): Header ascent
//   - footer-descent (named, length, default: auto): Footer descent
//   - background (named, content, default: none): Background content
//   - foreground (named, content, default: none): Foreground content
//   - fill (named, color, default: none): Page background color
//   - numbering (named, str, default: none): Page numbering pattern
//   - number-align (named, str, default: none): Page number alignment
//   - binding (named, str, default: auto): Binding side (left/right)
//   - columns (named, int, default: 1): Number of columns
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

	// Get optional margin argument (default for all sides)
	if marginArg := args.Find("margin"); marginArg != nil {
		if !IsAuto(marginArg.V) && !IsNone(marginArg.V) {
			if lv, ok := marginArg.V.(LengthValue); ok {
				m := lv.Length.Points
				elem.Margin = &m
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      marginArg.V.Type().String(),
					Span:     marginArg.Span,
				}
			}
		}
	}

	// Get optional margin-top argument
	if mtArg := args.Find("margin-top"); mtArg != nil {
		if !IsAuto(mtArg.V) && !IsNone(mtArg.V) {
			if lv, ok := mtArg.V.(LengthValue); ok {
				mt := lv.Length.Points
				elem.MarginTop = &mt
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      mtArg.V.Type().String(),
					Span:     mtArg.Span,
				}
			}
		}
	}

	// Get optional margin-bottom argument
	if mbArg := args.Find("margin-bottom"); mbArg != nil {
		if !IsAuto(mbArg.V) && !IsNone(mbArg.V) {
			if lv, ok := mbArg.V.(LengthValue); ok {
				mb := lv.Length.Points
				elem.MarginBottom = &mb
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      mbArg.V.Type().String(),
					Span:     mbArg.Span,
				}
			}
		}
	}

	// Get optional margin-left argument
	if mlArg := args.Find("margin-left"); mlArg != nil {
		if !IsAuto(mlArg.V) && !IsNone(mlArg.V) {
			if lv, ok := mlArg.V.(LengthValue); ok {
				ml := lv.Length.Points
				elem.MarginLeft = &ml
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      mlArg.V.Type().String(),
					Span:     mlArg.Span,
				}
			}
		}
	}

	// Get optional margin-right argument
	if mrArg := args.Find("margin-right"); mrArg != nil {
		if !IsAuto(mrArg.V) && !IsNone(mrArg.V) {
			if lv, ok := mrArg.V.(LengthValue); ok {
				mr := lv.Length.Points
				elem.MarginRight = &mr
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      mrArg.V.Type().String(),
					Span:     mrArg.Span,
				}
			}
		}
	}

	// Get optional margin-inside argument
	if miArg := args.Find("margin-inside"); miArg != nil {
		if !IsAuto(miArg.V) && !IsNone(miArg.V) {
			if lv, ok := miArg.V.(LengthValue); ok {
				mi := lv.Length.Points
				elem.MarginInside = &mi
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      miArg.V.Type().String(),
					Span:     miArg.Span,
				}
			}
		}
	}

	// Get optional margin-outside argument
	if moArg := args.Find("margin-outside"); moArg != nil {
		if !IsAuto(moArg.V) && !IsNone(moArg.V) {
			if lv, ok := moArg.V.(LengthValue); ok {
				mo := lv.Length.Points
				elem.MarginOutside = &mo
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      moArg.V.Type().String(),
					Span:     moArg.Span,
				}
			}
		}
	}

	// Get optional header argument
	if headerArg := args.Find("header"); headerArg != nil {
		if !IsNone(headerArg.V) {
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
		if !IsNone(footerArg.V) {
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
	if haArg := args.Find("header-ascent"); haArg != nil {
		if !IsAuto(haArg.V) && !IsNone(haArg.V) {
			if lv, ok := haArg.V.(LengthValue); ok {
				ha := lv.Length.Points
				elem.HeaderAscent = &ha
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      haArg.V.Type().String(),
					Span:     haArg.Span,
				}
			}
		}
	}

	// Get optional footer-descent argument
	if fdArg := args.Find("footer-descent"); fdArg != nil {
		if !IsAuto(fdArg.V) && !IsNone(fdArg.V) {
			if lv, ok := fdArg.V.(LengthValue); ok {
				fd := lv.Length.Points
				elem.FooterDescent = &fd
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      fdArg.V.Type().String(),
					Span:     fdArg.Span,
				}
			}
		}
	}

	// Get optional background argument
	if bgArg := args.Find("background"); bgArg != nil {
		if !IsNone(bgArg.V) {
			if cv, ok := bgArg.V.(ContentValue); ok {
				elem.Background = &cv.Content
			} else {
				return nil, &TypeMismatchError{
					Expected: "content or none",
					Got:      bgArg.V.Type().String(),
					Span:     bgArg.Span,
				}
			}
		}
	}

	// Get optional foreground argument
	if fgArg := args.Find("foreground"); fgArg != nil {
		if !IsNone(fgArg.V) {
			if cv, ok := fgArg.V.(ContentValue); ok {
				elem.Foreground = &cv.Content
			} else {
				return nil, &TypeMismatchError{
					Expected: "content or none",
					Got:      fgArg.V.Type().String(),
					Span:     fgArg.Span,
				}
			}
		}
	}

	// Get optional fill argument
	if fillArg := args.Find("fill"); fillArg != nil {
		if !IsNone(fillArg.V) {
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
	if numArg := args.Find("numbering"); numArg != nil {
		if !IsNone(numArg.V) {
			if s, ok := AsStr(numArg.V); ok {
				elem.Numbering = &s
			} else {
				return nil, &TypeMismatchError{
					Expected: "string or none",
					Got:      numArg.V.Type().String(),
					Span:     numArg.Span,
				}
			}
		}
	}

	// Get optional number-align argument
	if naArg := args.Find("number-align"); naArg != nil {
		if !IsNone(naArg.V) {
			if s, ok := AsStr(naArg.V); ok {
				// Validate alignment value
				if s != "start" && s != "center" && s != "end" && s != "left" && s != "right" && s != "top" && s != "bottom" {
					return nil, &TypeMismatchError{
						Expected: "alignment",
						Got:      "\"" + s + "\"",
						Span:     naArg.Span,
					}
				}
				elem.NumberAlign = &s
			} else {
				return nil, &TypeMismatchError{
					Expected: "alignment or none",
					Got:      naArg.V.Type().String(),
					Span:     naArg.Span,
				}
			}
		}
	}

	// Get optional binding argument
	if bindArg := args.Find("binding"); bindArg != nil {
		if !IsAuto(bindArg.V) && !IsNone(bindArg.V) {
			if s, ok := AsStr(bindArg.V); ok {
				// Validate binding value
				if s != "left" && s != "right" {
					return nil, &TypeMismatchError{
						Expected: "\"left\" or \"right\"",
						Got:      "\"" + s + "\"",
						Span:     bindArg.Span,
					}
				}
				elem.Binding = &s
			} else {
				return nil, &TypeMismatchError{
					Expected: "string or auto",
					Got:      bindArg.V.Type().String(),
					Span:     bindArg.Span,
				}
			}
		}
	}

	// Get optional columns argument
	if colArg := args.Find("columns"); colArg != nil {
		if !IsAuto(colArg.V) && !IsNone(colArg.V) {
			if iv, ok := AsInt(colArg.V); ok {
				cols := int(iv)
				if cols < 1 {
					return nil, &TypeMismatchError{
						Expected: "positive integer",
						Got:      "non-positive integer",
						Span:     colArg.Span,
					}
				}
				elem.Columns = &cols
			} else {
				return nil, &TypeMismatchError{
					Expected: "integer",
					Got:      colArg.V.Type().String(),
					Span:     colArg.Span,
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
		"page":     PageFunc(),
	}
}
