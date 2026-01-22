package eval

import (
	"fmt"

	"github.com/boergens/gotypst/library/foundations"
	"github.com/boergens/gotypst/library/layout"
	"github.com/boergens/gotypst/syntax"
)

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
					{Name: "paper", Type: TypeStr, Default: Str("a4"), Named: false},
					{Name: "width", Type: TypeLength, Default: Auto, Named: true},
					{Name: "height", Type: TypeLength, Default: Auto, Named: true},
					{Name: "flipped", Type: TypeBool, Default: False, Named: true},
					{Name: "margin", Type: TypeDyn, Default: Auto, Named: true},
					{Name: "columns", Type: TypeInt, Default: Int(1), Named: true},
					{Name: "fill", Type: TypeDyn, Default: Auto, Named: true},
					{Name: "numbering", Type: TypeDyn, Default: None, Named: true},
					{Name: "number-align", Type: TypeDyn, Default: None, Named: true},
					{Name: "header", Type: TypeDyn, Default: Auto, Named: true},
					{Name: "header-ascent", Type: TypeLength, Default: None, Named: true},
					{Name: "footer", Type: TypeDyn, Default: Auto, Named: true},
					{Name: "footer-descent", Type: TypeLength, Default: None, Named: true},
					{Name: "background", Type: TypeDyn, Default: None, Named: true},
					{Name: "foreground", Type: TypeDyn, Default: None, Named: true},
					{Name: "body", Type: TypeContent, Default: None, Named: false},
				},
			},
		},
	}
}

