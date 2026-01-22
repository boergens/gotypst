package layout

import (
	"github.com/boergens/gotypst/library/foundations"
	"github.com/boergens/gotypst/syntax"
)

// BoxElement represents an inline box container element.
// It can size its content, apply fills/strokes, and clip overflow.
//
// Reference: typst-reference/crates/typst-library/src/layout/container.rs
type BoxElement struct {
	// Width of the box (in points). If nil, auto-sizes to content.
	Width *float64
	// Height of the box (in points). If nil, auto-sizes to content.
	Height *float64
	// Baseline position (in points from bottom). If nil, uses content baseline.
	Baseline *float64
	// Fill color for the background. If nil, no fill.
	Fill foundations.Value
	// Stroke for the border. Can be length, color, or stroke dict. If nil, no stroke.
	Stroke foundations.Value
	// Radius for rounded corners. Can be single value or dictionary.
	Radius foundations.Value
	// Inset padding inside the box.
	Inset foundations.Value
	// Outset expansion outside the box.
	Outset foundations.Value
	// Whether to clip content that overflows the box.
	Clip bool
	// Body is the content inside the box.
	Body foundations.Content
}

func (*BoxElement) IsContentElement() {}

// BlockElement represents a block-level container element.
// It creates a new block in the document flow with optional sizing and styling.
//
// Reference: typst-reference/crates/typst-library/src/layout/container.rs
type BlockElement struct {
	// Width of the block (in points). If nil, auto-sizes.
	Width *float64
	// Height of the block (in points). If nil, auto-sizes.
	Height *float64
	// Whether the block can break across pages.
	Breakable *bool
	// Fill color for the background.
	Fill foundations.Value
	// Stroke for the border.
	Stroke foundations.Value
	// Radius for rounded corners.
	Radius foundations.Value
	// Inset padding inside the block.
	Inset foundations.Value
	// Outset expansion outside the block.
	Outset foundations.Value
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
	Body foundations.Content
}

func (*BlockElement) IsContentElement() {}

// BoxFunc creates the box element function.
func BoxFunc() *foundations.Func {
	name := "box"
	return &foundations.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: foundations.NativeFunc{
			Func: boxNative,
			Info: &foundations.FuncInfo{
				Name: "box",
				Params: []foundations.ParamInfo{
					{Name: "body", Type: foundations.TypeContent, Default: foundations.None, Named: false},
					{Name: "width", Type: foundations.TypeLength, Default: foundations.Auto, Named: true},
					{Name: "height", Type: foundations.TypeLength, Default: foundations.Auto, Named: true},
					{Name: "baseline", Type: foundations.TypeLength, Default: foundations.None, Named: true},
					{Name: "fill", Type: foundations.TypeColor, Default: foundations.None, Named: true},
					{Name: "stroke", Type: foundations.TypeDyn, Default: foundations.None, Named: true},
					{Name: "radius", Type: foundations.TypeDyn, Default: foundations.None, Named: true},
					{Name: "inset", Type: foundations.TypeDyn, Default: foundations.None, Named: true},
					{Name: "outset", Type: foundations.TypeDyn, Default: foundations.None, Named: true},
					{Name: "clip", Type: foundations.TypeBool, Default: foundations.False, Named: true},
				},
			},
		},
	}
}

