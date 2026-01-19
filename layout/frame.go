package layout

// Frame represents a laid-out frame containing positioned content.
// A frame is the fundamental unit of layout output.
type Frame struct {
	size   Size
	items  []FrameItem
	kind   FrameKind
	fill   *Paint
	stroke *Stroke
}

// NewFrame creates a new empty frame with the given size.
func NewFrame(size Size) *Frame {
	return &Frame{
		size: size,
		kind: FrameKindSoft,
	}
}

// Size returns the frame's dimensions.
func (f *Frame) Size() Size {
	return f.size
}

// Width returns the frame's width.
func (f *Frame) Width() Abs {
	return f.size.Width
}

// Height returns the frame's height.
func (f *Frame) Height() Abs {
	return f.size.Height
}

// SetSize sets the frame's size.
func (f *Frame) SetSize(size Size) {
	f.size = size
}

// Items returns the frame's items.
func (f *Frame) Items() []FrameItem {
	return f.items
}

// Push adds an item at a position.
func (f *Frame) Push(pos Point, item FrameItem) {
	f.items = append(f.items, PositionedItem{Position: pos, Item: item})
}

// PushFrame pushes a nested frame at a position.
func (f *Frame) PushFrame(pos Point, child *Frame) {
	f.Push(pos, GroupItem{Frame: child})
}

// IsEmpty returns true if the frame has no items.
func (f *Frame) IsEmpty() bool {
	return len(f.items) == 0
}

// Kind returns the frame's kind.
func (f *Frame) Kind() FrameKind {
	return f.kind
}

// SetKind sets the frame's kind.
func (f *Frame) SetKind(kind FrameKind) {
	f.kind = kind
}

// MakeKind sets the kind if it's currently soft.
func (f *Frame) MakeKind(kind FrameKind) {
	if f.kind == FrameKindSoft {
		f.kind = kind
	}
}

// Fill returns the frame's background fill.
func (f *Frame) Fill() *Paint {
	return f.fill
}

// SetFill sets the frame's background fill.
func (f *Frame) SetFill(fill *Paint) {
	f.fill = fill
}

// Stroke returns the frame's border stroke.
func (f *Frame) Stroke() *Stroke {
	return f.stroke
}

// SetStroke sets the frame's border stroke.
func (f *Frame) SetStroke(stroke *Stroke) {
	f.stroke = stroke
}

// Translate moves all items by an offset.
func (f *Frame) Translate(offset Point) {
	for i, item := range f.items {
		if pos, ok := item.(PositionedItem); ok {
			pos.Position.X += offset.X
			pos.Position.Y += offset.Y
			f.items[i] = pos
		}
	}
}

// FrameKind indicates whether a frame creates a hard boundary.
type FrameKind int

const (
	// FrameKindSoft means the frame does not create a boundary for gradients.
	FrameKindSoft FrameKind = iota
	// FrameKindHard means the frame creates a hard boundary for gradients.
	FrameKindHard
)

// FrameItem is the interface for items in a frame.
type FrameItem interface {
	isFrameItem()
}

// PositionedItem wraps an item with its position.
type PositionedItem struct {
	Position Point
	Item     FrameItem
}

func (PositionedItem) isFrameItem() {}

// GroupItem represents a nested frame.
type GroupItem struct {
	Frame     *Frame
	Transform *Transform
	Clips     []Clip
}

func (GroupItem) isFrameItem() {}

// TextItem represents shaped text.
type TextItem struct {
	Glyphs []Glyph
	// Additional text properties would go here
}

func (TextItem) isFrameItem() {}

// Glyph represents a single shaped glyph.
type Glyph struct {
	ID       uint16  // Glyph ID in the font
	XAdvance Em      // Horizontal advance
	XOffset  Em      // Horizontal offset
	YOffset  Em      // Vertical offset
	Cluster  int     // Character cluster index
}

// ShapeItem represents a geometric shape.
type ShapeItem struct {
	Shape Shape
	Fill  *Paint
	Stroke *Stroke
}

func (ShapeItem) isFrameItem() {}

// Shape is the interface for geometric shapes.
type Shape interface {
	isShape()
}

// RectShape represents a rectangle.
type RectShape struct {
	Size   Size
	Radius Corners[Abs]
}

func (RectShape) isShape() {}

// PathShape represents a path.
type PathShape struct {
	// Path data would go here
}

func (PathShape) isShape() {}

// Clip represents a clipping region.
type Clip struct {
	Shape Shape
}

// ImageItem represents an image.
type ImageItem struct {
	// Image data would go here
	Size Size
}

func (ImageItem) isFrameItem() {}

// LinkItem represents a hyperlink region.
type LinkItem struct {
	Dest string
	Size Size
}

func (LinkItem) isFrameItem() {}

// TagItem represents a metadata tag.
type TagItem struct {
	Tag Tag
}

func (TagItem) isFrameItem() {}

// Tag represents a metadata tag for introspection.
type Tag struct {
	// Tag data
}

// Fragment represents a sequence of frames that may span multiple pages/regions.
type Fragment struct {
	frames []*Frame
}

// NewFragment creates a new empty fragment.
func NewFragment() *Fragment {
	return &Fragment{}
}

// NewFragmentWithCapacity creates a fragment with pre-allocated capacity.
func NewFragmentWithCapacity(cap int) *Fragment {
	return &Fragment{frames: make([]*Frame, 0, cap)}
}

// Frames returns the frames in the fragment.
func (f *Fragment) Frames() []*Frame {
	return f.frames
}

// Len returns the number of frames.
func (f *Fragment) Len() int {
	return len(f.frames)
}

// IsEmpty returns true if the fragment has no frames.
func (f *Fragment) IsEmpty() bool {
	return len(f.frames) == 0
}

// First returns the first frame, or nil if empty.
func (f *Fragment) First() *Frame {
	if len(f.frames) == 0 {
		return nil
	}
	return f.frames[0]
}

// Last returns the last frame, or nil if empty.
func (f *Fragment) Last() *Frame {
	if len(f.frames) == 0 {
		return nil
	}
	return f.frames[len(f.frames)-1]
}

// Push adds a frame to the fragment.
func (f *Fragment) Push(frame *Frame) {
	f.frames = append(f.frames, frame)
}

// IntoFrame converts a single-frame fragment into a frame.
// Panics if the fragment has more than one frame.
func (f *Fragment) IntoFrame() *Frame {
	if len(f.frames) != 1 {
		panic("fragment must have exactly one frame")
	}
	return f.frames[0]
}

// Paint represents a fill paint (color, gradient, pattern).
type Paint struct {
	// For now just a solid color
	Color Color
}

// Color represents an RGBA color.
type Color struct {
	R, G, B, A uint8
}

// Stroke represents stroke properties for shapes.
type Stroke struct {
	Paint     Paint
	Thickness Abs
	LineCap   LineCap
	LineJoin  LineJoin
	DashArray []Abs
	DashPhase Abs
}

// LineCap specifies how line ends are rendered.
type LineCap int

const (
	LineCapButt LineCap = iota
	LineCapRound
	LineCapSquare
)

// LineJoin specifies how line joins are rendered.
type LineJoin int

const (
	LineJoinMiter LineJoin = iota
	LineJoinRound
	LineJoinBevel
)
