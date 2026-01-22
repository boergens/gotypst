package layout

import (
	"fmt"

	"github.com/boergens/gotypst/library/foundations"
	"github.com/boergens/gotypst/syntax"
)

// Paper represents a standard paper size.
//
// Reference: typst-reference/crates/typst-library/src/layout/page.rs
type Paper struct {
	Name   string
	Width  float64 // in points
	Height float64 // in points
}

// Standard paper sizes (width and height in points).
var Papers = map[string]Paper{
	"a0":         {Name: "a0", Width: 2383.94, Height: 3370.39},
	"a1":         {Name: "a1", Width: 1683.78, Height: 2383.94},
	"a2":         {Name: "a2", Width: 1190.55, Height: 1683.78},
	"a3":         {Name: "a3", Width: 841.89, Height: 1190.55},
	"a4":         {Name: "a4", Width: 595.28, Height: 841.89},
	"a5":         {Name: "a5", Width: 419.53, Height: 595.28},
	"a6":         {Name: "a6", Width: 297.64, Height: 419.53},
	"a7":         {Name: "a7", Width: 209.76, Height: 297.64},
	"a8":         {Name: "a8", Width: 147.40, Height: 209.76},
	"us-letter":  {Name: "us-letter", Width: 612, Height: 792},
	"us-legal":   {Name: "us-legal", Width: 612, Height: 1008},
	"us-tabloid": {Name: "us-tabloid", Width: 792, Height: 1224},
}

// PageElement represents a page layout element.
// It configures page properties like size, margins, headers, footers, etc.
//
// Reference: typst-reference/crates/typst-library/src/layout/page.rs
type PageElement struct {
	// Paper is a standard paper size name (e.g., "a4", "us-letter").
	Paper *string
	// Width is the page width in points. If nil, uses paper width.
	Width *float64
	// Height is the page height in points. If nil, uses paper height.
	// Can also be "auto" for infinite height.
	Height *float64
	// HeightAuto indicates height should grow to fit content.
	HeightAuto bool
	// Flipped indicates landscape orientation.
	Flipped bool
	// Margin is the page margins. Can be a single value or per-side.
	Margin foundations.Value
	// MarginLeft, MarginTop, MarginRight, MarginBottom are individual margins.
	MarginLeft   *float64
	MarginTop    *float64
	MarginRight  *float64
	MarginBottom *float64
	// Columns is the number of columns on the page.
	Columns *int
	// Fill is the page background fill.
	Fill foundations.Value
	// Numbering is the page numbering pattern.
	Numbering foundations.Value
	// NumberAlign is the alignment of page numbers.
	NumberAlign foundations.Value
	// Header is the page header content.
	Header foundations.Value
	// HeaderAscent is how much the header is raised into the margin.
	HeaderAscent *float64
	// Footer is the page footer content.
	Footer foundations.Value
	// FooterDescent is how much the footer is lowered into the margin.
	FooterDescent *float64
	// Background is content behind the page body.
	Background foundations.Value
	// Foreground is content in front of the page body.
	Foreground foundations.Value
	// Body is the page content (only when used as constructor).
	Body foundations.Content
}

func (*PageElement) IsContentElement() {}