// boxNative implements the box() function.
func boxNative(engine foundations.Engine, context foundations.Context, args *foundations.Args) (foundations.Value, error) {
	elem := &BoxElement{}

	// Get optional body argument (positional or named)
	bodyArg := args.Find("body")
	if bodyArg == nil {
		bodyArg2 := args.Eat()
		if bodyArg2 != nil {
			bodyArg = bodyArg2
		}
	}
	if bodyArg != nil && !foundations.IsNone(bodyArg.V) {
		if cv, ok := bodyArg.V.(foundations.ContentValue); ok {
			elem.Body = cv.Content
		} else {
			return nil, &foundations.TypeMismatchError{
				Expected: "content or none",
				Got:      bodyArg.V.Type().String(),
				Span:     bodyArg.Span,
			}
		}
	}

	// Get optional width argument
	if widthArg := args.Find("width"); widthArg != nil {
		if !foundations.IsAuto(widthArg.V) && !foundations.IsNone(widthArg.V) {
			if lv, ok := widthArg.V.(foundations.LengthValue); ok {
				w := lv.Length.Points
				elem.Width = &w
			} else if rv, ok := widthArg.V.(foundations.RelativeValue); ok {
				w := rv.Relative.Rel.Value * 100
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
		if !foundations.IsAuto(heightArg.V) && !foundations.IsNone(heightArg.V) {
			if lv, ok := heightArg.V.(foundations.LengthValue); ok {
				h := lv.Length.Points
				elem.Height = &h
			} else if rv, ok := heightArg.V.(foundations.RelativeValue); ok {
				h := rv.Relative.Rel.Value * 100
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

	// Get optional baseline argument
	if baselineArg := args.Find("baseline"); baselineArg != nil {
		if !foundations.IsAuto(baselineArg.V) && !foundations.IsNone(baselineArg.V) {
			if lv, ok := baselineArg.V.(foundations.LengthValue); ok {
				b := lv.Length.Points
				elem.Baseline = &b
			} else {
				return nil, &foundations.TypeMismatchError{
					Expected: "length or none",
					Got:      baselineArg.V.Type().String(),
					Span:     baselineArg.Span,
				}
			}
		}
	}

	// Get optional fill argument
	if fillArg := args.Find("fill"); fillArg != nil {
		if !foundations.IsNone(fillArg.V) {
			elem.Fill = fillArg.V
		}
	}

	// Get optional stroke argument
	if strokeArg := args.Find("stroke"); strokeArg != nil {
		if !foundations.IsNone(strokeArg.V) {
			elem.Stroke = strokeArg.V
		}
	}

	// Get optional radius argument
	if radiusArg := args.Find("radius"); radiusArg != nil {
		if !foundations.IsNone(radiusArg.V) {
			elem.Radius = radiusArg.V
		}
	}

	// Get optional inset argument
	if insetArg := args.Find("inset"); insetArg != nil {
		if !foundations.IsNone(insetArg.V) {
			elem.Inset = insetArg.V
		}
	}

	// Get optional outset argument
	if outsetArg := args.Find("outset"); outsetArg != nil {
		if !foundations.IsNone(outsetArg.V) {
			elem.Outset = outsetArg.V
		}
	}

	// Get optional clip argument (default: false)
	if clipArg := args.Find("clip"); clipArg != nil {
		if !foundations.IsNone(clipArg.V) && !foundations.IsAuto(clipArg.V) {
			clipVal, ok := foundations.AsBool(clipArg.V)
			if !ok {
				return nil, &foundations.TypeMismatchError{
					Expected: "bool",
					Got:      clipArg.V.Type().String(),
					Span:     clipArg.Span,
				}
			}
			elem.Clip = clipVal
		}
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	return foundations.ContentValue{Content: foundations.Content{
		Elements: []foundations.ContentElement{elem},
	}}, nil
}

// BlockFunc creates the block element function.
func BlockFunc() *foundations.Func {
	name := "block"
	return &foundations.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: foundations.NativeFunc{
			Func: blockNative,
			Info: &foundations.FuncInfo{
				Name: "block",
				Params: []foundations.ParamInfo{
					{Name: "body", Type: foundations.TypeContent, Default: foundations.None, Named: false},
					{Name: "width", Type: foundations.TypeLength, Default: foundations.Auto, Named: true},
					{Name: "height", Type: foundations.TypeLength, Default: foundations.Auto, Named: true},
					{Name: "breakable", Type: foundations.TypeBool, Default: foundations.True, Named: true},
					{Name: "fill", Type: foundations.TypeColor, Default: foundations.None, Named: true},
					{Name: "stroke", Type: foundations.TypeDyn, Default: foundations.None, Named: true},
					{Name: "radius", Type: foundations.TypeDyn, Default: foundations.None, Named: true},
					{Name: "inset", Type: foundations.TypeDyn, Default: foundations.None, Named: true},
					{Name: "outset", Type: foundations.TypeDyn, Default: foundations.None, Named: true},
					{Name: "spacing", Type: foundations.TypeLength, Default: foundations.None, Named: true},
					{Name: "above", Type: foundations.TypeLength, Default: foundations.Auto, Named: true},
					{Name: "below", Type: foundations.TypeLength, Default: foundations.Auto, Named: true},
					{Name: "clip", Type: foundations.TypeBool, Default: foundations.False, Named: true},
					{Name: "sticky", Type: foundations.TypeBool, Default: foundations.False, Named: true},
				},
			},
		},
	}
}

// blockNative implements the block() function.
func blockNative(engine foundations.Engine, context foundations.Context, args *foundations.Args) (foundations.Value, error) {
	elem := &BlockElement{}

	// Get optional body argument (positional or named)
	bodyArg := args.Find("body")
	if bodyArg == nil {
		bodyArg2 := args.Eat()
		if bodyArg2 != nil {
			bodyArg = bodyArg2
		}
	}
	if bodyArg != nil && !foundations.IsNone(bodyArg.V) {
		if cv, ok := bodyArg.V.(foundations.ContentValue); ok {
			elem.Body = cv.Content
		} else {
			return nil, &foundations.TypeMismatchError{
				Expected: "content or none",
				Got:      bodyArg.V.Type().String(),
				Span:     bodyArg.Span,
			}
		}
	}

	// Get optional width argument
	if widthArg := args.Find("width"); widthArg != nil {
		if !foundations.IsAuto(widthArg.V) && !foundations.IsNone(widthArg.V) {
			if lv, ok := widthArg.V.(foundations.LengthValue); ok {
				w := lv.Length.Points
				elem.Width = &w
			} else if rv, ok := widthArg.V.(foundations.RelativeValue); ok {
				w := rv.Relative.Rel.Value * 100
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
		if !foundations.IsAuto(heightArg.V) && !foundations.IsNone(heightArg.V) {
			if lv, ok := heightArg.V.(foundations.LengthValue); ok {
				h := lv.Length.Points
				elem.Height = &h
			} else if rv, ok := heightArg.V.(foundations.RelativeValue); ok {
				h := rv.Relative.Rel.Value * 100
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

	// Get optional breakable argument (default: true)
	if breakableArg := args.Find("breakable"); breakableArg != nil {
		if !foundations.IsNone(breakableArg.V) && !foundations.IsAuto(breakableArg.V) {
			breakVal, ok := foundations.AsBool(breakableArg.V)
			if !ok {
				return nil, &foundations.TypeMismatchError{
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
		if !foundations.IsNone(fillArg.V) {
			elem.Fill = fillArg.V
		}
	}

	// Get optional stroke argument
	if strokeArg := args.Find("stroke"); strokeArg != nil {
		if !foundations.IsNone(strokeArg.V) {
			elem.Stroke = strokeArg.V
		}
	}

	// Get optional radius argument
	if radiusArg := args.Find("radius"); radiusArg != nil {
		if !foundations.IsNone(radiusArg.V) {
			elem.Radius = radiusArg.V
		}
	}

	// Get optional inset argument
	if insetArg := args.Find("inset"); insetArg != nil {
		if !foundations.IsNone(insetArg.V) {
			elem.Inset = insetArg.V
		}
	}

	// Get optional outset argument
	if outsetArg := args.Find("outset"); outsetArg != nil {
		if !foundations.IsNone(outsetArg.V) {
			elem.Outset = outsetArg.V
		}
	}

	// Get optional spacing argument
	if spacingArg := args.Find("spacing"); spacingArg != nil {
		if !foundations.IsAuto(spacingArg.V) && !foundations.IsNone(spacingArg.V) {
			if lv, ok := spacingArg.V.(foundations.LengthValue); ok {
				s := lv.Length.Points
				elem.Spacing = &s
			} else {
				return nil, &foundations.TypeMismatchError{
					Expected: "length or auto",
					Got:      spacingArg.V.Type().String(),
					Span:     spacingArg.Span,
				}
			}
		}
	}

	// Get optional above argument
	if aboveArg := args.Find("above"); aboveArg != nil {
		if !foundations.IsAuto(aboveArg.V) && !foundations.IsNone(aboveArg.V) {
			if lv, ok := aboveArg.V.(foundations.LengthValue); ok {
				a := lv.Length.Points
				elem.Above = &a
			} else {
				return nil, &foundations.TypeMismatchError{
					Expected: "length or auto",
					Got:      aboveArg.V.Type().String(),
					Span:     aboveArg.Span,
				}
			}
		}
	}

	// Get optional below argument
	if belowArg := args.Find("below"); belowArg != nil {
		if !foundations.IsAuto(belowArg.V) && !foundations.IsNone(belowArg.V) {
			if lv, ok := belowArg.V.(foundations.LengthValue); ok {
				b := lv.Length.Points
				elem.Below = &b
			} else {
				return nil, &foundations.TypeMismatchError{
					Expected: "length or auto",
					Got:      belowArg.V.Type().String(),
					Span:     belowArg.Span,
				}
			}
		}
	}

	// Get optional clip argument (default: false)
	if clipArg := args.Find("clip"); clipArg != nil {
		if !foundations.IsNone(clipArg.V) && !foundations.IsAuto(clipArg.V) {
			clipVal, ok := foundations.AsBool(clipArg.V)
			if !ok {
				return nil, &foundations.TypeMismatchError{
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
		if !foundations.IsNone(stickyArg.V) && !foundations.IsAuto(stickyArg.V) {
			stickyVal, ok := foundations.AsBool(stickyArg.V)
			if !ok {
				return nil, &foundations.TypeMismatchError{
					Expected: "bool",
					Got:      stickyArg.V.Type().String(),
					Span:     stickyArg.Span,
				}
			}
			elem.Sticky = stickyVal
		}
	}

	if err := args.Finish(); err != nil {
		return nil, err
	}

	return foundations.ContentValue{Content: foundations.Content{
		Elements: []foundations.ContentElement{elem},
	}}, nil
}
