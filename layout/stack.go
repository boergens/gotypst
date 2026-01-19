package layout

import (
	"fmt"

	"github.com/boergens/gotypst/syntax"
)

// StackChild represents a child element of a stack.
type StackChild interface {
	isStackChild()
}

// StackSpacing represents spacing in a stack.
type StackSpacing struct {
	Amount Spacing
}

func (StackSpacing) isStackChild() {}

// StackBlock represents a block element in a stack.
type StackBlock struct {
	Content interface{} // The block content to layout
}

func (StackBlock) isStackChild() {}

// Spacing represents either absolute or fractional spacing.
type Spacing interface {
	isSpacing()
}

// RelSpacing represents relative (absolute) spacing.
type RelSpacing struct {
	Value Abs
}

func (RelSpacing) isSpacing() {}

// FrSpacing represents fractional spacing.
type FrSpacing struct {
	Value Fr
}

func (FrSpacing) isSpacing() {}

// StackElem represents the stack element configuration.
type StackElem struct {
	Dir      Dir
	Spacing  Spacing
	Children []StackChild
	Span     syntax.Span
}

// LayoutStack lays out a stack of elements.
//
// This function arranges children either horizontally or vertically based on
// the stack's direction, handling both absolute and fractional spacing.
func LayoutStack(
	elem *StackElem,
	layoutFunc LayoutFunc,
	styles StyleChain,
	regions *Regions,
) (Fragment, error) {
	layouter := newStackLayouter(elem.Span, elem.Dir, styles, regions)

	axis := layouter.dir.Axis()

	// Spacing to insert before the next block.
	var deferred Spacing

	for _, child := range elem.Children {
		switch c := child.(type) {
		case StackSpacing:
			layouter.layoutSpacing(c.Amount)
			deferred = nil

		case StackBlock:
			// Transparently handle `h` elements for horizontal stacks.
			if axis == AxisX {
				if h, ok := c.Content.(*HElem); ok {
					layouter.layoutSpacing(h.Amount)
					deferred = nil
					continue
				}
			}

			// Transparently handle `v` elements for vertical stacks.
			if axis == AxisY {
				if v, ok := c.Content.(*VElem); ok {
					layouter.layoutSpacing(v.Amount)
					deferred = nil
					continue
				}
			}

			if deferred != nil {
				layouter.layoutSpacing(deferred)
			}

			if err := layouter.layoutBlock(layoutFunc, c.Content, styles); err != nil {
				return nil, err
			}
			deferred = elem.Spacing
		}
	}

	return layouter.finish()
}

// HElem represents a horizontal spacing element.
type HElem struct {
	Amount Spacing
}

// VElem represents a vertical spacing element.
type VElem struct {
	Amount Spacing
}

// LayoutFunc is a function type for laying out content.
type LayoutFunc func(content interface{}, styles StyleChain, regions *Regions) (Fragment, error)

// StyleChain represents inherited styles (placeholder).
type StyleChain struct {
	// Styles would contain the actual style properties
}

// stackLayouter performs stack layout.
type stackLayouter struct {
	// span is the span to raise errors at during layout.
	span syntax.Span
	// dir is the stacking direction.
	dir Dir
	// axis is the axis of the stacking direction.
	axis Axis
	// styles contains the inherited styles.
	styles StyleChain
	// regions contains the layout regions.
	regions *Regions
	// expand indicates whether the stack should expand to fill the region.
	expand Axes[bool]
	// initial is the size of the current region before we started subtracting.
	initial Size
	// used is the generic size used by frames for the current region.
	used stackGenericSize
	// fr is the sum of fractions in the current region.
	fr Fr
	// items contains layouted items whose exact positions are not yet known.
	items []stackItem
	// finished contains frames for previous regions.
	finished []Frame
}

// stackItem is a prepared item in a stack layout.
type stackItem interface {
	isStackItem()
}

// stackItemAbsolute represents absolute spacing between other items.
type stackItemAbsolute struct {
	Value Abs
}

