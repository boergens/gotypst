package flow

import (
	"github.com/boergens/gotypst/layout"
)

// FlowMode represents the mode of flow layout.
type FlowMode int

const (
	// FlowModeRoot is for block elements with footnotes at the document root.
	FlowModeRoot FlowMode = iota
	// FlowModeBlock is for block-level children.
	FlowModeBlock
	// FlowModeInline is for inline-level children.
	FlowModeInline
)

// Stop represents control flow events during layout.
type Stop interface {
	isStop()
}

// StopFinish signals that the region should be finished.
type StopFinish struct {
	// Forced indicates if this is a forced break (e.g., column break).
	Forced bool
}

func (StopFinish) isStop() {}

// StopRelayout signals that the region needs to be relaid out due to an insertion.
type StopRelayout struct {
	Scope PlacementScope
}

func (StopRelayout) isStop() {}

// StopError represents a fatal error during layout.
type StopError struct {
	Err error
}

func (StopError) isStop() {}

// FlowResult represents the result of a flow operation.
type FlowResult[T any] struct {
	Value T
	Stop  Stop
}

// Ok creates a successful FlowResult.
func Ok[T any](value T) FlowResult[T] {
	return FlowResult[T]{Value: value}
}

// Err creates a FlowResult with a stop condition.
func Err[T any](stop Stop) FlowResult[T] {
	return FlowResult[T]{Stop: stop}
}

// IsOk returns true if the result is successful.
func (r FlowResult[T]) IsOk() bool {
	return r.Stop == nil
}

// PlacementScope represents where a placed element is scoped.
type PlacementScope int

const (
	// PlacementScopeColumn scopes to the current column.
	PlacementScopeColumn PlacementScope = iota
	// PlacementScopePage scopes to the current page.
	PlacementScopePage
)

// FixedAlignment represents a resolved alignment value.
type FixedAlignment int

const (
	// FixedAlignStart aligns to the start.
	FixedAlignStart FixedAlignment = iota
	// FixedAlignCenter aligns to the center.
	FixedAlignCenter
	// FixedAlignEnd aligns to the end.
	FixedAlignEnd
)

// Position calculates the offset for this alignment within the given free space.
func (a FixedAlignment) Position(free layout.Abs) layout.Abs {
	switch a {
	case FixedAlignStart:
		return 0
	case FixedAlignCenter:
		return free / 2
	case FixedAlignEnd:
		return free
	default:
		return 0
	}
}

// Max returns the maximum of two alignments.
func (a FixedAlignment) Max(other FixedAlignment) FixedAlignment {
	if other > a {
		return other
	}
	return a
}

// Axes holds a pair of values for horizontal (X) and vertical (Y) axes.
type Axes[T any] struct {
	X, Y T
}

// Tag represents an introspection tag.
type Tag struct {
	// Location uniquely identifies this tag.
	Location Location
}

// Clone creates a copy of the tag.
func (t *Tag) Clone() Tag {
	return Tag{Location: t.Location}
}

// Location uniquely identifies an element in the document.
type Location uint64

// Child represents a child element in flow layout.
type Child interface {
	isChild()
}

// TagChild represents an introspection tag.
type TagChild struct {
	Tag *Tag
}

func (TagChild) isChild() {}

// RelChild represents relative spacing.
type RelChild struct {
	Amount   Rel
	Weakness uint8
}

func (RelChild) isChild() {}

// FrChild represents fractional spacing.
type FrChild struct {
	Amount   layout.Fr
	Weakness uint8
}

func (FrChild) isChild() {}

// LineChild represents a line from a paragraph.
type LineChild struct {
	Frame Frame
	Align Axes[FixedAlignment]
	// Need includes the line's height plus following lines grouped by
	// widow/orphan prevention.
	Need layout.Abs
}

func (LineChild) isChild() {}

// SingleChild represents an unbreakable block.
type SingleChild struct {
	Align  Axes[FixedAlignment]
	Sticky bool
	Alone  bool
	Fr     *layout.Fr // nil if not fractionally sized
	frame  *Frame     // cached layout result
}

func (SingleChild) isChild() {}

// Layout lays out the single child in the given region.
func (s *SingleChild) Layout(engine *Engine, region Region) (Frame, error) {
	if s.frame != nil {
		return *s.frame, nil
	}
	// TODO: Implement actual layout
	return Frame{}, nil
}

