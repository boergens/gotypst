package flow

import (
	"github.com/boergens/gotypst/layout"
)

// Frame represents a laid out piece of content with positioned items.
type Frame struct {
	// Size is the dimensions of the frame.
	Size layout.Size
	// Baseline is the vertical position of the text baseline.
	Baseline layout.Abs
	// Items contains the positioned items in the frame.
	Items []FrameItem
	// Kind indicates whether the frame is hard or soft.
	Kind FrameKind
	// Parent optionally links to a parent location.
	Parent *FrameParent
}

// FrameKind distinguishes hard and soft frames.
type FrameKind int

const (
	// FrameSoft is a frame that can be merged with adjacent frames.
	FrameSoft FrameKind = iota
	// FrameHard is a frame that must remain separate (e.g., page boundaries).
	FrameHard
)

// NewSoftFrame creates a new soft frame with the given size.
func NewSoftFrame(size layout.Size) *Frame {
	return &Frame{
		Size:  size,
		Kind:  FrameSoft,
		Items: nil,
	}
}

// NewHardFrame creates a new hard frame with the given size.
func NewHardFrame(size layout.Size) *Frame {
	return &Frame{
		Size:  size,
		Kind:  FrameHard,
		Items: nil,
	}
}

// Width returns the frame's width.
func (f *Frame) Width() layout.Abs {
	return f.Size.Width
}

// Height returns the frame's height.
func (f *Frame) Height() layout.Abs {
	return f.Size.Height
}

// SizeMut returns a mutable reference to the size.
func (f *Frame) SizeMut() *layout.Size {
	return &f.Size
}

// IsEmpty returns true if the frame has no items.
func (f *Frame) IsEmpty() bool {
	return len(f.Items) == 0
}

// PushFrame adds a child frame at the given position.
func (f *Frame) PushFrame(pos layout.Point, child *Frame) {
	f.Items = append(f.Items, FrameItem{
		Pos:  pos,
		Item: GroupItem{Frame: child},
	})
}

// Translate moves all items by the given offset.
func (f *Frame) Translate(offset layout.Point) {
	for i := range f.Items {
		f.Items[i].Pos.X += offset.X
		f.Items[i].Pos.Y += offset.Y
	}
}

// SetParent sets the parent location for this frame.
func (f *Frame) SetParent(parent FrameParent) {
	f.Parent = &parent
}

// FrameItem represents a positioned item within a frame.
type FrameItem struct {
	Pos  layout.Point
	Item FrameItemKind
}

// FrameItemKind is the interface for frame item types.
type FrameItemKind interface {
	isFrameItemKind()
}

// GroupItem contains a nested frame.
type GroupItem struct {
	Frame *Frame
}

func (GroupItem) isFrameItemKind() {}

// TagItem represents a tag marker in the frame.
type TagItem struct {
	Tag Tag
}

func (TagItem) isFrameItemKind() {}

// Tag represents a content tag for tracking elements.
type Tag interface {
	isTag()
}

// StartTag marks the start of an element.
type StartTag struct {
	Elem     interface{} // Element reference
	Location Location
}

func (StartTag) isTag() {}

// EndTag marks the end of an element.
type EndTag struct {
	Elem interface{}
}

func (EndTag) isTag() {}

// FrameParent links a frame to its parent location.
type FrameParent struct {
	Location Location
	Inherit  Inherit
}

// NewFrameParent creates a new frame parent reference.
func NewFrameParent(loc Location, inherit Inherit) FrameParent {
	return FrameParent{Location: loc, Inherit: inherit}
}

// Inherit specifies whether styles are inherited.
type Inherit int

const (
	InheritNo Inherit = iota
	InheritYes
)

// Location uniquely identifies a position in the document.
type Location struct {
	hash uint64
}

// Variant creates a location variant for sub-elements.
func (l Location) Variant(n int) Location {
	// Simple hash combination
	return Location{hash: l.hash ^ uint64(n)}
}

// Fragment represents a sequence of frames (multi-region layout result).
type Fragment []*Frame

// IntoFrames converts the fragment into a slice of frames.
func (f Fragment) IntoFrames() []*Frame {
	return f
}

// Region represents a single layout region.
type Region struct {
	// Size is the available size for layout.
	Size layout.Size
	// Expand indicates which axes should expand to fill available space.
	Expand Axes[bool]
}

// NewRegion creates a new region with the given size and expansion flags.
func NewRegion(size layout.Size, expand Axes[bool]) Region {
	return Region{Size: size, Expand: expand}
}

// Regions represents a sequence of regions for multi-region layout.
type Regions struct {
	// Size is the size of the current region.
	Size layout.Size
	// Backlog contains sizes of subsequent regions.
	Backlog []layout.Abs
	// Expand indicates which axes should expand.
	Expand Axes[bool]
	// Full is the full size available (for relative sizing).
	Full layout.Size
	// Last indicates if this is the last region.
	Last bool
}

// Base returns the base size for relative sizing.
func (r *Regions) Base() layout.Size {
	return r.Full
}

// Iter returns an iterator over region sizes.
func (r *Regions) Iter() []layout.Size {
	result := make([]layout.Size, 0, 1+len(r.Backlog))
	result = append(result, r.Size)
	for _, h := range r.Backlog {
		result = append(result, layout.Size{Width: r.Size.Width, Height: h})
	}
	return result
}

