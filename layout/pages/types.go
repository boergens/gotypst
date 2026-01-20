package pages

import (
	"github.com/boergens/gotypst/eval"
	"github.com/boergens/gotypst/layout"
)

// ContentElement is a type alias for eval.ContentElement.
type ContentElement = eval.ContentElement

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

// ImageItem represents an embedded image.
type ImageItem struct {
	// Image contains the image data and metadata.
	Image Image
	// Size is the rendered size of the image.
	Size layout.Size
}

func (ImageItem) isFrameItem() {}

// TextItem represents shaped text in a frame.
type TextItem struct {
	// Glyphs contains the shaped glyph data.
	Glyphs []Glyph
	// Size is the bounding box size.
	Size layout.Size
}

func (TextItem) isFrameItem() {}

// Glyph represents a positioned glyph in a text run.
type Glyph struct {
	// ID is the glyph index in the font.
	ID uint16
	// XAdvance is the horizontal advance.
	XAdvance layout.Abs
	// XOffset is the horizontal offset.
	XOffset layout.Abs
	// YOffset is the vertical offset.
	YOffset layout.Abs
	// Range is the byte range in the source text.
	Range [2]int
}

// Image represents image data for embedding.
type Image struct {
	// Data is the raw image bytes.
	Data []byte
	// Format specifies the image format.
	Format ImageFormat
	// Width is the natural image width in pixels.
	Width int
	// Height is the natural image height in pixels.
	Height int
	// BitsPerComponent is typically 8.
	BitsPerComponent int
	// ColorSpace specifies the color space (e.g., DeviceRGB, DeviceGray).
	ColorSpace ColorSpace
	// Alpha contains optional alpha channel data.
	Alpha []byte
}

// ImageFormat specifies the image encoding format.
type ImageFormat int

const (
	// ImageFormatJPEG represents JPEG/DCTDecode encoded images.
	ImageFormatJPEG ImageFormat = iota
	// ImageFormatPNG represents PNG/FlateDecode encoded images.
	ImageFormatPNG
	// ImageFormatRaw represents uncompressed raw pixel data.
	ImageFormatRaw
)

// ColorSpace specifies the PDF color space.
type ColorSpace int

const (
	// ColorSpaceDeviceRGB represents the RGB color space.
	ColorSpaceDeviceRGB ColorSpace = iota
	// ColorSpaceDeviceGray represents the grayscale color space.
	ColorSpaceDeviceGray
	// ColorSpaceDeviceCMYK represents the CMYK color space.
	ColorSpaceDeviceCMYK
)

// String returns the PDF name for the color space.
func (cs ColorSpace) String() string {
	switch cs {
	case ColorSpaceDeviceRGB:
		return "DeviceRGB"
	case ColorSpaceDeviceGray:
		return "DeviceGray"
	case ColorSpaceDeviceCMYK:
		return "DeviceCMYK"
	default:
		return "DeviceRGB"
	}
}

// Tag represents an introspection tag.
type Tag struct {
	Kind     TagKind
	Location Location
	// Elem optionally holds element data for start tags.
	// This may contain a CounterUpdateElem for page counter updates.
	Elem TagElement
}

// TagElement is a marker interface for elements that can be embedded in tags.
type TagElement interface {
	isTagElement()
}

// TagKind indicates whether a tag is a start or end tag.
type TagKind int

const (
	TagStart TagKind = iota
	TagEnd
)

// CounterKey identifies which counter is being updated.
type CounterKey int

const (
	// CounterKeyPage is the page counter.
	CounterKeyPage CounterKey = iota
	// CounterKeyFigure is the figure counter.
	CounterKeyFigure
	// CounterKeyTable is the table counter.
	CounterKeyTable
	// CounterKeyEquation is the equation counter.
	CounterKeyEquation
	// CounterKeyHeading is the heading counter.
	CounterKeyHeading
)

// CounterUpdate represents an update operation on a counter.
type CounterUpdate interface {
	isCounterUpdate()
	// Apply applies the update to a counter state and returns the new value.
	Apply(current int) int
}

