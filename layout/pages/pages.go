package pages

import (
	"github.com/boergens/gotypst/layout"
)

// LayoutDocument lays out content into a paged document.
// This is the main entry point for document layout.
func LayoutDocument(engine *Engine, content *Content, styles StyleChain) (*PagedDocument, error) {
	locator := &Locator{Current: 0}
	splitLocator := locator.Split()

	// Convert content to pairs
	// TODO: This should realize the content through engine routines
	var children []Pair
	if content != nil {
		for _, elem := range content.Elements {
			children = append(children, Pair{
				Element: elem,
				Styles:  styles,
			})
		}
	}

	// Layout the pages
	pages, err := layoutPages(engine, children, splitLocator, styles)
	if err != nil {
		return nil, err
	}

	return &PagedDocument{
		Pages: pages,
		Info:  DocumentInfo{},
	}, nil
}

// layoutPages collects and lays out all pages.
func layoutPages(engine *Engine, children []Pair, locator *SplitLocator, styles StyleChain) ([]Page, error) {
	// Collect children into items
	items := Collect(children, locator, styles)

	// Extract run items for parallel layout
	var runItems []RunItem
	for _, item := range items {
		if run, ok := item.(RunItem); ok {
			runItems = append(runItems, run)
		}
	}

	// Layout all runs in parallel
	results := engine.Parallelize(runItems, func(e *Engine, run RunItem) ([]LayoutedPage, error) {
		return LayoutPageRun(e, run.Children, run.Locator, run.Initial)
	})

	// Process results and build pages
	var pages []Page
	var tags []Tag
	counter := NewManualPageCounter()
	runIdx := 0

	for _, item := range items {
		switch it := item.(type) {
		case RunItem:
			result := results[runIdx]
			runIdx++
			if result.err != nil {
				return nil, result.err
			}
			for _, layouted := range result.pages {
				page, err := Finalize(engine, counter, &tags, layouted)
				if err != nil {
					return nil, err
				}
				pages = append(pages, *page)
			}

		case ParityItem:
			if !it.Parity.Matches(len(pages)) {
				continue
			}
			layouted, err := LayoutBlankPage(engine, it.Locator, it.Initial)
			if err != nil {
				return nil, err
			}
			if layouted != nil {
				page, err := Finalize(engine, counter, &tags, *layouted)
				if err != nil {
					return nil, err
				}
				pages = append(pages, *page)
			}

		case TagsItem:
			for _, pair := range it.Children {
				if tagElem, ok := pair.Element.(*TagElem); ok {
					tags = append(tags, tagElem.Tag)
				}
			}
		}
	}

	// Add any remaining tags to the last page
	if len(tags) > 0 && len(pages) > 0 {
		last := &pages[len(pages)-1]
		pos := layout.Point{X: 0, Y: last.Frame.Height()}
		for _, tag := range tags {
			last.Frame.Push(pos, TagItem{Tag: tag})
		}
	}

	return pages, nil
}