// MayProgress returns true if there are more regions available.
func (r *Regions) MayProgress() bool {
	return len(r.Backlog) > 0 || !r.Last
}

// Next advances to the next region.
func (r *Regions) Next() {
	if len(r.Backlog) > 0 {
		r.Size.Height = r.Backlog[0]
		r.Backlog = r.Backlog[1:]
	}
}

// Clone creates a copy of the regions.
func (r *Regions) Clone() Regions {
	backlog := make([]layout.Abs, len(r.Backlog))
	copy(backlog, r.Backlog)
	return Regions{
		Size:    r.Size,
		Backlog: backlog,
		Expand:  r.Expand,
		Full:    r.Full,
		Last:    r.Last,
	}
}

// Axes represents a pair of values for X and Y axes.
type Axes[T any] struct {
	X, Y T
}

// NewAxes creates a new Axes with the given values.
func NewAxes[T any](x, y T) Axes[T] {
	return Axes[T]{X: x, Y: y}
}

// Splat creates an Axes with the same value for both axes.
func Splat[T any](v T) Axes[T] {
	return Axes[T]{X: v, Y: v}
}

// FlowMode indicates the context of flow layout.
type FlowMode int

const (
	// FlowModeRoot is the top-level flow (handles floats/footnotes).
	FlowModeRoot FlowMode = iota
	// FlowModeBlock is a nested block context.
	FlowModeBlock
	// FlowModeInline is an inline context.
	FlowModeInline
)

// FixedAlignment represents a fixed alignment value.
type FixedAlignment int

const (
	// FixedAlignmentStart aligns to the start.
	FixedAlignmentStart FixedAlignment = iota
	// FixedAlignmentCenter aligns to the center.
	FixedAlignmentCenter
	// FixedAlignmentEnd aligns to the end.
	FixedAlignmentEnd
)

// Position calculates the position for alignment within available space.
func (a FixedAlignment) Position(available layout.Abs) layout.Abs {
	switch a {
	case FixedAlignmentStart:
		return 0
	case FixedAlignmentCenter:
		return available / 2
	case FixedAlignmentEnd:
		return available
	}
	return 0
}

// Inv returns the inverse alignment.
func (a FixedAlignment) Inv() FixedAlignment {
	switch a {
	case FixedAlignmentStart:
		return FixedAlignmentEnd
	case FixedAlignmentEnd:
		return FixedAlignmentStart
	}
	return a
}

// PlacementScope indicates where a float should be placed.
type PlacementScope int

const (
	// PlacementScopeColumn places the float in the current column.
	PlacementScopeColumn PlacementScope = iota
	// PlacementScopeParent places the float in the parent (page).
	PlacementScopeParent
)

// Rel represents a relative length (absolute + ratio of parent).
type Rel struct {
	Abs   layout.Abs
	Ratio float64
}

// RelativeTo converts the relative length to absolute at the given base.
func (r Rel) RelativeTo(base layout.Abs) layout.Abs {
	return r.Abs + layout.Abs(r.Ratio*float64(base))
}

// RelAxes is a pair of relative lengths.
type RelAxes struct {
	X, Y Rel
}

// ZipMap applies RelativeTo to both axes.
func (r RelAxes) ZipMap(size layout.Size) layout.Point {
	return layout.Point{
		X: r.X.RelativeTo(size.Width),
		Y: r.Y.RelativeTo(size.Height),
	}
}

// ToPoint converts to a Point.
func (r RelAxes) ToPoint() layout.Point {
	return layout.Point{X: r.X.Abs, Y: r.Y.Abs}
}

// OuterHAlignment represents horizontal alignment at outer margins.
type OuterHAlignment int

const (
	// OuterHAlignmentStart is the start margin.
	OuterHAlignmentStart OuterHAlignment = iota
	// OuterHAlignmentEnd is the end margin.
	OuterHAlignmentEnd
)

// Resolve converts to FixedAlignment based on shared config.
func (a OuterHAlignment) Resolve(shared *SharedConfig) FixedAlignment {
	if a == OuterHAlignmentEnd {
		return FixedAlignmentEnd
	}
	return FixedAlignmentStart
}

// LineNumberingScope indicates when line numbers reset.
type LineNumberingScope int

const (
	// LineNumberingScopePage resets line numbers on each page.
	LineNumberingScopePage LineNumberingScope = iota
	// LineNumberingScopeDocument numbers continuously across pages.
	LineNumberingScopeDocument
)

// Smart represents a value that can be auto or custom.
type Smart[T any] struct {
	auto  bool
	value T
}

// SmartAuto creates an auto Smart value.
func SmartAuto[T any]() Smart[T] {
	return Smart[T]{auto: true}
}

// SmartCustom creates a custom Smart value.
func SmartCustom[T any](v T) Smart[T] {
	return Smart[T]{auto: false, value: v}
}

// IsAuto returns true if this is an auto value.
func (s Smart[T]) IsAuto() bool {
	return s.auto
}

// Get returns the custom value, panics if auto.
func (s Smart[T]) Get() T {
	if s.auto {
		panic("called Get on auto Smart value")
	}
	return s.value
}

// GetOr returns the custom value or the default.
func (s Smart[T]) GetOr(def T) T {
	if s.auto {
		return def
	}
	return s.value
}

// Numbering represents a numbering pattern.
type Numbering struct {
	Pattern string
}

// Clone creates a copy of the numbering.
func (n Numbering) Clone() Numbering {
	return n
}