// CounterUpdateSet sets the counter to a specific value.
type CounterUpdateSet struct {
	Value int
}

func (CounterUpdateSet) isCounterUpdate() {}

// Apply sets the counter to the specified value.
func (u CounterUpdateSet) Apply(_ int) int {
	return u.Value
}

// CounterUpdateStep increments the counter by one.
type CounterUpdateStep struct{}

func (CounterUpdateStep) isCounterUpdate() {}

// Apply increments the counter by one.
func (CounterUpdateStep) Apply(current int) int {
	return current + 1
}

// CounterUpdateElem represents a counter update element embedded in a tag.
type CounterUpdateElem struct {
	// Key identifies which counter to update.
	Key CounterKey
	// Update is the update operation to apply.
	Update CounterUpdate
}

func (CounterUpdateElem) isTagElement() {}

// CounterUpdateItem represents a counter update in a frame.
// This is used for running content like page numbering updates.
type CounterUpdateItem struct {
	// Counter identifies which counter to update ("page", "heading", etc.).
	Counter string
	// UpdateKind specifies the type of update (step, update to value, etc.).
	UpdateKind CounterUpdateKind
	// Value is the new value for absolute updates (used with CounterUpdateKindSet).
	Value int
}

func (CounterUpdateItem) isFrameItem() {}

// CounterUpdateKind specifies the type of counter update.
type CounterUpdateKind int

const (
	// CounterUpdateKindStep advances the counter by 1.
	CounterUpdateKindStep CounterUpdateKind = iota
	// CounterUpdateKindSet sets the counter to an absolute value.
	CounterUpdateKindSet
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
	Elements []eval.ContentElement
}

// TextElem represents text content.
type TextElem struct {
	// Text is the text string.
	Text string
}

func (*TextElem) IsContentElement() {}

// SpaceElem represents horizontal spacing.
type SpaceElem struct {
	// Width is the spacing amount.
	Width layout.Abs
}

func (*SpaceElem) IsContentElement() {}

// AlignElem represents alignment for content.
type AlignElem struct {
	// Align specifies the horizontal alignment (start, center, end).
	Align layout.Alignment
	// Body is the content to align.
	Body Content
}

func (*AlignElem) IsContentElement() {}

// NumberingElem represents a page number that will be formatted at layout time.
// This is used for running content in headers/footers.
type NumberingElem struct {
	// Pattern is the numbering pattern (e.g., "1", "i", "I", "a", "A").
	Pattern string
}

func (*NumberingElem) IsContentElement() {}

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
	// HeaderContent is the optional header content (laid out at finalization).
	HeaderContent *Content
	// FooterContent is the optional footer content (laid out at finalization).
	FooterContent *Content
	// HeaderSize is the available size for the header.
	HeaderSize layout.Size
	// FooterSize is the available size for the footer.
	FooterSize layout.Size
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
// Element can be any content element type (from eval or pages package).
type Pair struct {
	Element interface{}
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
// It recursively walks through all frame items looking for tags that
// contain page counter updates or CounterUpdateItem elements, and applies
// those updates to the logical page counter.
func (c *ManualPageCounter) Visit(frame *Frame) error {
	if frame == nil {
		return nil
	}

	for _, positioned := range frame.Items {
		switch item := positioned.Item.(type) {
		case GroupItem:
			// Recursively visit nested frames
			if err := c.Visit(&item.Frame); err != nil {
				return err
			}
		case TagItem:
			// Check for counter updates in start tags
			if item.Tag.Kind != TagStart {
				continue
			}
			elem, ok := item.Tag.Elem.(*CounterUpdateElem)
			if !ok || elem == nil {
				continue
			}
			// Only process page counter updates
			if elem.Key != CounterKeyPage {
				continue
			}
			// Apply the update to the logical counter
			c.logical = elem.Update.Apply(c.logical)
		case CounterUpdateItem:
			// Handle direct counter update items (for running content)
			if item.Counter == "page" {
				switch item.UpdateKind {
				case CounterUpdateKindStep:
					c.logical++
				case CounterUpdateKindSet:
					c.logical = item.Value
				}
			}
		}
	}
	return nil
}
