package flow

import (
	"sort"

	"github.com/boergens/gotypst/layout"
)

// FoundElement represents an element found in a frame with its Y position.
type FoundElement[T any] struct {
	Y       layout.Abs
	Element T
}

// FindInFrame collects elements of a specific type from a frame.
// Returns elements with their vertical positions.
func FindInFrame[T any](frame *Frame, matcher func(interface{}) (T, bool)) []FoundElement[T] {
	var output []FoundElement[T]
	findInFrameImpl(&output, frame, 0, matcher)
	return output
}

// findInFrameImpl recursively searches a frame for matching elements.
func findInFrameImpl[T any](
	output *[]FoundElement[T],
	frame *Frame,
	yOffset layout.Abs,
	matcher func(interface{}) (T, bool),
) {
	for _, item := range frame.Items {
		y := yOffset + item.Pos.Y
		switch it := item.Item.(type) {
		case GroupItem:
			findInFrameImpl(output, it.Frame, y, matcher)
		case TagItem:
			if start, ok := it.Tag.(StartTag); ok {
				if elem, ok := matcher(start.Elem); ok {
					*output = append(*output, FoundElement[T]{Y: y, Element: elem})
				}
			}
		}
	}
}

// FindInFrames collects elements from multiple frames.
func FindInFrames[T any](frames []*Frame, matcher func(interface{}) (T, bool)) []FoundElement[T] {
	var output []FoundElement[T]
	for _, frame := range frames {
		findInFrameImpl(&output, frame, 0, matcher)
	}
	return output
}

// SortByY sorts found elements by their Y position.
func SortByY[T any](elements []FoundElement[T]) {
	sort.Slice(elements, func(i, j int) bool {
		return elements[i].Y < elements[j].Y
	})
}

// LayoutLineNumbers positions line numbers in the output frame.
func LayoutLineNumbers(
	engine interface{},
	config *Config,
	lineConfig *LineNumberConfig,
	locator Locator,
	column int,
	output *Frame,
) error {
	splitLocator := locator.Split()

	// Reset counter at page start if scope is page
	if column == 0 && lineConfig.Scope == LineNumberingScopePage {
		reset, err := layoutLineNumberReset(engine, config, splitLocator)
		if err != nil {
			return err
		}
		output.PushFrame(layout.Point{}, reset)
	}

	// Find line markers in the frame
	markers := FindInFrame(output, func(elem interface{}) (*ParLineMarker, bool) {
		if m, ok := elem.(*ParLineMarker); ok {
			return m, true
		}
		return nil, false
	})

	if len(markers) == 0 {
		return nil
	}

	// Sort by Y position
	SortByY(markers)

	// Track layout state
	var maxNumberWidth layout.Abs
	var prevBottom *layout.Abs
	type lineNumberEntry struct {
		y      layout.Abs
		marker *ParLineMarker
		frame  *Frame
	}
	var lineNumbers []lineNumberEntry

	// Layout each line number
	for _, found := range markers {
		y := found.Y
		marker := found.Element

		// Skip if overlapping with previous line
		if prevBottom != nil && y < *prevBottom {
			continue
		}

		frame, err := layoutLineNumber(engine, config, splitLocator, &marker.Numbering)
		if err != nil {
			return err
		}

		bottom := y + maxAbs(frame.Height(), 1) // At least 1pt
		prevBottom = &bottom
		if frame.Width() > maxNumberWidth {
			maxNumberWidth = frame.Width()
		}
		lineNumbers = append(lineNumbers, lineNumberEntry{y: y, marker: marker, frame: frame})
	}

	// Position line numbers
	for _, entry := range lineNumbers {
		// Determine margin side
		margin := entry.marker.NumberMargin
		opposite := config.Columns.Count >= 2 && column+1 == config.Columns.Count
		var resolvedMargin FixedAlignment
		if opposite {
			resolvedMargin = OuterHAlignmentEnd.Resolve(config.Shared)
		} else {
			resolvedMargin = margin.Resolve(config.Shared)
		}

		// Get clearance
		var clearance layout.Abs
		if entry.marker.NumberClearance.IsAuto() {
			clearance = lineConfig.DefaultClearance
		} else {
			clearance = entry.marker.NumberClearance.Get().RelativeTo(output.Width())
		}

		// Calculate X position
		var x layout.Abs
		switch resolvedMargin {
		case FixedAlignmentStart:
			x = -maxNumberWidth - clearance
		case FixedAlignmentEnd:
			x = output.Width() + clearance
		}

		// Apply alignment within the number width
		var align FixedAlignment
		if entry.marker.NumberAlign != nil {
			align = *entry.marker.NumberAlign
		} else {
			align = resolvedMargin.Inv()
		}
		shift := align.Position(maxNumberWidth - entry.frame.Width())

		pos := layout.Point{X: x + shift, Y: entry.y}
		output.PushFrame(pos, entry.frame)
	}

	return nil
}

// layoutLineNumberReset creates a frame that resets the line number counter.
func layoutLineNumberReset(
	engine interface{},
	config *Config,
	locator *SplitLocator,
) (*Frame, error) {
	// Placeholder - would create counter reset content
	return NewSoftFrame(layout.Size{}), nil
}

// layoutLineNumber creates a frame for a single line number.
func layoutLineNumber(
	engine interface{},
	config *Config,
	locator *SplitLocator,
	numbering *Numbering,
) (*Frame, error) {
	// Placeholder - would create numbered content
	frame := NewSoftFrame(layout.Size{Width: 20, Height: 12})
	frame.Translate(layout.Point{Y: -frame.Baseline})
	return frame, nil
}

// FootnoteElem represents a footnote element (placeholder type).
type FootnoteElem struct {
	// Content is the footnote content.
	Content interface{}
	// Span is the source location.
	Span interface{}
	// location caches the element location.
	location *Location
}

// Location returns the element's location.
func (f *FootnoteElem) Location() *Location {
	return f.location
}

// IsRef returns true if this is a footnote reference.
func (f *FootnoteElem) IsRef() bool {
	return false
}

// FootnoteEntry wraps a footnote for layout.
type FootnoteEntry struct {
	Elem *FootnoteElem
}

// NewFootnoteEntry creates a new footnote entry.
func NewFootnoteEntry(elem *FootnoteElem) FootnoteEntry {
	return FootnoteEntry{Elem: elem}
}

// LayoutFootnoteSeparator creates the footnote separator frame.
func LayoutFootnoteSeparator(
	engine interface{},
	config *Config,
	base layout.Size,
) (*Frame, error) {
	// Placeholder - would layout the separator content from config
	return NewSoftFrame(layout.Size{Width: base.Width, Height: 1}), nil
}

// LayoutFootnote creates frames for a footnote entry.
func LayoutFootnote(
	engine interface{},
	config *Config,
	elem *FootnoteElem,
	pod Regions,
) (Fragment, error) {
	// Placeholder - would layout footnote content
	frame := NewSoftFrame(layout.Size{Width: pod.Size.Width, Height: 20})
	if loc := elem.Location(); loc != nil {
		frame.SetParent(NewFrameParent(*loc, InheritNo))
	}
	return Fragment{frame}, nil
}
