package pages

import (
	"github.com/boergens/gotypst/layout"
)

// PagedDocument represents a fully laid out document.
type PagedDocument struct {
	// Pages contains the laid out pages.
	Pages []Page
	// Info contains document metadata.
	Info DocumentInfo
}

// DocumentInfo contains document metadata.
type DocumentInfo struct {
	Title    *string
	Author   []string
	Keywords []string
	Date     *Date
}

// Date represents a date value.
type Date struct {
	Year  int
	Month int
	Day   int
}

// Page represents a single laid out page.
type Page struct {
	// Frame contains the page content.
	Frame Frame
	// Fill is the page background fill.
	Fill *Paint
	// Numbering is the page numbering pattern.
	Numbering *Numbering
	// Supplement is the page supplement content.
	Supplement Content
	// Number is the logical page number.
	Number int
}

// Frame represents a laid out frame of content.
type Frame struct {
	// Size is the frame dimensions.
	Size layout.Size
	// Items contains the positioned frame content.
	Items []PositionedItem
}

// Hard creates a frame with hard constraints (non-expandable).
func Hard(size layout.Size) Frame {
	return Frame{Size: size, Items: nil}
}

// Width returns the frame width.
func (f *Frame) Width() layout.Abs {
	return f.Size.Width
}

// Height returns the frame height.
func (f *Frame) Height() layout.Abs {
	return f.Size.Height
}

// Push adds an item to the frame at the given position.
func (f *Frame) Push(pos layout.Point, item FrameItem) {
	f.Items = append(f.Items, PositionedItem{Pos: pos, Item: item})
}

// PushFrame adds a nested frame at the given position.
func (f *Frame) PushFrame(pos layout.Point, frame Frame) {
	f.Items = append(f.Items, PositionedItem{Pos: pos, Item: GroupItem{Frame: frame}})
}

// PushMultiple adds multiple items at specified positions.
func (f *Frame) PushMultiple(items []PositionedItem) {
	f.Items = append(f.Items, items...)
}

// PositionedItem represents an item with its position in the frame.
type PositionedItem struct {
	Pos  layout.Point
	Item FrameItem
}

// FrameItem represents an item within a frame.
type FrameItem interface {
	isFrameItem()
}

// GroupItem represents a nested frame.
type GroupItem struct {
	Frame Frame
}

func (GroupItem) isFrameItem() {}

// TagItem represents an introspection tag.
type TagItem struct {
	Tag Tag
}

func (TagItem) isFrameItem() {}

// Tag represents an introspection tag.
type Tag struct {
	Kind     TagKind
	Location Location
}

// TagKind indicates whether a tag is a start or end tag.
type TagKind int

const (
	TagStart TagKind = iota
	TagEnd
)

// Location identifies an element location for introspection.
type Location uint64

// Paint represents a fill color or pattern.
type Paint struct {
	// Color is a solid color fill.
	Color *Color
}

// Color represents an RGBA color.
type Color struct {
	R, G, B, A uint8
}

// Numbering represents a page numbering pattern.
type Numbering struct {
	// Pattern is the numbering pattern string.
	Pattern string
}

// Content represents document content.
type Content struct {
	Elements []ContentElement
}

// ContentElement is a marker interface for content elements.
type ContentElement interface {
	isContentElement()
}

// Sides represents values for all four sides.
type Sides[T any] struct {
	Left, Top, Right, Bottom T
}

// SumByAxis returns the sum of sides along each axis.
func (s Sides[T]) SumByAxis() layout.Size {
	// Type assertion for Abs
	if left, ok := any(s.Left).(layout.Abs); ok {
		right := any(s.Right).(layout.Abs)
		top := any(s.Top).(layout.Abs)
		bottom := any(s.Bottom).(layout.Abs)
		return layout.Size{
			Width:  left + right,
			Height: top + bottom,
		}
	}
	return layout.Size{}
}

// Binding represents the page binding side.
type Binding int

const (
	BindingLeft Binding = iota
	BindingRight
)

// Swap returns true if margins should be swapped for this page number.
func (b Binding) Swap(pageNum int) bool {
	// For left-bound pages, swap on even pages (0-indexed)
	// For right-bound pages, swap on odd pages
	if b == BindingLeft {
		return pageNum%2 == 1 // even page (1-indexed)
	}
	return pageNum%2 == 0 // odd page (1-indexed)
}

// Parity represents desired page number parity.
type Parity int

const (
	ParityEven Parity = iota
	ParityOdd
)

// Matches returns true if the page count matches the desired parity.
func (p Parity) Matches(pageCount int) bool {
	isEven := pageCount%2 == 0
	if p == ParityEven {
		return !isEven // need to add page if count is odd
	}
	return isEven // need to add page if count is even
}

// LayoutedPage represents a mostly-finished page layout.
// It needs only knowledge of its exact page number to be finalized.
type LayoutedPage struct {
	// Inner is the main content frame.
	Inner Frame
	// Margin is the page margins.
	Margin Sides[layout.Abs]
	// Binding is the binding side.
	Binding Binding
	// TwoSided indicates if the document is two-sided.
	TwoSided bool
	// Header is the optional header frame.
	Header *Frame
	// Footer is the optional footer frame.
	Footer *Frame
	// Background is the optional background frame.
	Background *Frame
	// Foreground is the optional foreground frame.
	Foreground *Frame
	// Fill is the page fill.
	Fill *Paint
	// Numbering is the page numbering pattern.
	Numbering *Numbering
	// Supplement is the page supplement content.
	Supplement Content
}

// StyleChain represents a chain of styles.
type StyleChain struct {
	// Styles contains the style values.
	Styles map[string]interface{}
}

// Get retrieves a style value.
func (s StyleChain) Get(key string) interface{} {
	if s.Styles == nil {
		return nil
	}
	return s.Styles[key]
}

// Locator tracks element locations for introspection.
type Locator struct {
	// Current is the current location counter.
	Current uint64
}

// Split creates a SplitLocator from this locator.
func (l *Locator) Split() *SplitLocator {
	return &SplitLocator{base: l}
}

// SplitLocator allows splitting locations for parallel layout.
type SplitLocator struct {
	base    *Locator
	counter uint64
}

// Next returns a new locator for the next element.
func (s *SplitLocator) Next(span interface{}) Locator {
	s.counter++
	return Locator{Current: s.base.Current + s.counter}
}

// Relayout returns a locator for relayout purposes.
func (s *SplitLocator) Relayout() Locator {
	return Locator{Current: s.base.Current + s.counter}
}

// Pair represents a content element with its style chain.
type Pair struct {
	Element ContentElement
	Styles  StyleChain
}

// ManualPageCounter tracks manual page counter updates.
type ManualPageCounter struct {
	physical int
	logical  int
}

// NewManualPageCounter creates a new page counter.
func NewManualPageCounter() *ManualPageCounter {
	return &ManualPageCounter{physical: 0, logical: 1}
}

// Physical returns the physical page number (0-indexed).
func (c *ManualPageCounter) Physical() int {
	return c.physical
}

// Logical returns the logical page number.
func (c *ManualPageCounter) Logical() int {
	return c.logical
}

// Step advances the counter to the next page.
func (c *ManualPageCounter) Step() {
	c.physical++
	c.logical++
}

// Visit processes counter updates from the frame.
func (c *ManualPageCounter) Visit(frame *Frame) error {
	// TODO: Process counter updates within the frame
	return nil
}