// MultiChild represents a breakable block.
type MultiChild struct {
	Align  Axes[FixedAlignment]
	Sticky bool
	Alone  bool
}

func (MultiChild) isChild() {}

// Layout lays out the multi child across regions, returning the first frame
// and optional spill for remaining content.
func (m *MultiChild) Layout(engine *Engine, regions Regions) (Frame, *MultiSpill, error) {
	// TODO: Implement actual layout
	return Frame{}, nil, nil
}

// PlacedChild represents an absolutely or floatingly placed child.
type PlacedChild struct {
	AlignX    FixedAlignment
	AlignY    *FixedAlignment // nil means auto
	Scope     PlacementScope
	Float     bool
	Clearance layout.Abs
	Delta     Axes[Rel]
	location  Location
}

func (PlacedChild) isChild() {}

// Layout lays out the placed child at the given base size.
func (p *PlacedChild) Layout(engine *Engine, base layout.Size) (Frame, error) {
	// TODO: Implement actual layout
	return Frame{}, nil
}

// Location returns the location of this placed child.
func (p *PlacedChild) Location() Location {
	return p.location
}

// FlushChild signals that floats should be flushed.
type FlushChild struct{}

func (FlushChild) isChild() {}

// BreakChild represents a column/page break.
type BreakChild struct {
	// Weak indicates if this is a weak break (can be ignored if at region start).
	Weak bool
}

func (BreakChild) isChild() {}

// MultiSpill represents spillover from a breakable block.
type MultiSpill struct {
	// ExistNonEmptyFrame indicates if there are non-empty frames in the spill.
	ExistNonEmptyFrame bool
	multi              *MultiChild
	first              layout.Abs
	full               layout.Abs
	backlog            []layout.Abs
	minBacklogLen      int
}

// Layout continues layout of the spill in the given regions.
func (s *MultiSpill) Layout(engine *Engine, regions Regions) (Frame, *MultiSpill, error) {
	// TODO: Implement actual layout
	return Frame{}, nil, nil
}

// Align returns the alignment of the underlying multi child.
func (s *MultiSpill) Align() Axes[FixedAlignment] {
	return s.multi.Align
}

// Rel represents a relative length (combination of absolute and ratio).
type Rel struct {
	Abs   layout.Abs
	Ratio float64
}

// RelativeTo resolves the relative length to an absolute length.
func (r Rel) RelativeTo(base layout.Abs) layout.Abs {
	return r.Abs + layout.Abs(r.Ratio*float64(base))
}

// RelAxesToPoint converts axes of relative values to a point.
func RelAxesToPoint(a Axes[Rel], size layout.Size) layout.Point {
	return layout.Point{
		X: a.X.RelativeTo(size.Width),
		Y: a.Y.RelativeTo(size.Height),
	}
}

// Frame represents a laid out frame with positioned items.
type Frame struct {
	size  layout.Size
	items []FrameEntry
}

// FrameEntry represents a positioned item in a frame.
type FrameEntry struct {
	Pos  layout.Point
	Item FrameItem
}

// NewFrame creates a new frame with the given size.
func NewFrame(size layout.Size) Frame {
	return Frame{size: size}
}

// Soft creates a soft frame that can be resized.
func Soft(size layout.Size) Frame {
	return Frame{size: size}
}

// Size returns the frame's size.
func (f *Frame) Size() layout.Size {
	return f.size
}

// Width returns the frame's width.
func (f *Frame) Width() layout.Abs {
	return f.size.Width
}

// Height returns the frame's height.
func (f *Frame) Height() layout.Abs {
	return f.size.Height
}

// IsEmpty returns true if the frame has no items.
func (f *Frame) IsEmpty() bool {
	return len(f.items) == 0
}

// Items returns an iterator over the frame's items.
func (f *Frame) Items() []FrameEntry {
	return f.items
}

// Push adds an item to the frame at the given position.
func (f *Frame) Push(pos layout.Point, item FrameItem) {
	f.items = append(f.items, FrameEntry{Pos: pos, Item: item})
}

// PushFrame adds a nested frame at the given position.
func (f *Frame) PushFrame(pos layout.Point, frame Frame) {
	f.items = append(f.items, FrameEntry{Pos: pos, Item: FrameItemFrame{Frame: frame}})
}

