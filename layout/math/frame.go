package math

// Frame represents a laid-out frame of content.
type Frame struct {
	size     Size
	baseline *Abs
	items    []FrameItem
	hard     bool
}

// NewSoftFrame creates a new soft frame with the given size.
// Soft frames can be resized.
func NewSoftFrame(size Size) *Frame {
	return &Frame{size: size, hard: false}
}

// NewHardFrame creates a new hard frame with the given size.
// Hard frames cannot be resized.
func NewHardFrame(size Size) *Frame {
	return &Frame{size: size, hard: true}
}

// Size returns the frame's dimensions.
func (f *Frame) Size() Size {
	return f.size
}

// Width returns the frame's width.
func (f *Frame) Width() Abs {
	return f.size.X
}

// Height returns the frame's height.
func (f *Frame) Height() Abs {
	return f.size.Y
}

// SetSize sets the frame's size.
func (f *Frame) SetSize(size Size) {
	f.size = size
}

// SizeMut returns a mutable reference to the frame's size.
func (f *Frame) SizeMut() *Size {
	return &f.size
}

// Baseline returns the frame's baseline position from the top.
func (f *Frame) Baseline() Abs {
	if f.baseline != nil {
		return *f.baseline
	}
	return f.size.Y
}

// SetBaseline sets the baseline position.
func (f *Frame) SetBaseline(baseline Abs) {
	f.baseline = &baseline
}

// HasBaseline returns whether a baseline has been explicitly set.
func (f *Frame) HasBaseline() bool {
	return f.baseline != nil
}

// Ascent returns the distance from baseline to top.
func (f *Frame) Ascent() Abs {
	return f.Baseline()
}

// Descent returns the distance from baseline to bottom.
func (f *Frame) Descent() Abs {
	return f.size.Y - f.Baseline()
}

// Items returns the frame's items.
func (f *Frame) Items() []FrameItem {
	return f.items
}

// Push adds an item to the frame at the given position.
func (f *Frame) Push(pos Point, item FrameItem) {
	f.items = append(f.items, &PositionedItem{Position: pos, Item: item})
}

// PushFrame adds a child frame at the given position.
func (f *Frame) PushFrame(pos Point, child *Frame) {
	f.Push(pos, &GroupItem{Frame: child})
}

// PushText adds a text item at the given position.
func (f *Frame) PushText(pos Point, text *TextItem) {
	f.Push(pos, &TextFrameItem{Text: text})
}

// PushShape adds a shape item at the given position.
func (f *Frame) PushShape(pos Point, shape *Shape, span Span) {
	f.Push(pos, &ShapeFrameItem{Shape: shape, Span: span})
}

// PushTag adds a tag at the given position.
func (f *Frame) PushTag(pos Point, tag Tag) {
	f.Push(pos, &TagFrameItem{Tag: tag})
}

// PushMultiple adds multiple positioned items.
func (f *Frame) PushMultiple(items []struct {
	Pos  Point
	Item FrameItem
}) {
	for _, item := range items {
		f.Push(item.Pos, item.Item)
	}
}

// Translate moves all items by the given offset.
func (f *Frame) Translate(offset Point) {
	for i, item := range f.items {
		if pi, ok := item.(*PositionedItem); ok {
			f.items[i] = &PositionedItem{
				Position: Point{
					X: pi.Position.X + offset.X,
					Y: pi.Position.Y + offset.Y,
				},
				Item: pi.Item,
			}
		}
	}
}

// Transform applies a transformation to the frame.
func (f *Frame) Transform(transform Transform) {
	f.items = append(f.items, &TransformItem{Transform: transform})
}

// Resize resizes the frame and repositions content according to alignment.
// Returns the offset applied.
func (f *Frame) Resize(newSize Size, align Axes[FixedAlignment]) Point {
	oldSize := f.size
	f.size = newSize

	offsetX := align.X.Position(newSize.X - oldSize.X)
	offsetY := align.Y.Position(newSize.Y - oldSize.Y)
	offset := Point{X: offsetX, Y: offsetY}

	if offset.X != 0 || offset.Y != 0 {
		f.Translate(offset)
	}

	return offset
}

// IsEmpty returns whether the frame has no items.
func (f *Frame) IsEmpty() bool {
	return len(f.items) == 0
}

// Axes holds values for both X and Y axes.
type Axes[T any] struct {
	X, Y T
}

