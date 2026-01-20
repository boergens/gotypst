package pages

import (
	"github.com/boergens/gotypst/layout"
)

// Finalize pieces together the inner page frame and the marginals.
// We can only do this at the very end because inside/outside margins
// require knowledge of the physical page number, which is unknown
// during parallel layout. Headers and footers are laid out here so they
// can include running content like page numbers.
func Finalize(engine *Engine, counter *ManualPageCounter, tags *[]Tag, layouted LayoutedPage) (*Page, error) {
	margin := layouted.Margin

	// If two-sided, left becomes inside and right becomes outside.
	// Thus, for left-bound pages, we want to swap on even pages and
	// for right-bound pages, we want to swap on odd pages.
	if layouted.TwoSided && layouted.Binding.Swap(counter.Physical()) {
		margin.Left, margin.Right = margin.Right, margin.Left
	}

	// Get this page's number (needed for header/footer running content)
	pageNumber := counter.Logical()

	// Create a frame for the full page
	fullSize := layout.Size{
		Width:  layouted.Inner.Size.Width + margin.Left + margin.Right,
		Height: layouted.Inner.Size.Height + margin.Top + margin.Bottom,
	}
	frame := Hard(fullSize)

	// Add tags
	for _, tag := range *tags {
		frame.Push(layout.Point{X: 0, Y: 0}, TagItem{Tag: tag})
	}
	*tags = (*tags)[:0] // Clear the tags slice

	// Layout header/footer with actual page number for running content
	var headerFrame, footerFrame *Frame
	if layouted.HeaderContent != nil {
		headerFrame = layoutMarginal(engine, layouted.HeaderContent, layouted.HeaderSize, nil, pageNumber)
	}
	if layouted.FooterContent != nil {
		footerFrame = layoutMarginal(engine, layouted.FooterContent, layouted.FooterSize, nil, pageNumber)
	}

	// Add the "before" marginals. The order in which we push things here is
	// important as it affects the relative ordering of introspectable elements
	// and thus how counters resolve.
	if layouted.Background != nil {
		frame.PushFrame(layout.Point{X: 0, Y: 0}, *layouted.Background)
	}
	if headerFrame != nil {
		frame.PushFrame(layout.Point{X: margin.Left, Y: 0}, *headerFrame)
	}

	// Add the inner contents
	frame.PushFrame(layout.Point{X: margin.Left, Y: margin.Top}, layouted.Inner)

	// Add the "after" marginals
	if footerFrame != nil {
		y := fullSize.Height - footerFrame.Size.Height
		frame.PushFrame(layout.Point{X: margin.Left, Y: y}, *footerFrame)
	}
	if layouted.Foreground != nil {
		frame.PushFrame(layout.Point{X: 0, Y: 0}, *layouted.Foreground)
	}

	// Apply counter updates from within the page to the manual page counter
	if err := counter.Visit(&frame); err != nil {
		return nil, err
	}

	// Bump the counter for the next page
	counter.Step()

	return &Page{
		Frame:      frame,
		Fill:       layouted.Fill,
		Numbering:  layouted.Numbering,
		Supplement: layouted.Supplement,
		Number:     pageNumber,
	}, nil
}
