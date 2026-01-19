// Package flow implements block-level flow layout for Typst documents.
//
// Flow layout handles the positioning of block-level elements like paragraphs,
// headings, and containers. It manages page breaks, floats, and footnotes.
package flow

import (
	"github.com/boergens/gotypst/layout"
)

// BlockElem represents a block element with sizing and styling properties.
type BlockElem struct {
	// Sizing properties
	Width  layout.Sizing
	Height layout.Sizing

	// Inset (padding) on all sides
	Inset layout.Sides[layout.Rel]

	// Body content - one of Content, SingleLayouter, or MultiLayouter
	Body BlockBody

	// Visual styling
	Fill     *layout.Paint
	Stroke   *FixedStroke
	Radius   *layout.Corners[layout.Abs]
	Outset   *layout.Sides[layout.Rel]
	Clip     bool
	Breakable bool

	// Label for the block
	Label *layout.Tag
}

// BlockBody represents the content of a block element.
type BlockBody interface {
	isBlockBody()
}

// ContentBody contains realized content to layout.
type ContentBody struct {
	Content Content
}

func (ContentBody) isBlockBody() {}

// SingleLayouterBody contains a callback for single-region layout.
type SingleLayouterBody struct {
	Layouter func(engine *Engine, region layout.Region) (*layout.Frame, error)
}

func (SingleLayouterBody) isBlockBody() {}

// MultiLayouterBody contains a callback for multi-region layout.
type MultiLayouterBody struct {
	Layouter func(engine *Engine, regions *layout.Regions) (*layout.Fragment, error)
}

func (MultiLayouterBody) isBlockBody() {}

// Content represents realized document content.
// This is a placeholder type that would be filled in with actual content types.
type Content interface{}

// Engine represents the layout engine state.
// This is a placeholder that would contain the actual engine implementation.
type Engine struct {
	// Engine state would go here
}

// Locator tracks element positions for introspection.
type Locator struct {
	// Locator state would go here
}

// StyleChain represents cascaded styles.
type StyleChain struct {
	// Style chain would go here
}

// FixedStroke represents a stroke with all values resolved.
type FixedStroke struct {
	Paint     layout.Paint
	Thickness layout.Abs
	Cap       layout.LineCap
	Join      layout.LineJoin
	Dash      *StrokeDash
}

// StrokeDash represents dash pattern for strokes.
type StrokeDash struct {
	Array []layout.Abs
	Phase layout.Abs
}

// LayoutSingleBlock layouts an unbreakable (non-fragmenting) block element.
//
// This function:
// 1. Fetches sizing properties from the block element
// 2. Builds a pod region using UnbreakablePod
// 3. Layouts the body content
// 4. Applies insets, clipping, and visual styling
func LayoutSingleBlock(
	elem *BlockElem,
	engine *Engine,
	locator Locator,
	styles StyleChain,
	region layout.Region,
) (*layout.Frame, error) {
	// Resolve inset values against the region size
	inset := resolveInset(elem.Inset, styles, region.Size)

	// Build the pod region for layout
	pod := UnbreakablePod(elem.Width, elem.Height, inset, styles, region.Size)

	// Layout the body content
	var frame *layout.Frame
	var err error

	switch body := elem.Body.(type) {
	case nil:
		// Empty body - create zero-sized frame
		frame = layout.NewFrame(layout.ZeroSize)

	case ContentBody:
		// Layout content in the pod region
		frame, err = layoutFrame(engine, body.Content, locator, styles, pod)
		if err != nil {
			return nil, err
		}

	case SingleLayouterBody:
		// Call the single-region layouter
		frame, err = body.Layouter(engine, pod)
		if err != nil {
			return nil, err
		}

	case MultiLayouterBody:
		// Convert region to regions and call multi-layouter
		regions := regionToRegions(pod)
		fragment, err := body.Layouter(engine, regions)
		if err != nil {
			return nil, err
		}
		frame = fragment.IntoFrame()
	}

	// Set frame kind to Hard for explicit blocks (gradient boundary)
	frame.SetKind(layout.FrameKindHard)

	// Enforce size on expanded axes
	enforceSize(frame, elem.Width, elem.Height, pod)

	// Apply insets (padding)
	if !isZeroInset(inset) {
		frame = grow(frame, inset)
	}

	// Apply clipping
	if elem.Clip {
		frame = clipRect(frame, elem.Radius)
	}

	// Apply fill and stroke
	if elem.Fill != nil || elem.Stroke != nil {
		frame = fillAndStroke(frame, elem.Fill, elem.Stroke, elem.Outset, elem.Radius, region.Size, styles)
	}

	// Assign label if present
	if elem.Label != nil {
		frame.Push(layout.Origin, layout.TagItem{Tag: *elem.Label})
	}

	return frame, nil
}

