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
	// Width of the box. If nil, auto-sizes to content.
	Width *foundations.Relative `typst:"width,type=relative"`
	// Height of the box. If nil, auto-sizes to content.
	Height *foundations.Relative `typst:"height,type=relative"`
	// Baseline position. If nil, uses content baseline.
	Baseline *foundations.Length `typst:"baseline,type=length"`
	// Fill color for the background. If nil, no fill.
	Fill foundations.Value `typst:"fill"`
	// Stroke for the border. Can be length, color, or stroke dict. If nil, no stroke.
	Stroke foundations.Value `typst:"stroke"`
	// Radius for rounded corners. Can be single value or dictionary.
	Radius foundations.Value `typst:"radius"`
	// Inset padding inside the box.
	Inset foundations.Value `typst:"inset"`
	// Outset expansion outside the box.
	Outset foundations.Value `typst:"outset"`
	// Whether to clip content that overflows the box.
	Clip bool `typst:"clip,type=bool,default=false"`
	// Body is the content inside the box.
	Body foundations.Content `typst:"body,positional,type=content"`
}

func (*BoxElement) IsContentElement() {}

// BlockElement represents a block-level container element.
// It creates a new block in the document flow with optional sizing and styling.
//
// Reference: typst-reference/crates/typst-library/src/layout/container.rs
type BlockElement struct {
	// Width of the block. If nil, auto-sizes.
	Width *foundations.Relative `typst:"width,type=relative"`
	// Height of the block. If nil, auto-sizes.
	Height *foundations.Relative `typst:"height,type=relative"`
	// Whether the block can break across pages.
	Breakable *bool `typst:"breakable,type=bool,default=true"`
	// Fill color for the background.
	Fill foundations.Value `typst:"fill"`
	// Stroke for the border.
	Stroke foundations.Value `typst:"stroke"`
	// Radius for rounded corners.
	Radius foundations.Value `typst:"radius"`
	// Inset padding inside the block.
	Inset foundations.Value `typst:"inset"`
	// Outset expansion outside the block.
	Outset foundations.Value `typst:"outset"`
	// Spacing between adjacent blocks.
	Spacing *foundations.Length `typst:"spacing,type=length"`
	// Spacing above this block (overrides Spacing).
	Above *foundations.Length `typst:"above,type=length"`
	// Spacing below this block (overrides Spacing).
	Below *foundations.Length `typst:"below,type=length"`
	// Whether to clip content that overflows.
	Clip bool `typst:"clip,type=bool,default=false"`
	// Whether the block sticks to the next block.
	Sticky bool `typst:"sticky,type=bool,default=false"`
	// Body is the content inside the block.
	Body foundations.Content `typst:"body,positional,type=content"`
}

func (*BlockElement) IsContentElement() {}

// BoxDef is the registered element definition for box.
var BoxDef *foundations.ElementDef

// BlockDef is the registered element definition for block.
var BlockDef *foundations.ElementDef

func init() {
	BoxDef = foundations.RegisterElement[BoxElement]("box", nil)
	BlockDef = foundations.RegisterElement[BlockElement]("block", nil)
}

// WidthPts returns the width in points, or 0 if not set.
func (b *BoxElement) WidthPts() float64 {
	if b.Width == nil {
		return 0
	}
	return b.Width.Abs.Points
}

// HeightPts returns the height in points, or 0 if not set.
func (b *BoxElement) HeightPts() float64 {
	if b.Height == nil {
		return 0
	}
	return b.Height.Abs.Points
}

// BaselinePts returns the baseline in points, or 0 if not set.
func (b *BoxElement) BaselinePts() float64 {
	if b.Baseline == nil {
		return 0
	}
	return b.Baseline.Points
}

// WidthPts returns the width in points, or 0 if not set.
func (b *BlockElement) WidthPts() float64 {
	if b.Width == nil {
		return 0
	}
	return b.Width.Abs.Points
}

// HeightPts returns the height in points, or 0 if not set.
func (b *BlockElement) HeightPts() float64 {
	if b.Height == nil {
		return 0
	}
	return b.Height.Abs.Points
}

// IsBreakable returns whether the block can break across pages.
func (b *BlockElement) IsBreakable() bool {
	if b.Breakable == nil {
		return true
	}
	return *b.Breakable
}

// SpacingPts returns the spacing in points, or 0 if not set.
func (b *BlockElement) SpacingPts() float64 {
	if b.Spacing == nil {
		return 0
	}
	return b.Spacing.Points
}

// AbovePts returns the above spacing in points, or 0 if not set.
func (b *BlockElement) AbovePts() float64 {
	if b.Above == nil {
		return 0
	}
	return b.Above.Points
}

// BelowPts returns the below spacing in points, or 0 if not set.
func (b *BlockElement) BelowPts() float64 {
	if b.Below == nil {
		return 0
	}
	return b.Below.Points
}

// BoxFunc creates the box element function.
func BoxFunc() *foundations.Func {
	name := "box"
	return &foundations.Func{
		Name: &name,
		Span: syntax.Detached(),
		Repr: foundations.NativeFunc{
			Func: boxNative,
			Info: BoxDef.ToFuncInfo(),
		},
	}
}

// boxNative implements the box() function using the generic element parser.
func boxNative(engine foundations.Engine, context foundations.Context, args *foundations.Args) (foundations.Value, error) {
	elem, err := foundations.ParseElement[BoxElement](BoxDef, args)
	if err != nil {
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
			Info: BlockDef.ToFuncInfo(),
		},
	}
}

// blockNative implements the block() function using the generic element parser.
func blockNative(engine foundations.Engine, context foundations.Context, args *foundations.Args) (foundations.Value, error) {
	elem, err := foundations.ParseElement[BlockElement](BlockDef, args)
	if err != nil {
		return nil, err
	}

	return foundations.ContentValue{Content: foundations.Content{
		Elements: []foundations.ContentElement{elem},
	}}, nil
}
