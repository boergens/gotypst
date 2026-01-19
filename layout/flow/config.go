package flow

import (
	"github.com/boergens/gotypst/layout"
)

// Config holds configuration for flow layout.
type Config struct {
	// Mode indicates the flow layout context.
	Mode FlowMode
	// Columns holds column layout configuration.
	Columns ColumnConfig
	// Footnote holds footnote layout configuration.
	Footnote FootnoteConfig
	// LineNumbers holds line number configuration (nil if disabled).
	LineNumbers *LineNumberConfig
	// Shared holds shared layout configuration.
	Shared *SharedConfig
}

// ColumnConfig holds multi-column layout configuration.
type ColumnConfig struct {
	// Count is the number of columns.
	Count int
	// Width is the width of each column.
	Width layout.Abs
	// Gutter is the space between columns.
	Gutter layout.Abs
	// Dir is the column direction (LTR or RTL).
	Dir layout.Dir
}

// FootnoteConfig holds footnote layout configuration.
type FootnoteConfig struct {
	// Separator is the content to place between body and footnotes.
	Separator interface{} // Content type
	// Clearance is the minimum space before the separator.
	Clearance layout.Abs
	// Gap is the space between footnotes.
	Gap layout.Abs
	// Expand indicates if footnotes expand to fill width.
	Expand bool
}

// LineNumberConfig holds line number configuration.
type LineNumberConfig struct {
	// Scope indicates when line numbers reset.
	Scope LineNumberingScope
	// DefaultClearance is the default space between text and numbers.
	DefaultClearance layout.Abs
}

// SharedConfig holds configuration shared across layout operations.
type SharedConfig struct {
	// Styles is the current style chain.
	Styles interface{} // StyleChain type
	// Dir is the text direction.
	Dir layout.Dir
}

// Work holds the state for flow layout processing.
type Work struct {
	// Children contains unprocessed child elements.
	Children []Child
	// Spillover contains children that spilled from breakable blocks.
	Spillover []Child
	// Floats contains queued floating elements.
	Floats []*PlacedChild
	// Footnotes contains pending footnote elements.
	Footnotes []interface{} // Packed<FootnoteElem>
	// Skips tracks processed locations to avoid duplicates.
	Skips map[Location]struct{}
	// Tags contains collected tags.
	Tags []Tag
	// FootnoteSpill contains spillover footnote frames.
	FootnoteSpill []*Frame
}

// Clone creates a copy of the work state.
func (w *Work) Clone() Work {
	children := make([]Child, len(w.Children))
	copy(children, w.Children)
	spillover := make([]Child, len(w.Spillover))
	copy(spillover, w.Spillover)
	floats := make([]*PlacedChild, len(w.Floats))
	copy(floats, w.Floats)
	footnotes := make([]interface{}, len(w.Footnotes))
	copy(footnotes, w.Footnotes)
	skips := make(map[Location]struct{}, len(w.Skips))
	for k, v := range w.Skips {
		skips[k] = v
	}
	tags := make([]Tag, len(w.Tags))
	copy(tags, w.Tags)
	footnoteSpill := make([]*Frame, len(w.FootnoteSpill))
	copy(footnoteSpill, w.FootnoteSpill)

	return Work{
		Children:      children,
		Spillover:     spillover,
		Floats:        floats,
		Footnotes:     footnotes,
		Skips:         skips,
		Tags:          tags,
		FootnoteSpill: footnoteSpill,
	}
}

// ExtendSkips adds locations to the skip set.
func (w *Work) ExtendSkips(locs []Location) {
	for _, loc := range locs {
		w.Skips[loc] = struct{}{}
	}
}

// Child represents a child element in flow layout.
type Child interface {
	isChild()
}

// LineChild represents a line of text with spacing.
type LineChild struct {
	Frame   *Frame
	Leading layout.Abs
	Spacing layout.Abs
}

func (LineChild) isChild() {}

// SingleChild represents a single non-breakable block.
type SingleChild struct {
	Frame  *Frame
	Fr     layout.Fr
	Sticky bool
}

func (SingleChild) isChild() {}

// MultiChild represents a breakable block that may span regions.
type MultiChild struct {
	Layouter interface{} // Block layouter
	Sticky   bool
}

func (MultiChild) isChild() {}

// PlacedChild represents a floating element.
type PlacedChild struct {
	// Scope is where the float should be placed.
	Scope PlacementScope
	// AlignX is the horizontal alignment.
	AlignX FixedAlignment
	// AlignY is the optional vertical alignment (nil means auto).
	AlignY *FixedAlignment
	// Delta is the offset from aligned position.
	Delta RelAxes
	// Clearance is the minimum space around the float.
	Clearance layout.Abs
	// Content is the content to layout.
	Content interface{}
	// location caches the element location.
	location *Location
}

// Location returns the location of this placed child.
func (p *PlacedChild) Location() Location {
	if p.location == nil {
		// Generate a location based on content
		return Location{}
	}
	return *p.location
}

// Layout lays out the placed child content.
func (p *PlacedChild) Layout(engine interface{}, base layout.Size) (*Frame, error) {
	// Placeholder implementation
	return NewSoftFrame(base), nil
}

// SpacingChild represents explicit spacing between elements.
type SpacingChild struct {
	Amount layout.Abs
	Weak   bool
}

func (SpacingChild) isChild() {}

// ParLineMarker represents a paragraph line marker for line numbering.
type ParLineMarker struct {
	Numbering      Numbering
	NumberMargin   OuterHAlignment
	NumberClearance Smart[Rel]
	NumberAlign    *FixedAlignment
}