// LayoutMultiBlock layouts a breakable block element that can span multiple regions.
//
// This function:
// 1. Fetches sizing properties from the block element
// 2. Builds pod regions using BreakablePod
// 3. Layouts the body content across regions
// 4. Post-processes each frame with insets and styling
func LayoutMultiBlock(
	elem *BlockElem,
	engine *Engine,
	locator Locator,
	styles StyleChain,
	regions *layout.Regions,
) (*layout.Fragment, error) {
	// Resolve inset values
	inset := resolveInset(elem.Inset, styles, regions.Size)

	// Build pod regions for breakable layout
	backlog := make([]layout.Abs, 0, 2) // Small capacity optimization
	pod := BreakablePod(elem.Width, elem.Height, inset, styles, regions, &backlog)

	// Layout the body content
	var fragment *layout.Fragment
	var err error

	switch body := elem.Body.(type) {
	case nil:
		// Empty body - create frames for each backlog region
		fragment = layout.NewFragmentWithCapacity(len(backlog) + 1)
		fragment.Push(layout.NewFrame(layout.ZeroSize))
		for range backlog {
			fragment.Push(layout.NewFrame(layout.ZeroSize))
		}

	case ContentBody:
		// Layout content across regions
		fragment, err = layoutFragment(engine, body.Content, locator, styles, pod)
		if err != nil {
			return nil, err
		}

		// Handle width consistency for auto-sized content
		if layout.IsAutoOrFr(elem.Width) && fragment.Len() > 1 {
			// Check if widths are inconsistent
			maxWidth := layout.Abs(0)
			for _, f := range fragment.Frames() {
				if f.Width() > maxWidth {
					maxWidth = f.Width()
				}
			}

			// If widths vary, relayout with fixed width
			needsRelayout := false
			for _, f := range fragment.Frames() {
				if !f.Width().ApproxEq(maxWidth) {
					needsRelayout = true
					break
				}
			}

			if needsRelayout {
				// Rebuild pod with fixed width and expansion
				podFixed := pod.WithExpand(layout.Axes[bool]{X: true, Y: pod.Expand.Y})
				podFixed.Size.Width = maxWidth
				fragment, err = layoutFragment(engine, body.Content, locator, styles, podFixed)
				if err != nil {
					return nil, err
				}
			}
		}

	case SingleLayouterBody:
		// Call single layouter and convert to fragment
		frame, err := body.Layouter(engine, pod.First())
		if err != nil {
			return nil, err
		}
		fragment = layout.NewFragment()
		fragment.Push(frame)

	case MultiLayouterBody:
		// Call multi-region layouter directly
		fragment, err = body.Layouter(engine, pod)
		if err != nil {
			return nil, err
		}
	}

	// Post-process each frame
	for i, frame := range fragment.Frames() {
		// Set frame kind to Hard for explicit blocks
		frame.SetKind(layout.FrameKindHard)

		// Enforce size on expanded axes
		enforceSize(frame, elem.Width, elem.Height, pod.First())

		// Apply insets
		if !isZeroInset(inset) {
			fragment.Frames()[i] = grow(frame, inset)
			frame = fragment.Frames()[i]
		}

		// Apply clipping
		if elem.Clip {
			fragment.Frames()[i] = clipRect(frame, elem.Radius)
			frame = fragment.Frames()[i]
		}

		// Apply fill and stroke
		// Skip empty first frame if it's followed by non-empty frames (orphan prevention)
		skipFillStroke := i == 0 && frame.IsEmpty() && fragment.Len() > 1
		if !skipFillStroke && (elem.Fill != nil || elem.Stroke != nil) {
			fragment.Frames()[i] = fillAndStroke(
				frame, elem.Fill, elem.Stroke, elem.Outset, elem.Radius,
				regions.Size, styles,
			)
			frame = fragment.Frames()[i]
		}

		// Assign label (skip empty orphan frames)
		skipLabel := i == 0 && frame.IsEmpty() && fragment.Len() > 1
		if elem.Label != nil && !skipLabel {
			frame.Push(layout.Origin, layout.TagItem{Tag: *elem.Label})
		}
	}

	return fragment, nil
}