// NewAxes creates a new Axes with the given values.
func NewAxes[T any](x, y T) Axes[T] {
	return Axes[T]{X: x, Y: y}
}

// Splat creates Axes with the same value for both axes.
func Splat[T any](v T) Axes[T] {
	return Axes[T]{X: v, Y: v}
}

// FrameItem represents an item within a frame.
type FrameItem interface {
	isFrameItem()
}

// PositionedItem wraps an item with its position.
type PositionedItem struct {
	Position Point
	Item     FrameItem
}

func (*PositionedItem) isFrameItem() {}

// GroupItem represents a nested frame.
type GroupItem struct {
	Frame *Frame
}

func (*GroupItem) isFrameItem() {}

// TextFrameItem represents text within a frame.
type TextFrameItem struct {
	Text *TextItem
}

func (*TextFrameItem) isFrameItem() {}

// ShapeFrameItem represents a shape within a frame.
type ShapeFrameItem struct {
	Shape *Shape
	Span  Span
}

func (*ShapeFrameItem) isFrameItem() {}

// TagFrameItem represents a tag within a frame.
type TagFrameItem struct {
	Tag Tag
}

func (*TagFrameItem) isFrameItem() {}

// TransformItem represents a transformation applied to subsequent items.
type TransformItem struct {
	Transform Transform
}

func (*TransformItem) isFrameItem() {}

// Fragment represents a multi-frame layout result.
type Fragment struct {
	Frames []*Frame
}

// NewFragment creates a new fragment from frames.
func NewFragment(frames []*Frame) *Fragment {
	return &Fragment{Frames: frames}
}

// IntoFrame converts a single-frame fragment into a frame.
// For multi-frame fragments, this combines them.
func (f *Fragment) IntoFrame() *Frame {
	if len(f.Frames) == 0 {
		return NewSoftFrame(Size{})
	}
	if len(f.Frames) == 1 {
		return f.Frames[0]
	}

	// Combine multiple frames vertically
	var totalHeight Abs
	var maxWidth Abs
	for _, frame := range f.Frames {
		totalHeight += frame.Height()
		if frame.Width() > maxWidth {
			maxWidth = frame.Width()
		}
	}

	result := NewSoftFrame(Size{X: maxWidth, Y: totalHeight})
	var y Abs
	for _, frame := range f.Frames {
		result.PushFrame(Point{Y: y}, frame)
		y += frame.Height()
	}
	return result
}

// InlineItem represents an item in inline layout.
type InlineItem interface {
	isInlineItem()
}

// InlineFrame is a frame in inline layout.
type InlineFrame struct {
	Frame *Frame
}

func (*InlineFrame) isInlineItem() {}

// InlineSpace is space in inline layout.
type InlineSpace struct {
	Width    Abs
	Flexible bool
}

func (*InlineSpace) isInlineItem() {}

// Region represents a layout region.
type Region struct {
	Size   Size
	Expand Axes[bool]
}

// NewRegion creates a new region.
func NewRegion(size Size, expand Axes[bool]) Region {
	return Region{Size: size, Expand: expand}
}

// Regions represents multiple layout regions.
type Regions struct {
	size     Size
	current  int
	backlog  []Size
	lastSize Size
}

// Base returns the base size for the regions.
func (r *Regions) Base() Size {
	return r.size
}

// Next advances to the next region.
func (r *Regions) Next() {
	if r.current < len(r.backlog) {
		r.current++
	}
}

// MayProgress returns whether more regions are available.
func (r *Regions) MayProgress() bool {
	return r.current < len(r.backlog)
}

// MayBreak returns whether breaking is allowed.
func (r *Regions) MayBreak() bool {
	return true
}

// Sides represents values for all four sides.
type Sides[T any] struct {
	Left, Top, Right, Bottom T
}

// NewSides creates new Sides with the given values.
func NewSides[T any](left, top, right, bottom T) Sides[T] {
	return Sides[T]{Left: left, Top: top, Right: right, Bottom: bottom}
}

// Corners represents values for all four corners.
type Corners[T any] struct {
	TopLeft, TopRight, BottomRight, BottomLeft T
}

// SplatCorners creates Corners with the same value for all.
func SplatCorners[T any](v T) Corners[T] {
	return Corners[T]{TopLeft: v, TopRight: v, BottomRight: v, BottomLeft: v}
}