// PageFunc creates the page element function.
func PageFunc() *foundations.Func {
	name := "page"
	return &foundations.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: foundations.NativeFunc{
			Func: pageNative,
			Info: &foundations.FuncInfo{
				Name: "page",
				Params: []foundations.ParamInfo{
					{Name: "paper", Type: foundations.TypeStr, Default: foundations.Str("a4"), Named: false},
					{Name: "width", Type: foundations.TypeLength, Default: foundations.Auto, Named: true},
					{Name: "height", Type: foundations.TypeLength, Default: foundations.Auto, Named: true},
					{Name: "flipped", Type: foundations.TypeBool, Default: foundations.False, Named: true},
					{Name: "margin", Type: foundations.TypeDyn, Default: foundations.Auto, Named: true},
					{Name: "columns", Type: foundations.TypeInt, Default: foundations.Int(1), Named: true},
					{Name: "fill", Type: foundations.TypeDyn, Default: foundations.Auto, Named: true},
					{Name: "numbering", Type: foundations.TypeDyn, Default: foundations.None, Named: true},
					{Name: "number-align", Type: foundations.TypeDyn, Default: foundations.None, Named: true},
					{Name: "header", Type: foundations.TypeDyn, Default: foundations.Auto, Named: true},
					{Name: "header-ascent", Type: foundations.TypeLength, Default: foundations.None, Named: true},
					{Name: "footer", Type: foundations.TypeDyn, Default: foundations.Auto, Named: true},
					{Name: "footer-descent", Type: foundations.TypeLength, Default: foundations.None, Named: true},
					{Name: "background", Type: foundations.TypeDyn, Default: foundations.None, Named: true},
					{Name: "foreground", Type: foundations.TypeDyn, Default: foundations.None, Named: true},
					{Name: "body", Type: foundations.TypeContent, Default: foundations.None, Named: false},
				},
			},
		},
	}
}