// UnbreakablePod builds a region for an unbreakable sized container.
//
// This resolves sizing specifications into concrete dimensions:
// - Auto/Fr sizing uses the full base dimension
// - Rel sizing resolves the relative value against the base
func UnbreakablePod(
	width layout.Sizing,
	height layout.Sizing,
	inset layout.Sides[layout.Abs],
	styles StyleChain,
	base layout.Size,
) layout.Region {
	// Resolve width
	var podWidth layout.Abs
	switch w := width.(type) {
	case layout.SizingAuto, layout.SizingFr:
		podWidth = base.Width
	case layout.SizingRel:
		podWidth = w.Value.Resolve(base.Width)
	default:
		podWidth = base.Width
	}

	// Resolve height
	var podHeight layout.Abs
	switch h := height.(type) {
	case layout.SizingAuto, layout.SizingFr:
		podHeight = base.Height
	case layout.SizingRel:
		podHeight = h.Value.Resolve(base.Height)
	default:
		podHeight = base.Height
	}

	// Shrink by inset
	podWidth = (podWidth - layout.SumHorizontal(inset)).Max(0)
	podHeight = (podHeight - layout.SumVertical(inset)).Max(0)

	// Set expansion flags: expand if manually sized (not Auto) and finite
	expandX := !layout.IsAuto(width) && podWidth.IsFinite()
	expandY := !layout.IsAuto(height) && podHeight.IsFinite()

	return layout.Region{
		Size:   layout.Size{Width: podWidth, Height: podHeight},
		Expand: layout.Axes[bool]{X: expandX, Y: expandY},
	}
}

// BreakablePod builds regions for a breakable sized container.
//
// For auto/Fr height, it inherits regions from the input.
// For Rel height, it distributes the fixed height across regions.
func BreakablePod(
	width layout.Sizing,
	height layout.Sizing,
	inset layout.Sides[layout.Abs],
	styles StyleChain,
	regions *layout.Regions,
	backlog *[]layout.Abs,
) *layout.Regions {
	// Handle height-based region setup
	var pod *layout.Regions

	switch h := height.(type) {
	case layout.SizingAuto, layout.SizingFr:
		// Inherit from input regions
		pod = regions.Clone()

	case layout.SizingRel:
		// Manually sized - distribute height across regions
		resolvedHeight := h.Value.Resolve(regions.Full)
		first, rest := Distribute(resolvedHeight, regions, backlog)

		pod = &layout.Regions{
			Size:    layout.Size{Width: regions.Size.Width, Height: first},
			Full:    first,
			Backlog: rest,
			Last:    nil, // No repeatable region for fixed height
			Expand:  regions.Expand,
		}

	default:
		pod = regions.Clone()
	}

	// Resolve width
	switch w := width.(type) {
	case layout.SizingAuto, layout.SizingFr:
		// Keep full width
	case layout.SizingRel:
		resolvedWidth := w.Value.Resolve(regions.Size.Width)
		pod.Size.Width = resolvedWidth
	}

	// Shrink by inset
	pod = pod.ShrinkMultiple(inset)

	// Set expansion flags
	expandX := !layout.IsAuto(width) && pod.Size.Width.IsFinite()
	expandY := !layout.IsAuto(height) && pod.Size.Height.IsFinite()
	pod.Expand = layout.Axes[bool]{X: expandX, Y: expandY}

	return pod
}