func (stackItemAbsolute) isStackItem() {}

// stackItemFractional represents fractional spacing between other items.
type stackItemFractional struct {
	Value Fr
}

func (stackItemFractional) isStackItem() {}

// stackItemFrame represents a frame for a layouted block.
type stackItemFrame struct {
	Frame *Frame
	Align Axes[FixedAlignment]
}

func (stackItemFrame) isStackItem() {}

// newStackLayouter creates a new stack layouter.
func newStackLayouter(
	span syntax.Span,
	dir Dir,
	styles StyleChain,
	regions *Regions,
) *stackLayouter {
	axis := dir.Axis()
	expand := regions.Expand

	// Disable expansion along the block axis for children.
	childRegions := *regions
	if axis == AxisX {
		childRegions.Expand.X = false
	} else {
		childRegions.Expand.Y = false
	}

	return &stackLayouter{
		span:     span,
		dir:      dir,
		axis:     axis,
		styles:   styles,
		regions:  &childRegions,
		expand:   expand,
		initial:  regions.Size,
		used:     stackGenericSize{},
		fr:       FrZero(),
		items:    nil,
		finished: nil,
	}
}

// layoutSpacing adds spacing along the spacing direction.
func (l *stackLayouter) layoutSpacing(spacing Spacing) {
	switch s := spacing.(type) {
	case RelSpacing:
		// Resolve the spacing and limit it to the remaining space.
		resolved := s.Value
		var remaining *Abs
		if l.axis == AxisX {
			remaining = &l.regions.Size.X
		} else {
			remaining = &l.regions.Size.Y
		}
		limited := resolved.Min(*remaining)
		if l.dir.Axis() == AxisY {
			*remaining -= limited
		}
		l.used.Main += limited
		l.items = append(l.items, stackItemAbsolute{Value: resolved})

	case FrSpacing:
		l.fr += s.Value
		l.items = append(l.items, stackItemFractional{Value: s.Value})
	}
}

// layoutBlock lays out an arbitrary block.
func (l *stackLayouter) layoutBlock(
	layoutFunc LayoutFunc,
	block interface{},
	styles StyleChain,
) error {
	if l.regionIsFull() {
		if err := l.finishRegion(); err != nil {
			return err
		}
	}

	// Get alignment for the block (simplified - would check AlignElem in full impl)
	align := Axes[FixedAlignment]{X: AlignStart, Y: AlignStart}

	fragment, err := layoutFunc(block, styles, l.regions)
	if err != nil {
		return err
	}

	length := len(fragment)
	for i, frame := range fragment {
		// Grow our size, shrink the region and save the frame for later.
		specificSize := frame.Size

		if l.dir.Axis() == AxisY {
			l.regions.Size.Y -= specificSize.Y
		}

		var genericSize stackGenericSize
		if l.axis == AxisX {
			genericSize = stackGenericSize{Main: specificSize.Y, Cross: specificSize.X}
		} else {
			genericSize = stackGenericSize{Main: specificSize.X, Cross: specificSize.Y}
		}

		l.used.Main += genericSize.Main
		if genericSize.Cross > l.used.Cross {
			l.used.Cross = genericSize.Cross
		}

		frameCopy := frame
		l.items = append(l.items, stackItemFrame{Frame: &frameCopy, Align: align})

		if i+1 < length {
			if err := l.finishRegion(); err != nil {
				return err
			}
		}
	}

	return nil
}

// regionIsFull returns true if the current region is full.
func (l *stackLayouter) regionIsFull() bool {
	if l.axis == AxisX {
		return l.regions.Size.X <= 0
	}
	return l.regions.Size.Y <= 0
}