// pageNative implements the page() function.
func pageNative(engine foundations.Engine, context foundations.Context, args *foundations.Args) (foundations.Value, error) {
	elem := &PageElement{}

	// Get optional paper argument (positional or named)
	paperArg := args.Find("paper")
	if paperArg == nil {
		if peeked := args.Peek(); peeked != nil {
			if _, ok := foundations.AsStr(peeked.V); ok {
				paperArgSpanned, _ := args.Expect("paper")
				paperArg = &paperArgSpanned
			}
		}
	}
	if paperArg != nil {
		if !foundations.IsAuto(paperArg.V) && !foundations.IsNone(paperArg.V) {
			paperStr, ok := foundations.AsStr(paperArg.V)
			if !ok {
				return nil, &foundations.TypeMismatchError{
					Expected: "string",
					Got:      paperArg.V.Type().String(),
					Span:     paperArg.Span,
				}
			}
			paper, exists := Papers[paperStr]
			if !exists {
				return nil, &foundations.ConstructorError{
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
		if !foundations.IsAuto(widthArg.V) && !foundations.IsNone(widthArg.V) {
			if lv, ok := widthArg.V.(foundations.LengthValue); ok {
				w := lv.Length.Points
				elem.Width = &w
			} else {
				return nil, &foundations.TypeMismatchError{
					Expected: "length or auto",
					Got:      widthArg.V.Type().String(),
					Span:     widthArg.Span,
				}
			}
		}
	}

	// Get optional height argument
	if heightArg := args.Find("height"); heightArg != nil {
		if foundations.IsAuto(heightArg.V) {
			elem.HeightAuto = true
		} else if !foundations.IsNone(heightArg.V) {
			if lv, ok := heightArg.V.(foundations.LengthValue); ok {
				h := lv.Length.Points
				elem.Height = &h
			} else {
				return nil, &foundations.TypeMismatchError{
					Expected: "length or auto",
					Got:      heightArg.V.Type().String(),
					Span:     heightArg.Span,
				}
			}
		}
	}

	// Get optional flipped argument
	if flippedArg := args.Find("flipped"); flippedArg != nil {
		if !foundations.IsNone(flippedArg.V) && !foundations.IsAuto(flippedArg.V) {
			flipped, ok := foundations.AsBool(flippedArg.V)
			if !ok {
				return nil, &foundations.TypeMismatchError{
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
		if !foundations.IsAuto(marginArg.V) && !foundations.IsNone(marginArg.V) {
			switch m := marginArg.V.(type) {
			case foundations.LengthValue:
				pts := m.Length.Points
				elem.MarginLeft = &pts
				elem.MarginTop = &pts
				elem.MarginRight = &pts
				elem.MarginBottom = &pts
			case *foundations.Dict:
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
		if !foundations.IsAuto(colsArg.V) && !foundations.IsNone(colsArg.V) {
			cols, ok := foundations.AsInt(colsArg.V)
			if !ok {
				return nil, &foundations.TypeMismatchError{
					Expected: "integer",
					Got:      colsArg.V.Type().String(),
					Span:     colsArg.Span,
				}
			}
			colsInt := int(cols)
			if colsInt < 1 {
				return nil, &foundations.ConstructorError{
					Message: "column count must be at least 1",
					Span:    colsArg.Span,
				}
			}
			elem.Columns = &colsInt
		}
	}

	// Get optional fill argument
	if fillArg := args.Find("fill"); fillArg != nil {
		if !foundations.IsAuto(fillArg.V) && !foundations.IsNone(fillArg.V) {
			elem.Fill = fillArg.V
		}
	}

	// Get optional numbering argument
	if numbArg := args.Find("numbering"); numbArg != nil {
		if !foundations.IsNone(numbArg.V) {
			elem.Numbering = numbArg.V
		}
	}

	// Get optional number-align argument
	if naArg := args.Find("number-align"); naArg != nil {
		if !foundations.IsNone(naArg.V) && !foundations.IsAuto(naArg.V) {
			elem.NumberAlign = naArg.V
		}
	}

	// Get optional header argument
	if headerArg := args.Find("header"); headerArg != nil {
		if !foundations.IsAuto(headerArg.V) && !foundations.IsNone(headerArg.V) {
			elem.Header = headerArg.V
		}
	}

	// Get optional header-ascent argument
	if haArg := args.Find("header-ascent"); haArg != nil {
		if !foundations.IsNone(haArg.V) && !foundations.IsAuto(haArg.V) {
			if lv, ok := haArg.V.(foundations.LengthValue); ok {
				ha := lv.Length.Points
				elem.HeaderAscent = &ha
			}
		}
	}

	// Get optional footer argument
	if footerArg := args.Find("footer"); footerArg != nil {
		if !foundations.IsAuto(footerArg.V) && !foundations.IsNone(footerArg.V) {
			elem.Footer = footerArg.V
		}
	}

	// Get optional footer-descent argument
	if fdArg := args.Find("footer-descent"); fdArg != nil {
		if !foundations.IsNone(fdArg.V) && !foundations.IsAuto(fdArg.V) {
			if lv, ok := fdArg.V.(foundations.LengthValue); ok {
				fd := lv.Length.Points
				elem.FooterDescent = &fd
			}
		}
	}

	// Get optional background argument
	if bgArg := args.Find("background"); bgArg != nil {
		if !foundations.IsNone(bgArg.V) {
			elem.Background = bgArg.V
		}
	}

	// Get optional foreground argument
	if fgArg := args.Find("foreground"); fgArg != nil {
		if !foundations.IsNone(fgArg.V) {
			elem.Foreground = fgArg.V
		}
	}

	// Get optional body argument (positional)
	if bodyArg := args.Eat(); bodyArg != nil {
		if cv, ok := bodyArg.V.(foundations.ContentValue); ok {
			elem.Body = cv.Content
		} else if !foundations.IsNone(bodyArg.V) {
			return nil, &foundations.TypeMismatchError{
				Expected: "content or none",
				Got:      bodyArg.V.Type().String(),
				Span:     bodyArg.Span,
			}
		}
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	return foundations.ContentValue{Content: foundations.Content{
		Elements: []foundations.ContentElement{elem},
	}}, nil
}

// parseMarginDict parses a dictionary of margin values.
func parseMarginDict(d *foundations.Dict, elem *PageElement, span syntax.Span) error {
	getLength := func(key string) *float64 {
		val, ok := d.Get(key)
		if !ok {
			return nil
		}
		lv, ok := val.(foundations.LengthValue)
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