// Clone creates a copy of the frame.
func (f *Frame) Clone() Frame {
	items := make([]FrameEntry, len(f.items))
	copy(items, f.items)
	return Frame{size: f.size, items: items}
}

// FrameItem represents an item that can be placed in a frame.
type FrameItem interface {
	isFrameItem()
}

// FrameItemTag represents a tag in a frame.
type FrameItemTag struct {
	Tag Tag
}

func (FrameItemTag) isFrameItem() {}

// FrameItemLink represents a link in a frame.
type FrameItemLink struct {
	Dest string
	Size layout.Size
}

func (FrameItemLink) isFrameItem() {}

// FrameItemFrame represents a nested frame.
type FrameItemFrame struct {
	Frame Frame
}

func (FrameItemFrame) isFrameItem() {}

// FrameItemImage represents an image in a frame.
type FrameItemImage struct {
	// Data is the raw image bytes.
	Data []byte
	// Format specifies the image format (jpeg, png, raw).
	Format string
	// Width is the natural image width in pixels.
	Width int
	// Height is the natural image height in pixels.
	Height int
	// RenderSize is the size to render the image at.
	RenderSize layout.Size
}

func (FrameItemImage) isFrameItem() {}

// Region represents a single layout region.
type Region struct {
	Size   layout.Size
	Expand Axes[bool]
}

// NewRegion creates a new region with the given size and expansion settings.
func NewRegion(size layout.Size, expand Axes[bool]) Region {
	return Region{Size: size, Expand: expand}
}

// Regions represents multiple layout regions.
type Regions struct {
	Size    layout.Size
	Expand  Axes[bool]
	Full    layout.Size
	Backlog []layout.Abs
	Last    *layout.Size
}

// NewRegions creates new regions with the given parameters.
func NewRegions(size layout.Size, expand Axes[bool], full layout.Size) Regions {
	return Regions{
		Size:   size,
		Expand: expand,
		Full:   full,
	}
}

// Base returns the base size for relative measurements.
func (r *Regions) Base() layout.Size {
	return r.Full
}

// MayProgress returns true if moving to a subsequent region might improve things.
func (r *Regions) MayProgress() bool {
	return len(r.Backlog) > 0 || r.Last != nil
}

// IsFull returns true if the region is (over)full.
func (r *Regions) IsFull() bool {
	return r.Size.Height <= 0
}

// Iter returns an iterator over remaining region heights.
func (r *Regions) Iter() []layout.Abs {
	result := []layout.Abs{r.Size.Height}
	result = append(result, r.Backlog...)
	if r.Last != nil {
		result = append(result, r.Last.Height)
	}
	return result
}

// Engine represents the layout engine context.
type Engine struct {
	// TODO: Add engine fields as needed
}

// Work tracks the progress of flow layout.
type Work struct {
	children []Child
	index    int
	Spill    *MultiSpill
	Floats   []*PlacedChild
	Tags     []*Tag
	Skips    map[Location]struct{}
}

// NewWork creates a new work tracker for the given children.
func NewWork(children []Child) *Work {
	return &Work{
		children: children,
		Skips:    make(map[Location]struct{}),
	}
}

// Head returns the current child, or nil if done.
func (w *Work) Head() Child {
	if w.index >= len(w.children) {
		return nil
	}
	return w.children[w.index]
}

// Advance moves to the next child.
func (w *Work) Advance() {
	w.index++
}

// Done returns true if all children have been processed.
func (w *Work) Done() bool {
	return w.index >= len(w.children) && w.Spill == nil
}

// Clone creates a copy of the work state.
func (w *Work) Clone() Work {
	floats := make([]*PlacedChild, len(w.Floats))
	copy(floats, w.Floats)
	tags := make([]*Tag, len(w.Tags))
	copy(tags, w.Tags)
	skips := make(map[Location]struct{}, len(w.Skips))
	for k, v := range w.Skips {
		skips[k] = v
	}
	return Work{
		children: w.children,
		index:    w.index,
		Spill:    w.Spill,
		Floats:   floats,
		Tags:     tags,
		Skips:    skips,
	}
}

// Config holds shared flow configuration.
type Config struct {
	Mode FlowMode
	// TODO: Add more configuration fields as needed
}

// PlacedFloat represents a float that has been laid out and is ready for placement.
type PlacedFloat struct {
	Placed *PlacedChild
	Frame  Frame
}

