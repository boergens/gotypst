package eval

import (
	"github.com/boergens/gotypst/syntax"
)

// ----------------------------------------------------------------------------
// Box Element
// ----------------------------------------------------------------------------
// Reference: typst-reference/crates/typst-library/src/layout/container.rs

// BoxElement represents an inline box container element.
// It can size its content, apply fills/strokes, and clip overflow.
type BoxElement struct {
	// Width of the box (in points). If nil, auto-sizes to content.
	Width *float64
	// Height of the box (in points). If nil, auto-sizes to content.
	Height *float64
	// Baseline position (in points from bottom). If nil, uses content baseline.
	Baseline *float64
	// Fill color for the background. If nil, no fill.
	Fill Value
	// Stroke for the border. Can be length, color, or stroke dict. If nil, no stroke.
	Stroke Value
	// Radius for rounded corners. Can be single value or dictionary.
	Radius Value
	// Inset padding inside the box.
	Inset Value
	// Outset expansion outside the box.
	Outset Value
	// Whether to clip content that overflows the box.
	Clip bool
	// Body is the content inside the box.
	Body Content
}

func (*BoxElement) IsContentElement() {}

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
					{Name: "baseline", Type: TypeLength, Default: None, Named: true},
					{Name: "fill", Type: TypeColor, Default: None, Named: true},
					{Name: "stroke", Type: TypeDyn, Default: None, Named: true},
					{Name: "radius", Type: TypeDyn, Default: None, Named: true},
					{Name: "inset", Type: TypeDyn, Default: None, Named: true},
					{Name: "outset", Type: TypeDyn, Default: None, Named: true},
					{Name: "clip", Type: TypeBool, Default: False, Named: true},
				},
			},
		},
	}
}

// boxNative implements the box() function.
// Creates a BoxElement with optional sizing and styling.
//
// Arguments:
//   - body (positional, content, default: none): The content inside the box
//   - width (named, length, default: auto): Box width
//   - height (named, length, default: auto): Box height
//   - baseline (named, length, default: none): Baseline position from bottom
//   - fill (named, color, default: none): Background fill
//   - stroke (named, various, default: none): Border stroke
//   - radius (named, various, default: none): Corner radius
//   - inset (named, various, default: none): Inner padding
//   - outset (named, various, default: none): Outer expansion
//   - clip (named, bool, default: false): Whether to clip overflow
func boxNative(vm *Vm, args *Args) (Value, error) {
	elem := &BoxElement{}

	// Get optional body argument (positional or named)
	bodyArg := args.Find("body")
	if bodyArg == nil {
		bodyArg2 := args.Eat()
		if bodyArg2 != nil {
			bodyArg = bodyArg2
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
				w := lv.Length.Points
				elem.Width = &w
			} else if rv, ok := widthArg.V.(RelativeValue); ok {
				// Handle relative values (percentages)
				w := rv.Relative.Rel.Value * 100 // Store as percentage for now
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
			} else if rv, ok := heightArg.V.(RelativeValue); ok {
				h := rv.Relative.Rel.Value * 100
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

	// Get optional baseline argument
	if baselineArg := args.Find("baseline"); baselineArg != nil {
		if !IsAuto(baselineArg.V) && !IsNone(baselineArg.V) {
			if lv, ok := baselineArg.V.(LengthValue); ok {
				b := lv.Length.Points
				elem.Baseline = &b
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or none",
					Got:      baselineArg.V.Type().String(),
					Span:     baselineArg.Span,
				}
			}
		}
	}

	// Get optional fill argument
	if fillArg := args.Find("fill"); fillArg != nil {
		if !IsNone(fillArg.V) {
			elem.Fill = fillArg.V
		}
	}

	// Get optional stroke argument
	if strokeArg := args.Find("stroke"); strokeArg != nil {
		if !IsNone(strokeArg.V) {
			elem.Stroke = strokeArg.V
		}
	}

	// Get optional radius argument
	if radiusArg := args.Find("radius"); radiusArg != nil {
		if !IsNone(radiusArg.V) {
			elem.Radius = radiusArg.V
		}
	}

	// Get optional inset argument
	if insetArg := args.Find("inset"); insetArg != nil {
		if !IsNone(insetArg.V) {
			elem.Inset = insetArg.V
		}
	}

	// Get optional outset argument
	if outsetArg := args.Find("outset"); outsetArg != nil {
		if !IsNone(outsetArg.V) {
			elem.Outset = outsetArg.V
		}
	}

	// Get optional clip argument (default: false)
	if clipArg := args.Find("clip"); clipArg != nil {
		if !IsNone(clipArg.V) && !IsAuto(clipArg.V) {
			clipVal, ok := AsBool(clipArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "bool",
					Got:      clipArg.V.Type().String(),
					Span:     clipArg.Span,
				}
			}
			elem.Clip = clipVal
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

// ----------------------------------------------------------------------------
// Block Element
// ----------------------------------------------------------------------------

// BlockElement represents a block-level container element.
// It creates a new block in the document flow with optional sizing and styling.
type BlockElement struct {
	// Width of the block (in points). If nil, auto-sizes.
	Width *float64
	// Height of the block (in points). If nil, auto-sizes.
	Height *float64
	// Whether the block can break across pages.
	Breakable *bool
	// Fill color for the background.
	Fill Value
	// Stroke for the border.
	Stroke Value
	// Radius for rounded corners.
	Radius Value
	// Inset padding inside the block.
	Inset Value
	// Outset expansion outside the block.
	Outset Value
	// Spacing between adjacent blocks.
	Spacing *float64
	// Spacing above this block (overrides Spacing).
	Above *float64
	// Spacing below this block (overrides Spacing).
	Below *float64
	// Whether to clip content that overflows.
	Clip bool
	// Whether the block sticks to the next block.
	Sticky bool
	// Body is the content inside the block.
	Body Content
}

func (*BlockElement) IsContentElement() {}

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
					{Name: "breakable", Type: TypeBool, Default: True, Named: true},
					{Name: "fill", Type: TypeColor, Default: None, Named: true},
					{Name: "stroke", Type: TypeDyn, Default: None, Named: true},
					{Name: "radius", Type: TypeDyn, Default: None, Named: true},
					{Name: "inset", Type: TypeDyn, Default: None, Named: true},
					{Name: "outset", Type: TypeDyn, Default: None, Named: true},
					{Name: "spacing", Type: TypeLength, Default: None, Named: true},
					{Name: "above", Type: TypeLength, Default: Auto, Named: true},
					{Name: "below", Type: TypeLength, Default: Auto, Named: true},
					{Name: "clip", Type: TypeBool, Default: False, Named: true},
					{Name: "sticky", Type: TypeBool, Default: False, Named: true},
				},
			},
		},
	}
}