// pageNative implements the page() function.
// Creates a PageElement with the given configuration.
//
// Arguments:
//   - paper (positional, string, default: "a4"): Standard paper size
//   - width (named, length, default: auto): Page width
//   - height (named, length, default: auto): Page height
//   - flipped (named, bool, default: false): Landscape orientation
//   - margin (named, length or dict, default: auto): Page margins
//   - columns (named, int, default: 1): Number of columns
//   - fill (named, color, default: auto): Background fill
//   - numbering (named, string or function, default: none): Page numbering
//   - number-align (named, alignment, default: center+bottom): Number alignment
//   - header (named, content, default: auto): Page header
//   - header-ascent (named, length, default: 30%): Header ascent
//   - footer (named, content, default: auto): Page footer
//   - footer-descent (named, length, default: 30%): Footer descent
//   - background (named, content, default: none): Background content
//   - foreground (named, content, default: none): Foreground content
//   - body (positional, content, default: none): Page body content
func pageNative(engine foundations.Engine, context foundations.Context, args *Args) (Value, error) {
	elem := &layout.PageElement{}

	// Get optional paper argument (positional or named)
	paperArg := args.Find("paper")
	if paperArg == nil {
		// Try to peek at first positional arg
		if peeked := args.Peek(); peeked != nil {
			if _, ok := AsStr(peeked.V); ok {
				paperArgSpanned, _ := args.Expect("paper")
				paperArg = &paperArgSpanned
			}
		}
	}
	if paperArg != nil {
		if !IsAuto(paperArg.V) && !IsNone(paperArg.V) {
			paperStr, ok := AsStr(paperArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "string",
					Got:      paperArg.V.Type().String(),
					Span:     paperArg.Span,
				}
			}
			paper, exists := layout.Papers[paperStr]
			if !exists {
				return nil, &ConstructorError{
					Message: fmt.Sprintf("unknown paper size: %q", paperStr),
					Span:    paperArg.Span,
				}
			}
			elem.Paper = &paper.Name
			elem.Width = &paper.Width
			elem.Height = &paper.Height
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
		if IsAuto(heightArg.V) {
			elem.HeightAuto = true
		} else if !IsNone(heightArg.V) {
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
		if !IsNone(flippedArg.V) && !IsAuto(flippedArg.V) {
			flipped, ok := AsBool(flippedArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "bool",
					Got:      flippedArg.V.Type().String(),
					Span:     flippedArg.Span,
				}
			}
			elem.Flipped = flipped
		}
	}

	// Get optional margin argument
	if marginArg := args.Find("margin"); marginArg != nil {
		if !IsAuto(marginArg.V) && !IsNone(marginArg.V) {
			// Can be a single length or a dictionary
			switch m := marginArg.V.(type) {
			case LengthValue:
				pts := m.Length.Points
				elem.MarginLeft = &pts
				elem.MarginTop = &pts
				elem.MarginRight = &pts
				elem.MarginBottom = &pts
			case *DictValue:
				// Parse individual margins from dict
				if err := parseMarginDict(m, elem, marginArg.Span); err != nil {
					return nil, err
				}
			default:
				elem.Margin = marginArg.V
			}
		}
	}

	// Get optional columns argument
	if colsArg := args.Find("columns"); colsArg != nil {
		if !IsAuto(colsArg.V) && !IsNone(colsArg.V) {
			cols, ok := AsInt(colsArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "integer",
					Got:      colsArg.V.Type().String(),
					Span:     colsArg.Span,
				}
			}
			colsInt := int(cols)
			if colsInt < 1 {
				return nil, &ConstructorError{
					Message: "column count must be at least 1",
					Span:    colsArg.Span,
				}
			}
			elem.Columns = &colsInt
		}
	}

	// Get optional fill argument
	if fillArg := args.Find("fill"); fillArg != nil {
		if !IsAuto(fillArg.V) && !IsNone(fillArg.V) {
			elem.Fill = fillArg.V
		}
	}

	// Get optional numbering argument
	if numbArg := args.Find("numbering"); numbArg != nil {
		if !IsNone(numbArg.V) {
			elem.Numbering = numbArg.V
		}
	}

	// Get optional number-align argument
	if naArg := args.Find("number-align"); naArg != nil {
		if !IsNone(naArg.V) && !IsAuto(naArg.V) {
			elem.NumberAlign = naArg.V
		}
	}

	// Get optional header argument
	if headerArg := args.Find("header"); headerArg != nil {
		if !IsAuto(headerArg.V) && !IsNone(headerArg.V) {
			elem.Header = headerArg.V
		}
	}

	// Get optional header-ascent argument
	if haArg := args.Find("header-ascent"); haArg != nil {
		if !IsNone(haArg.V) && !IsAuto(haArg.V) {
			if lv, ok := haArg.V.(LengthValue); ok {
				ha := lv.Length.Points
				elem.HeaderAscent = &ha
			}
		}
	}

	// Get optional footer argument
	if footerArg := args.Find("footer"); footerArg != nil {
		if !IsAuto(footerArg.V) && !IsNone(footerArg.V) {
			elem.Footer = footerArg.V
		}
	}

	// Get optional footer-descent argument
	if fdArg := args.Find("footer-descent"); fdArg != nil {
		if !IsNone(fdArg.V) && !IsAuto(fdArg.V) {
			if lv, ok := fdArg.V.(LengthValue); ok {
				fd := lv.Length.Points
				elem.FooterDescent = &fd
			}
		}
	}

	// Get optional background argument
	if bgArg := args.Find("background"); bgArg != nil {
		if !IsNone(bgArg.V) {
			elem.Background = bgArg.V
		}
	}

	// Get optional foreground argument
	if fgArg := args.Find("foreground"); fgArg != nil {
		if !IsNone(fgArg.V) {
			elem.Foreground = fgArg.V
		}
	}

	// Get optional body argument (positional)
	if bodyArg := args.Eat(); bodyArg != nil {
		if cv, ok := bodyArg.V.(ContentValue); ok {
			elem.Body = cv.Content
		} else if !IsNone(bodyArg.V) {
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

	// Create the PageElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}

// parseMarginDict parses a dictionary of margin values.
func parseMarginDict(d *DictValue, elem *layout.PageElement, span syntax.Span) error {
	// Helper to get and parse a length from the dict
	getLength := func(key string) *float64 {
		val, ok := d.Get(key)
		if !ok {
			return nil
		}
		lv, ok := val.(LengthValue)
		if !ok {
			return nil
		}
		pts := lv.Length.Points
		return &pts
	}

	// Check for "rest" first (applies to all unset sides)
	if pts := getLength("rest"); pts != nil {
		elem.MarginLeft = pts
		elem.MarginTop = pts
		elem.MarginRight = pts
		elem.MarginBottom = pts
	}

	// Check for "x" and "y" shortcuts
	if pts := getLength("x"); pts != nil {
		elem.MarginLeft = pts
		elem.MarginRight = pts
	}
	if pts := getLength("y"); pts != nil {
		elem.MarginTop = pts
		elem.MarginBottom = pts
	}

	// Check individual sides (override shortcuts)
	if pts := getLength("left"); pts != nil {
		elem.MarginLeft = pts
	}
	if pts := getLength("top"); pts != nil {
		elem.MarginTop = pts
	}
	if pts := getLength("right"); pts != nil {
		elem.MarginRight = pts
	}
	if pts := getLength("bottom"); pts != nil {
		elem.MarginBottom = pts
	}

	return nil
}