// finishRegion advances to the next region.
func (l *stackLayouter) finishRegion() error {
	// Determine the size of the stack in this region depending on whether
	// the region expands.
	var size Size
	usedAxes := l.used.ToAxes(l.axis)
	if l.expand.X {
		size.X = l.initial.X
	} else {
		size.X = usedAxes.X
	}
	if l.expand.Y {
		size.Y = l.initial.Y
	} else {
		size.Y = usedAxes.Y
	}
	size = size.Min(l.initial)

	// Get full size and remaining space.
	var full Abs
	if l.axis == AxisX {
		full = l.initial.X
	} else {
		full = l.initial.Y
	}
	remaining := full - l.used.Main

	// Expand fully if there are fr spacings.
	if float64(l.fr) > 0.0 && full.IsFinite() {
		l.used.Main = full
		if l.axis == AxisX {
			size.X = full
		} else {
			size.Y = full
		}
	}

	if !size.X.IsFinite() || !size.Y.IsFinite() {
		return fmt.Errorf("stack spacing is infinite at %v", l.span)
	}

	output := NewFrameHard(size)
	cursor := AbsZero()
	ruler := l.dirStart()

	// Place all frames.
	for _, item := range l.items {
		switch it := item.(type) {
		case stackItemAbsolute:
			cursor += it.Value

		case stackItemFractional:
			cursor += it.Value.Share(l.fr, remaining)

		case stackItemFrame:
			frame := it.Frame
			align := it.Align

			var alignAxis FixedAlignment
			if l.axis == AxisX {
				alignAxis = align.X
			} else {
				alignAxis = align.Y
			}

			if l.dir.IsPositive() {
				if alignAxis > ruler {
					ruler = alignAxis
				}
			} else {
				if alignAxis < ruler {
					ruler = alignAxis
				}
			}

			// Align along the main axis.
			var parent, child Abs
			if l.axis == AxisX {
				parent = size.X
				child = frame.Size.X
			} else {
				parent = size.Y
				child = frame.Size.Y
			}

			main := ruler.Position(child, parent-l.used.Main)
			if l.dir.IsPositive() {
				main += cursor
			} else {
				main += l.used.Main - child - cursor
			}

			// Align along the cross axis.
			other := l.axis.Other()
			var crossAlign FixedAlignment
			var crossParent, crossChild Abs
			if other == AxisX {
				crossAlign = align.X
				crossParent = size.X
				crossChild = frame.Size.X
			} else {
				crossAlign = align.Y
				crossParent = size.Y
				crossChild = frame.Size.Y
			}
			cross := crossAlign.Position(crossChild, crossParent-crossChild)

			var pos Point
			if l.axis == AxisX {
				pos = NewPoint(main, cross)
			} else {
				pos = NewPoint(cross, main)
			}

			cursor += child
			output.PushFrame(pos, frame)
		}
	}

	// Advance to the next region.
	l.advanceRegion()
	l.initial = l.regions.Size
	l.used = stackGenericSize{}
	l.fr = FrZero()
	l.items = nil
	l.finished = append(l.finished, *output)

	return nil
}

// dirStart returns the starting alignment for the direction.
func (l *stackLayouter) dirStart() FixedAlignment {
	if l.dir.IsPositive() {
		return AlignStart
	}
	return AlignEnd
}

// advanceRegion moves to the next region.
func (l *stackLayouter) advanceRegion() {
	if len(l.regions.Backlog) > 0 {
		l.regions.Size = l.regions.Backlog[0]
		l.regions.Backlog = l.regions.Backlog[1:]
	} else if !l.regions.Last {
		l.regions.Size = l.regions.Full
	}
}

// finish completes layouting and returns the resulting frames.
func (l *stackLayouter) finish() (Fragment, error) {
	if err := l.finishRegion(); err != nil {
		return nil, err
	}
	return l.finished, nil
}

// stackGenericSize is a generic size with main and cross axes.
type stackGenericSize struct {
	Cross Abs
	Main  Abs
}

// ToAxes converts to the specific representation given the current main axis.
func (g stackGenericSize) ToAxes(main Axis) Axes[Abs] {
	if main == AxisX {
		return Axes[Abs]{X: g.Main, Y: g.Cross}
	}
	return Axes[Abs]{X: g.Cross, Y: g.Main}
}