// blockNative implements the block() function.
// Creates a BlockElement with optional sizing and styling.
//
// Arguments:
//   - body (positional, content, default: none): The content inside the block
//   - width (named, length, default: auto): Block width
//   - height (named, length, default: auto): Block height
//   - breakable (named, bool, default: true): Whether block can break across pages
//   - fill (named, color, default: none): Background fill
//   - stroke (named, various, default: none): Border stroke
//   - radius (named, various, default: none): Corner radius
//   - inset (named, various, default: none): Inner padding
//   - outset (named, various, default: none): Outer expansion
//   - spacing (named, length, default: 1.2em): Spacing between blocks
//   - above (named, length, default: auto): Spacing above this block
//   - below (named, length, default: auto): Spacing below this block
//   - clip (named, bool, default: false): Whether to clip overflow
//   - sticky (named, bool, default: false): Whether to stick to next block
func blockNative(vm *Vm, args *Args) (Value, error) {
	elem := &BlockElement{}

	// Get optional body argument (positional or named)
	bodyArg := args.Find("body")
	if bodyArg == nil {
		bodyArg2 := args.Eat()
		if bodyArg2 != nil {
			bodyArg = bodyArg2
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
				w := lv.Length.Points
				elem.Width = &w
			} else if rv, ok := widthArg.V.(RelativeValue); ok {
				w := rv.Relative.Rel.Value * 100
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
			} else if rv, ok := heightArg.V.(RelativeValue); ok {
				h := rv.Relative.Rel.Value * 100
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

	// Get optional breakable argument (default: true)
	if breakableArg := args.Find("breakable"); breakableArg != nil {
		if !IsNone(breakableArg.V) && !IsAuto(breakableArg.V) {
			breakVal, ok := AsBool(breakableArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "bool",
					Got:      breakableArg.V.Type().String(),
					Span:     breakableArg.Span,
				}
			}
			elem.Breakable = &breakVal
		}
	}

	// Get optional fill argument
	if fillArg := args.Find("fill"); fillArg != nil {
		if !IsNone(fillArg.V) {
			elem.Fill = fillArg.V
		}
	}

	// Get optional stroke argument
	if strokeArg := args.Find("stroke"); strokeArg != nil {
		if !IsNone(strokeArg.V) {
			elem.Stroke = strokeArg.V
		}
	}

	// Get optional radius argument
	if radiusArg := args.Find("radius"); radiusArg != nil {
		if !IsNone(radiusArg.V) {
			elem.Radius = radiusArg.V
		}
	}

	// Get optional inset argument
	if insetArg := args.Find("inset"); insetArg != nil {
		if !IsNone(insetArg.V) {
			elem.Inset = insetArg.V
		}
	}

	// Get optional outset argument
	if outsetArg := args.Find("outset"); outsetArg != nil {
		if !IsNone(outsetArg.V) {
			elem.Outset = outsetArg.V
		}
	}

	// Get optional spacing argument
	if spacingArg := args.Find("spacing"); spacingArg != nil {
		if !IsAuto(spacingArg.V) && !IsNone(spacingArg.V) {
			if lv, ok := spacingArg.V.(LengthValue); ok {
				s := lv.Length.Points
				elem.Spacing = &s
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      spacingArg.V.Type().String(),
					Span:     spacingArg.Span,
				}
			}
		}
	}

	// Get optional above argument
	if aboveArg := args.Find("above"); aboveArg != nil {
		if !IsAuto(aboveArg.V) && !IsNone(aboveArg.V) {
			if lv, ok := aboveArg.V.(LengthValue); ok {
				a := lv.Length.Points
				elem.Above = &a
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
				b := lv.Length.Points
				elem.Below = &b
			} else {
				return nil, &TypeMismatchError{
					Expected: "length or auto",
					Got:      belowArg.V.Type().String(),
					Span:     belowArg.Span,
				}
			}
		}
	}

	// Get optional clip argument (default: false)
	if clipArg := args.Find("clip"); clipArg != nil {
		if !IsNone(clipArg.V) && !IsAuto(clipArg.V) {
			clipVal, ok := AsBool(clipArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "bool",
					Got:      clipArg.V.Type().String(),
					Span:     clipArg.Span,
				}
			}
			elem.Clip = clipVal
		}
	}

	// Get optional sticky argument (default: false)
	if stickyArg := args.Find("sticky"); stickyArg != nil {
		if !IsNone(stickyArg.V) && !IsAuto(stickyArg.V) {
			stickyVal, ok := AsBool(stickyArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "bool",
					Got:      stickyArg.V.Type().String(),
					Span:     stickyArg.Span,
				}
			}
			elem.Sticky = stickyVal
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
