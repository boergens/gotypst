package eval

import (
	"github.com/boergens/gotypst/syntax"
)

// ----------------------------------------------------------------------------
// Layout Element Functions
// ----------------------------------------------------------------------------
// This file contains layout element functions that correspond to Typst's
// layout module:
// - stack() - stacking layout
// - align() - alignment container
// - columns() - multi-column layout
// - box() - inline box container
// - block() - block-level container
// - pad() - padding container
//
// Reference: typst-reference/crates/typst-library/src/layout/

// ----------------------------------------------------------------------------
// Stack Element
// ----------------------------------------------------------------------------

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

// ----------------------------------------------------------------------------
// Align Element
// ----------------------------------------------------------------------------

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
// Columns Element
// ----------------------------------------------------------------------------

// ColumnsElement represents a multi-column layout element.
// It arranges its body content into multiple columns.
type ColumnsElement struct {
	// Count is the number of columns.
	// If nil, defaults to 2.
	Count *int
	// Gutter is the gap between columns (in points).
	// If nil, defaults to 4% of page width.
	Gutter *float64
	// Body is the content to arrange in columns.
	Body Content
}

func (*ColumnsElement) IsContentElement() {}

// ColumnsFunc creates the columns element function.
func ColumnsFunc() *Func {
	name := "columns"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: columnsNative,
			Info: &FuncInfo{
				Name: "columns",
				Params: []ParamInfo{
					{Name: "count", Type: TypeInt, Default: Int(2), Named: false},
					{Name: "gutter", Type: TypeRelative, Default: None, Named: true},
					{Name: "body", Type: TypeContent, Named: false},
				},
			},
		},
	}
}

// columnsNative implements the columns() function.
// Creates a ColumnsElement with the given column count, gutter, and body.
//
// Arguments:
//   - count (positional, int, default: 2): The number of columns
//   - gutter (named, relative, default: none): The gap between columns
//   - body (positional, content): The content to arrange in columns
func columnsNative(vm *Vm, args *Args) (Value, error) {
	// Get optional count argument (default: 2)
	count := 2
	countArg := args.Find("count")
	if countArg == nil {
		// Try to get positional count if it's an integer
		if peeked := args.Take(); peeked != nil {
			if _, ok := AsInt(peeked.V); ok {
				countArgSpanned, _ := args.Expect("count")
				countArg = &countArgSpanned
			}
		}
	}
	if countArg != nil {
		if !IsNone(countArg.V) && !IsAuto(countArg.V) {
			countVal, ok := AsInt(countArg.V)
			if !ok {
				return nil, &TypeMismatchError{
					Expected: "integer",
					Got:      countArg.V.Type().String(),
					Span:     countArg.Span,
				}
			}
			count = int(countVal)
			if count < 1 {
				return nil, &ConstructorError{
					Message: "column count must be at least 1",
					Span:    countArg.Span,
				}
			}
		}
	}

	// Get optional gutter argument
	var gutter *float64
	if gutterArg := args.Find("gutter"); gutterArg != nil {
		if !IsNone(gutterArg.V) && !IsAuto(gutterArg.V) {
			switch g := gutterArg.V.(type) {
			case LengthValue:
				gutter = &g.Length.Points
			case RelativeValue:
				// For relative, we store the absolute part for now
				// Full relative support would need layout context
				gutter = &g.Relative.Abs.Points
			case RatioValue:
				// Convert ratio to a representative value
				// (full conversion needs layout context)
				pts := g.Ratio.Value * 100 // Scale for visibility
				gutter = &pts
			default:
				return nil, &TypeMismatchError{
					Expected: "length or relative",
					Got:      gutterArg.V.Type().String(),
					Span:     gutterArg.Span,
				}
			}
		}
	}

	// Get required body argument (positional)
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

	// Create the ColumnsElement wrapped in ContentValue
	elem := &ColumnsElement{
		Gutter: gutter,
		Body:   body,
	}
	if count != 2 {
		elem.Count = &count
	}

	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}

// ----------------------------------------------------------------------------
// Box Element
// ----------------------------------------------------------------------------

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

// ----------------------------------------------------------------------------
// Pad Element
// ----------------------------------------------------------------------------

// PadElement represents a padding container element.
// It adds spacing around its content.
type PadElement struct {
	// Left padding (in points).
	Left *float64
	// Top padding (in points).
	Top *float64
	// Right padding (in points).
	Right *float64
	// Bottom padding (in points).
	Bottom *float64
	// Body is the content to pad.
	Body Content
}

func (*PadElement) IsContentElement() {}

// PadFunc creates the pad element function.
func PadFunc() *Func {
	name := "pad"
	return &Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: NativeFunc{
			Func: padNative,
			Info: &FuncInfo{
				Name: "pad",
				Params: []ParamInfo{
					{Name: "body", Type: TypeContent, Named: false},
					{Name: "left", Type: TypeLength, Default: None, Named: true},
					{Name: "top", Type: TypeLength, Default: None, Named: true},
					{Name: "right", Type: TypeLength, Default: None, Named: true},
					{Name: "bottom", Type: TypeLength, Default: None, Named: true},
					{Name: "x", Type: TypeLength, Default: None, Named: true},
					{Name: "y", Type: TypeLength, Default: None, Named: true},
					{Name: "rest", Type: TypeLength, Default: None, Named: true},
				},
			},
		},
	}
}

