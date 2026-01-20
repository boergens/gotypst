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

// ----------------------------------------------------------------------------
// Page Element Function
// ----------------------------------------------------------------------------

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
					{Name: "paper", Type: TypeStr, Default: None, Named: true},
					{Name: "width", Type: TypeAuto, Default: Auto, Named: true},
					{Name: "height", Type: TypeAuto, Default: Auto, Named: true},
					{Name: "flipped", Type: TypeBool, Default: False, Named: true},
					{Name: "margin", Type: TypeAuto, Default: Auto, Named: true},
					{Name: "binding", Type: TypeAuto, Default: Auto, Named: true},
					{Name: "columns", Type: TypeInt, Default: Int(1), Named: true},
					{Name: "fill", Type: TypeAuto, Default: Auto, Named: true},
					{Name: "numbering", Type: TypeStr, Default: None, Named: true},
					{Name: "number-align", Type: TypeAuto, Default: Auto, Named: true},
					{Name: "header", Type: TypeContent, Default: None, Named: true},
					{Name: "header-ascent", Type: TypeRelative, Default: None, Named: true},
					{Name: "footer", Type: TypeContent, Default: None, Named: true},
					{Name: "footer-descent", Type: TypeRelative, Default: None, Named: true},
					{Name: "background", Type: TypeContent, Default: None, Named: true},
					{Name: "foreground", Type: TypeContent, Default: None, Named: true},
					{Name: "body", Type: TypeContent, Named: false},
				},
			},
		},
	}
}