// Distribute allocates a fixed height across existing regions.
// Returns the first region height and the backlog of remaining heights.
func Distribute(
	height layout.Abs,
	regions *layout.Regions,
	buf *[]layout.Abs,
) (layout.Abs, []layout.Abs) {
	*buf = (*buf)[:0] // Clear buffer

	if height <= 0 {
		return 0, *buf
	}

	remaining := height
	current := regions.Clone()

	for {
		// Clamp to current region's available space
		available := current.Size.Height
		used := remaining.Min(available)
		*buf = append(*buf, used)
		remaining -= used

		// Check termination conditions
		if remaining.ApproxEq(0) {
			// Fully distributed
			break
		}

		if !current.CanBreak() {
			// Cannot break to next region
			break
		}

		if available.ApproxEq(0) && current.Size.Height.ApproxEq(0) {
			// No progress possible (empty region and no space)
			break
		}

		// Advance to next region
		if !current.Next() {
			break
		}
	}

	// If remaining height exists, add it to the last buffer entry (overflow)
	if remaining > 0 && len(*buf) > 0 {
		(*buf)[len(*buf)-1] += remaining
	}

	if len(*buf) == 0 {
		return 0, *buf
	}

	// Return first element as first region height, rest as backlog
	first := (*buf)[0]
	rest := (*buf)[1:]
	return first, rest
}

// Helper functions - these would be implemented in separate files

// resolveInset resolves relative inset values to absolute values.
func resolveInset(inset layout.Sides[layout.Rel], styles StyleChain, base layout.Size) layout.Sides[layout.Abs] {
	return layout.Sides[layout.Abs]{
		Left:   inset.Left.Resolve(base.Width),
		Top:    inset.Top.Resolve(base.Height),
		Right:  inset.Right.Resolve(base.Width),
		Bottom: inset.Bottom.Resolve(base.Height),
	}
}

// isZeroInset checks if all inset values are zero.
func isZeroInset(inset layout.Sides[layout.Abs]) bool {
	return inset.Left.IsZero() && inset.Top.IsZero() &&
		inset.Right.IsZero() && inset.Bottom.IsZero()
}

// regionToRegions converts a single region to a Regions structure.
func regionToRegions(region layout.Region) *layout.Regions {
	return &layout.Regions{
		Size:   region.Size,
		Full:   region.Size.Height,
		Expand: region.Expand,
	}
}

// enforceSize enforces the frame size on expanded axes.
func enforceSize(frame *layout.Frame, width, height layout.Sizing, pod layout.Region) {
	size := frame.Size()
	if pod.Expand.X && !layout.IsAuto(width) {
		size.Width = pod.Size.Width
	}
	if pod.Expand.Y && !layout.IsAuto(height) {
		size.Height = pod.Size.Height
	}
	frame.SetSize(size)
}

// Placeholder functions that would call into the main layout engine

// layoutFrame layouts content in a single region.
func layoutFrame(engine *Engine, content Content, locator Locator, styles StyleChain, region layout.Region) (*layout.Frame, error) {
	// This would call into the main layout engine
	// For now, return an empty frame as a placeholder
	return layout.NewFrame(region.Size), nil
}