// padNative implements the pad() function.
// Creates a PadElement with padding around content.
//
// Arguments:
//   - body (positional, content): The content to pad
//   - left (named, length, default: 0pt): Left padding
//   - top (named, length, default: 0pt): Top padding
//   - right (named, length, default: 0pt): Right padding
//   - bottom (named, length, default: 0pt): Bottom padding
//   - x (named, length, default: 0pt): Horizontal padding (sets left and right)
//   - y (named, length, default: 0pt): Vertical padding (sets top and bottom)
//   - rest (named, length, default: 0pt): Padding for all sides
func padNative(vm *Vm, args *Args) (Value, error) {
	elem := &PadElement{}

	// Get rest argument first (applies to all sides)
	if restArg := args.Find("rest"); restArg != nil {
		if !IsNone(restArg.V) && !IsAuto(restArg.V) {
			if lv, ok := restArg.V.(LengthValue); ok {
				r := lv.Length.Points
				elem.Left = &r
				elem.Top = &r
				elem.Right = &r
				elem.Bottom = &r
			} else {
				return nil, &TypeMismatchError{
					Expected: "length",
					Got:      restArg.V.Type().String(),
					Span:     restArg.Span,
				}
			}
		}
	}

	// Get x argument (sets left and right)
	if xArg := args.Find("x"); xArg != nil {
		if !IsNone(xArg.V) && !IsAuto(xArg.V) {
			if lv, ok := xArg.V.(LengthValue); ok {
				x := lv.Length.Points
				elem.Left = &x
				elem.Right = &x
			} else {
				return nil, &TypeMismatchError{
					Expected: "length",
					Got:      xArg.V.Type().String(),
					Span:     xArg.Span,
				}
			}
		}
	}

	// Get y argument (sets top and bottom)
	if yArg := args.Find("y"); yArg != nil {
		if !IsNone(yArg.V) && !IsAuto(yArg.V) {
			if lv, ok := yArg.V.(LengthValue); ok {
				y := lv.Length.Points
				elem.Top = &y
				elem.Bottom = &y
			} else {
				return nil, &TypeMismatchError{
					Expected: "length",
					Got:      yArg.V.Type().String(),
					Span:     yArg.Span,
				}
			}
		}
	}

	// Get individual side arguments (override shorthands)
	if leftArg := args.Find("left"); leftArg != nil {
		if !IsNone(leftArg.V) && !IsAuto(leftArg.V) {
			if lv, ok := leftArg.V.(LengthValue); ok {
				l := lv.Length.Points
				elem.Left = &l
			} else {
				return nil, &TypeMismatchError{
					Expected: "length",
					Got:      leftArg.V.Type().String(),
					Span:     leftArg.Span,
				}
			}
		}
	}

	if topArg := args.Find("top"); topArg != nil {
		if !IsNone(topArg.V) && !IsAuto(topArg.V) {
			if lv, ok := topArg.V.(LengthValue); ok {
				t := lv.Length.Points
				elem.Top = &t
			} else {
				return nil, &TypeMismatchError{
					Expected: "length",
					Got:      topArg.V.Type().String(),
					Span:     topArg.Span,
				}
			}
		}
	}

	if rightArg := args.Find("right"); rightArg != nil {
		if !IsNone(rightArg.V) && !IsAuto(rightArg.V) {
			if lv, ok := rightArg.V.(LengthValue); ok {
				r := lv.Length.Points
				elem.Right = &r
			} else {
				return nil, &TypeMismatchError{
					Expected: "length",
					Got:      rightArg.V.Type().String(),
					Span:     rightArg.Span,
				}
			}
		}
	}

	if bottomArg := args.Find("bottom"); bottomArg != nil {
		if !IsNone(bottomArg.V) && !IsAuto(bottomArg.V) {
			if lv, ok := bottomArg.V.(LengthValue); ok {
				b := lv.Length.Points
				elem.Bottom = &b
			} else {
				return nil, &TypeMismatchError{
					Expected: "length",
					Got:      bottomArg.V.Type().String(),
					Span:     bottomArg.Span,
				}
			}
		}
	}

	// Get required body argument (positional)
	bodyArg := args.Find("body")
	if bodyArg == nil {
		bodyArgSpanned, err := args.Expect("body")
		if err != nil {
			return nil, err
		}
		bodyArg = &bodyArgSpanned
	}

	if cv, ok := bodyArg.V.(ContentValue); ok {
		elem.Body = cv.Content
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

	// Create the PadElement wrapped in ContentValue
	return ContentValue{Content: Content{
		Elements: []ContentElement{elem},
	}}, nil
}