// Composer handles flow composition including floats and footnotes.
type Composer struct {
	Engine *Engine
	Work   *Work
	Config *Config

	// placedFloats holds floats that have been laid out and will be positioned
	// in the output frame. These are tracked separately from queued floats.
	placedFloats []PlacedFloat
}

// Footnotes processes footnotes discovered in a frame.
func (c *Composer) Footnotes(
	regions *Regions,
	frame *Frame,
	flowNeed layout.Abs,
	breakable bool,
	migratable bool,
) error {
	// TODO: Implement footnote handling
	return nil
}

// Float processes a floating placed child.
// It lays out the float and either places it immediately or queues it
// for subsequent regions if it doesn't fit.
func (c *Composer) Float(
	placed *PlacedChild,
	regions *Regions,
	clearance bool,
	migratable bool,
) error {
	// Skip if this float has already been processed (e.g., during relayout).
	if _, ok := c.Work.Skips[placed.Location()]; ok {
		return nil
	}

	// First, process any queued floats that might now fit.
	if err := c.processQueuedFloats(regions); err != nil {
		return err
	}

	// Layout the float at the base size.
	frame, err := placed.Layout(c.Engine, regions.Base())
	if err != nil {
		return err
	}

	// Check if the float fits in the current region.
	// For clearance, we need space after existing content.
	requiredHeight := frame.Height()
	if clearance {
		requiredHeight += placed.Clearance
	}

	// Check if the float fits.
	if regions.Size.Height.Fits(requiredHeight) {
		// Float fits - place it.
		c.placedFloats = append(c.placedFloats, PlacedFloat{
			Placed: placed,
			Frame:  frame,
		})
		c.Work.Skips[placed.Location()] = struct{}{}

		// Reduce available height for inline floats (AlignY == nil).
		// Top/bottom floats don't consume flow space.
		if placed.AlignY == nil {
			regions.Size.Height -= requiredHeight
		}
	} else if migratable && regions.MayProgress() {
		// Float doesn't fit but can be queued for subsequent regions.
		c.Work.Floats = append(c.Work.Floats, placed)
	} else {
		// Float doesn't fit and cannot be queued - place it anyway.
		// This can happen when there's no room but we must proceed.
		c.placedFloats = append(c.placedFloats, PlacedFloat{
			Placed: placed,
			Frame:  frame,
		})
		c.Work.Skips[placed.Location()] = struct{}{}
	}

	return nil
}

// processQueuedFloats attempts to place any queued floats that might now fit.
func (c *Composer) processQueuedFloats(regions *Regions) error {
	if len(c.Work.Floats) == 0 {
		return nil
	}

	// Try to place queued floats.
	remaining := make([]*PlacedChild, 0, len(c.Work.Floats))
	for _, queued := range c.Work.Floats {
		// Skip if already processed.
		if _, ok := c.Work.Skips[queued.Location()]; ok {
			continue
		}

		// Layout the float.
		frame, err := queued.Layout(c.Engine, regions.Base())
		if err != nil {
			return err
		}

		// Check if it fits.
		if regions.Size.Height.Fits(frame.Height()) {
			c.placedFloats = append(c.placedFloats, PlacedFloat{
				Placed: queued,
				Frame:  frame,
			})
			c.Work.Skips[queued.Location()] = struct{}{}

			// Reduce available height for inline floats.
			if queued.AlignY == nil {
				regions.Size.Height -= frame.Height()
			}
		} else {
			// Still doesn't fit - keep in queue.
			remaining = append(remaining, queued)
		}
	}
	c.Work.Floats = remaining

	return nil
}

// PlacedFloats returns the floats that have been placed in this composition.
func (c *Composer) PlacedFloats() []PlacedFloat {
	return c.placedFloats
}

// ClearPlacedFloats clears the list of placed floats.
func (c *Composer) ClearPlacedFloats() {
	c.placedFloats = nil
}

// InsertionWidth returns the width occupied by insertions (floats/footnotes).
// This is used to ensure the output frame is wide enough to accommodate floats.
func (c *Composer) InsertionWidth() layout.Abs {
	var maxWidth layout.Abs
	for _, pf := range c.placedFloats {
		if pf.Frame.Width() > maxWidth {
			maxWidth = pf.Frame.Width()
		}
	}
	return maxWidth
}