// layoutFragment layouts content across multiple regions.
func layoutFragment(engine *Engine, content Content, locator Locator, styles StyleChain, regions *layout.Regions) (*layout.Fragment, error) {
	// This would call into the main layout engine
	// For now, return a single-frame fragment as a placeholder
	fragment := layout.NewFragment()
	fragment.Push(layout.NewFrame(regions.Size))
	return fragment, nil
}

// grow adds inset padding to a frame.
func grow(frame *layout.Frame, inset layout.Sides[layout.Abs]) *layout.Frame {
	// Create a new frame with increased size
	newSize := layout.Size{
		Width:  frame.Width() + layout.SumHorizontal(inset),
		Height: frame.Height() + layout.SumVertical(inset),
	}
	newFrame := layout.NewFrame(newSize)
	newFrame.SetKind(frame.Kind())

	// Offset the original content by the inset
	offset := layout.Point{X: inset.Left, Y: inset.Top}
	newFrame.PushFrame(offset, frame)

	return newFrame
}

// clipRect creates a clipping rectangle for the frame.
func clipRect(frame *layout.Frame, radius *layout.Corners[layout.Abs]) *layout.Frame {
	// Create a group with clipping
	clip := layout.Clip{
		Shape: layout.RectShape{
			Size:   frame.Size(),
			Radius: cornerOrZero(radius),
		},
	}

	group := layout.GroupItem{
		Frame: frame,
		Clips: []layout.Clip{clip},
	}

	newFrame := layout.NewFrame(frame.Size())
	newFrame.SetKind(frame.Kind())
	newFrame.Push(layout.Origin, group)
	return newFrame
}

// fillAndStroke applies fill and stroke styling to a frame.
func fillAndStroke(
	frame *layout.Frame,
	fill *layout.Paint,
	stroke *FixedStroke,
	outset *layout.Sides[layout.Rel],
	radius *layout.Corners[layout.Abs],
	base layout.Size,
	styles StyleChain,
) *layout.Frame {
	// Calculate shape size (may include outset)
	shapeSize := frame.Size()
	shapeOffset := layout.Origin

	if outset != nil {
		resolved := layout.Sides[layout.Abs]{
			Left:   outset.Left.Resolve(base.Width),
			Top:    outset.Top.Resolve(base.Height),
			Right:  outset.Right.Resolve(base.Width),
			Bottom: outset.Bottom.Resolve(base.Height),
		}
		shapeSize = layout.Size{
			Width:  shapeSize.Width + layout.SumHorizontal(resolved),
			Height: shapeSize.Height + layout.SumVertical(resolved),
		}
		shapeOffset = layout.Point{X: -resolved.Left, Y: -resolved.Top}
	}

	// Create the shape
	shape := layout.RectShape{
		Size:   shapeSize,
		Radius: cornerOrZero(radius),
	}

	// Create shape item with fill and/or stroke
	var strokeLayout *layout.Stroke
	if stroke != nil {
		strokeLayout = &layout.Stroke{
			Paint:     stroke.Paint,
			Thickness: stroke.Thickness,
			LineCap:   stroke.Cap,
			LineJoin:  stroke.Join,
		}
		if stroke.Dash != nil {
			strokeLayout.DashArray = stroke.Dash.Array
			strokeLayout.DashPhase = stroke.Dash.Phase
		}
	}

	shapeItem := layout.ShapeItem{
		Shape:  shape,
		Fill:   fill,
		Stroke: strokeLayout,
	}

	// Create new frame with shape as background
	newFrame := layout.NewFrame(frame.Size())
	newFrame.SetKind(frame.Kind())
	newFrame.Push(shapeOffset, shapeItem)

	// Copy original frame content
	for _, item := range frame.Items() {
		newFrame.Push(layout.Origin, item)
	}

	return newFrame
}

// cornerOrZero returns corners or zero corners.
func cornerOrZero(corners *layout.Corners[layout.Abs]) layout.Corners[layout.Abs] {
	if corners != nil {
		return *corners
	}
	return layout.Corners[layout.Abs]{}
}