// pageNative implements the page() function.
// Creates a PageElement with the specified configuration.
//
// Arguments:
//   - paper (named, str or none): Paper size name (e.g., "a4", "us-letter")
//   - width (named, auto or length): Explicit page width
//   - height (named, auto or length): Explicit page height
//   - flipped (named, bool): Whether to flip width/height (landscape)
//   - margin (named, auto or length or dict): Page margins
//   - binding (named, auto or alignment): Binding side for two-sided documents
//   - columns (named, int): Number of text columns
//   - fill (named, auto or none or color): Page background fill
//   - numbering (named, str or none): Page numbering pattern
//   - number-align (named, auto or alignment): Page number alignment
//   - header (named, content or none): Page header
//   - header-ascent (named, relative or none): Header vertical offset
//   - footer (named, content or none): Page footer
//   - footer-descent (named, relative or none): Footer vertical offset
//   - background (named, content or none): Background content
//   - foreground (named, content or none): Foreground content
//   - body (positional, content): Page content
func pageNative(_ *Vm, args *Args) (Value, error) {
	elem := &PageElement{
		Columns: 1, // Default to 1 column
	}

	// Get optional paper argument
	if paperArg := args.Find("paper"); paperArg != nil && !IsNone(paperArg.V) {
		paper, ok := AsStr(paperArg.V)
		if !ok {
			return nil, &TypeMismatchError{
				Expected: "string or none",
				Got:      paperArg.V.Type().String(),
				Span:     paperArg.Span,
			}
		}
		elem.Paper = paper
	}

	// Get optional width argument
	if widthArg := args.Find("width"); widthArg != nil && !IsAuto(widthArg.V) {
		rel, err := parseRelative(widthArg)
		if err != nil {
			return nil, err
		}
		elem.Width = rel
	}

	// Get optional height argument
	if heightArg := args.Find("height"); heightArg != nil && !IsAuto(heightArg.V) {
		rel, err := parseRelative(heightArg)
		if err != nil {
			return nil, err
		}
		elem.Height = rel
	}

	// Get optional flipped argument
	if flippedArg := args.Find("flipped"); flippedArg != nil {
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

	// Get optional margin argument
	if marginArg := args.Find("margin"); marginArg != nil && !IsAuto(marginArg.V) {
		margin, err := parsePageMargin(marginArg)
		if err != nil {
			return nil, err
		}
		elem.Margin = margin
	}

	// Get optional binding argument
	if bindingArg := args.Find("binding"); bindingArg != nil && !IsAuto(bindingArg.V) {
		binding, err := parsePageBinding(bindingArg)
		if err != nil {
			return nil, err
		}
		elem.Binding = binding
	}

	// Get optional columns argument
	if columnsArg := args.Find("columns"); columnsArg != nil {
		cols, ok := AsInt(columnsArg.V)
		if !ok {
			return nil, &TypeMismatchError{
				Expected: "int",
				Got:      columnsArg.V.Type().String(),
				Span:     columnsArg.Span,
			}
		}
		if cols < 1 {
			cols = 1
		}
		elem.Columns = int(cols)
	}

	// Get optional fill argument
	if fillArg := args.Find("fill"); fillArg != nil && !IsAuto(fillArg.V) && !IsNone(fillArg.V) {
		elem.Fill = fillArg.V
	}

	// Get optional numbering argument
	if numberingArg := args.Find("numbering"); numberingArg != nil && !IsNone(numberingArg.V) {
		numbering, ok := AsStr(numberingArg.V)
		if !ok {
			return nil, &TypeMismatchError{
				Expected: "string or none",
				Got:      numberingArg.V.Type().String(),
				Span:     numberingArg.Span,
			}
		}
		elem.Numbering = &numbering
	}

	// Get optional number-align argument
	if alignArg := args.Find("number-align"); alignArg != nil && !IsAuto(alignArg.V) {
		align, err := parseAlignment(alignArg)
		if err != nil {
			return nil, err
		}
		elem.NumberAlign = align
	}

	// Get optional header argument
	if headerArg := args.Find("header"); headerArg != nil && !IsNone(headerArg.V) {
		if c, ok := headerArg.V.(ContentValue); ok {
			elem.Header = &c.Content
		} else {
			return nil, &TypeMismatchError{
				Expected: "content or none",
				Got:      headerArg.V.Type().String(),
				Span:     headerArg.Span,
			}
		}
	}

	// Get optional header-ascent argument
	if ascentArg := args.Find("header-ascent"); ascentArg != nil && !IsNone(ascentArg.V) {
		rel, err := parseRelative(ascentArg)
		if err != nil {
			return nil, err
		}
		elem.HeaderAscent = rel
	}

	// Get optional footer argument
	if footerArg := args.Find("footer"); footerArg != nil && !IsNone(footerArg.V) {
		if c, ok := footerArg.V.(ContentValue); ok {
			elem.Footer = &c.Content
		} else {
			return nil, &TypeMismatchError{
				Expected: "content or none",
				Got:      footerArg.V.Type().String(),
				Span:     footerArg.Span,
			}
		}
	}

	// Get optional footer-descent argument
	if descentArg := args.Find("footer-descent"); descentArg != nil && !IsNone(descentArg.V) {
		rel, err := parseRelative(descentArg)
		if err != nil {
			return nil, err
		}
		elem.FooterDescent = rel
	}

	// Get optional background argument
	if bgArg := args.Find("background"); bgArg != nil && !IsNone(bgArg.V) {
		if c, ok := bgArg.V.(ContentValue); ok {
			elem.Background = &c.Content
		} else {
			return nil, &TypeMismatchError{
				Expected: "content or none",
				Got:      bgArg.V.Type().String(),
				Span:     bgArg.Span,
			}
		}
	}

	// Get optional foreground argument
	if fgArg := args.Find("foreground"); fgArg != nil && !IsNone(fgArg.V) {
		if c, ok := fgArg.V.(ContentValue); ok {
			elem.Foreground = &c.Content
		} else {
			return nil, &TypeMismatchError{
				Expected: "content or none",
				Got:      fgArg.V.Type().String(),
				Span:     fgArg.Span,
			}
		}
	}

	// Get body content (positional argument)
	bodyArg := args.Find("body")
	if bodyArg == nil {
		bodyArgSpanned, err := args.Expect("body")
		if err != nil {
			// Body is optional for set rules
			elem.Body = Content{}
		} else {
			bodyArg = &bodyArgSpanned
		}
	}
	if bodyArg != nil {
		if c, ok := bodyArg.V.(ContentValue); ok {
			elem.Body = c.Content
		} else if !IsNone(bodyArg.V) {
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

	// Create the PageElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}

// parseRelative parses a relative length value from an argument.
func parseRelative(arg *syntax.Spanned[Value]) (*Relative, error) {
	switch v := arg.V.(type) {
	case LengthValue:
		return &Relative{Abs: v.Length}, nil
	case RatioValue:
		return &Relative{Rel: v.Ratio}, nil
	case RelativeValue:
		return &v.Relative, nil
	default:
		return nil, &TypeMismatchError{
			Expected: "length or ratio",
			Got:      arg.V.Type().String(),
			Span:     arg.Span,
		}
	}
}

// parsePageMargin parses margin values from an argument.
func parsePageMargin(arg *syntax.Spanned[Value]) (*PageMargin, error) {
	margin := &PageMargin{}

	switch v := arg.V.(type) {
	case LengthValue:
		rel := &Relative{Abs: v.Length}
		margin.Uniform = rel
		return margin, nil
	case RatioValue:
		rel := &Relative{Rel: v.Ratio}
		margin.Uniform = rel
		return margin, nil
	case RelativeValue:
		rel := v.Relative
		margin.Uniform = &rel
		return margin, nil
	case DictValue:
		// Parse dictionary margins
		for _, key := range v.Keys() {
			val, _ := v.Get(key)
			relArg := &syntax.Spanned[Value]{V: val, Span: arg.Span}
			rel, err := parseRelative(relArg)
			if err != nil {
				return nil, err
			}
			switch key {
			case "left":
				margin.Left = rel
			case "right":
				margin.Right = rel
			case "top":
				margin.Top = rel
			case "bottom":
				margin.Bottom = rel
			case "inside":
				margin.Inside = rel
			case "outside":
				margin.Outside = rel
			case "x":
				margin.X = rel
			case "y":
				margin.Y = rel
			case "rest":
				margin.Rest = rel
			}
		}
		return margin, nil
	default:
		return nil, &TypeMismatchError{
			Expected: "length, ratio, or dictionary",
			Got:      arg.V.Type().String(),
			Span:     arg.Span,
		}
	}
}

// parsePageBinding parses a binding value from an argument.
func parsePageBinding(arg *syntax.Spanned[Value]) (*PageBinding, error) {
	if s, ok := AsStr(arg.V); ok {
		switch s {
		case "left":
			binding := PageBindingLeft
			return &binding, nil
		case "right":
			binding := PageBindingRight
			return &binding, nil
		default:
			return nil, &TypeMismatchError{
				Expected: "\"left\" or \"right\"",
				Got:      "\"" + s + "\"",
				Span:     arg.Span,
			}
		}
	}
	return nil, &TypeMismatchError{
		Expected: "alignment or string",
		Got:      arg.V.Type().String(),
		Span:     arg.Span,
	}
}

// parseAlignment parses an alignment value from an argument.
func parseAlignment(arg *syntax.Spanned[Value]) (*Alignment, error) {
	align := &Alignment{}
	if s, ok := AsStr(arg.V); ok {
		switch s {
		case "start":
			h := HAlignStart
			align.Horizontal = &h
		case "center":
			h := HAlignCenter
			align.Horizontal = &h
		case "end":
			h := HAlignEnd
			align.Horizontal = &h
		case "left":
			h := HAlignLeft
			align.Horizontal = &h
		case "right":
			h := HAlignRight
			align.Horizontal = &h
		case "top":
			v := VAlignTop
			align.Vertical = &v
		case "horizon":
			v := VAlignHorizon
			align.Vertical = &v
		case "bottom":
			v := VAlignBottom
			align.Vertical = &v
		default:
			return nil, &TypeMismatchError{
				Expected: "alignment",
				Got:      "\"" + s + "\"",
				Span:     arg.Span,
			}
		}
		return align, nil
	}
	// TODO: Handle alignment type directly
	return nil, &TypeMismatchError{
		Expected: "alignment",
		Got:      arg.V.Type().String(),
		Span:     arg.Span,
	}
}

// ----------------------------------------------------------------------------
// Paper Sizes
// ----------------------------------------------------------------------------

// PaperSize represents a standard paper size.
type PaperSize struct {
	Name   string
	Width  float64 // in points
	Height float64 // in points
}

// PaperSizes contains standard paper size definitions.
var PaperSizes = map[string]PaperSize{
	// ISO A series
	"a0":  {Name: "a0", Width: 2383.94, Height: 3370.39},
	"a1":  {Name: "a1", Width: 1683.78, Height: 2383.94},
	"a2":  {Name: "a2", Width: 1190.55, Height: 1683.78},
	"a3":  {Name: "a3", Width: 841.89, Height: 1190.55},
	"a4":  {Name: "a4", Width: 595.28, Height: 841.89},
	"a5":  {Name: "a5", Width: 419.53, Height: 595.28},
	"a6":  {Name: "a6", Width: 297.64, Height: 419.53},
	"a7":  {Name: "a7", Width: 209.76, Height: 297.64},
	"a8":  {Name: "a8", Width: 147.40, Height: 209.76},
	"a9":  {Name: "a9", Width: 104.88, Height: 147.40},
	"a10": {Name: "a10", Width: 73.70, Height: 104.88},
	"a11": {Name: "a11", Width: 51.02, Height: 73.70},
	// ISO B series
	"iso-b1": {Name: "iso-b1", Width: 2004.09, Height: 2834.65},
	"iso-b2": {Name: "iso-b2", Width: 1417.32, Height: 2004.09},
	"iso-b3": {Name: "iso-b3", Width: 1000.63, Height: 1417.32},
	"iso-b4": {Name: "iso-b4", Width: 708.66, Height: 1000.63},
	"iso-b5": {Name: "iso-b5", Width: 498.90, Height: 708.66},
	"iso-b6": {Name: "iso-b6", Width: 354.33, Height: 498.90},
	"iso-b7": {Name: "iso-b7", Width: 249.45, Height: 354.33},
	"iso-b8": {Name: "iso-b8", Width: 175.75, Height: 249.45},
	// ISO C series (envelopes)
	"iso-c3": {Name: "iso-c3", Width: 918.43, Height: 1298.27},
	"iso-c4": {Name: "iso-c4", Width: 649.13, Height: 918.43},
	"iso-c5": {Name: "iso-c5", Width: 459.21, Height: 649.13},
	"iso-c6": {Name: "iso-c6", Width: 323.15, Height: 459.21},
	"iso-c7": {Name: "iso-c7", Width: 229.61, Height: 323.15},
	"iso-c8": {Name: "iso-c8", Width: 161.57, Height: 229.61},
	// US paper sizes
	"us-letter":  {Name: "us-letter", Width: 612.0, Height: 792.0},
	"us-legal":   {Name: "us-legal", Width: 612.0, Height: 1008.0},
	"us-tabloid": {Name: "us-tabloid", Width: 792.0, Height: 1224.0},
	"us-ledger":  {Name: "us-ledger", Width: 1224.0, Height: 792.0},
	// US ANSI sizes
	"us-executive":   {Name: "us-executive", Width: 522.0, Height: 756.0},
	"us-foolscap":    {Name: "us-foolscap", Width: 612.0, Height: 936.0},
	"us-statement":   {Name: "us-statement", Width: 396.0, Height: 612.0},
	"us-quarto":      {Name: "us-quarto", Width: 609.45, Height: 779.53},
	"us-government":  {Name: "us-government", Width: 576.0, Height: 756.0},
	"us-business":    {Name: "us-business", Width: 252.0, Height: 612.0},
	"us-digest":      {Name: "us-digest", Width: 396.0, Height: 612.0},
	"us-trade":       {Name: "us-trade", Width: 432.0, Height: 648.0},
	// Japanese JIS sizes
	"jis-b0": {Name: "jis-b0", Width: 2919.69, Height: 4127.24},
	"jis-b1": {Name: "jis-b1", Width: 2063.62, Height: 2919.69},
	"jis-b2": {Name: "jis-b2", Width: 1459.84, Height: 2063.62},
	"jis-b3": {Name: "jis-b3", Width: 1031.81, Height: 1459.84},
	"jis-b4": {Name: "jis-b4", Width: 728.50, Height: 1031.81},
	"jis-b5": {Name: "jis-b5", Width: 515.91, Height: 728.50},
	"jis-b6": {Name: "jis-b6", Width: 362.83, Height: 515.91},
	"jis-b7": {Name: "jis-b7", Width: 257.95, Height: 362.83},
	// DIN paper sizes
	"din-d3": {Name: "din-d3", Width: 779.53, Height: 1105.51},
	"din-d4": {Name: "din-d4", Width: 552.76, Height: 779.53},
	"din-d5": {Name: "din-d5", Width: 389.76, Height: 552.76},
	"din-d6": {Name: "din-d6", Width: 275.91, Height: 389.76},
	// Presentation
	"presentation-16-9":  {Name: "presentation-16-9", Width: 720.0, Height: 405.0},
	"presentation-4-3":   {Name: "presentation-4-3", Width: 720.0, Height: 540.0},
}

// GetPaperSize returns the paper size for the given name, or nil if not found.
func GetPaperSize(name string) *PaperSize {
	if size, ok := PaperSizes[name]; ok {
		return &size
	}
	return nil
}

// ----------------------------------------------------------------------------
// Library Registration
// ----------------------------------------------------------------------------

// RegisterElementFunctions registers all element functions in the given scope.
// Call this when setting up the standard library scope.
func RegisterElementFunctions(scope *Scope) {
	// Register raw element function
	scope.DefineFunc("raw", RawFunc())
	// Register page element function
	scope.DefineFunc("page", PageFunc())
}

// ElementFunctions returns a map of all element function names to their functions.
// This is useful for introspection and testing.
func ElementFunctions() map[string]*Func {
	return map[string]*Func{
		"raw":  RawFunc(),
		"page": PageFunc(),
	}
}
