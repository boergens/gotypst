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

// BoxFunc creates the box element function.
func BoxFunc() *Func {
	name := "box"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: boxNative,
			Info: &FuncInfo{
				Name: "box",
				Params: []ParamInfo{
					{Name: "body", Type: TypeContent, Default: None, Named: false},
					{Name: "width", Type: TypeLength, Default: Auto, Named: true},
					{Name: "height", Type: TypeLength, Default: Auto, Named: true},
					{Name: "baseline", Type: TypeLength, Default: Auto, Named: true},
					{Name: "fill", Type: TypeColor, Default: None, Named: true},
					{Name: "inset", Type: TypeLength, Default: None, Named: true},
					{Name: "outset", Type: TypeLength, Default: None, Named: true},
					{Name: "radius", Type: TypeLength, Default: None, Named: true},
					{Name: "clip", Type: TypeBool, Default: True, Named: true},
				},
			},
		},
	}
}

// boxNative implements the box() function.
// Creates a BoxElement with optional sizing and styling properties.
//
// Arguments:
//   - body (positional, content, default: none): The box content
//   - width (named, length, default: auto): Width of the box
//   - height (named, length, default: auto): Height of the box
//   - baseline (named, length, default: auto): Baseline offset
//   - fill (named, color, default: none): Background fill
//   - inset (named, length, default: none): Inner padding
//   - outset (named, length, default: none): Outer padding
//   - radius (named, length, default: none): Corner radius
//   - clip (named, bool, default: true): Whether to clip content
func boxNative(vm *Vm, args *Args) (Value, error) {
	// Create element with defaults
	elem := &BoxElement{}

	// Get optional body argument (positional)
	bodyArg := args.Find("body")
	if bodyArg == nil {
		// Try positional
		if bodyArgSpanned, err := args.Expect("body"); err == nil {
			bodyArg = &bodyArgSpanned
		}
	}
	if bodyArg != nil && !IsNone(bodyArg.V) {
		if cv, ok := bodyArg.V.(ContentValue); ok {
			elem.Body = cv.Content
		} else {
			return nil, &TypeMismatchError{
				Expected: "content or none",
				Got:      bodyArg.V.Type().String(),
				Span:     bodyArg.Span,
			}
		}
	}

	// Get optional width argument
	if widthArg := args.Find("width"); widthArg != nil {
		if !IsAuto(widthArg.V) && !IsNone(widthArg.V) {
			if lv, ok := widthArg.V.(LengthValue); ok {
				width := lv.Length.Points
				elem.Width = &width
			} else if rv, ok := widthArg.V.(RatioValue); ok {
				// Convert ratio to a relative value (percentage of container)
				width := rv.Ratio.Value * 100 // Store as percentage
				elem.Width = &width
			} else {
				return nil, &TypeMismatchError{
					Expected: "length, ratio, or auto",
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
				height := lv.Length.Points
				elem.Height = &height
			} else if rv, ok := heightArg.V.(RatioValue); ok {
				height := rv.Ratio.Value * 100
				elem.Height = &height
			} else if fv, ok := heightArg.V.(FractionValue); ok {
				height := fv.Fraction.Value
				elem.Height = &height
			} else {
				return nil, &TypeMismatchError{
					Expected: "length, ratio, fraction, or auto",
					Got:      heightArg.V.Type().String(),
					Span:     heightArg.Span,
				}
			}
		}
	}

	// Get optional baseline argument
	if baselineArg := args.Find("baseline"); baselineArg != nil {
		if !IsAuto(baselineArg.V) && !IsNone(baselineArg.V) {
			if lv, ok := baselineArg.V.(LengthValue); ok {
				baseline := lv.Length.Points
				elem.Baseline = &baseline
			} else if rv, ok := baselineArg.V.(RatioValue); ok {
				baseline := rv.Ratio.Value * 100
				elem.Baseline = &baseline
			} else {
				return nil, &TypeMismatchError{
					Expected: "length, ratio, or auto",
					Got:      baselineArg.V.Type().String(),
					Span:     baselineArg.Span,
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

	// Get optional inset argument
	if insetArg := args.Find("inset"); insetArg != nil {
		if !IsNone(insetArg.V) {
			if lv, ok := insetArg.V.(LengthValue); ok {
				inset := lv.Length.Points
				elem.Inset = &inset
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      insetArg.V.Type().String(),
					Span:     insetArg.Span,
				}
			}
		}
	}

	// Get optional outset argument
	if outsetArg := args.Find("outset"); outsetArg != nil {
		if !IsNone(outsetArg.V) {
			if lv, ok := outsetArg.V.(LengthValue); ok {
				outset := lv.Length.Points
				elem.Outset = &outset
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      outsetArg.V.Type().String(),
					Span:     outsetArg.Span,
				}
			}
		}
	}

	// Get optional radius argument
	if radiusArg := args.Find("radius"); radiusArg != nil {
		if !IsNone(radiusArg.V) {
			if lv, ok := radiusArg.V.(LengthValue); ok {
				radius := lv.Length.Points
				elem.Radius = &radius
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      radiusArg.V.Type().String(),
					Span:     radiusArg.Span,
				}
			}
		}
	}

	// Get optional clip argument
	if clipArg := args.Find("clip"); clipArg != nil {
		if !IsAuto(clipArg.V) && !IsNone(clipArg.V) {
			if cv, ok := AsBool(clipArg.V); ok {
				elem.Clip = &cv
			} else {
				return nil, &TypeMismatchError{
					Expected: "bool",
					Got:      clipArg.V.Type().String(),
					Span:     clipArg.Span,
				}
			}
		}
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Create the BoxElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}

// BlockFunc creates the block element function.
func BlockFunc() *Func {
	name := "block"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: blockNative,
			Info: &FuncInfo{
				Name: "block",
				Params: []ParamInfo{
					{Name: "body", Type: TypeContent, Default: None, Named: false},
					{Name: "width", Type: TypeLength, Default: Auto, Named: true},
					{Name: "height", Type: TypeLength, Default: Auto, Named: true},
					{Name: "fill", Type: TypeColor, Default: None, Named: true},
					{Name: "inset", Type: TypeLength, Default: None, Named: true},
					{Name: "outset", Type: TypeLength, Default: None, Named: true},
					{Name: "radius", Type: TypeLength, Default: None, Named: true},
					{Name: "clip", Type: TypeBool, Default: True, Named: true},
					{Name: "breakable", Type: TypeBool, Default: True, Named: true},
					{Name: "above", Type: TypeLength, Default: Auto, Named: true},
					{Name: "below", Type: TypeLength, Default: Auto, Named: true},
				},
			},
		},
	}
}

// blockNative implements the block() function.
// Creates a BlockElement with optional sizing, spacing, and styling properties.
//
// Arguments:
//   - body (positional, content, default: none): The block content
//   - width (named, length, default: auto): Width of the block
//   - height (named, length, default: auto): Height of the block
//   - fill (named, color, default: none): Background fill
//   - inset (named, length, default: none): Inner padding
//   - outset (named, length, default: none): Outer padding
//   - radius (named, length, default: none): Corner radius
//   - clip (named, bool, default: true): Whether to clip content
//   - breakable (named, bool, default: true): Whether block can break across pages
//   - above (named, length, default: auto): Spacing above block
//   - below (named, length, default: auto): Spacing below block
func blockNative(vm *Vm, args *Args) (Value, error) {
	// Create element with defaults
	elem := &BlockElement{}

	// Get optional body argument (positional)
	bodyArg := args.Find("body")
	if bodyArg == nil {
		// Try positional
		if bodyArgSpanned, err := args.Expect("body"); err == nil {
			bodyArg = &bodyArgSpanned
		}
	}
	if bodyArg != nil && !IsNone(bodyArg.V) {
		if cv, ok := bodyArg.V.(ContentValue); ok {
			elem.Body = cv.Content
		} else {
			return nil, &TypeMismatchError{
				Expected: "content or none",
				Got:      bodyArg.V.Type().String(),
				Span:     bodyArg.Span,
			}
		}
	}

	// Get optional width argument
	if widthArg := args.Find("width"); widthArg != nil {
		if !IsAuto(widthArg.V) && !IsNone(widthArg.V) {
			if lv, ok := widthArg.V.(LengthValue); ok {
				width := lv.Length.Points
				elem.Width = &width
			} else if rv, ok := widthArg.V.(RatioValue); ok {
				width := rv.Ratio.Value * 100
				elem.Width = &width
			} else {
				return nil, &TypeMismatchError{
					Expected: "length, ratio, or auto",
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
				height := lv.Length.Points
				elem.Height = &height
			} else if rv, ok := heightArg.V.(RatioValue); ok {
				height := rv.Ratio.Value * 100
				elem.Height = &height
			} else if fv, ok := heightArg.V.(FractionValue); ok {
				height := fv.Fraction.Value
				elem.Height = &height
			} else {
				return nil, &TypeMismatchError{
					Expected: "length, ratio, fraction, or auto",
					Got:      heightArg.V.Type().String(),
					Span:     heightArg.Span,
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

	// Get optional inset argument
	if insetArg := args.Find("inset"); insetArg != nil {
		if !IsNone(insetArg.V) {
			if lv, ok := insetArg.V.(LengthValue); ok {
				inset := lv.Length.Points
				elem.Inset = &inset
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      insetArg.V.Type().String(),
					Span:     insetArg.Span,
				}
			}
		}
	}

	// Get optional outset argument
	if outsetArg := args.Find("outset"); outsetArg != nil {
		if !IsNone(outsetArg.V) {
			if lv, ok := outsetArg.V.(LengthValue); ok {
				outset := lv.Length.Points
				elem.Outset = &outset
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      outsetArg.V.Type().String(),
					Span:     outsetArg.Span,
				}
			}
		}
	}

	// Get optional radius argument
	if radiusArg := args.Find("radius"); radiusArg != nil {
		if !IsNone(radiusArg.V) {
			if lv, ok := radiusArg.V.(LengthValue); ok {
				radius := lv.Length.Points
				elem.Radius = &radius
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      radiusArg.V.Type().String(),
					Span:     radiusArg.Span,
				}
			}
		}
	}

	// Get optional clip argument
	if clipArg := args.Find("clip"); clipArg != nil {
		if !IsAuto(clipArg.V) && !IsNone(clipArg.V) {
			if cv, ok := AsBool(clipArg.V); ok {
				elem.Clip = &cv
			} else {
				return nil, &TypeMismatchError{
					Expected: "bool",
					Got:      clipArg.V.Type().String(),
					Span:     clipArg.Span,
				}
			}
		}
	}

	// Get optional breakable argument
	if breakableArg := args.Find("breakable"); breakableArg != nil {
		if !IsAuto(breakableArg.V) && !IsNone(breakableArg.V) {
			if bv, ok := AsBool(breakableArg.V); ok {
				elem.Breakable = &bv
			} else {
				return nil, &TypeMismatchError{
					Expected: "bool",
					Got:      breakableArg.V.Type().String(),
					Span:     breakableArg.Span,
				}
			}
		}
	}

	// Get optional above argument
	if aboveArg := args.Find("above"); aboveArg != nil {
		if !IsAuto(aboveArg.V) && !IsNone(aboveArg.V) {
			if lv, ok := aboveArg.V.(LengthValue); ok {
				above := lv.Length.Points
				elem.Above = &above
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      aboveArg.V.Type().String(),
					Span:     aboveArg.Span,
				}
			}
		}
	}

	// Get optional below argument
	if belowArg := args.Find("below"); belowArg != nil {
		if !IsAuto(belowArg.V) && !IsNone(belowArg.V) {
			if lv, ok := belowArg.V.(LengthValue); ok {
				below := lv.Length.Points
				elem.Below = &below
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      belowArg.V.Type().String(),
					Span:     belowArg.Span,
				}
			}
		}
	}

	// Check for unexpected arguments
	if err := args.Finish(); err != nil {
		return nil, err
	}

	// Create the BlockElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
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
	// Register box element function
	scope.DefineFunc("box", BoxFunc())
	// Register block element function
	scope.DefineFunc("block", BlockFunc())
}

// ElementFunctions returns a map of all element function names to their functions.
// This is useful for introspection and testing.
func ElementFunctions() map[string]*Func {
	return map[string]*Func{
		"raw":      RawFunc(),
		"par":      ParFunc(),
		"parbreak": ParbreakFunc(),
		"box":      BoxFunc(),
		"block":    BlockFunc(),
	}
}
